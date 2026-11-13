package brain

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// hookMaxBytes caps the notes text one hook call adds to the agent's
// context. Measured on Windows with 3,000 cards (see the stage 8 commit):
// the cost of the call is reading the cards, not the size of the output,
// so the cap is about not flooding the conversation. Claude Code itself
// cuts additionalContext at 10,000 characters.
const hookMaxBytes = 4096

// hookStateMaxAge is how long a session's delivered-notes list is kept.
const hookStateMaxAge = 7 * 24 * time.Hour

// HookOptions tunes FileContextHook. The zero value is what the CLI uses.
type HookOptions struct {
	StateDir string // per-session state; "" = <os temp>/mobiai-brain-hook
	MaxBytes int    // notes text cap; 0 = hookMaxBytes
}

// hookInput is the subset of the hook payload we read. Claude Code sends
// session_id and tool_input.file_path; Cursor sends conversation_id and a
// tool_input whose file key its docs don't name, so several are tried.
type hookInput struct {
	SessionID      string         `json:"session_id"`
	ConversationID string         `json:"conversation_id"`
	HookEventName  string         `json:"hook_event_name"`
	Cwd            string         `json:"cwd"`
	ToolInput      map[string]any `json:"tool_input"`
	FilePath       string         `json:"file_path"`
}

// FileContextHook implements `mobiai brain hook file-context`: given the
// hook payload of a file read or edit, it returns the hook output that
// hands the agent the brain notes applying to that file's folder, or nil
// when there is nothing to say. It never writes the brain (no migration)
// and never fails loudly: a hook must not get in the agent's way.
//
// Each note is delivered at most once per session, and the notes text is
// capped (HookOptions.MaxBytes); notes that don't fit are listed by id.
func FileContextHook(input []byte, opts HookOptions) []byte {
	var in hookInput
	if err := json.Unmarshal(input, &in); err != nil {
		return nil
	}
	file := in.FilePath
	for _, key := range []string{"file_path", "path", "target_file", "filePath", "file"} {
		if s, ok := in.ToolInput[key].(string); ok && s != "" {
			file = s
			break
		}
	}
	if file == "" {
		return nil
	}
	if !filepath.IsAbs(file) {
		if in.Cwd == "" {
			return nil
		}
		file = filepath.Join(in.Cwd, file)
	}
	root := findBrainRoot(filepath.Dir(file))
	if root == "" {
		return nil
	}
	rel, err := filepath.Rel(root, file)
	if err != nil || strings.HasPrefix(rel, "..") {
		return nil
	}
	rel = filepath.ToSlash(rel)

	notes, err := LoadNotes(NewBrainPaths(root))
	if err != nil {
		return nil
	}
	type match struct {
		note        MemoryEntry
		specificity int
	}
	var matches []match
	for _, n := range notes {
		if strings.EqualFold(strings.TrimSpace(n.Get("status")), StatusDeprecated) {
			continue
		}
		best := -1
		for _, pattern := range appliesTo(n) {
			if matchAppliesTo(pattern, rel) && len(pattern) > best {
				best = len(pattern)
			}
		}
		if best >= 0 {
			matches = append(matches, match{n, best})
		}
	}
	session := in.SessionID
	if session == "" {
		session = in.ConversationID
	}
	if session != "" {
		session += "\x00" + root
	}
	state := loadHookState(opts.stateDir(), session)
	delivered := map[string]bool{}
	for _, id := range state.Delivered {
		delivered[id] = true
	}
	pending := matches[:0]
	for _, m := range matches {
		if !delivered[m.note.Get("id")] {
			pending = append(pending, m)
		}
	}
	matches = pending
	if len(matches) == 0 {
		return nil
	}

	// Which notes are unconfirmed costs several git calls, so it is worked
	// out once per session (when there is first something to deliver) and
	// kept in the session state. Cards saved later in the session are the
	// user's own, so they are confirmed anyway.
	if !state.Checked {
		p := NewBrainPaths(root)
		cfg, err := LoadConfig(p)
		if err != nil {
			cfg = DefaultConfig(root)
		}
		conf := loadConfirmation(p, cfg)
		for _, n := range notes {
			if !conf.confirmed(n) {
				state.Unconfirmed = append(state.Unconfirmed, n.Get("id"))
			}
		}
		state.Checked = true
	}
	unconfirmed := map[string]bool{}
	for _, id := range state.Unconfirmed {
		unconfirmed[id] = true
	}
	for i := range notes {
		notes[i].Flags.Unconfirmed = unconfirmed[notes[i].Get("id")]
	}
	applySupersedes(notes)
	flags := map[string]NoteFlags{}
	for _, n := range notes {
		flags[n.Get("id")] = n.Flags
	}
	live := matches[:0]
	for _, m := range matches {
		m.note.Flags = flags[m.note.Get("id")]
		if m.note.Flags.SupersededBy == "" {
			live = append(live, m)
		}
	}
	matches = live
	if len(matches) == 0 {
		saveHookState(opts.stateDir(), session, state)
		return nil
	}

	sort.SliceStable(matches, func(i, j int) bool {
		if matches[i].specificity != matches[j].specificity {
			return matches[i].specificity > matches[j].specificity
		}
		return matches[i].note.Get("date") > matches[j].note.Get("date")
	})

	maxBytes := opts.MaxBytes
	if maxBytes <= 0 {
		maxBytes = hookMaxBytes
	}
	var b strings.Builder
	fmt.Fprintf(&b, "MobiAI Brain notes that apply to %s (each note is shown once per session; notes marked `unconfirmed` were written by someone else and not reviewed yet; `mobiai brain context` also tells which may be out of date):\n\n", rel)
	header := b.Len()
	var more []string
	for _, m := range matches {
		id := m.note.Get("id")
		state.Delivered = append(state.Delivered, id)
		var entry strings.Builder
		writeEntry(&entry, m.note)
		text := entry.String() + "\n"
		switch {
		case b.Len()+len(text) <= header+maxBytes:
			b.WriteString(text)
		case b.Len() == header:
			// The first note alone is over the cap: show its beginning.
			b.WriteString(truncateUTF8(text, maxBytes) + "…\n\n")
		default:
			more = append(more, id)
		}
	}
	if b.Len() == header && len(more) == 0 {
		return nil
	}
	if len(more) > 0 {
		fmt.Fprintf(&b, "More notes apply to this folder: %s. Read their files or run `mobiai brain show <id>`.\n", strings.Join(more, ", "))
	}
	saveHookState(opts.stateDir(), session, state)

	var out any
	if in.ConversationID != "" || in.HookEventName == "postToolUse" {
		out = map[string]string{"additional_context": b.String()}
	} else {
		out = map[string]any{"hookSpecificOutput": map[string]string{
			"hookEventName":     "PostToolUse",
			"additionalContext": b.String(),
		}}
	}
	data, err := json.Marshal(out)
	if err != nil {
		return nil
	}
	return data
}

