package coderef

import (
	"os/exec"
	"path/filepath"
	"strings"
)

// runGit runs one git subcommand rooted at root and returns its trimmed stdout.
func runGit(root string, args ...string) (string, error) {
	out, err := exec.Command("git", append([]string{"-C", root}, args...)...).Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}

// ScanRepo lists the source files git tracks at the given revision.
func ScanRepo(name, root, rev string) (Repo, error) {
	out, err := runGit(root, "ls-tree", "-r", "--name-only", rev)
	if err != nil {
		return Repo{}, err
	}
	r := Repo{Name: name, Root: root}
	for _, line := range strings.Split(strings.TrimRight(out, "\n"), "\n") {
		if line != "" && sourceExt[strings.ToLower(filepath.Ext(line))] {
			r.Files = append(r.Files, filepath.ToSlash(line))
		}
	}
	return r, nil
}

// HeadSHA resolves a revision to a full sha, which is what a note's drift_checked_at
// records and what the next run diffs against.
func HeadSHA(root, rev string) (string, error) {
	out, err := runGit(root, "rev-parse", rev)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}

// ChangedFiles is drift's cheap gate: the files touched between two revisions.
func ChangedFiles(root, since, head string) ([]string, error) {
	out, err := runGit(root, "diff", "--name-only", "--no-renames", since+".."+head)
	if err != nil {
		return nil, err
	}
	var files []string
	for _, line := range strings.Split(strings.TrimRight(out, "\n"), "\n") {
		if line != "" {
			files = append(files, filepath.ToSlash(line))
		}
	}
	return files, nil
}

// ChangedFile is one path git reported touched between two revisions, with enough of
// --name-status's verdict to tell a same-commit deletion from an edit.
type ChangedFile struct {
	Path    string
	Deleted bool
}

// ChangedFilesStatus is ChangedFiles' status-aware sibling: it can tell "this path was
// edited" from "this path is gone" without a second git subprocess.
func ChangedFilesStatus(root, since, head string) ([]ChangedFile, error) {
	out, err := runGit(root, "diff", "--name-status", "--no-renames", since+".."+head)
	if err != nil {
		return nil, err
	}
	var files []ChangedFile
	for _, line := range strings.Split(strings.TrimRight(out, "\n"), "\n") {
		status, p, ok := strings.Cut(line, "\t")
		if !ok || !sourceExt[strings.ToLower(filepath.Ext(p))] {
			continue
		}
		files = append(files, ChangedFile{
			Path:    filepath.ToSlash(p),
			Deleted: strings.HasPrefix(status, "D"),
		})
	}
	return files, nil
}
