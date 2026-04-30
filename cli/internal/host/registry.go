package host

import "fmt"

// Registry holds the set of HostAdapters known to the CLI.
// In this phase the default registry contains the four tier-1 adapters;
// tier-2 adapters and user-defined ones plug in here in later phases.
type Registry struct {
	adapters []HostAdapter
}

// NewDefaultRegistry returns the built-in tier-1 registry.
func NewDefaultRegistry() *Registry {
	return &Registry{
		adapters: []HostAdapter{
			// Tier 1 — adapter dedicado, testeado a fondo.
			newClaudeCode(),
			newCursor(),
			newGemini(),
			newCodex(),
			// Tier 2 — best-effort vía el standard agentskills.io.
			newGoose(),
			newGitHubCopilot(),
			newVSCode(),
			newOpenCode(),
			newJunie(),
			newRooCode(),
		},
	}
}

// Adapters returns all registered adapters in registration order.
func (r *Registry) Adapters() []HostAdapter {
	out := make([]HostAdapter, len(r.adapters))
	copy(out, r.adapters)
	return out
}

// Get returns the adapter with the given ID.
func (r *Registry) Get(id string) (HostAdapter, error) {
	for _, a := range r.adapters {
		if a.ID() == id {
			return a, nil
		}
	}
	return nil, fmt.Errorf("host adapter %q not found", id)
}

// Detect runs Detect() on every registered adapter and returns the subset
// whose Detect.Found is true.
func (r *Registry) Detect() []HostAdapter {
	var out []HostAdapter
	for _, a := range r.adapters {
		if a.Detect().Found {
			out = append(out, a)
		}
	}
	return out
}
