package main

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/mimir45/Knowledge-Forge/pkg/codeindex"
	"github.com/mimir45/Knowledge-Forge/pkg/coderef"
	"github.com/mimir45/Knowledge-Forge/pkg/drift"
	"github.com/mimir45/Knowledge-Forge/pkg/sentinel"
)

// markerStyles is deliberately narrower than sentinel's Style set.
var markerStyles = map[string]sentinel.Style{
	"java":       sentinel.Slash,
	"typescript": sentinel.Slash,
}

// marker is one symbol-shaped citation resolved to a place in the target repo.
type marker struct {
	path  string
	sym   codeindex.Symbol
	slugs []string
}

// resolveMarkers folds every note's symbol-shaped refs into one marker per.
func resolveMarkers(rels []noteRef, src symbolFinder, repoName string) []marker {
	byKey := map[string]*marker{}
	var order []string
	for _, nr := range rels {
		for _, ref := range nr.refs {
			if ref.Symbol == "" {
				continue
			}
			repo, path, sym, ok := src.Find(ref.Symbol, "")
			if !ok || repo != repoName {
				continue
			}
			key := path + "\x00" + sym.Name
			if m, exists := byKey[key]; exists {
				if !contains(m.slugs, nr.slug) {
					m.slugs = append(m.slugs, nr.slug)
				}
				continue
			}
			byKey[key] = &marker{path: path, sym: sym, slugs: []string{nr.slug}}
			order = append(order, key)
		}
	}
	out := make([]marker, 0, len(order))
	for _, key := range order {
		out = append(out, *byKey[key])
	}
	return out
}

func contains(ss []string, s string) bool {
	for _, v := range ss {
		if v == s {
			return true
		}
	}
	return false
}

// noteRef pairs one note's slug with the refs its body/frontmatter cite — the input
// resolveMarkers needs, projected once so it doesn't reload the vault twice.
type noteRef struct {
	slug string
	refs []coderef.Ref
}

func noteRefsOf(vaultRoot string) ([]noteRef, error) {
	notes, err := loadNotes(vaultRoot)
	if err != nil {
		return nil, err
	}
	slugs := slugsOf(notes)
	out := make([]noteRef, 0, len(notes))
	for _, n := range driftNotes(notes) {
		out = append(out, noteRef{slug: slugs[n.Rel], refs: n.Refs})
	}
	return out, nil
}

func markerID(sym codeindex.Symbol) string { return "logback:" + sym.Name }

func markerBody(slugs []string) string {
	links := make([]string, len(slugs))
	for i, s := range slugs {
		links[i] = "[[" + s + "]]"
	}
	return "forge: " + strings.Join(links, " ")
}

// writeInlineMarkers is opt-in (static.logback.inline_markers) and degrades cleanly on
// a CGO_ENABLED=0 build: no symbol table means no anchor line.
func writeInlineMarkers(vaultRoot string, r drift.Repo, src symbolFinder, dryRun bool) bool {
	return applyMarkers(vaultRoot, r, src, dryRun, "inline markers", "written",
		func(abs, id string, style sentinel.Style, m marker) error {
			return sentinel.UpsertBefore(abs, id, style, markerBody(m.slugs), m.sym.Start)
		})
}

// removeMarkers strips every marker forge logback could have written.
func removeMarkers(vaultRoot string, r drift.Repo, src symbolFinder, dryRun bool) bool {
	return applyMarkers(vaultRoot, r, src, dryRun, "--remove-markers", "removed",
		func(abs, id string, style sentinel.Style, m marker) error {
			return sentinel.Remove(abs, id, style)
		})
}

// applyMarkers is the walk writeInlineMarkers and removeMarkers share.
func applyMarkers(vaultRoot string, r drift.Repo, src symbolFinder, dryRun bool,
	errLabel, verb string, apply func(abs, id string, style sentinel.Style, m marker) error) bool {
	if !codeindex.Available() {
		fmt.Printf("%s: %s skipped: %v\n", r.Name, errLabel, codeindex.ErrUnavailable)
		return true
	}
	refs, err := noteRefsOf(vaultRoot)
	if err != nil {
		fmt.Printf("%s: %s: %v\n", r.Name, errLabel, err)
		return false
	}
	ok := true
	for _, m := range resolveMarkers(refs, src, r.Name) {
		if !applyOneMarker(r, m, dryRun, verb, apply) {
			ok = false
		}
	}
	return ok
}

// applyOneMarker applies (or dry-run announces) one marker.
func applyOneMarker(r drift.Repo, m marker, dryRun bool, verb string,
	apply func(abs, id string, style sentinel.Style, m marker) error) bool {
	style, known := markerStyles[codeindex.Lang(m.path)]
	if !known {
		return true
	}
	abs := filepath.Join(r.Root, m.path)
	if dryRun {
		fmt.Printf("%s: %s: marker for %s would be %s (dry run)\n", r.Name, m.path, m.sym.Name, verb)
		return true
	}
	if err := apply(abs, markerID(m.sym), style, m); err != nil {
		fmt.Printf("%s: %s: %v\n", r.Name, m.path, err)
		return false
	}
	fmt.Printf("%s: %s: marker for %s %s\n", r.Name, m.path, m.sym.Name, verb)
	return true
}
