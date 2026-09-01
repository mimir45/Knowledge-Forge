package drift

import (
	"testing"

	"github.com/mimir45/Knowledge-Forge/pkg/codeindex"
	"github.com/mimir45/Knowledge-Forge/pkg/coderef"
)

// fakeSource stands in for git so the ladder can be tested one rung at a time. Every
// verdict below is a pure function of what this map says at two revisions, which is the
// same contract GitSource honours against real objects.
type fakeSource struct {
	files map[string]codeindex.File // "repo@rev@path"
	revs  map[string]string         // "repo@date"
}

const headRev = "headsha"

func (f *fakeSource) Head(string) string { return headRev }

func (f *fakeSource) RevBefore(repo, date string) string { return f.revs[repo+"@"+date] }

func (f *fakeSource) At(repo, path, rev string) (codeindex.File, bool) {
	c, ok := f.files[repo+"@"+rev+"@"+path]
	return c, ok
}

func (f *fakeSource) Find(name, asOf string) (string, string, codeindex.Symbol, bool) {
	rev := headRev
	if asOf != "" {
		rev = f.revs["app@"+asOf]
	}
	for key, file := range f.files {
		if s, ok := file.Lookup(name); ok && key == "app@"+rev+"@"+file.Path {
			return "app", file.Path, s, true
		}
	}
	return "", "", codeindex.Symbol{}, false
}

func (f *fakeSource) ResolveAt(ref coderef.Ref, asOf string) coderef.Resolution {
	rev := headRev
	if asOf != "" {
		rev = f.revs["app@"+asOf]
	}
	if _, ok := f.files["app@"+rev+"@"+ref.Path]; ok {
		ref.Repo = "app"
		return coderef.Resolution{Ref: ref, Status: coderef.Resolved, RepoPath: ref.Path}
	}
	return coderef.Resolution{Ref: ref, Status: coderef.Unresolved}
}

func sym(name string, start int, body string) codeindex.Symbol {
	return codeindex.Symbol{Name: name, Kind: "method", Start: start, BodyHash: body}
}

func file(path string, syms ...codeindex.Symbol) codeindex.File {
	return codeindex.File{Path: path, Lang: "java", Symbols: syms}
}

const p = "src/main/java/Order.java"

// src builds a source where the note's verified-era tree ("old") and HEAD differ only in
// the ways each case is about.
func src(now, then codeindex.File) *fakeSource {
	f := &fakeSource{
		files: map[string]codeindex.File{},
		revs:  map[string]string{"app@2026-01-01": "oldsha"},
	}
	if now.Path != "" {
		f.files["app@"+headRev+"@"+now.Path] = now
	}
	if then.Path != "" {
		f.files["app@oldsha@"+then.Path] = then
	}
	return f
}

func registry() *coderef.Registry {
	return coderef.NewRegistry([]coderef.Repo{{Name: "app", Root: "/app", Files: []string{p}}})
}

func TestCheckLadder(t *testing.T) {
	was := file(p, sym("Order.place", 10, "h1"))
	cases := []struct {
		name string
		ref  coderef.Ref
		now  codeindex.File
		want Verdict
	}{
		{"file deleted", pathRef(10), codeindex.File{}, Broken},
		{"symbol removed", pathRef(10), file(p, sym("Order.cancel", 10, "h1")), Broken},
		{"line moved", pathRef(10), file(p, sym("Order.place", 42, "h1")), Repaired},
		{"body changed", pathRef(10), file(p, sym("Order.place", 10, "h2")), Suspect},
		{"unchanged", pathRef(10), was, OK},
		// A deleted file must not report as a moved line: the ladder's order is the test.
		{"deleted beats moved", pathRef(99), codeindex.File{}, Broken},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			n := Note{Rel: "notes/a.md", Verified: "2026-01-01", Refs: []coderef.Ref{c.ref}}
			got := Check([]Note{n}, registry(), src(c.now, was), nil, Opts{})
			assertOne(t, got, c.want)
		})
	}
}

func pathRef(line int) coderef.Ref {
	return coderef.Ref{Raw: p, Kind: coderef.KindPath, Path: p, Symbol: "Order.place", Line: line}
}

