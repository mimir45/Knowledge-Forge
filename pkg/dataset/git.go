// Package dataset builds offline training datasets from
// what is already in the vault's git history. Nothing here calls a model; D3 is a pure
// function of (commit, tree state), the same anchoring rule drift follows.
package dataset

import (
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// Git is a minimal read-only shell over the git CLI, scoped to one repository. The D3
// capture runs inside a git hook, so git is already on the path and already the process
// that woke us; pulling go-git in for six plumbing reads would cost more than it buys.
// The library dependency belongs in pkg/gitsig (Phase 2b), which needs blame and churn.
type Git struct{ Dir string }

func (g Git) run(args ...string) (string, error) {
	cmd := exec.Command("git", append([]string{"-C", g.Dir}, args...)...)
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git %s: %w", strings.Join(args, " "), err)
	}
	return strings.TrimRight(string(out), "\n"), nil
}

// CommitMeta resolves a revision to its full sha and committer date. Committer date, not
// author date: a rebase rewrites the former and leaves the latter, and the window we are
// measuring is "how long the note sat in the repo before a human touched it".
func (g Git) CommitMeta(rev string) (sha string, when time.Time, err error) {
	out, err := g.run("show", "-s", "--format=%H%x00%cI", rev)
	if err != nil {
		return "", time.Time{}, err
	}
	sha, iso, ok := strings.Cut(out, "\x00")
	if !ok {
		return "", time.Time{}, fmt.Errorf("unparseable commit meta %q", out)
	}
	when, err = time.Parse(time.RFC3339, iso)
	return sha, when, err
}

// Trailers returns the commit message's trailer block, one "Key: value" per line.
func (g Git) Trailers(rev string) (string, error) {
	return g.run("show", "-s", "--format=%(trailers)", rev)
}

// ModifiedFiles lists paths a commit changed in the given ways ("M", "MR", …), reporting
// the post-rename spelling. -M matters because the realistic human correction retitles the
// note and renames the file in the same commit; without it git splits that into an add and
// a delete and neither looks like an edit. A root commit has no parent to diff against, so
// it reports nothing rather than erroring.
func (g Git) ModifiedFiles(sha, filter string) ([]string, error) {
	if _, err := g.run("rev-parse", "--verify", "--quiet", sha+"^"); err != nil {
		return nil, nil
	}
	out, err := g.run("diff-tree", "--no-commit-id", "--name-only", "-r", "-M",
		"--diff-filter="+filter, sha)
	if err != nil || out == "" {
		return nil, err
	}
	return strings.Split(out, "\n"), nil
}

// Origin locates the commit that introduced a path, following it across renames, and
// reports the sha, the date, and the path as it was spelled back then. Following matters:
// the Phase 1 migration moved every note, and without --follow every pre-existing note
// would look like it was created that day. The old spelling matters for the same reason —
// Show reads by (sha, path), and the post-rename path does not exist in that commit.
type Origin struct {
	SHA  string
	Path string
	When time.Time
}

func (g Git) FirstAdded(sha, path string) (Origin, error) {
	out, err := g.run("log", "--follow", "--diff-filter=A", "--format=%H%x00%cI",
		"--name-only", "-1", sha, "--", path)
	if err != nil || out == "" {
		return Origin{}, err
	}
	return parseFirstAdded(out)
}

func parseFirstAdded(out string) (Origin, error) {
	lines := strings.Split(out, "\n")
	addSHA, iso, ok := strings.Cut(lines[0], "\x00")
	if !ok || len(lines) < 2 {
		return Origin{}, fmt.Errorf("unparseable log output %q", out)
	}
	when, err := time.Parse(time.RFC3339, iso)
	// --name-only appends a blank line then the paths; the last non-empty one is ours.
	old := ""
	for _, l := range lines[1:] {
		if strings.TrimSpace(l) != "" {
			old = l
		}
	}
	return Origin{SHA: addSHA, Path: old, When: when}, err
}

// Show reads a path's contents at a revision, straight from the object store, so it is
// unaffected by whatever the working tree happens to hold.
func (g Git) Show(sha, path string) (string, error) {
	return g.run("show", sha+":"+path)
}
