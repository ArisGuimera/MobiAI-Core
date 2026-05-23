package graph

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Write serializes idx to path atomically: writes to <path>.tmp, fsyncs, then
// renames over the destination. Parent directories are created if needed and
// the JSON is indented with two spaces for diff-friendly storage.
func Write(path string, idx *Index) error {
	if idx == nil {
		return fmt.Errorf("índice nulo")
	}

	data, err := json.MarshalIndent(idx, "", "  ")
	if err != nil {
		return fmt.Errorf("serializar índice: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("crear directorio padre: %w", err)
	}

	tmp := path + ".tmp"
	f, err := os.OpenFile(tmp, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		return fmt.Errorf("crear archivo temporal: %w", err)
	}
	if _, err := f.Write(data); err != nil {
		f.Close()
		os.Remove(tmp)
		return fmt.Errorf("escribir archivo temporal: %w", err)
	}
	if err := f.Sync(); err != nil {
		f.Close()
		os.Remove(tmp)
		return fmt.Errorf("sync archivo temporal: %w", err)
	}
	if err := f.Close(); err != nil {
		os.Remove(tmp)
		return fmt.Errorf("cerrar archivo temporal: %w", err)
	}

	if err := os.Rename(tmp, path); err != nil {
		os.Remove(tmp)
		return fmt.Errorf("renombrar archivo temporal: %w", err)
	}
	return nil
}

// Read loads and unmarshals an Index from path. It returns an error if the
// file cannot be read, the JSON is malformed, or the stored Version does not
// match IndexVersion (on-disk format incompatibility).
func Read(path string) (*Index, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("leer índice: %w", err)
	}

	var idx Index
	if err := json.Unmarshal(data, &idx); err != nil {
		return nil, fmt.Errorf("parsear índice: %w", err)
	}

	if idx.Version != IndexVersion {
		return nil, fmt.Errorf("versión del índice incompatible: %d (esperado %d)", idx.Version, IndexVersion)
	}
	return &idx, nil
}
