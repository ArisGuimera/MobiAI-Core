package state

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// CatalogMeta tracks the freshness of the local catalog cache.
// Persisted as <cache>/meta.json.
type CatalogMeta struct {
	LastSync time.Time `json:"last_sync"`
	Version  string    `json:"version,omitempty"`
	ETag     string    `json:"etag,omitempty"`
}

// LoadMeta reads meta.json. Returns a zero-valued CatalogMeta if
// the file does not exist.
func LoadMeta(p Paths) (*CatalogMeta, error) {
	path := p.MetaFile()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &CatalogMeta{}, nil
		}
		return nil, fmt.Errorf("read meta.json: %w", err)
	}
	var m CatalogMeta
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("parse meta.json: %w", err)
	}
	return &m, nil
}

// Save writes meta.json atomically.
func (m *CatalogMeta) Save(p Paths) error {
	if err := os.MkdirAll(p.Cache(), 0o755); err != nil {
		return fmt.Errorf("mkdir cache dir: %w", err)
	}
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal meta: %w", err)
	}
	final := p.MetaFile()
	tmp, err := os.CreateTemp(filepath.Dir(final), "meta-*.json.tmp")
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

// IsStale reports whether the catalog cache is older than maxAge.
// A zero LastSync is considered stale (no sync yet).
func (m *CatalogMeta) IsStale(maxAge time.Duration) bool {
	if m.LastSync.IsZero() {
		return true
	}
	return time.Since(m.LastSync) > maxAge
}
