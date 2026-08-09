package report

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

// CodeGroup is one module or package of a repository, as moc/codebase.md needs it.
type CodeGroup struct {
	Name      string
	Files     int
	Commits   int      // churn over the window
	Owners    []string // most-committing authors first
	DependsOn []string
	Notes     []string // slugs of notes that cite code in this group
}

// Uncovered is code that churns and has nothing written about it. This is §B.5's last
// line and the only output of the whole suite that is about the codebase rather than the
// wiki: high churn, real size, zero notes, ranked.
type Uncovered struct {
	Symbol  string
	Path    string
	LOC     int
	Commits int
}

// CodebaseInput is what moc/codebase.md renders from.
//
// The types here are local rather than pkg/codeindex's on purpose. codeindex is the one
// package that needs cgo for go-tree-sitter, and importing it would drag that requirement
// into pkg/report and from there into everything that renders a file. The caller in
// cmd/forge does the projection, which is where the cgo boundary already sits.
type CodebaseInput struct {
	Repo      string
	Groups    []CodeGroup
	Uncovered []Uncovered
	Days      int // the churn window, e.g. 90
	Now       time.Time
}

// RenderCodebase produces moc/codebase.md — the map from code to the notes about it.
//
// It is written into moc/ rather than reports/ because it is a map of content: a way into
// the vault organised by the shape of the system rather than by note type. That has a
// consequence the vault contract has to allow for — moc/ is in the link graph, so the
// wikilinks below genuinely de-orphan the notes they point at, which is what a MOC is for,
// and moc/ is exempt from the note contract because `type:` has no value for a map.
func RenderCodebase(in CodebaseInput) []byte {
	var b strings.Builder
	fmt.Fprintf(&b, "# Codebase map — %s — %s\n\n", in.Repo, in.Now.Format("2006-01-02"))
	fmt.Fprintf(&b, "%d %s · %d documented · churn over %d days\n",
		len(in.Groups), plural(len(in.Groups), "module", "modules"),
		documented(in.Groups), in.Days)
	writeGroups(&b, in)
	writeUndocumented(&b, in)
	return []byte(b.String())
}

func documented(gs []CodeGroup) int {
	n := 0
	for _, g := range gs {
		if len(g.Notes) > 0 {
			n++
		}
	}
	return n
}

func writeGroups(b *strings.Builder, in CodebaseInput) {
	if len(in.Groups) == 0 {
		empty(b, "no module could be resolved in this repository")
		return
	}
	for _, g := range in.Groups {
		writeGroup(b, g)
	}
}

func writeGroup(b *strings.Builder, g CodeGroup) {
	fmt.Fprintf(b, "\n## %s  (%d %s · %d commits", g.Name, g.Files,
		plural(g.Files, "file", "files"), g.Commits)
	if len(g.Owners) > 0 {
		fmt.Fprintf(b, " · owners: %s", strings.Join(head(g.Owners, 3), ", "))
	}
	b.WriteString(")\n")
	if len(g.DependsOn) > 0 {
		fmt.Fprintf(b, "Depends on: %s\n", strings.Join(g.DependsOn, ", "))
	}
	writeGroupNotes(b, g)
}

func writeGroupNotes(b *strings.Builder, g CodeGroup) {
	if len(g.Notes) == 0 {
		b.WriteString("Notes: _none_\n")
		return
	}
	links := make([]string, 0, len(g.Notes))
	for _, slug := range g.Notes {
		links = append(links, "[["+slug+"]]")
	}
	fmt.Fprintf(b, "Notes: %s\n", strings.Join(links, " · "))
}

// writeUndocumented ranks by commits first and size second: code that changes often is code
// whose behaviour is not settled, and undocumented churn is what actually costs a team.
// A large file nobody has touched in a year is not urgent, whatever its line count.
//
// The name keeps its distance from coverage.go's writeUncovered on purpose: that one lists
// stacks the wiki never mentions, this one lists symbols that move. Two different absences.
func writeUndocumented(b *strings.Builder, in CodebaseInput) {
	u := append([]Uncovered(nil), in.Uncovered...)
	sortUncovered(u)
	fmt.Fprintf(b, "\n## Undocumented and moving — %d\n", len(u))
	if len(u) == 0 {
		empty(b, "every symbol that changed in the window has a note")
		return
	}
	fmt.Fprintf(b, "\nHigh churn, real size, no note. Ranked by how often each changed "+
		"in the last %d days.\n\n", in.Days)
	for _, s := range head(u, 25) {
		fmt.Fprintf(b, "- ⚠️ `%s` — %d LOC, changed %dx, 0 notes — `%s`\n",
			s.Symbol, s.LOC, s.Commits, s.Path)
	}
}

func sortUncovered(u []Uncovered) {
	sort.Slice(u, func(i, j int) bool {
		if u[i].Commits != u[j].Commits {
			return u[i].Commits > u[j].Commits
		}
		if u[i].LOC != u[j].LOC {
			return u[i].LOC > u[j].LOC
		}
		return u[i].Symbol < u[j].Symbol
	})
}
