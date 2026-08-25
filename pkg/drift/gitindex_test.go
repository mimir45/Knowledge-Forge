package drift

import (
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"testing"

	"knowledge-forge/pkg/codeindex"
	"knowledge-forge/pkg/coderef"
)

// TestNameMapOrdersDeclarationsWithinOneFile is the bug drift.md showed on the real vault:
// its verdict count flipped between 9 and 10 notes across runs on an unchanged tree. One
// Java file declares both `Order.Builder` and `OrderItem.Builder`, the short-name hits tied
// on (repo, path), and sort.Slice settled the tie by map iteration order. Find must pick
// the same declaration every time or a verdict is not a function of the tree.
func TestNameMapOrdersDeclarationsWithinOneFile(t *testing.T) {
	want := []loc{
		{"a", "src/Order.java", codeindex.Symbol{Name: "Order.Builder", Start: 10}},
		{"a", "src/Order.java", codeindex.Symbol{Name: "OrderItem.Builder", Start: 90}},
		{"b", "src/Other.java", codeindex.Symbol{Name: "Other.Builder", Start: 5}},
	}
	for run := 0; run < 50; run++ {
		ls := append([]loc(nil), want...)
		rand.Shuffle(len(ls), func(i, j int) { ls[i], ls[j] = ls[j], ls[i] })
		m := &nameMap{full: map[string][]loc{}, short: map[string][]loc{"Builder": ls}}
		m.sort()
		if got := m.short["Builder"]; !sameOrder(got, want) {
			t.Fatalf("run %d: %+v, want %+v", run, got, want)
		}
	}
}

// TestResolveAtFindsFileDeletedFromHistory is B-026's real-git twin of
// TestUnresolvedPathFallback: everything in that test answers ResolveAt by map
// membership, which cannot fail the way production code can. Only GitSource.registryAt
// exercises coderef.ScanRepo against real git objects at a past revision, so this test
// writes an actual .java file, deletes it, and confirms ResolveAt tells the two revisions
// apart — and that registryAt memoises rather than rescanning per call. It needs no
// codeindex.Available() guard: ResolveAt never touches codeindex, so it must pass
// identically on both build lanes.
func TestResolveAtFindsFileDeletedFromHistory(t *testing.T) {
	repo := t.TempDir()
	writeRepo(t, repo, orderV1)
	commitDated(t, repo, "add Order", "2020-01-01T12:00:00")
	git(t, repo, "rm", "-q", "src/main/java/Order.java")
	commitDated(t, repo, "delete Order", "2020-06-01T12:00:00")

	gs := NewGitSource([]Repo{{Name: "app", Root: repo}}, "")
	ref := coderef.Ref{Raw: "Order.java", Kind: coderef.KindPath, Path: "src/main/java/Order.java"}
	assertResolveAt(t, gs, ref, "2020-01-01", coderef.Resolved)
	assertResolveAt(t, gs, ref, "", coderef.Unresolved)

	if rg1, rg2 := gs.registryAt(""), gs.registryAt(""); rg1 != rg2 {
		t.Error("registryAt(HEAD) not memoised: distinct *Registry on repeat calls")
	}
}

func assertResolveAt(t *testing.T, gs *GitSource, ref coderef.Ref, asOf string,
	want coderef.Status) {

	t.Helper()
	if res := gs.ResolveAt(ref, asOf); res.Status != want {
		t.Fatalf("ResolveAt(asOf=%q) = %+v, want status %s", asOf, res, want)
	}
}

// commitDated is commit (rollback_test.go), sharing its ensureGitRepo setup, but with an
// explicit commit date, needed here so a "verified at" asOf can land strictly between the
// add and the delete regardless of when the test itself runs.
func commitDated(t *testing.T, root, msg, date string) string {
	t.Helper()
	ensureGitRepo(t, root)
	git(t, root, "add", "-A")
	cmd := exec.Command("git", "-C", root, "commit", "-q", "-m", msg)
	cmd.Env = append(os.Environ(), "GIT_AUTHOR_DATE="+date, "GIT_COMMITTER_DATE="+date)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git commit: %v\n%s", err, out)
	}
	return head(t, root)
}

func sameOrder(got, want []loc) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i] != want[i] {
			return false
		}
	}
	return true
}

// TestGitSourceRebuildsFromScratchOnStaleExtractor is B-015's own cache-invalidation
// check, over the real hook path rather than codeindex.Load in isolation: a cache file
// stamped by an older Extractor must make GitSource.build take the full-rebuild branch,
// not Patch the bogus stale entry forward as if it still described the tree.
func TestGitSourceRebuildsFromScratchOnStaleExtractor(t *testing.T) {
	if !codeindex.Available() {
		t.Skip("built without cgo: no symbol table to rebuild")
	}
	repo, cacheDir := t.TempDir(), t.TempDir()
	writeRepo(t, repo, orderV1)
	commit(t, repo, "add Order")

	stale := codeindex.Index{Repo: "app", Commit: "not-a-real-commit", Extractor: codeindex.Extractor - 1,
		Files: map[string]codeindex.File{"src/main/java/Order.java": {Path: "bogus.java"}}}
	if err := codeindex.Save(filepath.Join(cacheDir, "code-index-app.json"), stale); err != nil {
		t.Fatal(err)
	}

	gs := NewGitSource([]Repo{{Name: "app", Root: repo}}, cacheDir)
	ix := gs.Index("app", "HEAD")
	if ix.Extractor != codeindex.Extractor {
		t.Fatalf("Extractor = %d, want current %d — stale cache was not discarded",
			ix.Extractor, codeindex.Extractor)
	}
	f, ok := ix.Files["src/main/java/Order.java"]
	if !ok {
		t.Fatalf("files = %v, want a fresh parse of src/main/java/Order.java", ix.Files)
	}
	if _, ok := f.Lookup("Order.place"); !ok {
		t.Errorf("symbols = %+v, want the real Order.place — got the stale cache's shape", f.Symbols)
	}
}

// TestLessLocIsATotalOrder: a comparator that reports two distinct declarations equal is
// exactly what let an unstable sort choose between them.
func TestLessLocIsATotalOrder(t *testing.T) {
	ls := []loc{
		{"a", "p.java", codeindex.Symbol{Name: "X", Start: 1}},
		{"a", "p.java", codeindex.Symbol{Name: "Y", Start: 1}},
		{"a", "p.java", codeindex.Symbol{Name: "Z", Start: 2}},
	}
	sort.Slice(ls, func(i, j int) bool { return lessLoc(ls[i], ls[j]) })
	for i := 0; i+1 < len(ls); i++ {
		if !lessLoc(ls[i], ls[i+1]) {
			t.Errorf("lessLoc reports %+v and %+v equal", ls[i], ls[i+1])
		}
	}
}
