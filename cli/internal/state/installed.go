package state

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

// Installed tracks which packs are installed in which hosts.
// Persisted as <state>/installed.json.
type Installed struct {
	// Packs maps pack name → list of host IDs (sorted).
	Packs map[string][]string `json:"packs"`
}

// LoadInstalled reads installed.json. Returns an empty Installed if
// the file does not exist.
func LoadInstalled(p Paths) (*Installed, error) {
	path := p.InstalledFile()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Installed{Packs: map[string][]string{}}, nil
		}
		return nil, fmt.Errorf("read installed.json: %w", err)
	}
	var s Installed
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("parse installed.json: %w", err)
	}
	if s.Packs == nil {
		s.Packs = map[string][]string{}
	}
	return &s, nil
}

// Save writes installed.json atomically (write-tmp + rename).
func (s *Installed) Save(p Paths) error {
	if err := os.MkdirAll(p.State(), 0o755); err != nil {
		return fmt.Errorf("mkdir state dir: %w", err)
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal installed: %w", err)
	}
	final := p.InstalledFile()
	tmp, err := os.CreateTemp(filepath.Dir(final), "installed-*.json.tmp")
	if err != nil {
		return fmt.Errorf("create tmp file: %w", err)
	}
	tmpName := tmp.Name()
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return fmt.Errorf("write tmp: %w", err)
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("close tmp: %w", err)
	}
	if err := os.Rename(tmpName, final); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("rename tmp to %s: %w", final, err)
	}
	return nil
}

// Add records pack as installed in host. No-op if already present.
// Hosts list is kept sorted.
func (s *Installed) Add(pack, host string) {
	if s.Packs == nil {
		s.Packs = map[string][]string{}
	}
	for _, h := range s.Packs[pack] {
		if h == host {
			return
		}
	}
	s.Packs[pack] = append(s.Packs[pack], host)
	sort.Strings(s.Packs[pack])
}

// Remove drops host from pack's host list. If the list becomes empty,
// the pack key is deleted.
func (s *Installed) Remove(pack, host string) {
	hosts, ok := s.Packs[pack]
	if !ok {
		return
	}
	out := hosts[:0]
	for _, h := range hosts {
		if h != host {
			out = append(out, h)
		}
	}
	if len(out) == 0 {
		delete(s.Packs, pack)
		return
	}
	s.Packs[pack] = out
}

// HostsFor returns the (sorted) list of hosts where pack is installed,
// or nil if the pack has no record.
func (s *Installed) HostsFor(pack string) []string {
	hosts, ok := s.Packs[pack]
	if !ok || len(hosts) == 0 {
		return nil
	}
	out := append([]string(nil), hosts...)
	return out
}
