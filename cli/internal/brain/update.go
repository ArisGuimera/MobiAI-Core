package brain

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// UpdateOptions specifies which metadata fields to update on an entry.
// Empty string values mean "leave this field unchanged"; to remove
// review_after entirely, set ClearReviewAfter to true.
//
// Combining ReviewAfter and ClearReviewAfter is rejected (contradictory).
type UpdateOptions struct {
	Status           string // new status: active|temporary|deprecated
	ReviewAfter      string // new review_after (YYYY-MM-DD)
	ClearReviewAfter bool   // if true, drop the review_after line
}

// UpdateResult bundles the data callers need to confirm a successful
// update — what file/section the entry lives in, its title for echo,
// and the previous values of any fields that changed (so the CLI/MCP
// can show "active → deprecated" style diffs).
type UpdateResult struct {
	Section         string // canonical section (e.g. "bugfixes")
	File            string // basename (e.g. "bugfixes.md")
	Title           string
	PrevStatus      string
	NewStatus       string
	PrevReviewAfter string
	NewReviewAfter  string
}

// UpdateEntry locates the entry with id across all memory files and
// rewrites its metadata block atomically. The entry's body and the
// rest of the file are preserved byte-perfect.
//
// Returns a clear error when id isn't found, when the requested status
// is invalid, or when ReviewAfter has a bad format. The brain must be
// initialized (same guard as save/search).
//
// For bugfix entries, the `type:` field is re-derived from the new
// status (temporary → platform_workaround, active|deprecated → bug_fix)
// to keep the invariant maintained by AppendEntry.
func UpdateEntry(p BrainPaths, id string, opts UpdateOptions) (*UpdateResult, error) {
	if !p.Exists() {
		return nil, fmt.Errorf("brain no inicializado en %s — corré `mobiai brain init` primero", p.Root)
	}
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, fmt.Errorf("id es requerido")
	}
	if opts.Status != "" {
		if _, ok := validStatuses[opts.Status]; !ok {
			return nil, fmt.Errorf("--status %q inválido (esperado: active|temporary|deprecated)", opts.Status)
		}
	}
	if opts.ReviewAfter != "" {
		if _, err := time.Parse(reviewAfterLayout, opts.ReviewAfter); err != nil {
			return nil, fmt.Errorf("--review-after debe ser YYYY-MM-DD: %w", err)
		}
	}
	if opts.ReviewAfter != "" && opts.ClearReviewAfter {
		return nil, fmt.Errorf("no podés pasar --review-after y --clear-review-after a la vez")
	}
	if opts.Status == "" && opts.ReviewAfter == "" && !opts.ClearReviewAfter {
		return nil, fmt.Errorf("nada que actualizar: pasá al menos --status, --review-after o --clear-review-after")
	}

	// Walk every memory file in canonical order until we find a `## `
	// section whose meta block contains `- id: <id>`.
	for _, mf := range MemoryFiles {
		fpath := filepath.Join(p.MemoriesDir, mf.Name)
		data, err := os.ReadFile(fpath)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, fmt.Errorf("leer %s: %w", fpath, err)
		}
		updated, res, found, err := rewriteEntryByID(string(data), id, mf.Name, opts)
		if err != nil {
			return nil, err
		}
		if !found {
			continue
		}
		if err := atomicWrite(fpath, []byte(updated)); err != nil {
			return nil, err
		}
		res.Section = canonicalSection(mf.Name)
		res.File = mf.Name
		return res, nil
	}
	return nil, fmt.Errorf("no encontré ninguna entrada con id %q en el brain", id)
}

