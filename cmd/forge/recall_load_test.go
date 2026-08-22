package main

import (
	"testing"

	"knowledge-forge/pkg/store"
)

// closedStore is a Store whose handle is gone — the cheapest stand-in for the conditions
// refresh actually has to survive in the field (a concurrent forge intent holding the
// SQLite write lock, a full disk), none of which are constructible here.
func closedStore(t *testing.T) *store.Store {
	t.Helper()
	st, err := store.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if err := st.Close(); err != nil {
		t.Fatal(err)
	}
	return st
}

// B-029. writeRows must report what the old refresh() dropped. This is the whole value of
// the split: the error has to exist before ignoring it can be a decision rather than an
// accident, and before this fix three of them — Begin, Put and Commit — were discarded in
// a function that then returned nil on every path.
func TestWriteRowsReportsAFailedCacheWrite(t *testing.T) {
	if err := writeRows(closedStore(t), []store.Row{{Rel: "notes/concept/a.md"}}); err == nil {
		t.Error("writeRows on a closed store returned nil")
	}
}

// refresh swallows that error on purpose, and must keep doing so. Propagating it would
// reach runRecall, which prints and exits 1 without emitting candidates it has already
// scored correctly — trading a stale cache, which rowFor heals on the next run, for a
// discarded correct answer. The assertion is that this call is survivable, not that it
// succeeds.
func TestRefreshSwallowsAFailedCacheWrite(t *testing.T) {
	refresh(closedStore(t), []store.Row{{Rel: "notes/concept/a.md"}})
}