// The cheap gate is a hard requirement, not an optimisation: a citation whose file git
// did not touch must produce no finding at all, so the hook path never re-scans.
func TestChangedGateSkipsUntouchedFiles(t *testing.T) {
	was := file(p, sym("Order.place", 10, "h1"))
	n := Note{Rel: "notes/a.md", Verified: "2026-01-01", Refs: []coderef.Ref{pathRef(10)}}
	got := Check([]Note{n}, registry(), src(codeindex.File{}, was),
		&Changed{Touched: map[string]bool{"other/File.java": true}}, Opts{})
	if len(got) != 0 {
		t.Fatalf("findings = %+v, want none: the file was outside the gate", got)
	}
}

func TestSymbolOnlyCitations(t *testing.T) {
	was := file(p, sym("OrderConsumer", 3, "h1"))
	ref := coderef.Ref{Raw: "OrderConsumer", Kind: coderef.KindSymbol, Symbol: "OrderConsumer"}
	n := Note{Rel: "notes/a.md", Verified: "2026-01-01", Refs: []coderef.Ref{ref}}

	t.Run("present at head", func(t *testing.T) {
		assertOne(t, Check([]Note{n}, registry(), src(was, was), nil, Opts{}), OK)
	})
	// On the hook path an unknown name is far more often a library type than a deletion.
	t.Run("absent, shallow", func(t *testing.T) {
		assertOne(t, Check([]Note{n}, registry(), src(codeindex.File{}, was), nil, Opts{}), Skipped)
	})
	// Only the weekly pass pays to prove the name was once declared here.
	t.Run("absent, deep", func(t *testing.T) {
		got := Check([]Note{n}, registry(), src(codeindex.File{}, was), nil, Opts{Deep: true})
		assertOne(t, got, Broken)
	})
	t.Run("absent at both revisions", func(t *testing.T) {
		got := Check([]Note{n}, registry(), src(codeindex.File{}, codeindex.File{}), nil,
			Opts{Deep: true})
		assertOne(t, got, Skipped)
	})
}

// p2 is a second path, never registered by registry(), so a citation naming it always
// resolves Unresolved against the registry — the routing bug's trigger condition — while
// still being locatable in a fakeSource's own files/revs for the ResolveAt fallback.
const p2 = "src/main/java/OldOrder.java"

func p2Ref() coderef.Ref { return coderef.Ref{Raw: p2, Kind: coderef.KindPath, Path: p2} }

// TestUnresolvedPathFallback covers unresolvedPath's shallow/deep, gate-ordering and
// no-baseline branches. Every case resolves Unresolved against registry() (p2 is not
// registered); what varies is whether the deep-sweep fallback then finds p2 in history.
func TestUnresolvedPathFallback(t *testing.T) {
	wasP2 := file(p2, sym("OldOrder.cancel", 5, "h1"))
	deleted := src(codeindex.File{}, wasP2) // gone at HEAD, present at the verified date
	n := Note{Rel: "notes/a.md", Verified: "2026-01-01", Refs: []coderef.Ref{p2Ref()}}

	// (a) regression test for the gate-ordering bug: a partial run (changed != nil, and
	// its Deleted set does not name p2) must produce no finding at all, not SKIPPED —
	// SKIPPED would let Apply's restore path flip the note back up on an unrelated later
	// commit, which is exactly the regression the hook-path deletion fix must not reintroduce.
	t.Run("partial run produces no finding outside its own deletion evidence", func(t *testing.T) {
		got := Check([]Note{n}, registry(), deleted,
			&Changed{Touched: map[string]bool{"other/File.java": true}}, Opts{Deep: true})
		if len(got) != 0 {
			t.Fatalf("findings = %+v, want none: outside the gate, nothing to report", got)
		}
	})

	// (b) the true full sweep: changed == nil, deep, history confirms the file existed.
	t.Run("full sweep finds the deletion", func(t *testing.T) {
		got := Check([]Note{n}, registry(), deleted, nil, Opts{Deep: true})
		assertOne(t, got, Broken)
	})

	// (c) full sweep, but history has no record of p2 either — absent at both revisions.
	t.Run("absent at both revisions stays skipped", func(t *testing.T) {
		neverExisted := src(codeindex.File{}, codeindex.File{})
		got := Check([]Note{n}, registry(), neverExisted, nil, Opts{Deep: true})
		assertOne(t, got, Skipped)
	})

	// (d) Deep alone gates this, independent of the gate-ordering check above.
	t.Run("shallow run stays skipped regardless of gate", func(t *testing.T) {
		got := Check([]Note{n}, registry(), deleted, nil, Opts{Deep: false})
		assertOne(t, got, Skipped)
	})

	// (e) no verified date: no baseline to prove "it was there."
	t.Run("no verified date stays skipped", func(t *testing.T) {
		unverified := Note{Rel: "notes/a.md", Refs: []coderef.Ref{p2Ref()}}
		got := Check([]Note{unverified}, registry(), deleted, nil, Opts{Deep: true})
		assertOne(t, got, Skipped)
	})
}

