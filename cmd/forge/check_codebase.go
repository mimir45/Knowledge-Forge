package main

import (
	"fmt"
	"path"
	"path/filepath"
	"sort"

	"knowledge-forge/pkg/codeindex"
	"knowledge-forge/pkg/coderef"
	"knowledge-forge/pkg/drift"
	"knowledge-forge/pkg/gitsig"
	"knowledge-forge/pkg/report"
)

// These make ADDENDUM section B.5's "high churn, real size" concrete. A file touched once
// is not churning and a six-line accessor with no note is not a documentation gap;
// listing either would bury the handful of symbols that genuinely are one.
const (
	minSymbolCommits = 2
	minSymbolLOC     = 15
	maxGroups        = 20
)

func (d *checkData) driftAndCode() {
	if len(d.cfg.repos) == 0 {
		return
	}
	rg, err := registryOf(d.cfg.repos)
	if err != nil {
		d.repoErr = err
		return
	}
	src := drift.NewGitSource(d.cfg.repos, filepath.Join(d.root, ".forge"))
	d.findings = drift.Check(driftNotes(d.notes), rg, src, nil, drift.Opts{Deep: true})
	d.code, d.codeErr = d.codebases(rg, src)
}

// codebases builds one section per repository. It is the only collector that needs
// pkg/codeindex, and therefore cgo: a binary built with CGO_ENABLED=0 renders every other
// report and reports this one as skipped, rather than writing a map that claims the
// codebase is fully documented because nothing could parse it.
func (d *checkData) codebases(rg *coderef.Registry, src symbolFinder) ([]report.CodebaseInput, error) {
	if !codeindex.Available() {
		return nil, codeindex.ErrUnavailable
	}
	cited := d.citedPaths(rg, src)
	out := make([]report.CodebaseInput, 0, len(d.cfg.repos))
	for _, r := range d.cfg.repos {
		in, err := d.oneCodebase(r.Name, r.Root, cited[r.Name])
		if err != nil {
			return nil, fmt.Errorf("%s: %w", r.Name, err)
		}
		out = append(out, in)
	}
	return out, nil
}

func (d *checkData) oneCodebase(name, root string, cited map[string][]string) (report.CodebaseInput, error) {
	commits, err := gitsig.Log(root, d.now.AddDate(0, 0, -d.cfg.days))
	if err != nil {
		return report.CodebaseInput{}, err
	}
	scanned, err := coderef.ScanRepo(name, root, "HEAD")
	if err != nil {
		return report.CodebaseInput{}, err
	}
	ix, err := codeindex.Build(name, root, "HEAD", scanned.Files)
	if err != nil {
		return report.CodebaseInput{}, err
	}
	st := gitsig.Analyze(commits)
	return report.CodebaseInput{Repo: name, Days: d.cfg.days, Now: d.now,
		Groups: groupsOf(scanned.Files, st, cited), Uncovered: uncoveredOf(ix, st, cited)}, nil
}

// symbolFinder is the single question coverage asks of drift's symbol table. Naming it
// here rather than taking drift.Source whole says what this collector depends on, and
// keeps the fake in the test to one method.
type symbolFinder interface {
	Find(name, asOf string) (repo, path string, sym codeindex.Symbol, ok bool)
}

// citedPaths resolves every note's code citations to repo-relative paths, keyed by what
// the repository calls a file rather than by the shorthand the note wrote. Resolution is
// the whole point: AUDIT NF-4 found 14 of 19 path-shaped refs matching no file, and an
// unresolved ref would make a documented module look undocumented.
func (d *checkData) citedPaths(rg *coderef.Registry, src symbolFinder) map[string]map[string][]string {
	out := map[string]map[string][]string{}
	for _, n := range driftNotes(d.notes) {
		for _, ref := range n.Refs {
			if repo, p, ok := locate(ref, rg, src); ok {
				if out[repo] == nil {
					out[repo] = map[string][]string{}
				}
				out[repo][p] = append(out[repo][p], d.slugs[n.Rel])
			}
		}
	}
	return out
}

