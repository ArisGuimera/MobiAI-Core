// Package catalog reads the MobiAI marketplace + per-plugin manifests
// from a root directory containing .claude-plugin/marketplace.json.
package catalog

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Marketplace mirrors .claude-plugin/marketplace.json.
type Marketplace struct {
	Name     string           `json:"name"`
	Owner    MarketplaceOwner `json:"owner"`
	Metadata MarketplaceMeta  `json:"metadata"`
	Plugins  []PluginRef      `json:"plugins"`
}

type MarketplaceOwner struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type MarketplaceMeta struct {
	Description string `json:"description"`
	Version     string `json:"version"`
}

// PluginRef is a single entry under "plugins" in marketplace.json.
type PluginRef struct {
	Name        string   `json:"name"`
	Source      string   `json:"source"`
	Description string   `json:"description"`
	Category    string   `json:"category"`
	Tags        []string `json:"tags,omitempty"`
}

// LoadMarketplace reads <rootDir>/.claude-plugin/marketplace.json.
func LoadMarketplace(rootDir string) (*Marketplace, error) {
	path := filepath.Join(rootDir, ".claude-plugin", "marketplace.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read marketplace.json: %w", err)
	}
	var mp Marketplace
	if err := json.Unmarshal(data, &mp); err != nil {
		return nil, fmt.Errorf("parse marketplace.json: %w", err)
	}
	return &mp, nil
}
