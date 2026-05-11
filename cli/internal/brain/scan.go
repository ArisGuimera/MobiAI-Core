package brain

import (
	"encoding/json"
	"fmt"
	"os"
)

// ScanVersion is the schema version of scan.json.
const ScanVersion = 1

// Project type constants used in Config and Scan.
const (
	ProjectTypeUnknown     = "unknown"
	ProjectTypeAndroid     = "android"
	ProjectTypeIOS         = "ios"
	ProjectTypeKMP         = "kmp"
	ProjectTypeFlutter     = "flutter"
	ProjectTypeReactNative = "react_native"
)

// Platform constants.
const (
	PlatformAndroid = "android"
	PlatformIOS     = "ios"
	PlatformShared  = "shared"
)

// Scan is the persisted shape of <brain>/scan.json. Field names match the
// spec in MobiAI Brain Phase 1.
type Scan struct {
	Version       int               `json:"version"`
	ProjectName   string            `json:"project_name"`
	ProjectType   string            `json:"project_type"`
	Platforms     []string          `json:"platforms"`
	BuildSystems  []string          `json:"build_systems"`
	UI            []string          `json:"ui"`
	DI            []string          `json:"di"`
	Network       []string          `json:"network"`
	Persistence   []string          `json:"persistence"`
	Serialization []string          `json:"serialization"`
	Testing       []string          `json:"testing"`
	Integrations  []string          `json:"integrations"`
	CICD          []string          `json:"ci_cd"`
	DetectedFiles map[string]string `json:"detected_files"`
	Warnings      []string          `json:"warnings"`
	CreatedAt     string            `json:"created_at"`
	UpdatedAt     string            `json:"updated_at"`
}

// LoadScan reads <brain>/scan.json. Returns (nil, nil) if the file does
// not exist (a fresh brain that hasn't been scanned yet).
func LoadScan(p BrainPaths) (*Scan, error) {
	data, err := os.ReadFile(p.ScanFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read scan.json: %w", err)
	}
	var s Scan
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("parse scan.json: %w", err)
	}
	return &s, nil
}

// Save writes scan.json atomically, refreshing UpdatedAt and seeding
// CreatedAt the first time around.
func (s *Scan) Save(p BrainPaths) error {
	now := nowISO()
	if s.CreatedAt == "" {
		s.CreatedAt = now
	}
	s.UpdatedAt = now
	if err := os.MkdirAll(p.Dir, 0o755); err != nil {
		return fmt.Errorf("crear %s: %w", p.Dir, err)
	}
	return writeJSONAtomic(p.ScanFile, s)
}

// newScan returns an empty Scan with allocated slices/maps so detectors
// can append without nil checks.
func newScan(projectName string) *Scan {
	return &Scan{
		Version:       ScanVersion,
		ProjectName:   projectName,
		ProjectType:   ProjectTypeUnknown,
		Platforms:     []string{},
		BuildSystems:  []string{},
		UI:            []string{},
		DI:            []string{},
		Network:       []string{},
		Persistence:   []string{},
		Serialization: []string{},
		Testing:       []string{},
		Integrations:  []string{},
		CICD:          []string{},
		DetectedFiles: map[string]string{},
		Warnings:      []string{},
	}
}
