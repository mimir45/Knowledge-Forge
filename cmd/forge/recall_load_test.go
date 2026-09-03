package main

import (
	"testing"

	"github.com/mimir45/Knowledge-Forge/pkg/store"
)

// closedStore is a Store whose handle is gone.
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

// writeRows must report the errors an earlier version of this cache refresh silently
// dropped.
func TestWriteRowsReportsAFailedCacheWrite(t *testing.T) {
	if err := writeRows(closedStore(t), []store.Row{{Rel: "notes/concept/a.md"}}); err == nil {
		t.Error("writeRows on a closed store returned nil")
	}
}

// refresh swallows that error on purpose, and must keep doing so.
func TestRefreshSwallowsAFailedCacheWrite(t *testing.T) {
	refresh(closedStore(t), []store.Row{{Rel: "notes/concept/a.md"}})
}
