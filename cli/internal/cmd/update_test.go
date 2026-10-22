package cmd

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ArisGuimera/MobiAI-Core/cli/internal/state"
)

func TestUpdate_FromLocalRoot_UpdatesMeta(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("MOBIAI_HOME", filepath.Join(tmp, ".mobiai"))
	root := filepath.Join("..", "catalog", "testdata", "sample")

	cmd := NewUpdateCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"--catalog-root", root})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}

	if !strings.Contains(out.String(), "Skills catalog updated") {
		t.Errorf("output: %q", out.String())
	}

	paths, _ := state.NewPaths()
	meta, err := state.LoadMeta(paths)
	if err != nil {
		t.Fatal(err)
	}
	if time.Since(meta.LastSync) > time.Minute {
		t.Errorf("LastSync should be recent: %v", meta.LastSync)
	}
	if meta.Version == "" {
		t.Error("Version should be set after update")
	}
}

// makeBareCatalogRepo creates a bare git repo seeded from the project's
// catalog testdata fixture. Returns the file:// URL the bare repo can be
// cloned from. Skips the test if `git` is not installed.
func makeBareCatalogRepo(t *testing.T) string {
	t.Helper()
	_, url := makeCatalogRepos(t)
	return url
}

// makeCatalogRepos is makeBareCatalogRepo but also returns the working repo,
// so a test can push follow-up commits that the next `mobiai update` pulls.
func makeCatalogRepos(t *testing.T) (work, url string) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not on PATH; skipping remote-catalog test")
	}

	src, err := filepath.Abs(filepath.Join("..", "catalog", "testdata", "sample"))
	if err != nil {
		t.Fatal(err)
	}

	// Create a working repo from the fixture.
	work = filepath.Join(t.TempDir(), "work")
	if err := os.MkdirAll(work, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := copyDirContents(src, work); err != nil {
		t.Fatalf("seed work: %v", err)
	}
	mustRun(t, work, "git", "init", "--initial-branch=main")
	mustRun(t, work, "git", "-c", "user.email=t@t", "-c", "user.name=t", "add", ".")
	mustRun(t, work, "git", "-c", "user.email=t@t", "-c", "user.name=t", "commit", "-m", "seed")

	// Create a bare repo and push to it.
	// Use --initial-branch=main so HEAD points to main (not master), matching
	// the working repo's branch name.
	bare := filepath.Join(t.TempDir(), "bare.git")
	mustRun(t, "", "git", "init", "--bare", "--initial-branch=main", bare)
	mustRun(t, work, "git", "remote", "add", "origin", bare)
	mustRun(t, work, "git", "push", "origin", "main")

	// Convert filesystem path to a file:// URL portable across OSes.
	abs, err := filepath.Abs(bare)
	if err != nil {
		t.Fatal(err)
	}
	return work, "file:///" + filepath.ToSlash(abs)
}

func copyDirContents(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(src, path)
		target := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, data, 0o644)
	})
}

func mustRun(t *testing.T, dir, name string, args ...string) {
	t.Helper()
	c := exec.Command(name, args...)
	if dir != "" {
		c.Dir = dir
	}
	out, err := c.CombinedOutput()
	if err != nil {
		t.Fatalf("%s %v: %v\n%s", name, args, err, out)
	}
}

func TestUpdate_SkipBinary_DoesNotSelfUpdate(t *testing.T) {
	// Stamp a non-dev version so the self-update path would normally engage;
	// --skip-binary must short-circuit it (no network, just the skip notice).
	prev := version
	SetVersion("0.0.1")
	t.Cleanup(func() { SetVersion(prev) })

	tmp := t.TempDir()
	t.Setenv("MOBIAI_HOME", filepath.Join(tmp, ".mobiai"))
	root := filepath.Join("..", "catalog", "testdata", "sample")

	cmd := NewUpdateCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"--catalog-root", root, "--skip-binary"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}

	got := out.String()
	if !strings.Contains(got, "--skip-binary") {
		t.Errorf("expected skip-binary notice; got: %q", got)
	}
	// It must not have attempted a download / self-update. "downloading" and
	// the "→" arrow appear only on the self-update path.
	if strings.Contains(got, "downloading") || strings.Contains(got, "→") {
		t.Errorf("--skip-binary should not self-update; got: %q", got)
	}
}

