package dataset

import (
	"fmt"
	"path/filepath"
	"sort"

	"github.com/mimir45/Knowledge-Forge/pkg/codeindex"
	"github.com/mimir45/Knowledge-Forge/pkg/coderef"
	"github.com/mimir45/Knowledge-Forge/pkg/vault"
)

// D6 is the sixth dataset: (repo symbol or module → the note explaining it).
const (
	D6Kind = "d6-code-knowledge"
	D6Tag  = "d6"
)

// D6Pair is one JSONL record.
type D6Pair struct {
	Kind   string `json:"kind"`
	Repo   string `json:"repo"`
	Path   string `json:"path,omitempty"`
	Symbol string `json:"symbol,omitempty"`
	Note   string `json:"note"` // vault-relative path of the note citing it
}

// loadD6 is loadTier's D6 case.
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

// loadIndexes reads every .forge/code-index-<repo>.json this machine currently holds.
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

// fileList sorts an Index's file set.
func fileList(ix codeindex.Index) []string {
	out := make([]string, 0, len(ix.Files))
	for p := range ix.Files {
		out = append(out, p)
	}
	sort.Strings(out)
	return out
}

// loadContentNotes is loadNotes (cmd/forge/index.go) inlined for pkg/dataset.
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

// pairsFromNotes resolves every note's citations and dedupes on the resulting
// quadruple.
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

// resolveD6 mirrors locate() (cmd/forge/check_codebase.go) exactly rather than
// inventing a more lenient strategy.
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

// findSymbol searches every loaded index for name.
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
