package catalog

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// PluginManifest mirrors <source>/.claude-plugin/plugin.json.
// Skills defaults to ["./skills/"] when omitted (handled by Catalog.Load).
type PluginManifest struct {
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	Version      string   `json:"version"`
	License      string   `json:"license,omitempty"`
	Dependencies []string `json:"dependencies,omitempty"`
	Skills       []string `json:"skills,omitempty"`
	Keywords     []string `json:"keywords,omitempty"`
}

// LoadPlugin reads <rootDir>/<source>/.claude-plugin/plugin.json.
// source is the relative path from marketplace.json (e.g., "./skills/core").
func LoadPlugin(rootDir, source string) (*PluginManifest, error) {
	path := filepath.Join(rootDir, filepath.FromSlash(source), ".claude-plugin", "plugin.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read plugin.json at %s: %w", path, err)
	}
	var pm PluginManifest
	if err := json.Unmarshal(data, &pm); err != nil {
		return nil, fmt.Errorf("parse plugin.json at %s: %w", path, err)
	}
	return &pm, nil
}
