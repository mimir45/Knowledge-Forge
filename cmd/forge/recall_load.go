package main

import (
	"os"
	"path/filepath"

	"github.com/mimir45/Knowledge-Forge/pkg/recall"
	"github.com/mimir45/Knowledge-Forge/pkg/store"
	"github.com/mimir45/Knowledge-Forge/pkg/vault"
)

// loadDocs builds recall's view of the vault, preferring the SQLite cache.
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
	refresh(st, stale)
	return docs, nil
}

// rowFor returns the cached row when the file on disk still matches it.
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

// docOf resolves the freshness window.
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

// refresh writes back the rows that were re-parsed, and deliberately reports nothing.
func refresh(st *store.Store, rows []store.Row) {
	// The common run has every row fresh from the cache, and it must cost no transaction.
	if len(rows) == 0 {
		return
	}
	_ = writeRows(st, rows) // the one deliberate swallow; see above for why it is not fatal
}

// writeRows is refresh's checked body.
func writeRows(st *store.Store, rows []store.Row) error {
	tx, err := st.DB.Begin()
	if err != nil {
		return err
	}
	for _, r := range rows {
		if err := store.Put(tx, r); err != nil {
			tx.Rollback() //nolint:errcheck // the Put error is the one worth returning; index.go:persist ignores it for the same reason
			return err
		}
	}
	return commit(tx)
}
