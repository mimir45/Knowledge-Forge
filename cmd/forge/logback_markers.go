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

// markerStyles is deliberately narrower than sentinel's Style set: codeindex.Lang only
// parses Java and TypeScript (AUDIT §7's file-count call, documented on codeindex.Lang
// itself), so a Python entry here would gate markers on a language this binary can never
// find symbols in. Extending codeindex's grammars extends this table too.
var markerStyles = map[string]sentinel.Style{
	"java":       sentinel.Slash,
	"typescript": sentinel.Slash,
}

// marker is one symbol-shaped citation resolved to a place in the target repo. Only
// symbol citations get an inline marker — a file/path-only ref has no line to anchor to
// and is already covered by docs/knowledge-map.md and the module's CLAUDE.md fragment.
type marker struct {
	path  string
	sym   codeindex.Symbol
	slugs []string
}

// resolveMarkers folds every note's symbol-shaped refs into one marker per (path, symbol),
// merging slugs when more than one note cites the same symbol so a shared symbol gets one
// marker, not a stack of them. It keys off ref.Symbol rather than ref.Kind == KindSymbol:
// the canonical `code_refs:` form is path-shaped *and* carries a symbol
// ("app:Order.java#Order.place", Kind == KindPath), and that is the common case a note
// actually writes — filtering to bare-symbol citations only would skip nearly every real
// one. For each ref it looks the symbol up in the live symbol table (src, at HEAD — the
// same lookup locate() uses for coverage) and keeps only the ones that land in this
// repository.
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

// writeInlineMarkers is opt-in (static.logback.inline_markers) and degrades cleanly on a
// CGO_ENABLED=0 build: no symbol table means no anchor line, so it reports skipped rather
// than erroring the whole run.
func writeInlineMarkers(vaultRoot string, r drift.Repo, src symbolFinder, dryRun bool) bool {
	return applyMarkers(vaultRoot, r, src, dryRun, "inline markers", "written",
		func(abs, id string, style sentinel.Style, m marker) error {
			return sentinel.UpsertBefore(abs, id, style, markerBody(m.slugs), m.sym.Start)
		})
}

// removeMarkers strips every marker forge logback could have written, independent of the
// current inline_markers config — --remove-markers is meant to work even after the config
// gate was turned back off, which is when a leftover marker is most likely to be found.
func removeMarkers(vaultRoot string, r drift.Repo, src symbolFinder, dryRun bool) bool {
	return applyMarkers(vaultRoot, r, src, dryRun, "--remove-markers", "removed",
		func(abs, id string, style sentinel.Style, m marker) error {
			return sentinel.Remove(abs, id, style)
		})
}

// applyMarkers is the walk writeInlineMarkers and removeMarkers share: resolve every
// symbol citation in this repo to a marker, then apply one sentinel operation to each.
// errLabel, verb and apply are the only things that differ between a write and a remove.
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

// applyOneMarker applies (or dry-run announces) one marker, skipping languages
// markerStyles doesn't cover. It reports false only on a real apply error — an unknown
// language or a dry run both count as "nothing to fail."
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
