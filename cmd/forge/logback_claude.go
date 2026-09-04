package main

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/mimir45/Knowledge-Forge/pkg/drift"
	"github.com/mimir45/Knowledge-Forge/pkg/report"
	"github.com/mimir45/Knowledge-Forge/pkg/sentinel"
)

// logbackSentinelID names the managed block every logback-written fragment and inline
// marker uses.
const logbackSentinelID = "logback"

// writeClaudeFragments upserts one managed block per documented module's CLAUDE.md.
func writeClaudeFragments(r drift.Repo, groups []report.CodeGroup, dryRun bool) bool {
	ok := true
	for _, g := range groups {
		if len(g.Notes) == 0 {
			continue
		}
		path := filepath.Join(r.Root, g.Name, "CLAUDE.md")
		if dryRun {
			fmt.Printf("%s: %s would be written (dry run)\n", r.Name, relOrDot(r.Root, path))
			continue
		}
		if err := sentinel.Upsert(path, logbackSentinelID, sentinel.Markdown, fragmentBody(g.Notes)); err != nil {
			fmt.Printf("%s: %s: %v\n", r.Name, relOrDot(r.Root, path), err)
			ok = false
			continue
		}
		fmt.Printf("%s: %s updated\n", r.Name, relOrDot(r.Root, path))
	}
	return ok
}

// fragmentBody is the block's inner text — one line, so a hand-edited CLAUDE.md that
// already has its own sections gains exactly one new paragraph, not a sprawling list.
func fragmentBody(notes []string) string {
	links := make([]string, len(notes))
	for i, s := range notes {
		links[i] = "[[" + s + "]]"
	}
	return "Relevant notes: " + strings.Join(links, " · ")
}

func relOrDot(root, path string) string {
	if rel, err := filepath.Rel(root, path); err == nil {
		return rel
	}
	return path
}
