package host

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ArisGuimera/MobiAI-Core/cli/internal/catalog"
)

// genericAdapter implements every HostAdapter method against a configurable
// home subdirectory. Tier-1 adapters are thin factories around it.
type genericAdapter struct {
	id         string
	name       string
	homepage   string
	homeSubdir string // e.g., ".claude"
	caps       Caps
}

func (a *genericAdapter) ID() string         { return a.id }
func (a *genericAdapter) Name() string       { return a.name }
func (a *genericAdapter) Homepage() string   { return a.homepage }
func (a *genericAdapter) Capabilities() Caps { return a.caps }

func (a *genericAdapter) hostDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, a.homeSubdir)
}

func (a *genericAdapter) SkillsDir() string {
	return filepath.Join(a.hostDir(), "skills")
}

func (a *genericAdapter) Detect() DetectResult {
	dir := a.hostDir()
	if dir == "" {
		return DetectResult{Found: false, Searched: []string{"<unresolved home>"}}
	}
	info, err := os.Stat(dir)
	if err == nil && info.IsDir() {
		return DetectResult{Found: true, Path: dir, Searched: []string{dir}}
	}
	return DetectResult{Found: false, Searched: []string{dir}}
}

// Install copies each skill's source directory into <SkillsDir>/<skill.ID>/.
// Existing files in the target are overwritten (Install is idempotent).
func (a *genericAdapter) Install(skills []catalog.Skill) error {
	if err := os.MkdirAll(a.SkillsDir(), 0o755); err != nil {
		return fmt.Errorf("mkdir skills dir: %w", err)
	}
	for _, s := range skills {
		dst := filepath.Join(a.SkillsDir(), s.ID)
		if err := copyDir(s.AbsPath, dst); err != nil {
			return fmt.Errorf("install skill %q: %w", s.ID, err)
		}
	}
	return nil
}

// Uninstall removes <SkillsDir>/<skillID>/ for each ID. Missing directories
// are silently ignored (idempotent).
func (a *genericAdapter) Uninstall(skillIDs []string) error {
	for _, id := range skillIDs {
		dir := filepath.Join(a.SkillsDir(), id)
		if err := os.RemoveAll(dir); err != nil {
			return fmt.Errorf("uninstall skill %q: %w", id, err)
		}
	}
	return nil
}

// List returns every direct subdirectory of SkillsDir as an InstalledSkill.
// Loose files in SkillsDir are ignored.
func (a *genericAdapter) List() ([]InstalledSkill, error) {
	dir := a.SkillsDir()
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read skills dir %s: %w", dir, err)
	}
	var out []InstalledSkill
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		out = append(out, InstalledSkill{
			ID:   e.Name(),
			Path: filepath.Join(dir, e.Name()),
		})
	}
	return out, nil
}

// Verify is a stub. Real drift detection ships with `mobiai doctor` polish.
func (a *genericAdapter) Verify() DriftReport {
	return DriftReport{}
}
