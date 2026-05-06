package host

import "fmt"

// Registry holds the set of HostAdapters known to the CLI.
//
// Adapters fall into three tiers:
//   - Tier 1 (Claude Code, Cursor, Gemini CLI, Codex): adapter dedicado, testeado a fondo.
//   - Tier 2 (Goose, Copilot CLI, OpenCode, Junie, Roo Code): best-effort vía agentskills.io.
//   - Tier 3 (todo lo demás): paths especulativos a confirmar por la comunidad.
//
// Auto-detection (Detect) excludes tier-3 adapters by default to avoid
// hijacking unrelated directories like ~/.mux or ~/.factory. Users opt
// in via SetIncludeExperimental(true) (wired to --include-experimental
// or the MOBIAI_INCLUDE_EXPERIMENTAL env var). Get(id) and Adapters()
// return tier-3 adapters regardless, so --host=<id> still works.
type Registry struct {
	adapters            []HostAdapter
	includeExperimental bool
}

// tier3IDs is the set of adapter IDs whose detection is gated behind
// the experimental opt-in. Keep this in sync with the adapter factories
// in this package.
var tier3IDs = map[string]bool{
	"autohand":         true,
	"openhands":        true,
	"mux":              true,
	"amp":              true,
	"letta":            true,
	"firebender":       true,
	"claude-desktop":   true,
	"piebald":          true,
	"factory":          true,
	"pi":               true,
	"databricks-genie": true,
	"agentman":         true,
	"trae":             true,
	"spring-ai":        true,
	"mistral-vibe":     true,
	"command-code":     true,
	"ona":              true,
	"vt-code":          true,
	"qodo":             true,
	"laravel-boost":    true,
	"emdash":           true,
	"snowflake-cortex": true,
	"kiro":             true,
	"workshop":         true,
	"ai-edge-gallery":  true,
	"nanobot":          true,
	"fast-agent":       true,
}

// NewDefaultRegistry returns the built-in registry with all 36 adapters
// known to the CLI. Tier-3 adapters are loaded but excluded from Detect()
// until SetIncludeExperimental(true) is called.
func NewDefaultRegistry() *Registry {
	adapters := []HostAdapter{
		// Tier 1 — adapter dedicado, testeado a fondo.
		newClaudeCode(),
		newCursor(),
		newGemini(),
		newCodex(),
		// Tier 2 — best-effort vía el standard agentskills.io.
		newGoose(),
		newGitHubCopilot(),
		newOpenCode(),
		newJunie(),
		newRooCode(),
		// Tier 3 — speculativo, paths a confirmar por la comunidad.
		newAutohand(),
		newOpenHands(),
		newMux(),
		newAmp(),
		newLetta(),
		newFirebender(),
		newClaudeDesktop(),
		newPiebald(),
		newFactory(),
		newPi(),
		newDatabricksGenie(),
		newAgentman(),
		newTRAE(),
		newSpringAI(),
		newMistralVibe(),
		newCommandCode(),
		newOna(),
		newVTCode(),
		newQodo(),
		newLaravelBoost(),
		newEmdash(),
		newSnowflakeCortex(),
		newKiro(),
		newWorkshop(),
		newAIEdgeGallery(),
		newNanobot(),
		newFastAgent(),
	}
	for _, a := range adapters {
		if tier3IDs[a.ID()] {
			if g, ok := a.(*genericAdapter); ok {
				g.experimental = true
			}
		}
	}
	return &Registry{adapters: adapters}
}

// SetIncludeExperimental toggles whether Detect() considers experimental
// (tier-3) adapters.
func (r *Registry) SetIncludeExperimental(include bool) {
	r.includeExperimental = include
}

// Adapters returns all registered adapters in registration order, regardless
// of experimental status (intended for debug/visibility commands).
func (r *Registry) Adapters() []HostAdapter {
	out := make([]HostAdapter, len(r.adapters))
	copy(out, r.adapters)
	return out
}

// Get returns the adapter with the given ID. Experimental adapters are
// resolvable so that --host=<id> can target them explicitly.
func (r *Registry) Get(id string) (HostAdapter, error) {
	for _, a := range r.adapters {
		if a.ID() == id {
			return a, nil
		}
	}
	return nil, fmt.Errorf("host adapter %q not found", id)
}

// Detect runs Detect() on every non-experimental adapter and returns the
// subset whose Detect.Found is true. Experimental adapters are included
// only when SetIncludeExperimental(true) has been called.
func (r *Registry) Detect() []HostAdapter {
	var out []HostAdapter
	for _, a := range r.adapters {
		if a.Experimental() && !r.includeExperimental {
			continue
		}
		if a.Detect().Found {
			out = append(out, a)
		}
	}
	return out
}
