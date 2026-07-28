// Package state manages the ~/.mobiai/ directory: paths, the installed.json
// file (which packs are installed in which hosts), and the catalog
// freshness metadata.
package state

import (
	"fmt"
	"os"
	"path/filepath"
)

// Paths describes the filesystem locations the CLI uses at runtime.
// Override the root via MOBIAI_HOME (mainly for testing).
type Paths struct {
	Home string
}

// NewPaths resolves Home from MOBIAI_HOME, falling back to ~/.mobiai.
func NewPaths() (Paths, error) {
	if h := os.Getenv("MOBIAI_HOME"); h != "" {
		return Paths{Home: h}, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return Paths{}, fmt.Errorf("resolve user home dir: %w", err)
	}
	return Paths{Home: filepath.Join(home, ".mobiai")}, nil
}

func (p Paths) Bin() string           { return filepath.Join(p.Home, "bin") }
func (p Paths) Cache() string         { return filepath.Join(p.Home, "cache") }
func (p Paths) CatalogDir() string    { return filepath.Join(p.Cache(), "catalog") }
func (p Paths) State() string         { return filepath.Join(p.Home, "state") }
func (p Paths) Logs() string          { return filepath.Join(p.Home, "logs") }
func (p Paths) InstalledFile() string { return filepath.Join(p.State(), "installed.json") }
func (p Paths) MetaFile() string      { return filepath.Join(p.Cache(), "meta.json") }
func (p Paths) ConfigFile() string    { return filepath.Join(p.Home, "config.json") }

// EnsureDirs creates Bin, Cache, State, and Logs if they don't exist.
// Permissions: 0o755.
func (p Paths) EnsureDirs() error {
	for _, dir := range []string{p.Bin(), p.Cache(), p.State(), p.Logs()} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("mkdir %s: %w", dir, err)
		}
	}
	return nil
}