// TestUnresolvedPathSameCommitDeletion covers same-commit deletion detection: the hook path (changed != nil,
// Deep == false) now gets same-commit deletion evidence straight from the gate, with no
// historical registry scan and no wait for the next full sweep.
func TestUnresolvedPathSameCommitDeletion(t *testing.T) {
	n := Note{Rel: "notes/a.md", Refs: []coderef.Ref{p2Ref()}}
	shallow := src(codeindex.File{}, codeindex.File{}) // no historical baseline needed

	t.Run("gate reports the exact path deleted", func(t *testing.T) {
		gate := &Changed{Deleted: map[string]string{p2: "app"}}
		got := Check([]Note{n}, registry(), shallow, gate, Opts{})
		assertOne(t, got, Broken)
	})
	t.Run("gate reports a different basename deleted elsewhere", func(t *testing.T) {
		gate := &Changed{Deleted: map[string]string{"other/Unrelated.java": "app"}}
		got := Check([]Note{n}, registry(), shallow, gate, Opts{})
		if len(got) != 0 {
			t.Fatalf("findings = %+v, want none: no deletion evidence for this citation", got)
		}
	})
	t.Run("gate touched but did not delete", func(t *testing.T) {
		gate := &Changed{Touched: map[string]bool{p2: true}}
		got := Check([]Note{n}, registry(), shallow, gate, Opts{})
		if len(got) != 0 {
			t.Fatalf("findings = %+v, want none: touched is not deleted", got)
		}
	})
	// The whole point of this detection path: no --deep, no verified date, still BROKEN immediately.
	t.Run("shallow, no verified date, still catches it", func(t *testing.T) {
		gate := &Changed{Deleted: map[string]string{p2: "app"}}
		got := Check([]Note{n}, registry(), shallow, gate, Opts{Deep: false})
		assertOne(t, got, Broken)
	})
}

func TestUnresolvedPathIsSkippedNotBroken(t *testing.T) {
	ref := coderef.Ref{Raw: "~/.continue/config.ts", Kind: coderef.KindPath,
		Path: "~/.continue/config.ts"}
	n := Note{Rel: "notes/a.md", Refs: []coderef.Ref{ref}}
	assertOne(t, Check([]Note{n}, registry(), src(codeindex.File{}, codeindex.File{}), nil,
		Opts{}), Skipped)
}

// Only BROKEN costs a note its confidence. SUSPECT is a prompt to re-read, not evidence
// the note is wrong, and demoting on it would make every refactor a vault-wide downgrade.
func TestOnlyBrokenDemotes(t *testing.T) {
	for v, want := range map[Verdict]bool{Broken: true, Suspect: false, Repaired: false,
		OK: false, Skipped: false} {
		if got := (Finding{Verdict: v}).Demoting(); got != want {
			t.Errorf("%s demoting = %v, want %v", v, got, want)
		}
	}
}

func assertOne(t *testing.T, got []Finding, want Verdict) {
	t.Helper()
	if len(got) != 1 {
		t.Fatalf("findings = %+v, want exactly one", got)
	}
	if got[0].Verdict != want {
		t.Fatalf("verdict = %s (%s), want %s", got[0].Verdict, got[0].Reason, want)
	}
}

func BenchmarkCheck(b *testing.B) {
	was := file(p, sym("Order.place", 10, "h1"))
	s, rg := src(was, was), registry()
	notes := make([]Note, 200)
	for i := range notes {
		notes[i] = Note{Rel: "notes/a.md", Verified: "2026-01-01", Refs: []coderef.Ref{pathRef(10)}}
	}
	for b.Loop() {
		Check(notes, rg, s, nil, Opts{})
	}
}
