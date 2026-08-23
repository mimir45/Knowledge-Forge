package drift

import (
	"path/filepath"
	"sort"
	"strings"

	"knowledge-forge/pkg/codeindex"
	"knowledge-forge/pkg/coderef"
)

// build prefers the persisted index and patches it forward, so the only tree-sitter work
// on the hook path is the handful of files the commit touched. A cache one commit behind
// HEAD is the *normal* case — the hook fires post-commit — and patching is what makes
// "the symbol table at HEAD" true rather than approximately true.
func (g *GitSource) build(repo, rev string) codeindex.Index {
	root := g.root(repo)
	if ix, ok := codeindex.Load(g.cachePath(repo)); ok && ix.Commit != "" {
		if changed, err := coderef.ChangedFiles(root, ix.Commit, rev); err == nil {
			out, _ := codeindex.Patch(ix, root, rev, changed)
			g.persist(repo, rev, out)
			return out
		}
	}
	return g.full(repo, root, rev)
}

func (g *GitSource) full(repo, root, rev string) codeindex.Index {
	empty := codeindex.Index{Repo: repo, Commit: rev, Files: map[string]codeindex.File{}}
	r, err := coderef.ScanRepo(repo, root, rev)
	if err != nil {
		return empty
	}
	ix, err := codeindex.Build(repo, root, rev, r.Files)
	if err != nil {
		return ix // ErrUnavailable on a no-cgo build: no symbols, so every citation SKIPs
	}
	g.persist(repo, rev, ix)
	return ix
}

// persist caches HEAD only. A historical revision is resolved once per run by forge check
// and never asked for again; caching it would trade disk for nothing.
//
// The file is .forge/code-index-<repo>.json, one per configured --repo name=path, not the
// singular .forge/code-index.json ADDENDUM §B.6 and DESIGN §15 describe. The suffix is
// required, not cosmetic: forge drift/check/logback all take --repo repeatably, so a
// single shared name would let the second repo's index overwrite the first's on the very
// next run. See BACKLOG B-027 — the docs are the stale side, deliberately unedited.
func (g *GitSource) persist(repo, rev string, ix codeindex.Index) {
	if g.cache == "" || rev != g.Head(repo) || len(ix.Files) == 0 {
		return
	}
	codeindex.Save(g.cachePath(repo), ix) //nolint:errcheck // the cache is derived; a failed write costs the next run a rebuild and nothing else
}

func (g *GitSource) cachePath(repo string) string {
	if g.cache == "" {
		return ""
	}
	return filepath.Join(g.cache, "code-index-"+repo+".json")
}

// nameMap answers a symbol-only citation. full holds declarations under their qualified
// name, short under the bare member name; a lookup tries full first, because `render`
// matching every component in the tree is not evidence about `AccountsLoader.render`.
type nameMap struct {
	full  map[string][]loc
	short map[string][]loc
}

type loc struct {
	repo, path string
	sym        codeindex.Symbol
}

// Find resolves a name across every registered repository. Where two repositories declare
// the same name the first in (repo, path) order wins — deterministic, which is what
// verdict purity requires, though only the "is this still declared anywhere" half of the
// answer is trustworthy in that case.
func (g *GitSource) Find(name, asOf string) (string, string, codeindex.Symbol, bool) {
	nm := g.nameIndex(asOf)
	hits := nm.full[name]
	if len(hits) == 0 {
		hits = nm.short[name]
	}
	if len(hits) == 0 {
		return "", "", codeindex.Symbol{}, false
	}
	return hits[0].repo, hits[0].path, hits[0].sym, true
}

// nameIndex flattens every repository's symbol table into one map so a symbol-only
// citation costs a lookup rather than a scan. Memoised per asOf: forge check resolves a
// different revision for each distinct verified date, and rebuilding an index per note
// would put the weekly pass in minutes rather than seconds.
func (g *GitSource) nameIndex(asOf string) *nameMap {
	if m, ok := g.names[asOf]; ok {
		return m
	}
	m := &nameMap{full: map[string][]loc{}, short: map[string][]loc{}}
	for _, r := range g.repos {
		g.collect(m, r.Name, g.revFor(r.Name, asOf))
	}
	m.sort()
	g.names[asOf] = m
	return m
}

func (g *GitSource) collect(m *nameMap, repo, rev string) {
	if rev == "" {
		return
	}
	for p, f := range g.indexAt(repo, rev).Files {
		for _, s := range f.Symbols {
			m.full[s.Name] = append(m.full[s.Name], loc{repo, p, s})
			if i := strings.LastIndex(s.Name, "."); i > 0 {
				m.short[s.Name[i+1:]] = append(m.short[s.Name[i+1:]], loc{repo, p, s})
			}
		}
	}
}

// ResolveAt answers a path-shaped citation against one repository's file list as it stood
// at asOf. It goes through coderef.ScanRepo, not codeindex.Build, so — unlike Find and At
// — it answers identically on both the cgo and the pure-Go build lane.
func (g *GitSource) ResolveAt(ref coderef.Ref, asOf string) coderef.Resolution {
	return g.registryAt(asOf).Resolve(ref)
}

// registryAt is memoised per asOf the same way nameIndex memoises per asOf: forge check
// resolves one revision per distinct verified date across the whole vault, not once per
// citation, so this costs one git ls-tree per distinct (repo, date) pair, not per note.
func (g *GitSource) registryAt(asOf string) *coderef.Registry {
	if rg, ok := g.registries[asOf]; ok {
		return rg
	}
	out := make([]coderef.Repo, 0, len(g.repos))
	for _, r := range g.repos {
		if rev := g.revFor(r.Name, asOf); rev != "" {
			if scanned, err := coderef.ScanRepo(r.Name, r.Root, rev); err == nil {
				out = append(out, scanned)
			}
		}
	}
	rg := coderef.NewRegistry(out)
	g.registries[asOf] = rg
	return rg
}

func (g *GitSource) revFor(repo, asOf string) string {
	if asOf == "" {
		return g.Head(repo)
	}
	return g.RevBefore(repo, asOf)
}

// sort makes the winner of an ambiguous name a property of the tree rather than of Go's
// map iteration order.
func (m *nameMap) sort() {
	for _, hits := range []map[string][]loc{m.full, m.short} {
		for _, ls := range hits {
			sort.Slice(ls, func(i, j int) bool { return lessLoc(ls[i], ls[j]) })
		}
	}
}

// lessLoc has to order two declarations that share a file, not just two that share a name.
// One Java file routinely declares `Order.Builder` and `OrderItem.Builder`, and under the
// short name those tie on (repo, path); sort.Slice is not stable, so the tie was settled
// by whatever order the Files map happened to yield. That made drift.md oscillate between
// runs on an unchanged tree — a note came and went from the suspect list because `Builder`
// resolved to a different declaration each time. Declaration order inside the file is the
// tiebreak, with the name behind it, because the pair is unique in the tree.
func lessLoc(a, b loc) bool {
	switch {
	case a.repo != b.repo:
		return a.repo < b.repo
	case a.path != b.path:
		return a.path < b.path
	case a.sym.Start != b.sym.Start:
		return a.sym.Start < b.sym.Start
	}
	return a.sym.Name < b.sym.Name
}