func TestUpdate_GitClone_FreshCache(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("MOBIAI_HOME", filepath.Join(tmp, ".mobiai"))
	t.Setenv("MOBIAI_CATALOG_GIT_URL", makeBareCatalogRepo(t))

	cmd := NewUpdateCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}

	// Verify the repo was cloned to ~/.mobiai/cache/catalog/.
	cacheRoot := filepath.Join(tmp, ".mobiai", "cache", "catalog")
	if _, err := os.Stat(filepath.Join(cacheRoot, ".claude-plugin", "marketplace.json")); err != nil {
		t.Errorf("expected marketplace.json in cloned cache: %v", err)
	}
	if !strings.Contains(out.String(), "Cloning catalog") {
		t.Errorf("expected clone message; got: %q", out.String())
	}
	if !strings.Contains(out.String(), "Skills catalog updated") {
		t.Errorf("expected success message; got: %q", out.String())
	}

	// CatalogMeta should be fresh.
	paths, _ := state.NewPaths()
	meta, err := state.LoadMeta(paths)
	if err != nil {
		t.Fatal(err)
	}
	if time.Since(meta.LastSync) > time.Minute {
		t.Errorf("LastSync should be recent: %v", meta.LastSync)
	}
}

func TestUpdate_GitPull_ExistingCache(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("MOBIAI_HOME", filepath.Join(tmp, ".mobiai"))
	url := makeBareCatalogRepo(t)
	t.Setenv("MOBIAI_CATALOG_GIT_URL", url)

	// First update: clone.
	first := NewUpdateCmd()
	var out1 bytes.Buffer
	first.SetOut(&out1)
	first.SetErr(&out1)
	if err := first.Execute(); err != nil {
		t.Fatalf("first execute: %v", err)
	}

	// Second update: pull.
	second := NewUpdateCmd()
	var out2 bytes.Buffer
	second.SetOut(&out2)
	second.SetErr(&out2)
	if err := second.Execute(); err != nil {
		t.Fatalf("second execute: %v", err)
	}

	if !strings.Contains(out2.String(), "Syncing catalog (git pull)") {
		t.Errorf("expected pull message on 2nd run; got: %q", out2.String())
	}
}

// runCmd executes a fresh command with args and returns its combined output.
func runCmd(t *testing.T, c interface {
	SetOut(io.Writer)
	SetErr(io.Writer)
	SetArgs([]string)
	Execute() error
}, args ...string) string {
	t.Helper()
	var out bytes.Buffer
	c.SetOut(&out)
	c.SetErr(&out)
	c.SetArgs(args)
	if err := c.Execute(); err != nil {
		t.Fatalf("execute %v: %v\n%s", args, err, out.String())
	}
	return out.String()
}