// rewriteEntryByID scans the raw file content line-by-line, locates the
// entry whose meta block contains `- id: <id>`, mutates the meta block
// in place, and returns the rewritten file plus a result describing
// the change. The body, the rest of the file, and any meta keys not
// targeted by opts are left untouched.
//
// Operates on strings rather than bytes to keep newline handling simple
// — Go's bufio splits on `\n`, and we preserve a trailing newline iff
// the original ended with one.
func rewriteEntryByID(content, id, basename string, opts UpdateOptions) (string, *UpdateResult, bool, error) {
	lines := splitLinesPreserve(content)

	// Phase 1: locate the entry header (`## Title`) whose meta block
	// includes `- id: <id>`. We need both the header index and the
	// indices of the meta block (contiguous bullets after the header,
	// tolerating blank lines in between, like the parser).
	for i := 0; i < len(lines); i++ {
		if !strings.HasPrefix(lines[i], "## ") {
			continue
		}
		title := strings.TrimSpace(strings.TrimPrefix(lines[i], "## "))
		// Find end-of-entry (next ## or EOF).
		entryEnd := len(lines)
		for j := i + 1; j < len(lines); j++ {
			if strings.HasPrefix(lines[j], "## ") {
				entryEnd = j
				break
			}
		}
		// Find the meta block: lines starting from i+1 that match the
		// `- key: value` pattern, allowing blank lines between bullets
		// until the first non-blank non-meta line.
		metaStart := i + 1
		metaEnd := metaStart
		metas := make([]metaLine, 0, 8)
		for j := metaStart; j < entryEnd; j++ {
			if k, v, ok := parseMetaLine(lines[j]); ok {
				metas = append(metas, metaLine{key: k, val: v, raw: lines[j]})
				metaEnd = j + 1
				continue
			}
			if strings.TrimSpace(lines[j]) == "" {
				// Blank lines inside the meta block are tolerated.
				metaEnd = j + 1
				continue
			}
			// First real content line ends the meta block.
			break
		}

		// Check if this is our entry.
		gotID := ""
		for _, m := range metas {
			if m.key == "id" {
				gotID = m.val
				break
			}
		}
		if gotID != id {
			continue
		}

		// Found it. Apply mutations + render new meta block.
		res := &UpdateResult{Title: title}
		newMetas, err := applyUpdate(metas, basename, opts, res)
		if err != nil {
			return "", nil, true, err
		}

		// Count leading/trailing blank lines wrapping the meta block so
		// we can reproduce the exact spacing save.renderEntry uses
		// (one blank between `## Title` and the first meta, and one
		// blank between the last meta and the body). Without this the
		// updated entry visually drifts from freshly saved ones.
		leadingBlanks := 0
		for k := metaStart; k < metaEnd && strings.TrimSpace(lines[k]) == ""; k++ {
			leadingBlanks++
		}
		trailingBlanks := 0
		for k := metaEnd - 1; k >= metaStart+leadingBlanks && strings.TrimSpace(lines[k]) == ""; k-- {
			trailingBlanks++
		}

		newMetaLines := renderMetaLines(newMetas)
		wrapped := make([]string, 0, leadingBlanks+len(newMetaLines)+trailingBlanks)
		for k := 0; k < leadingBlanks; k++ {
			wrapped = append(wrapped, "")
		}
		wrapped = append(wrapped, newMetaLines...)
		for k := 0; k < trailingBlanks; k++ {
			wrapped = append(wrapped, "")
		}

		out := make([]string, 0, len(lines)+len(wrapped)-(metaEnd-metaStart))
		out = append(out, lines[:metaStart]...)
		out = append(out, wrapped...)
		out = append(out, lines[metaEnd:]...)
		return strings.Join(out, "\n"), res, true, nil
	}
	return content, nil, false, nil
}

// metaLine holds a parsed meta entry along with its original raw form.
// We don't strictly need raw — we always re-render — but keeping it
// around makes debugging easier and would let us preserve unusual
// formatting in the future.
type metaLine struct {
	key string
	val string
	raw string
}

