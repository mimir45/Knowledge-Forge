package coderef

import (
	"os/exec"
	"path/filepath"
	"strings"
)

// ScanRepo lists the source files git tracks at the given revision. Tracked files only,
// and at a named revision rather than the working tree: drift is a pure function of
// (note refs, tree state), so an untracked scratch file must not make a citation
// resolve and a half-finished edit must not make one break.
func ScanRepo(name, root, rev string) (Repo, error) {
	out, err := exec.Command("git", "-C", root, "ls-tree", "-r", "--name-only", rev).Output()
	if err != nil {
		return Repo{}, err
	}
	r := Repo{Name: name, Root: root}
	for _, line := range strings.Split(strings.TrimRight(string(out), "\n"), "\n") {
		if line != "" && sourceExt[strings.ToLower(filepath.Ext(line))] {
			r.Files = append(r.Files, filepath.ToSlash(line))
		}
	}
	return r, nil
}

// HeadSHA resolves a revision to a full sha, which is what a note's drift_checked_at
// records and what the next run diffs against.
func HeadSHA(root, rev string) (string, error) {
	out, err := exec.Command("git", "-C", root, "rev-parse", rev).Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// ChangedFiles is drift's cheap gate: the files touched between two revisions. AST
// comparison runs only on this set, never on the whole repository and never on the
// vault. On a typical commit it is single digits.
//
// --no-renames is load-bearing. With rename detection on, git reports only a rename's
// destination, and a note citing the old path would then sit outside the gate and be
// scored OK while its file no longer exists.
func ChangedFiles(root, since, head string) ([]string, error) {
	out, err := exec.Command("git", "-C", root, "diff", "--name-only", "--no-renames",
		since+".."+head).Output()
	if err != nil {
		return nil, err
	}
	var files []string
	for _, line := range strings.Split(strings.TrimRight(string(out), "\n"), "\n") {
		if line != "" {
			files = append(files, filepath.ToSlash(line))
		}
	}
	return files, nil
}
