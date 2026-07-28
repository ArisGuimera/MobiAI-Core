package catalog

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// CommunityPack is the catalog name of the community pack — the only pack
// whose skills are installable individually (the picker drills into it and the
// CLI accepts `community/<skill-id>` specs). Every other pack installs whole.
const CommunityPack = "community"

// Catalog is the aggregated view: marketplace + every plugin manifest +
// resolved skill list per pack. All file IO happens in Load.
type Catalog struct {
	Root        string
	Marketplace *Marketplace
	Packs       []Pack
	byName      map[string]int // pack name → Packs index
}

// Pack pairs a marketplace entry with its resolved plugin.json.
type Pack struct {
	Ref      PluginRef
	Manifest PluginManifest
}

// Skill is a single skill discovered under a pack's skills directories.
// AbsPath is the absolute path to the skill directory (parent of SKILL.md).
type Skill struct {
	ID          string // directory name (e.g., "using-mobiai")
	AbsPath     string
	Description string // `description:` from SKILL.md frontmatter (may be empty)
}

// Load reads the marketplace and every plugin manifest under rootDir.
// Returns an error if marketplace.json or any referenced plugin.json is missing.
func Load(rootDir string) (*Catalog, error) {
	mp, err := LoadMarketplace(rootDir)
	if err != nil {
		return nil, err
	}
	c := &Catalog{
		Root:        rootDir,
		Marketplace: mp,
		Packs:       make([]Pack, 0, len(mp.Plugins)),
		byName:      make(map[string]int, len(mp.Plugins)),
	}
	for i, ref := range mp.Plugins {
		pm, err := LoadPlugin(rootDir, ref.Source)
		if err != nil {
			return nil, fmt.Errorf("pack %q: %w", ref.Name, err)
		}
		c.Packs = append(c.Packs, Pack{Ref: ref, Manifest: *pm})
		c.byName[ref.Name] = i
	}
	return c, nil
}

// Get returns the pack with the given name.
func (c *Catalog) Get(name string) (*Pack, error) {
	idx, ok := c.byName[name]
	if !ok {
		return nil, fmt.Errorf("pack %q not found in catalog", name)
	}
	return &c.Packs[idx], nil
}

// Has reports whether a pack with the given name exists.
func (c *Catalog) Has(name string) bool {
	_, ok := c.byName[name]
	return ok
}

// Reindex rebuilds the internal name → index map from c.Packs.
// Useful when callers (notably tests) construct a Catalog by hand
// instead of going through Load.
func (c *Catalog) Reindex() {
	c.byName = make(map[string]int, len(c.Packs))
	for i, p := range c.Packs {
		c.byName[p.Ref.Name] = i
	}
}

// Skills enumerates the skills declared by pack p.
// A skill is any direct subdirectory of one of the pack's skill dirs that
// contains a SKILL.md file. If the manifest's Skills field is empty,
// "./skills/" is used as the single default.
func (c *Catalog) Skills(p *Pack) ([]Skill, error) {
	dirs := p.Manifest.Skills
	if len(dirs) == 0 {
		dirs = []string{"./skills/"}
	}
	var out []Skill
	seen := make(map[string]bool)
	for _, rel := range dirs {
		base := filepath.Join(c.Root, filepath.FromSlash(p.Ref.Source), filepath.FromSlash(rel))
		entries, err := os.ReadDir(base)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, fmt.Errorf("read skills dir %s: %w", base, err)
		}
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			abs := filepath.Join(base, e.Name())
			if _, err := os.Stat(filepath.Join(abs, "SKILL.md")); err != nil {
				continue
			}
			if seen[abs] {
				continue
			}
			seen[abs] = true
			out = append(out, Skill{
				ID:          e.Name(),
				AbsPath:     abs,
				Description: readSkillDescription(filepath.Join(abs, "SKILL.md")),
			})
		}
	}
	return out, nil
}

// readSkillDescription extracts the `description:` value from a SKILL.md
// YAML frontmatter block (the lines between the first two `---` markers).
// It is intentionally tolerant: any IO/format problem yields an empty string
// so skill enumeration never fails just because a description is missing.
func readSkillDescription(skillMdPath string) string {
	f, err := os.Open(skillMdPath)
	if err != nil {
		return ""
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	inFrontmatter := false
	for sc.Scan() {
		line := sc.Text()
		trimmed := strings.TrimSpace(line)
		if !inFrontmatter {
			if trimmed == "---" {
				inFrontmatter = true
				continue
			}
			// No frontmatter at the top of the file → nothing to read.
			return ""
		}
		if trimmed == "---" {
			return "" // closed the block without finding a description
		}
		if rest, ok := strings.CutPrefix(line, "description:"); ok {
			return strings.TrimSpace(rest)
		}
	}
	return ""
}