// applyUpdate mutates metas according to opts, fills the result's
// Prev/New fields, and re-derives the type: field when status changes
// on a bugfix entry.
//
// Order of operations:
//  1. Read current values into res.Prev*.
//  2. Apply Status mutation (recompute type if appropriate).
//  3. Apply ReviewAfter mutation (or clear).
//  4. Re-read final values into res.New*.
func applyUpdate(metas []metaLine, basename string, opts UpdateOptions, res *UpdateResult) ([]metaLine, error) {
	get := func(key string) string {
		for _, m := range metas {
			if m.key == key {
				return m.val
			}
		}
		return ""
	}
	set := func(key, val string) []metaLine {
		for i, m := range metas {
			if m.key == key {
				metas[i].val = val
				return metas
			}
		}
		return append(metas, metaLine{key: key, val: val})
	}
	drop := func(key string) []metaLine {
		out := metas[:0]
		for _, m := range metas {
			if m.key == key {
				continue
			}
			out = append(out, m)
		}
		return out
	}

	res.PrevStatus = get("status")
	res.PrevReviewAfter = get("review_after")

	if opts.Status != "" {
		metas = set("status", opts.Status)
		// For bugfix entries, keep the type: field consistent with
		// status. We can derive SaveType from the file name.
		if st := saveTypeForFile(basename); st == SaveTypeBugfix {
			metas = set("type", st.fullTypeLabel(opts.Status))
		}
	}
	if opts.ClearReviewAfter {
		metas = drop("review_after")
	} else if opts.ReviewAfter != "" {
		metas = set("review_after", opts.ReviewAfter)
	}

	res.NewStatus = get("status")
	res.NewReviewAfter = get("review_after")
	return metas, nil
}

// renderMetaLines turns the meta slice back into bullet lines, ordered
// canonically (id, type, status, platform, area, date, review_after,
// then anything else in original order). Matches the layout produced
// by save.renderEntry so updated entries look indistinguishable from
// freshly saved ones.
func renderMetaLines(metas []metaLine) []string {
	canonical := []string{"id", "type", "status", "platform", "area", "date", "review_after"}
	byKey := make(map[string]metaLine, len(metas))
	order := make([]string, 0, len(metas))
	for _, m := range metas {
		if _, seen := byKey[m.key]; !seen {
			order = append(order, m.key)
		}
		byKey[m.key] = m
	}

	out := make([]string, 0, len(metas))
	for _, k := range canonical {
		if m, ok := byKey[k]; ok {
			out = append(out, fmt.Sprintf("- %s: %s", m.key, m.val))
			delete(byKey, k)
		}
	}
	// Append any non-canonical keys in original order.
	for _, k := range order {
		if m, ok := byKey[k]; ok {
			out = append(out, fmt.Sprintf("- %s: %s", m.key, m.val))
		}
	}
	return out
}

// saveTypeForFile maps a memory file basename back to its SaveType. We
// need this to re-derive the `type:` label on status changes. Returns
// empty SaveType for non-save sections (integrations, releases).
func saveTypeForFile(basename string) SaveType {
	switch basename {
	case "decisions.md":
		return SaveTypeDecision
	case "bugfixes.md":
		return SaveTypeBugfix
	case "testing.md":
		return SaveTypeTesting
	}
	return ""
}

// splitLinesPreserve splits content on `\n` without dropping the empty
// trailing element when content ends with `\n` (unlike strings.Split,
// which adds one — we want to know whether the file ended with a
// newline so we can preserve it on write).
func splitLinesPreserve(content string) []string {
	if content == "" {
		return nil
	}
	// strings.Split on "\n" already produces what we want: a trailing
	// newline yields an empty final element; no trailing newline does
	// not. We restore that in joinLinesPreserve.
	return strings.Split(content, "\n")
}

// Note: we use strings.Join directly above instead of a helper because
// strings.Split already gives us a trailing empty element when content
// ended with a newline, so strings.Join reconstructs the trailing
// newline automatically. No special handling needed.

// atomicWrite writes data to path using the temp+rename pattern so a
// crash mid-write never leaves a corrupted file. Same atomicity strategy
// as the config/scan writers in this package.
func atomicWrite(path string, data []byte) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".update-*.tmp")
	if err != nil {
		return fmt.Errorf("crear tmp en %s: %w", dir, err)
	}
	tmpName := tmp.Name()
	cleanup := func() { _ = os.Remove(tmpName) }
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		cleanup()
		return fmt.Errorf("escribir tmp: %w", err)
	}
	if err := tmp.Close(); err != nil {
		cleanup()
		return fmt.Errorf("cerrar tmp: %w", err)
	}
	if err := os.Rename(tmpName, path); err != nil {
		cleanup()
		return fmt.Errorf("rename %s → %s: %w", tmpName, path, err)
	}
	return nil
}
