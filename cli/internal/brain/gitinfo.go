package brain

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"strings"
	"time"
)

// gitTimeout bounds every git call the brain makes, so a hung git (a
// credential prompt, a huge repo) never blocks a command forever.
const gitTimeout = 10 * time.Second

// runGit runs git in dir and returns its stdout. stdin, when not empty, is
// fed to the process. Tests replace it to count or fake calls.
var runGit = func(dir, stdin string, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), gitTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0", "GIT_OPTIONAL_LOCKS=0")
	if stdin != "" {
		cmd.Stdin = strings.NewReader(stdin)
	}
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	return out.String(), err
}

// branchCandidates is the detection order for the main branch after the
// remote's default (refs/remotes/origin/HEAD).
var branchCandidates = []string{
	"refs/remotes/origin/main",
	"refs/remotes/origin/master",
	"refs/heads/main",
	"refs/heads/master",
}

// detectConfirmedBranch returns the branch a note must reach to count as
// confirmed: config's confirmed_branch when it exists, else the remote's
// default branch (origin/HEAD, e.g. develop in git-flow teams), else
// origin/main, origin/master, main, master. "" when there is no git or no
// such branch: then every note counts as confirmed.
func detectConfirmedBranch(root string, cfg Config) string {
	if b := strings.TrimSpace(cfg.ConfirmedBranch); b != "" {
		if _, err := runGit(root, "", "rev-parse", "--verify", "--quiet", b+"^{commit}"); err == nil {
			return b
		}
	}
	args := append([]string{"for-each-ref", "--format=%(refname)%00%(symref)", "refs/remotes/origin/HEAD"}, branchCandidates...)
	out, err := runGit(root, "", args...)
	if err != nil {
		return ""
	}
	present := map[string]bool{}
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		ref, target, _ := strings.Cut(line, "\x00")
		if ref == "refs/remotes/origin/HEAD" && target != "" {
			return shortRef(target)
		}
		present[ref] = true
	}
	for _, ref := range branchCandidates {
		if present[ref] {
			return shortRef(ref)
		}
	}
	return ""
}

// shortRef turns refs/heads/x into x and refs/remotes/o/x into o/x.
func shortRef(ref string) string {
	if s, ok := strings.CutPrefix(ref, "refs/heads/"); ok {
		return s
	}
	if s, ok := strings.CutPrefix(ref, "refs/remotes/"); ok {
		return s
	}
	return ref
}
