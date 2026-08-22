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
	refresh(st, stale)
	return docs, nil
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

// refresh writes back the rows that were re-parsed, and deliberately reports nothing.
//
// A failure here is not fatal: the cache is derived, and a run that scored correctly off
// fresh markdown is still correct. rowFor re-parses whatever the cache does not match, so
// a write that never lands costs a re-parse on the next run and nothing else.
//
// BACKLOG B-029 filed this as a correctness bug and prescribed propagating the error
// through the `commit` helper. Tracing the callers says otherwise: loadDocs' error reaches
// runRecall (recall.go:77), which prints it and returns 1 **without emitting the
// candidates it has already scored correctly**. A concurrent `forge intent` on the
// UserPromptSubmit hook holding the SQLite write lock is enough to produce that, so
// propagating would trade a stale cache for a discarded correct answer. The entry was
// sized from errcheck output rather than from that call trace.
//
// What was actually wrong is the signature. Every path returned nil, so
// `return docs, refresh(...)` read as propagation while guaranteeing the opposite — and a
// later edit that started returning a real error would have turned a transient lock into
// an exit 1 with no reviewer noticing. The return is gone, so the promise matches the
// behaviour, and writeRows checks the three errors the old body dropped.
func refresh(st *store.Store, rows []store.Row) {
	// The common run has every row fresh from the cache, and it must cost no transaction.
	if len(rows) == 0 {
		return
	}
	_ = writeRows(st, rows) // the one deliberate swallow; see above for why it is not fatal
}

// writeRows is refresh's checked body, split out so that ignoring the failure happens at
// exactly one site instead of at three. It also closes the gap errcheck could not see: the
// old body dropped st.DB.Begin()'s error on an assignment errcheck does not flag, which is
// why the entry's finding count understated this function. commit is index.go's helper,
// reused for the same reason persist uses it rather than a bare tx.Commit().
func writeRows(st *store.Store, rows []store.Row) error {
	tx, err := st.DB.Begin()
	if err != nil {
		return err
	}
	for _, r := range rows {
		if err := store.Put(tx, r); err != nil {
			tx.Rollback()
			return err
		}
	}
	return commit(tx)
}