// An update must re-copy the packs the user already installed: new skills
// land, changed skills get the new content, and skills that left the catalog
// are removed from the host.
func TestUpdate_RefreshesInstalledPacks(t *testing.T) {
	tmp := communityTestHome(t)
	work, url := makeCatalogRepos(t)
	t.Setenv("MOBIAI_CATALOG_GIT_URL", url)

	runCmd(t, NewUpdateCmd(), "--skip-binary")
	runCmd(t, NewSkillsCmd(), "add", "android", "--yes")

	skillsDir := filepath.Join(tmp, ".claude", "skills")
	if _, err := os.Stat(filepath.Join(skillsDir, "android-build", "SKILL.md")); err != nil {
		t.Fatalf("precondition: android-build should be installed: %v", err)
	}

	// Upstream catalog changes: edge-to-edge is edited, android-build leaves
	// the pack and new-skill joins it. android-build exists only in this
	// fixture (not in the catalog embedded in the test binary), so removing it
	// proves the pre-update snapshot of the cached catalog drives the pruning.
	android := filepath.Join(work, "skills", "android", "skills")
	edgeSkill := filepath.Join(android, "google", "edge-to-edge", "SKILL.md")
	if err := os.WriteFile(edgeSkill, []byte("---\nname: edge-to-edge\ndescription: test fixture\n---\nv2 body\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.RemoveAll(filepath.Join(android, "android-build")); err != nil {
		t.Fatal(err)
	}
	newSkill := filepath.Join(android, "google", "new-skill")
	if err := os.MkdirAll(newSkill, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(newSkill, "SKILL.md"), []byte("---\nname: new-skill\ndescription: test fixture\n---\nnew body\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	mustRun(t, work, "git", "-c", "user.email=t@t", "-c", "user.name=t", "add", "-A")
	mustRun(t, work, "git", "-c", "user.email=t@t", "-c", "user.name=t", "commit", "-m", "catalog v2")
	mustRun(t, work, "git", "push", "origin", "main")

	out := runCmd(t, NewUpdateCmd(), "--skip-binary")

	got, err := os.ReadFile(filepath.Join(skillsDir, "edge-to-edge", "SKILL.md"))
	if err != nil || !strings.Contains(string(got), "v2 body") {
		t.Errorf("edge-to-edge should have the new content; got %q (err %v)", got, err)
	}
	if _, err := os.Stat(filepath.Join(skillsDir, "new-skill", "SKILL.md")); err != nil {
		t.Errorf("new-skill should be installed by the update: %v", err)
	}
	if _, err := os.Stat(filepath.Join(skillsDir, "android-build")); !os.IsNotExist(err) {
		t.Errorf("android-build left the catalog and should be removed; stat err: %v", err)
	}
	if !strings.Contains(out, "android") {
		t.Errorf("output should report the refreshed pack; got: %q", out)
	}
}

// A skill that moves from one installed pack to another must survive the
// refresh: it is stale for the old pack but still shipped by the new one.
// community → android is the order-sensitive case, because "android" is
// refreshed first and "community" would then delete what it just installed.
func TestUpdate_RefreshKeepsSkillMovedBetweenPacks(t *testing.T) {
	tmp := communityTestHome(t)
	work, url := makeCatalogRepos(t)
	t.Setenv("MOBIAI_CATALOG_GIT_URL", url)

	runCmd(t, NewUpdateCmd(), "--skip-binary")
	runCmd(t, NewSkillsCmd(), "add", "android", "community", "--yes")

	from := filepath.Join(work, "skills", "community", "skills", "alpha-skill")
	to := filepath.Join(work, "skills", "android", "skills", "alpha-skill")
	if err := os.Rename(from, to); err != nil {
		t.Fatal(err)
	}
	mustRun(t, work, "git", "-c", "user.email=t@t", "-c", "user.name=t", "add", "-A")
	mustRun(t, work, "git", "-c", "user.email=t@t", "-c", "user.name=t", "commit", "-m", "move alpha-skill to android")
	mustRun(t, work, "git", "push", "origin", "main")

	runCmd(t, NewUpdateCmd(), "--skip-binary")

	if _, err := os.Stat(filepath.Join(tmp, ".claude", "skills", "alpha-skill", "SKILL.md")); err != nil {
		t.Errorf("alpha-skill moved to android and must stay installed: %v", err)
	}
}

// A host recorded in installed.json but no longer present (its config dir was
// deleted) must be skipped, not recreated by the refresh.
func TestUpdate_RefreshSkipsMissingHosts(t *testing.T) {
	tmp := communityTestHome(t)
	paths, err := state.NewPaths()
	if err != nil {
		t.Fatal(err)
	}
	installed := &state.Installed{Packs: map[string][]string{"android": {"codex"}}}
	if err := installed.Save(paths); err != nil {
		t.Fatal(err)
	}

	runCmd(t, NewUpdateCmd(), "--catalog-root", sampleCatalogRoot(t), "--skip-binary")

	if _, err := os.Stat(filepath.Join(tmp, ".codex")); !os.IsNotExist(err) {
		t.Errorf("refresh must not create the config dir of a missing host; stat err: %v", err)
	}
}
