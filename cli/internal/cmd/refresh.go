package cmd

import (
	"fmt"
	"io"
	"path/filepath"
	"sort"

	"github.com/ArisGuimera/MobiAI-Core/cli/internal/catalog"
	"github.com/ArisGuimera/MobiAI-Core/cli/internal/embedded"
	"github.com/ArisGuimera/MobiAI-Core/cli/internal/i18n"
	"github.com/ArisGuimera/MobiAI-Core/cli/internal/state"
)

// previousSkillIDs records, before `mobiai update` replaces them, which skill
// IDs each pack shipped in the catalogs the installed skills may have come
// from: the cached remote catalog and the one embedded in the running binary.
// It is only used to learn which skills left the catalog, so they can be
// removed. The snapshot must be taken eagerly: catalog.Skills reads the pack
// directories when called, so reading the cached catalog after the git pull
// would already return the new content. Sources that fail to load are skipped.
func previousSkillIDs(paths state.Paths) map[string]map[string]bool {
	ids := map[string]map[string]bool{}
	record := func(root string) {
		c, err := catalog.Load(root)
		if err != nil {
			return
		}
		for i := range c.Packs {
			skills, err := c.Skills(&c.Packs[i])
			if err != nil {
				continue
			}
			name := c.Packs[i].Ref.Name
			if ids[name] == nil {
				ids[name] = map[string]bool{}
			}
			for _, s := range skills {
				ids[name][s.ID] = true
			}
		}
	}
	record(paths.CatalogDir())
	if !embedded.IsEmpty() {
		root := filepath.Join(paths.Cache(), "embedded")
		if ensureEmbeddedExtracted(root) == nil {
			record(root)
		}
	}
	return ids
}

// refreshPlan is what one installed.json entry needs on each of its hosts.
type refreshPlan struct {
	key     string // installed.json key: a pack name or "community/<id>"
	install []catalog.Skill
	stale   []string // skill IDs to remove: they left the catalog
}

// refreshInstalledPacks re-copies every installed pack from c into the hosts
// recorded in installed.json, so `mobiai update` delivers new and changed
// skills without a separate `mobiai skills add`. Skills a previous catalog
// shipped but c no longer does are removed. Hosts that are no longer present
// are skipped instead of recreated. Per-host failures are reported as warnings,
// like a failed binary self-update, because the catalog sync already succeeded.
func refreshInstalledPacks(out io.Writer, c *catalog.Catalog, previous map[string]map[string]bool, g GlobalFlags, paths state.Paths) error {
	installed, err := state.LoadInstalled(paths)
	if err != nil {
		return err
	}
	if len(installed.Packs) == 0 {
		return nil
	}

	keys := make([]string, 0, len(installed.Packs))
	for k := range installed.Packs {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var plans []refreshPlan
	for _, key := range keys {
		packName, only := key, ""
		if id, ok := state.ParseCommunitySkillKey(key); ok {
			packName, only = catalog.CommunityPack, id
		}
		pack, err := c.Get(packName)
		if err != nil {
			fmt.Fprintf(out, i18n.T("  %s is no longer in the catalog; left as it is.\n"), key)
			continue
		}
		skills, err := c.Skills(pack)
		if err != nil {
			return fmt.Errorf("enumerate skills for %s: %w", packName, err)
		}
		plan := refreshPlan{key: key}
		if only != "" {
			plan.install = filterCatalogSkills(skills, []string{only})
			if len(plan.install) == 0 {
				plan.stale = []string{only}
			}
		} else {
			plan.install = skills
			plan.stale = staleSkillIDs(skills, previous[packName])
		}
		plans = append(plans, plan)
	}

	// A skill that moved between packs is stale for the old pack but wanted by
	// the new one; never remove what another installed pack still ships.
	wanted := map[string]map[string]bool{}
	for _, p := range plans {
		for _, h := range installed.Packs[p.key] {
			if wanted[h] == nil {
				wanted[h] = map[string]bool{}
			}
			for _, s := range p.install {
				wanted[h][s.ID] = true
			}
		}
	}

	registry := newRegistry(g)
	fmt.Fprint(out, i18n.T("Refreshing installed skills...\n"))
	for _, p := range plans {
		for _, hostID := range append([]string(nil), installed.Packs[p.key]...) {
			if len(g.Hosts) > 0 && !containsString(g.Hosts, hostID) {
				continue
			}
			h, err := registry.Get(hostID)
			if err != nil || !h.Detect().Found {
				fmt.Fprintf(out, i18n.T("  %s: skipped %s (not found on this machine).\n"), p.key, hostID)
				continue
			}
			var stale []string
			for _, id := range p.stale {
				if !wanted[hostID][id] {
					stale = append(stale, id)
				}
			}
			if len(stale) > 0 {
				if err := h.Uninstall(stale); err != nil {
					fmt.Fprintf(out, i18n.T("Warning: could not refresh %s in %s: %v\n"), p.key, h.Name(), err)
					continue
				}
			}
			if len(p.install) > 0 {
				if err := h.Install(p.install); err != nil {
					fmt.Fprintf(out, i18n.T("Warning: could not refresh %s in %s: %v\n"), p.key, h.Name(), err)
					continue
				}
			} else {
				// A per-skill community entry whose skill left the catalog.
				installed.Remove(p.key, hostID)
			}
			fmt.Fprintf(out, "✓ %s → %s\n", p.key, h.Name())
		}
	}
	return installed.Save(paths)
}

// staleSkillIDs returns the IDs in previous (what a pack used to ship) that the
// current version of the pack no longer ships.
func staleSkillIDs(current []catalog.Skill, previous map[string]bool) []string {
	keep := make(map[string]bool, len(current))
	for _, s := range current {
		keep[s.ID] = true
	}
	var stale []string
	for id := range previous {
		if !keep[id] {
			stale = append(stale, id)
		}
	}
	sort.Strings(stale)
	return stale
}
