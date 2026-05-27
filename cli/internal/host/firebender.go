package host

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ArisGuimera/MobiAI-Core/cli/internal/catalog"
)

// newFirebender returns a HostAdapter for Firebender.
func newFirebender() HostAdapter {
	return &firebenderAdapter{
		genericAdapter: genericAdapter{
			id:         "firebender",
			name:       "Firebender",
			homepage:   "https://firebender.com",
			homeSubdir: ".firebender",
			caps:       Caps{Skills: true},
		},
	}
}

type firebenderAdapter struct {
	genericAdapter
}

func (a *firebenderAdapter) projectDir() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("resolve current working directory for firebender: %w", err)
	}
	return filepath.Join(cwd, ".firebender"), nil
}

func (a *firebenderAdapter) SkillsDir() string {
	dir, err := a.projectDir()
	if err != nil {
		return ""
	}
	return filepath.Join(dir, "skills")
}

func (a *firebenderAdapter) Detect() DetectResult {
	dir, err := a.projectDir()
	if err != nil {
		return DetectResult{Found: false, Searched: []string{"<unresolved cwd>"}}
	}
	info, statErr := os.Stat(dir)
	if statErr == nil && info.IsDir() {
		return DetectResult{Found: true, Path: dir, Searched: []string{dir}}
	}
	return DetectResult{Found: false, Searched: []string{dir}}
}

func (a *firebenderAdapter) Install(skills []catalog.Skill) error {
	skillsDir := a.SkillsDir()
	if skillsDir == "" {
		return fmt.Errorf("resolve firebender skills dir")
	}
	if err := os.MkdirAll(skillsDir, 0o755); err != nil {
		return fmt.Errorf("mkdir skills dir: %w", err)
	}
	for _, s := range skills {
		dst := filepath.Join(skillsDir, s.ID)
		if err := copyDir(s.AbsPath, dst); err != nil {
			return fmt.Errorf("install skill %q: %w", s.ID, err)
		}
	}
	return nil
}

func (a *firebenderAdapter) Uninstall(skillIDs []string) error {
	skillsDir := a.SkillsDir()
	if skillsDir == "" {
		return fmt.Errorf("resolve firebender skills dir")
	}
	for _, id := range skillIDs {
		dir := filepath.Join(skillsDir, id)
		if err := os.RemoveAll(dir); err != nil {
			return fmt.Errorf("uninstall skill %q: %w", id, err)
		}
	}
	return nil
}

func (a *firebenderAdapter) List() ([]InstalledSkill, error) {
	dir := a.SkillsDir()
	if dir == "" {
		return nil, fmt.Errorf("resolve firebender skills dir")
	}
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
