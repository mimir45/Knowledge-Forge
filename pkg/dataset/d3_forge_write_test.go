package dataset

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"
)

// fixtureSrc mirrors cmd/forge/e2e_test.go's constant: testdata/vault must never be
// written to or git-init-ed in place, so every test that mutates a vault
// works on a copy staged into a temp dir.
const fixtureSrc = "../../testdata/vault"

// seedFromFixture stages testdata/vault into a fresh git repo and commits it as the
// starting state, so this test exercises D3 against the real fixture topology rather
// than a synthetic one-file repo.
func seedFromFixture(t *testing.T) *repo {
	t.Helper()
	r := newRepo(t)
	if err := filepath.WalkDir(fixtureSrc, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		rel, _ := filepath.Rel(fixtureSrc, p)
		b, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		r.write(rel, string(b))
		return nil
	}); err != nil {
		t.Fatalf("staging fixture: %v", err)
	}
	r.commit(day0, "seed from testdata/vault")
	return r
}

// TestForgeWriteTrailerSuppressesCapture is the acceptance test for a core D3 invariant:
// forge-librarian stamps Forge-Write on its own commits, and those commits must not enter D3
// as (model, model) pairs — see d3.go's ForgeTrailer doc comment. Step 2 is the pin: it
// proves step 1's zero-pairs result is the trailer guard firing, not a setup bug that
// would report zero regardless of what the commit trailer says.
func TestForgeWriteTrailerSuppressesCapture(t *testing.T) {
	r := seedFromFixture(t)
	r.write("notes/concept/b007-smoke.md", note("ask", "B007 Smoke", "generated text"))
	r.commit(day0, "generate a note")

	// 1. An edit committed with the trailer must be treated as forge's own, not a human
	// correction — zero pairs, even though the edit is otherwise indistinguishable from a
	// real correction.
	r.write("notes/concept/b007-smoke.md", note("ask", "B007 Smoke", "forge-rewritten text"))
	r.commit(day1, "forge-librarian regenerates\n\n"+ForgeTrailer+": true")
	if pairs := capture(t, r); len(pairs) != 0 {
		t.Fatalf("with %s trailer: got %d pairs, want 0: %+v", ForgeTrailer, len(pairs), pairs)
	}

	// 2. The identical shape of edit, minus the trailer, must be captured — proving (1)
	// wasn't zero for some unrelated reason (wrong window, wrong origin, wrong path).
	r.write("notes/concept/b007-smoke.md", note("ask", "B007 Smoke", "human-corrected text"))
	r.commit(day1, "a human corrects the note")
	pairs := capture(t, r)
	if len(pairs) != 1 {
		t.Fatalf("without trailer: got %d pairs, want 1: %+v", len(pairs), pairs)
	}
	if pairs[0].Note != "notes/concept/b007-smoke.md" {
		t.Errorf("pair is for %q, want notes/concept/b007-smoke.md", pairs[0].Note)
	}
}