// locate answers which file a citation is about, and it has to answer for symbol-only
// citations too — most of the vault's references name a class and no path. coderef leaves
// those without a RepoPath by design, so pkg/drift looks them up in the symbol table; a
// coverage pass that skipped them instead reported `SignUpPage` as "0 notes" in the same
// run where drift.md named two notes citing it.
//
// The lookup is drift's own, at HEAD, so both reports attribute a citation to the same
// file. Where two repositories declare one name that means crediting only the first in
// (repo, path) order — the arbitration drift already makes, and agreeing with it is worth
// more here than a second, differently-guessed answer.
func locate(ref coderef.Ref, rg *coderef.Registry, src symbolFinder) (repo, p string, ok bool) {
	if ref.Kind == coderef.KindSymbol {
		repo, p, _, ok := src.Find(ref.Symbol, "")
		return repo, p, ok
	}
	res := rg.Resolve(ref)
	return res.Ref.Repo, res.RepoPath, res.RepoPath != ""
}

// groupsOf takes the directory as the module. It is a poor abstraction for a Java build
// layout and an honest one: nothing in the index knows about Maven modules or Go packages,
// and inventing a grouping the code does not declare would put files in modules their
// authors never wrote. Only the busiest maxGroups survive — a map of 400 directories is
// not a map.
func groupsOf(files []string, st *gitsig.Stats, cited map[string][]string) []report.CodeGroup {
	byDir := map[string][]string{}
	for _, f := range files {
		byDir[path.Dir(f)] = append(byDir[path.Dir(f)], f)
	}
	out := make([]report.CodeGroup, 0, len(byDir))
	for dir, fs := range byDir {
		out = append(out, groupOf(dir, fs, st, cited))
	}
	sortGroups(out)
	if len(out) > maxGroups {
		out = out[:maxGroups]
	}
	return out
}

func groupOf(dir string, files []string, st *gitsig.Stats, cited map[string][]string) report.CodeGroup {
	g := report.CodeGroup{Name: dir, Files: len(files)}
	owners := map[string]int{}
	slugs := map[string]bool{}
	for _, f := range files {
		g.Commits += st.Churn[f]
		if o, _ := st.Owner(f); o != "" {
			owners[o] += st.Churn[f]
		}
		for _, s := range cited[f] {
			slugs[s] = true
		}
	}
	g.Owners, g.Notes = topKeys(owners), sortedSet(slugs)
	return g
}

func sortGroups(gs []report.CodeGroup) {
	sort.Slice(gs, func(i, j int) bool {
		if gs[i].Commits != gs[j].Commits {
			return gs[i].Commits > gs[j].Commits
		}
		return gs[i].Name < gs[j].Name
	})
}

// uncoveredOf is the last line of section B.5: high churn, real size, zero notes.
//
// Coverage is decided per file, not per symbol. A note that cites a class covers the
// methods inside it in every sense a reader cares about, and demanding a citation per
// symbol would report a well-documented file as forty gaps.
func uncoveredOf(ix codeindex.Index, st *gitsig.Stats, cited map[string][]string) []report.Uncovered {
	var out []report.Uncovered
	for p, f := range ix.Files {
		if len(cited[p]) > 0 || st.Churn[p] < minSymbolCommits {
			continue
		}
		for _, s := range f.Symbols {
			if loc := s.End - s.Start + 1; loc >= minSymbolLOC {
				out = append(out, report.Uncovered{Symbol: s.Name, Path: p, LOC: loc,
					Commits: st.Churn[p]})
			}
		}
	}
	return out
}

func topKeys(counts map[string]int) []string {
	out := make([]string, 0, len(counts))
	for k := range counts {
		out = append(out, k)
	}
	sort.Slice(out, func(i, j int) bool {
		if counts[out[i]] != counts[out[j]] {
			return counts[out[i]] > counts[out[j]]
		}
		return out[i] < out[j]
	})
	return out
}

func sortedSet(set map[string]bool) []string {
	out := make([]string, 0, len(set))
	for k := range set {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
