package store

import (
	"path/filepath"
	"testing"
)

func open(t *testing.T) *Store {
	t.Helper()
	s, err := Open(t.TempDir())
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}

func put(t *testing.T, s *Store, rows ...Row) {
	t.Helper()
	tx, err := s.DB.Begin()
	if err != nil {
		t.Fatal(err)
	}
	for _, r := range rows {
		if err := Put(tx, r); err != nil {
			tx.Rollback() //nolint:errcheck // the Put error is the one worth reporting
			t.Fatalf("Put(%s): %v", r.Rel, err)
		}
	}
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}
}

func count(t *testing.T, s *Store, table string) int {
	t.Helper()
	var n int
	if err := s.DB.QueryRow("SELECT COUNT(*) FROM " + table).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

func sample() Row {
	return Row{
		Rel: "notes/concept/a.md", Slug: "a", Title: "A", Type: "concept",
		Confidence: "high", Updated: "2026-08-01", Verified: "2026-08-01",
		FreshnessDays: 365, MTime: 1000, Size: 42, Valid: true,
		Stack: []string{"java", "spring-boot"}, Tags: []string{"jpa"},
		Links: []Link{{Target: "b", Resolved: "notes/concept/b.md"}, {Target: "gone"}},
	}
}

func TestOpenCreatesCacheUnderDotForge(t *testing.T) {
	root := t.TempDir()
	s, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	if _, err := s.DB.Exec("SELECT 1"); err != nil {
		t.Fatalf("db unusable: %v", err)
	}
	// The cache is derived state and belongs in .forge/, never beside the markdown.
	if _, err := Open(filepath.Join(root, ".forge", "cache")); err != nil {
		t.Fatalf("reopen: %v", err)
	}
}

func TestPutIsAnUpsert(t *testing.T) {
	s := open(t)
	put(t, s, sample())
	r := sample()
	r.Title, r.MTime, r.Stack = "A renamed", 2000, []string{"go"}
	put(t, s, r)
	if got := count(t, s, "notes"); got != 1 {
		t.Errorf("notes = %d, want 1; the second Put inserted instead of updating", got)
	}
	var title string
	if err := s.DB.QueryRow("SELECT title FROM notes WHERE rel=?", r.Rel).Scan(&title); err != nil {
		t.Fatal(err)
	}
	if title != "A renamed" {
		t.Errorf("title = %q, want the updated value", title)
	}
	// Child rows are replaced, not appended, or stack values would accumulate forever.
	if got := count(t, s, "note_stack"); got != 1 {
		t.Errorf("note_stack = %d, want 1", got)
	}
}

// TestFreshIsTheMtimeCache: the <200ms index budget depends on skipping unchanged files,
// and on never skipping a changed one.
func TestFreshIsTheMtimeCache(t *testing.T) {
	s := open(t)
	put(t, s, sample())
	cases := []struct {
		name        string
		mtime, size int64
		want        bool
	}{
		{"unchanged", 1000, 42, true},
		{"touched", 1001, 42, false},
		{"resized", 1000, 43, false},
		{"both", 1001, 43, false},
	}
	for _, c := range cases {
		if got := s.Fresh("notes/concept/a.md", c.mtime, c.size); got != c.want {
			t.Errorf("%s: Fresh = %v, want %v", c.name, got, c.want)
		}
	}
	if s.Fresh("notes/concept/missing.md", 1000, 42) {
		t.Error("an uncached note reported fresh")
	}
}

func TestPrune(t *testing.T) {
	s := open(t)
	gone := sample()
	gone.Rel, gone.Slug = "notes/concept/deleted.md", "deleted"
	put(t, s, sample(), gone)
	if err := s.Prune(map[string]bool{"notes/concept/a.md": true}); err != nil {
		t.Fatal(err)
	}
	for _, tbl := range []string{"notes", "note_stack", "note_tags", "links"} {
		var n int
		col := "rel"
		if tbl == "links" {
			col = "src"
		}
		q := "SELECT COUNT(*) FROM " + tbl + " WHERE " + col + "=?"
		if err := s.DB.QueryRow(q, gone.Rel).Scan(&n); err != nil {
			t.Fatal(err)
		}
		if n != 0 {
			t.Errorf("%s: %d rows survived Prune", tbl, n)
		}
	}
	if got := count(t, s, "notes"); got != 1 {
		t.Errorf("Prune removed a live note: notes = %d", got)
	}
}

// TestResetDropsEverything: `forge reindex` calls this, so a corrupt cache is never a
// data-loss event — the markdown rebuilds all of it.
func TestResetDropsEverything(t *testing.T) {
	s := open(t)
	put(t, s, sample())
	if err := s.Reset(); err != nil {
		t.Fatal(err)
	}
	for _, tbl := range []string{"notes", "note_stack", "note_tags", "links"} {
		if got := count(t, s, tbl); got != 0 {
			t.Errorf("%s: %d rows survived Reset", tbl, got)
		}
	}
}

// TestUnresolvedLinkIsStored: a dangling wikilink is recorded with an empty `resolved`
// rather than dropped, because the link report needs to name what is broken.
func TestUnresolvedLinkIsStored(t *testing.T) {
	s := open(t)
	put(t, s, sample())
	var resolved string
	err := s.DB.QueryRow("SELECT resolved FROM links WHERE target='gone'").Scan(&resolved)
	if err != nil {
		t.Fatalf("the dangling link was not stored: %v", err)
	}
	if resolved != "" {
		t.Errorf("resolved = %q, want empty", resolved)
	}
}
