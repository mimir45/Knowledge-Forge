package main

import (
	"fmt"
	"maps"
	"path"
	"path/filepath"
	"slices"
	"sort"

	"github.com/mimir45/Knowledge-Forge/pkg/codeindex"
	"github.com/mimir45/Knowledge-Forge/pkg/coderef"
	"github.com/mimir45/Knowledge-Forge/pkg/drift"
	"github.com/mimir45/Knowledge-Forge/pkg/gitsig"
	"github.com/mimir45/Knowledge-Forge/pkg/report"
)

// These make the original spec's "high churn, real size" concrete.
const (
	minSymbolCommits = 2
	minSymbolLOC     = 15
	maxGroups        = 20
)

func (d *checkData) driftAndCode() {
	if len(d.cfg.repos) == 0 {
		return
	}
	scans, err := scanRepos(d.cfg.repos)
	if err != nil {
		d.repoErr = err
		return
	}
	rg := newRegistryFrom(d.cfg.repos, scans)
	src := drift.NewGitSource(d.cfg.repos, filepath.Join(d.root, ".forge"))
	d.findings = drift.Check(driftNotes(d.notes), rg, src, nil, drift.Opts{Deep: true})
	d.code, d.codeErr = d.codebases(rg, src, scans)
}

// codebases builds one section per repository.
func (d *checkData) codebases(rg *coderef.Registry, src symbolFinder,
	scans map[string]coderef.Repo) ([]report.CodebaseInput, error) {
	if !codeindex.Available() {
		return nil, codeindex.ErrUnavailable
	}
	cited := d.citedPaths(rg, src)
	out := make([]report.CodebaseInput, 0, len(d.cfg.repos))
	for _, r := range d.cfg.repos {
		in, err := d.oneCodebase(r.Name, r.Root, cited[r.Name], scans[r.Name].Files, src)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", r.Name, err)
		}
		out = append(out, in)
	}
	return out, nil
}

// oneCodebase reads the symbol table through src rather than calling codeindex.Build
// itself: driftAndCode's GitSource already built.
func (d *checkData) oneCodebase(name, root string, cited map[string][]string,
	files []string, src symbolFinder) (report.CodebaseInput, error) {
	commits, err := gitsig.Log(root, d.now.AddDate(0, 0, -d.cfg.days))
	if err != nil {
		return report.CodebaseInput{}, err
	}
	st := gitsig.Analyze(commits)
	ix := src.Index(name, "HEAD")
	return report.CodebaseInput{Repo: name, Days: d.cfg.days, Now: d.now,
		Groups:    groupsOf(files, st, cited, dependsOn(ix, files)),
		Uncovered: uncoveredOf(ix, st, cited)}, nil
}

// symbolFinder is the single question coverage asks of drift's symbol table.
type symbolFinder interface {
	Find(name, asOf string) (repo, path string, sym codeindex.Symbol, ok bool)
	Index(repo, rev string) codeindex.Index
}

// citedPaths resolves every note's code citations to repo-relative paths.
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

// locate answers which file a citation is about.
func locate(ref coderef.Ref, rg *coderef.Registry, src symbolFinder) (repo, p string, ok bool) {
	if ref.Kind == coderef.KindSymbol {
		repo, p, _, ok := src.Find(ref.Symbol, "")
		return repo, p, ok
	}
	res := rg.Resolve(ref)
	return res.Ref.Repo, res.RepoPath, res.RepoPath != ""
}

// groupsOf takes the directory as the module.
func groupsOf(files []string, st *gitsig.Stats, cited map[string][]string,
	deps map[string][]string) []report.CodeGroup {

	byDir := map[string][]string{}
	for _, f := range files {
		byDir[path.Dir(f)] = append(byDir[path.Dir(f)], f)
	}
	out := make([]report.CodeGroup, 0, len(byDir))
	for dir, fs := range byDir {
		out = append(out, groupOf(dir, fs, st, cited, deps[dir]))
	}
	sortGroups(out)
	if len(out) > maxGroups {
		out = out[:maxGroups]
	}
	return out
}

func groupOf(dir string, files []string, st *gitsig.Stats, cited map[string][]string,
	dependsOn []string) report.CodeGroup {

	g := report.CodeGroup{Name: dir, Files: len(files), DependsOn: dependsOn}
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
	g.Owners, g.Notes = topKeys(owners), slices.Sorted(maps.Keys(slugs))
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
