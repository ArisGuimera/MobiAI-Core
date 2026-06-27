package state

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config holds user preferences persisted at <home>/config.json.
// Currently just the CLI language; extend with new fields as needed.
type Config struct {
	Lang string `json:"lang,omitempty"` // "en" | "es" (empty = default)
}

// LoadConfig reads config.json. Returns a zero-valued Config (no error) if the
// file does not exist, so callers get default behavior on a fresh install.
func LoadConfig(p Paths) (*Config, error) {
	data, err := os.ReadFile(p.ConfigFile())
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{}, nil
		}
		return nil, fmt.Errorf("read config.json: %w", err)
	}
	var c Config
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, fmt.Errorf("parse config.json: %w", err)
	}
	return &c, nil
}

// Save writes config.json atomically (temp file + rename), mirroring meta.go.
func (c *Config) Save(p Paths) error {
	if err := os.MkdirAll(p.Home, 0o755); err != nil {
		return fmt.Errorf("mkdir home dir: %w", err)
	}
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	final := p.ConfigFile()
	tmp, err := os.CreateTemp(filepath.Dir(final), "config-*.json.tmp")
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
