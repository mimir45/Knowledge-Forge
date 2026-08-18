package drift

import (
	"os/exec"
	"strings"

	"knowledge-forge/pkg/codeindex"
	"knowledge-forge/pkg/coderef"
)

// Repo names a code repository drift is allowed to look at.
type Repo struct{ Name, Root string }

// GitSource is the production Source: every answer it gives is read out of a git object
// store, so two runs at the same tree state give the same verdicts and a revert restores
// a demoted note without an undo log.
//
// It is memoised, not concurrent — Check drives it from a single goroutine.
type GitSource struct {
	repos []Repo
	cache string // directory for persisted indexes; "" disables the cache
	heads map[string]string
	revs  map[string]string          // repo@date -> sha
	idx   map[string]codeindex.Index // repo@rev
	names map[string]*nameMap        // asOf -> flattened symbol lookup

	registries map[string]*coderef.Registry // asOf -> path registry
}

func NewGitSource(repos []Repo, cacheDir string) *GitSource {
	return &GitSource{
		repos: repos, cache: cacheDir,
		heads: map[string]string{}, revs: map[string]string{},
		idx: map[string]codeindex.Index{}, names: map[string]*nameMap{},

		registries: map[string]*coderef.Registry{},
	}
}

func (g *GitSource) root(repo string) string {
	for _, r := range g.repos {
		if r.Name == repo {
			return r.Root
		}
	}
	return ""
}

func (g *GitSource) Head(repo string) string {
	if h, ok := g.heads[repo]; ok {
		return h
	}
	h, err := coderef.HeadSHA(g.root(repo), "HEAD")
	if err != nil {
		h = ""
	}
	g.heads[repo] = h
	return h
}

// RevBefore picks the commit the note was actually written against. The note records a
// date, not a sha, so "the tree as it stood when this was verified" is the last commit of
// that day — inclusive, because a note is written after the code it describes.
func (g *GitSource) RevBefore(repo, date string) string {
	key := repo + "@" + date
	if r, ok := g.revs[key]; ok {
		return r
	}
	out, err := exec.Command("git", "-C", g.root(repo), "rev-list", "-1",
		"--before="+date+"T23:59:59", g.Head(repo)).Output()
	rev := ""
	if err == nil {
		rev = strings.TrimSpace(string(out))
	}
	g.revs[key] = rev
	return rev
}

func (g *GitSource) At(repo, path, rev string) (codeindex.File, bool) {
	f, ok := g.indexAt(repo, rev).Files[path]
	return f, ok
}

// Index exposes the same cache-preferring, patch-forward index indexAt already builds for
// At and Find, so a caller that also wants the full symbol table (coverage reporting, for
// one) doesn't pay for a second full tree-sitter parse to get what this package already
// has in hand.
func (g *GitSource) Index(repo, rev string) codeindex.Index {
	return g.indexAt(repo, rev)
}

func (g *GitSource) indexAt(repo, rev string) codeindex.Index {
	key := repo + "@" + rev
	if ix, ok := g.idx[key]; ok {
		return ix
	}
	ix := g.build(repo, rev)
	g.idx[key] = ix
	return ix
}
