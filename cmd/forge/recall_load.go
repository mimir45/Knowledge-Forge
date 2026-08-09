package main

import (
	"os"
	"path/filepath"

	"knowledge-forge/pkg/recall"
	"knowledge-forge/pkg/store"
	"knowledge-forge/pkg/vault"
)

// loadDocs builds recall's view of the vault, preferring the SQLite cache. A row is
// reused when store.Fresh holds for (mtime, size); otherwise the markdown is re-parsed
// and the row rewritten, so `forge recall` self-heals a cold or partial cache instead
// of requiring `forge index` to have run first.
func loadDocs(root string) ([]recall.Doc, error) {
	st, err := store.Open(root)
	if err != nil {
		return nil, err
	}
	defer st.Close()
	cached, err := st.Rows()
	if err != nil {
		return nil, err
	}
	rels, err := vault.Walk(root)
	if err != nil {
		return nil, err
	}
	return collect(root, rels, cached, st)
}

func collect(root string, rels []string, cached map[string]store.Row, st *store.Store) ([]recall.Doc, error) {
	schema, err := vault.LoadSchema()
	if err != nil {
		return nil, err
	}
	// The link index is built from the path list alone, and it must be built: rowOf
	// recomputes a note's links, and store.Put replaces them. Writing a row back with an
	// empty index would mark every wikilink in it unresolved and corrupt the graph that
	// `forge index` populated.
	ix := vault.NewIndex(rels)
	var docs []recall.Doc
	var stale []store.Row
	for _, rel := range rels {
		if !vault.IsContentNote(rel) {
			continue
		}
		row, fresh := rowFor(root, rel, cached, st, ix)
		if row.Rel == "" {
			continue
		}
		if !fresh {
			stale = append(stale, row)
		}
		docs = append(docs, docOf(root, row, schema))
	}
	return docs, refresh(st, stale)
}

// rowFor returns the cached row when the file on disk still matches it, and otherwise
// parses the markdown. The second return reports which happened, so the caller can write
// back only what changed — rewriting an unchanged row would churn the cache for nothing.
func rowFor(root, rel string, cached map[string]store.Row, st *store.Store, ix *vault.Index) (store.Row, bool) {
	fi, err := os.Stat(filepath.Join(root, rel))
	if err != nil {
		return store.Row{}, true
	}
	if row, ok := cached[rel]; ok && st.Fresh(rel, fi.ModTime().Unix(), fi.Size()) {
		return row, true
	}
	n, err := vault.Load(filepath.Join(root, rel), rel)
	if err != nil {
		return store.Row{}, true
	}
	return rowOf(n, ix), false
}

// docOf resolves the freshness window: the note's own freshness_days when it carries
// one, else the type default from references/schema.yaml. Both can legitimately be 0,
// which means never stale (DESIGN §10 — decisions get superseded, not expired).
func docOf(root string, r store.Row, s *vault.Schema) recall.Doc {
	days := r.FreshnessDays
	if days == 0 {
		days = s.FreshnessDefault(r.Type)
	}
	abs := filepath.Join(root, r.Rel)
	return recall.Doc{
		Rel: r.Rel, Slug: r.Slug, Title: r.Title,
		Tags: r.Tags, Stack: r.Stack,
		Updated: r.Updated, Verified: r.Verified, FreshnessDays: days,
		LoadBody: func() []byte { b, _ := os.ReadFile(abs); return b },
	}
}

// refresh writes back the rows that were re-parsed. A failure here is not fatal: the
// cache is derived, and a run that scored correctly off fresh markdown is still correct.
func refresh(st *store.Store, rows []store.Row) error {
	if len(rows) == 0 {
		return nil
	}
	tx, err := st.DB.Begin()
	if err != nil {
		return nil
	}
	for _, r := range rows {
		if err := store.Put(tx, r); err != nil {
			tx.Rollback()
			return nil
		}
	}
	tx.Commit()
	return nil
}
