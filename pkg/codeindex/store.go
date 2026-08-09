package codeindex

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Save writes the index to .forge/code-index.json. Like every other file under .forge/
// it is a derived cache: deleting it costs a rebuild and nothing else.
func Save(path string, ix Index) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	b, err := json.Marshal(ix)
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, b, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path) // atomic: a hook killed mid-write leaves the old index intact
}

// Load returns a persisted index, or ok=false for anything unreadable. A corrupt cache is
// not an error worth reporting to a git hook — it is a rebuild.
func Load(path string) (Index, bool) {
	b, err := os.ReadFile(path)
	if err != nil {
		return Index{}, false
	}
	var ix Index
	if json.Unmarshal(b, &ix) != nil || ix.Files == nil || ix.Extractor != Extractor {
		return Index{}, false // a cache from an older extractor is missing symbols, not stale by a commit
	}
	return ix, true
}

// Patch brings a persisted index forward to a newer revision by re-parsing only the files
// git reports as changed. This is the hook path's entire performance story: the brief's
// cheap gate is `git diff --name-only`, and honouring it here is what keeps `forge drift`
// inside 100ms on a repository whose full index takes seconds to build.
//
// It stays a pure function of tree state because every re-parsed blob is read at rev — the
// patched index is byte-identical to a full build at that revision.
func Patch(ix Index, root, rev string, changed []string) (Index, error) {
	out := Index{Repo: ix.Repo, Commit: rev, Extractor: Extractor,
		Files: map[string]File{}, Deps: Deps(root)}
	for p, f := range ix.Files {
		out.Files[p] = f
	}
	for _, p := range changed {
		delete(out.Files, p) // deleted files stay deleted; survivors are re-added below
	}
	if len(changed) == 0 {
		return out, nil
	}
	fresh, err := Build(ix.Repo, root, rev, changed)
	for p, f := range fresh.Files {
		out.Files[p] = f
	}
	return out, err
}