// findBrainRoot walks up from dir to the folder holding
// .mobiai/brain/config.json ("" when there is none).
func findBrainRoot(dir string) string {
	for {
		if isFile(filepath.Join(dir, ".mobiai", "brain", "config.json")) {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}

// matchAppliesTo reports whether a repo-relative file matches one entry of
// a note's "applies to": a folder ("dir/", or a plain path without glob
// characters, which also matches the file itself), or a glob where `**`
// spans any number of folders.
func matchAppliesTo(pattern, name string) bool {
	pattern = strings.TrimPrefix(strings.TrimSpace(pattern), "./")
	if !strings.ContainsAny(pattern, "*?[") {
		dir := strings.TrimSuffix(pattern, "/")
		return name == dir || strings.HasPrefix(name, dir+"/")
	}
	return matchSegments(strings.Split(pattern, "/"), strings.Split(name, "/"))
}

func matchSegments(pattern, name []string) bool {
	if len(pattern) == 0 {
		return len(name) == 0
	}
	if pattern[0] == "**" {
		for i := 0; i <= len(name); i++ {
			if matchSegments(pattern[1:], name[i:]) {
				return true
			}
		}
		return false
	}
	if len(name) == 0 {
		return false
	}
	if ok, err := path.Match(pattern[0], name[0]); err != nil || !ok {
		return false
	}
	return matchSegments(pattern[1:], name[1:])
}

func (o HookOptions) stateDir() string {
	if o.StateDir != "" {
		return o.StateDir
	}
	return filepath.Join(os.TempDir(), "mobiai-brain-hook")
}

// hookStateFile is where a session's delivered note ids are kept.
func hookStateFile(dir, session string) string {
	sum := sha256.Sum256([]byte(session))
	return filepath.Join(dir, hex.EncodeToString(sum[:])[:16]+".json")
}

// hookState is what the hook remembers about one session in one brain.
type hookState struct {
	Delivered   []string `json:"delivered"`   // note ids already shown
	Checked     bool     `json:"checked"`     // Unconfirmed was computed
	Unconfirmed []string `json:"unconfirmed"` // ids of unconfirmed notes
}

// loadHookState returns what this session already got. Without a session
// id nothing is remembered.
func loadHookState(dir, session string) hookState {
	var state hookState
	if session == "" {
		return state
	}
	data, err := os.ReadFile(hookStateFile(dir, session))
	if err != nil {
		return state
	}
	_ = json.Unmarshal(data, &state)
	return state
}

// saveHookState records the session state and drops state files of
// sessions older than hookStateMaxAge.
func saveHookState(dir, session string, state hookState) {
	if session == "" {
		return
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return
	}
	data, _ := json.Marshal(state)
	_ = os.WriteFile(hookStateFile(dir, session), data, 0o644)

	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	for _, e := range entries {
		if info, err := e.Info(); err == nil && time.Since(info.ModTime()) > hookStateMaxAge {
			_ = os.Remove(filepath.Join(dir, e.Name()))
		}
	}
}

// truncateUTF8 cuts s to at most n bytes without splitting a character.
func truncateUTF8(s string, n int) string {
	if len(s) <= n {
		return s
	}
	for n > 0 && n < len(s) && s[n]&0xC0 == 0x80 {
		n--
	}
	return s[:n]
}
