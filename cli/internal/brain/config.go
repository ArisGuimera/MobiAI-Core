package brain

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// ConfigVersion is the current schema version of config.json.
// Bumped when the schema changes in a non-backwards-compatible way.
const ConfigVersion = 1

// DefaultRules is the seed set of rules written by `brain init` when no
// config.json exists. Users edit these freely afterwards; the brain treats
// the rules as opaque strings.
var DefaultRules = []string{
	"Prioritize clean architecture",
	"Prepare code for testing",
	"Do not introduce hacks",
	"Respect existing project conventions",
}

// Config is the persisted shape of <brain>/config.json.
type Config struct {
	Version     int      `json:"version"`
	ProjectName string   `json:"project_name"`
	ProjectType string   `json:"project_type"`
	Platforms   []string `json:"platforms"`
	Rules       []string `json:"rules"`
	CreatedAt   string   `json:"created_at"`
	UpdatedAt   string   `json:"updated_at"`
}

// DefaultConfig returns the template used by `brain init` for a brand-new
// project. ProjectName is filename-safe (the basename of the project root).
func DefaultConfig(projectRoot string) Config {
	now := nowISO()
	return Config{
		Version:     ConfigVersion,
		ProjectName: filepath.Base(projectRoot),
		ProjectType: ProjectTypeUnknown,
		Platforms:   []string{},
		Rules:       append([]string(nil), DefaultRules...),
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// LoadConfig reads <brain>/config.json. Returns os.ErrNotExist if missing.
func LoadConfig(p BrainPaths) (Config, error) {
	data, err := os.ReadFile(p.ConfigFile)
	if err != nil {
		return Config{}, err
	}
	var c Config
	if err := json.Unmarshal(data, &c); err != nil {
		return Config{}, fmt.Errorf("parse %s: %w", p.ConfigFile, err)
	}
	return c, nil
}

// Save writes <brain>/config.json atomically (temp + rename).
func (c *Config) Save(p BrainPaths) error {
	c.UpdatedAt = nowISO()
	if err := os.MkdirAll(p.Dir, 0o755); err != nil {
		return fmt.Errorf("create %s: %w", p.Dir, err)
	}
	return writeJSONAtomic(p.ConfigFile, c)
}

// nowISO returns the current time in RFC3339 UTC. Wrapped so tests can
// freeze it via timeNow.
var timeNow = time.Now

func nowISO() string {
	return timeNow().UTC().Format(time.RFC3339)
}

// writeJSONAtomic mirrors the temp+rename pattern in state/installed.go so
// we never leave a half-written JSON file on disk.
func writeJSONAtomic(path string, v interface{}) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), filepath.Base(path)+".tmp-*")
	if err != nil {
		return fmt.Errorf("tmp file: %w", err)
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
	if err := os.Rename(tmpName, path); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("rename → %s: %w", path, err)
	}
	return nil
}
