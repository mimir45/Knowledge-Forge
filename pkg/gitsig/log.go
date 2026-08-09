// Package gitsig reads signals out of a repository's history: which files churn, who owns
// them, and which files keep changing together. It feeds churn.md and moc/codebase.md.
//
// It shells out to git rather than using go-git, which STACK names. That is a deliberate
// deviation and it is for consistency, not preference: pkg/coderef, pkg/drift, pkg/dataset
// and pkg/codeindex all already shell out, they shipped in merged phases, and drift clears
// its 100ms hook budget doing it. A second, differently-behaved git implementation in the
// same binary would be a source of disagreements about what a revision means.
//
// Everything here reads committed history only. Like drift, it never looks at the working
// tree — a half-finished edit is not churn.
package gitsig

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// Commit is one commit and the files it touched.
type Commit struct {
	SHA    string
	Author string
	When   time.Time
	Files  []string
}

// logFormat delimits records with NUL so a commit body can never be mistaken for the next
// record, and fields with a byte no author name or date contains.
const logFormat = "--pretty=format:%x00%H\x1f%an\x1f%aI"

// Log reads every commit touching the repository since a cutoff. A zero cutoff reads the
// whole history.
func Log(root string, since time.Time) ([]Commit, error) {
	args := []string{"-C", root, "log", logFormat, "--name-only", "--no-renames", "--no-merges"}
	if !since.IsZero() {
		args = append(args, "--since="+since.Format(time.RFC3339))
	}
	out, err := exec.Command("git", args...).Output()
	if err != nil {
		// git's own message is the only thing that distinguishes "not a repository" from a
		// bad revision or a missing directory, and "exit status 128" distinguishes nothing.
		return nil, fmt.Errorf("git log in %s: %w", root, withStderr(err))
	}
	return parseLog(string(out)), nil
}

// withStderr promotes git's diagnostic out of the ExitError, where nothing prints it.
func withStderr(err error) error {
	var ee *exec.ExitError
	if errors.As(err, &ee) && len(ee.Stderr) > 0 {
		return errors.New(strings.TrimSpace(string(ee.Stderr)))
	}
	return err
}

func parseLog(out string) []Commit {
	var commits []Commit
	for _, rec := range strings.Split(out, "\x00") {
		if c, ok := parseRecord(rec); ok {
			commits = append(commits, c)
		}
	}
	return commits
}

// parseRecord reads one NUL-delimited record: a header line of SHA/author/date, then the
// touched paths, one per line.
func parseRecord(rec string) (Commit, bool) {
	head, rest, _ := strings.Cut(strings.TrimLeft(rec, "\n"), "\n")
	f := strings.Split(head, "\x1f")
	if len(f) != 3 {
		return Commit{}, false
	}
	when, err := time.Parse(time.RFC3339, f[2])
	if err != nil {
		return Commit{}, false
	}
	return Commit{SHA: f[0], Author: f[1], When: when, Files: nonEmptyLines(rest)}, true
}

func nonEmptyLines(s string) []string {
	var out []string
	for _, l := range strings.Split(s, "\n") {
		if l = strings.TrimSpace(l); l != "" {
			out = append(out, l)
		}
	}
	return out
}
