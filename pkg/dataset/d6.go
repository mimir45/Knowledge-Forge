package dataset

import (
	"fmt"
	"path/filepath"
	"sort"

	"github.com/mimir45/Knowledge-Forge/pkg/codeindex"
	"github.com/mimir45/Knowledge-Forge/pkg/coderef"
	"github.com/mimir45/Knowledge-Forge/pkg/vault"
)

// D6 is ADDENDUM §D.1's sixth dataset: (repo symbol or module → the note explaining it).
// It is a derivation rather
// than a capture tier — D1-D5 each have a write path and accumulate forward; D6 has
// none and needs none, because `forge logback` already resolves the same
// (symbol → note) mapping every time it runs. There is deliberately no AppendD6: adding
// a sixth capture path would defeat the point of deriving it instead.
//
// Volume is "= note count" per ADDENDUM, but that describes distinct notes, not distinct
// pairs — a note citing five symbols yields five records, one per citation that
// resolves. Every export's datasheet says so.
const (
	D6Kind = "d6-code-knowledge"
	D6Tag  = "d6"
)

// D6Pair is one JSONL record. Unlike D1-D5, anonymization is refused outright for this
// tier (see refuseDerivedOptions) rather than attempted: Repo, Path and Symbol are the
// entire feature D6 exists to carry, and they are also the most employer-identifying
// strings in the system. Every export is raw text.
type D6Pair struct {
	Kind   string `json:"kind"`
	Repo   string `json:"repo"`
	Path   string `json:"path,omitempty"`
	Symbol string `json:"symbol,omitempty"`
	Note   string `json:"note"` // vault-relative path of the note citing it
}

// loadD6 is loadTier's D6 case. It fails closed on a cache Load can't read (stale
// Extractor, corrupt JSON) rather than silently deriving a smaller corpus from whatever
// did load — the same reasoning read.go already applied to a truncated capture line: a
// cache nobody could parse is a cache nobody could vouch for, so the whole export
// refuses rather than reporting success over an incomplete symbol table.
func loadD6(root string) ([]any, error) {
	ixs, err := loadIndexes(root)
	if err != nil {
		return nil, err
	}
	notes, err := loadContentNotes(root)
	if err != nil {
		return nil, err
	}
	rg := coderef.NewRegistry(reposOf(ixs))
	return boxed(pairsFromNotes(notes, rg, ixs), nil)
}

// loadIndexes reads every .forge/code-index-<repo>.json this machine currently holds, in
// sorted (repo name) order — the same determinism discipline this codebase applies
// elsewhere, so a symbol declared in two repos resolves to the same one on every run. An
// absent cache is a repo nobody has run `forge logback`/`check`/`drift` against yet and
// is not an error; a present-but-unreadable one is.
func loadIndexes(root string) ([]codeindex.Index, error) {
	paths, err := filepath.Glob(filepath.Join(root, ".forge", "code-index-*.json"))
	if err != nil {
		return nil, err
	}
	sort.Strings(paths)
	out := make([]codeindex.Index, 0, len(paths))
	for _, p := range paths {
		ix, ok := codeindex.Load(p)
		if !ok {
			return nil, fmt.Errorf("dataset: d6: %s: unreadable, or cached by an older "+
				"Extractor; run forge logback (or forge check/drift) against this repo to "+
				"refresh it before exporting d6", p)
		}
		out = append(out, ix)
	}
	return out, nil
}

func reposOf(ixs []codeindex.Index) []coderef.Repo {
	out := make([]coderef.Repo, len(ixs))
	for i, ix := range ixs {
		out[i] = coderef.Repo{Name: ix.Repo, Files: fileList(ix)}
	}
	return out
}

// fileList sorts an Index's file set. Everything downstream that walks a repo's files —
// this, and findSymbol below — reads them in this order, so a tie (two files declaring
// the same trailing member) resolves the same way on every run rather than however Go's
// map iteration happens to land.
func fileList(ix codeindex.Index) []string {
	out := make([]string, 0, len(ix.Files))
	for p := range ix.Files {
		out = append(out, p)
	}
	sort.Strings(out)
	return out
}

// loadContentNotes is loadNotes (cmd/forge/index.go) inlined for pkg/dataset: cmd depends
// on pkg, never the reverse, so the walk-and-load pair is duplicated here rather than
// shared.
func loadContentNotes(root string) ([]*vault.Note, error) {
	rels, err := vault.Walk(root)
	if err != nil {
		return nil, err
	}
	var out []*vault.Note
	for _, rel := range rels {
		if !vault.IsContentNote(rel) {
			continue
		}
		if n, err := vault.Load(filepath.Join(root, rel), rel); err == nil {
			out = append(out, n)
		}
	}
	return out, nil
}

// pairsFromNotes resolves every note's citations and dedupes on the resulting quadruple —
// a note citing the same symbol once in frontmatter and once in prose must not double its
// own weight in the corpus.
func pairsFromNotes(notes []*vault.Note, rg *coderef.Registry, ixs []codeindex.Index) []D6Pair {
	seen := map[string]bool{}
	var out []D6Pair
	for _, n := range notes {
		for _, ref := range refsOf(n) {
			p, ok := resolveD6(ref, n.Rel, rg, ixs)
			if !ok {
				continue
			}
			key := p.Repo + "|" + p.Path + "|" + p.Symbol + "|" + p.Note
			if !seen[key] {
				seen[key] = true
				out = append(out, p)
			}
		}
	}
	return out
}

// refsOf is driftNotes' (cmd/forge/drift.go) per-note half: both citation shapes, body
// prose first, frontmatter's canonical form taking precedence by being appended in front.
func refsOf(n *vault.Note) []coderef.Ref {
	refs := coderef.FromBody(n.Rel, n.Body)
	if n.FM != nil {
		refs = append(coderef.FromFrontmatter(n.Rel, n.FM.List("code_refs")), refs...)
	}
	return refs
}

// resolveD6 mirrors locate() (cmd/forge/check_codebase.go) exactly rather than inventing
// a more lenient strategy: a symbol-kind ref resolves through the symbol table, anything
// else through the registry, and a citation resolving to neither contributes no pair —
// same as it contributes nothing to moc/codebase.md's coverage today.
func resolveD6(ref coderef.Ref, note string, rg *coderef.Registry, ixs []codeindex.Index) (D6Pair, bool) {
	var repo, path string
	if ref.Kind == coderef.KindSymbol {
		repo, path, _ = findSymbol(ixs, ref.Symbol)
	} else if res := rg.Resolve(ref); res.RepoPath != "" {
		repo, path = res.Ref.Repo, res.RepoPath
	}
	if repo == "" {
		return D6Pair{}, false
	}
	return D6Pair{Kind: D6Kind, Repo: repo, Path: path, Symbol: ref.Symbol, Note: note}, true
}

// findSymbol searches every loaded index for name, repos and files both in sorted order,
// via File.Lookup — exact qualified name, then trailing member, the same rule
// codeindex.File.Lookup gives forge check's own coverage numbers. Deliberately not
// Index.FindSymbol: that method ranges over Index.Files, a map, so on a repo where two
// files both declare the same trailing member it would pick whichever Go's map iteration
// happened to visit first — a different file on different runs.
func findSymbol(ixs []codeindex.Index, name string) (repo, path string, ok bool) {
	for _, ix := range ixs {
		for _, p := range fileList(ix) {
			if _, found := ix.Files[p].Lookup(name); found {
				return ix.Repo, p, true
			}
		}
	}
	return "", "", false
}
