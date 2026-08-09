// Package store is the derived SQLite cache. Markdown is the only source of truth;
// nothing here is authoritative and `forge reindex` rebuilds all of it from the vault.
// Pure Go driver (modernc.org/sqlite) so the binary stays static under CGO_ENABLED=0.
//
// One exception: budget.go's table. AUDIT §8.4 D-8 makes per-day USD spend the one
// value SQLite holds that markdown does not — Reset() never lists it, on purpose.
package store

import (
	"database/sql"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

// Store wraps the cache database.
type Store struct{ DB *sql.DB }

const schemaSQL = `
CREATE TABLE IF NOT EXISTS notes (
  rel            TEXT PRIMARY KEY,
  slug           TEXT,
  title          TEXT,
  type           TEXT,
  confidence     TEXT,
  updated        TEXT,
  verified       TEXT,
  freshness_days INTEGER,
  mtime          INTEGER NOT NULL,
  size           INTEGER NOT NULL,
  valid          INTEGER NOT NULL DEFAULT 0
);
CREATE TABLE IF NOT EXISTS note_stack (rel TEXT, value TEXT);
CREATE TABLE IF NOT EXISTS note_tags  (rel TEXT, value TEXT);
CREATE TABLE IF NOT EXISTS links      (src TEXT, target TEXT, resolved TEXT);
CREATE INDEX IF NOT EXISTS idx_stack   ON note_stack(value);
CREATE INDEX IF NOT EXISTS idx_tags    ON note_tags(value);
CREATE INDEX IF NOT EXISTS idx_lnk_tgt ON links(resolved);
`

// Open creates or opens <vault>/.forge/cache/index.db.
func Open(vaultRoot string) (*Store, error) {
	dir := filepath.Join(vaultRoot, ".forge", "cache")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	db, err := sql.Open("sqlite", filepath.Join(dir, "index.db"))
	if err != nil {
		return nil, err
	}
	if _, err := db.Exec(schemaSQL); err != nil {
		db.Close()
		return nil, err
	}
	if err := ensureBudgetSchema(db); err != nil {
		db.Close()
		return nil, err
	}
	return &Store{DB: db}, nil
}

// Close releases the database handle.
func (s *Store) Close() error { return s.DB.Close() }

// Reset drops every derived row. This is what makes the cache disposable: `forge
// reindex` calls it and repopulates purely from markdown, so a corrupt or stale DB is
// never a data-loss event.
func (s *Store) Reset() error {
	for _, t := range []string{"notes", "note_stack", "note_tags", "links"} {
		if _, err := s.DB.Exec("DELETE FROM " + t); err != nil {
			return err
		}
	}
	return nil
}

// Row is one note as the cache holds it.
type Row struct {
	Rel, Slug, Title, Type, Confidence string
	Updated, Verified                  string
	FreshnessDays                      int
	MTime, Size                        int64
	Valid                              bool
	Stack, Tags                        []string
	Links                              []Link
}

// Link is one outbound wikilink. Resolved is "" when the target does not exist.
type Link struct{ Target, Resolved string }

const upsertSQL = `INSERT INTO notes
 (rel,slug,title,type,confidence,updated,verified,freshness_days,mtime,size,valid)
 VALUES (?,?,?,?,?,?,?,?,?,?,?)
 ON CONFLICT(rel) DO UPDATE SET
   slug=excluded.slug, title=excluded.title, type=excluded.type,
   confidence=excluded.confidence, updated=excluded.updated,
   verified=excluded.verified, freshness_days=excluded.freshness_days,
   mtime=excluded.mtime, size=excluded.size, valid=excluded.valid`

// Put writes one note's derived rows inside the caller's transaction.
func Put(tx *sql.Tx, r Row) error {
	_, err := tx.Exec(upsertSQL, r.Rel, r.Slug, r.Title, r.Type, r.Confidence,
		r.Updated, r.Verified, r.FreshnessDays, r.MTime, r.Size, r.Valid)
	if err != nil {
		return err
	}
	return putChildren(tx, r)
}

func putChildren(tx *sql.Tx, r Row) error {
	for _, q := range []string{"note_stack", "note_tags", "links"} {
		if _, err := tx.Exec("DELETE FROM "+q+" WHERE "+srcCol(q)+"=?", r.Rel); err != nil {
			return err
		}
	}
	if err := insertValues(tx, "note_stack", r.Rel, r.Stack); err != nil {
		return err
	}
	if err := insertValues(tx, "note_tags", r.Rel, r.Tags); err != nil {
		return err
	}
	return insertLinks(tx, r)
}

func srcCol(table string) string {
	if table == "links" {
		return "src"
	}
	return "rel"
}

func insertValues(tx *sql.Tx, table, rel string, vals []string) error {
	for _, v := range vals {
		if _, err := tx.Exec("INSERT INTO "+table+" (rel,value) VALUES (?,?)", rel, v); err != nil {
			return err
		}
	}
	return nil
}

func insertLinks(tx *sql.Tx, r Row) error {
	for _, l := range r.Links {
		_, err := tx.Exec("INSERT INTO links (src,target,resolved) VALUES (?,?,?)",
			r.Rel, l.Target, l.Resolved)
		if err != nil {
			return err
		}
	}
	return nil
}

// Fresh reports whether the cached row for rel already matches the file on disk.
// This is the mtime cache the latency budget depends on.
func (s *Store) Fresh(rel string, mtime, size int64) bool {
	var m, sz int64
	err := s.DB.QueryRow("SELECT mtime,size FROM notes WHERE rel=?", rel).Scan(&m, &sz)
	return err == nil && m == mtime && sz == size
}

// Prune deletes cached notes that no longer exist in the vault.
func (s *Store) Prune(live map[string]bool) error {
	rows, err := s.DB.Query("SELECT rel FROM notes")
	if err != nil {
		return err
	}
	var dead []string
	for rows.Next() {
		var rel string
		if err := rows.Scan(&rel); err == nil && !live[rel] {
			dead = append(dead, rel)
		}
	}
	rows.Close()
	return s.deleteAll(dead)
}

func (s *Store) deleteAll(rels []string) error {
	for _, rel := range rels {
		for _, t := range []string{"notes", "note_stack", "note_tags", "links"} {
			if _, err := s.DB.Exec("DELETE FROM "+t+" WHERE "+srcCol(t)+"=?", rel); err != nil {
				return err
			}
		}
	}
	return nil
}
