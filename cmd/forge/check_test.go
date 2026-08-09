package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"knowledge-forge/pkg/codeindex"
	"knowledge-forge/pkg/coderef"
	"knowledge-forge/pkg/gitsig"
	"knowledge-forge/pkg/linkcheck"
	"knowledge-forge/pkg/vault"
)

// gitVault stages the fixture and commits it, because two of the reports read history:
// churn.md is the vault's own commits, and a vault that is not a repo must degrade to a
// skipped file rather than a failed run.
func gitVault(t *testing.T) string {
	t.Helper()
	root := fixtureCopy(t)
	for _, args := range [][]string{
		{"init"}, {"config", "user.email", "t@example.com"},
		{"config", "user.name", "T"}, {"add", "-A"},
		{"-c", "commit.gpgsign=false", "commit", "-m", "fixture"},
	} {
		cmd := exec.Command("git", args...)
		cmd.Dir = root
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	return root
}

// TestCheckWritesTheVaultReports: without --repo there is nothing to check code against,
// so the run must produce the nine vault reports and skip drift.md and the codebase map
// rather than writing them empty.
func TestCheckWritesTheVaultReports(t *testing.T) {
	root := gitVault(t)
	if code := cmdCheck([]string{"--vault", root, "--offline"}); code != 0 {
		t.Fatalf("forge check exit %d", code)
	}
	for _, name := range []string{"coverage", "staleness", "duplicates", "orphans",
		"gaps", "graph-health", "churn", "deadlinks", "cost"} {
		if _, err := os.Stat(filepath.Join(root, "reports", name+".md")); err != nil {
			t.Errorf("reports/%s.md: %v", name, err)
		}
	}
	for _, rel := range []string{"reports/drift.md", "moc/codebase.md"} {
		if _, err := os.Stat(filepath.Join(root, rel)); err == nil {
			t.Errorf("%s written without --repo", rel)
		}
	}
}

// TestCheckIsIdempotentOnDisk: the headers carry a date, not a timestamp, so a second run
// on the same day must leave every byte and every mtime alone — a vault that rewrites its
// own reports every run buries real changes in its git diff.
func TestCheckIsIdempotentOnDisk(t *testing.T) {
	root := gitVault(t)
	cmdCheck([]string{"--vault", root, "--offline"})
	before := mtimes(t, filepath.Join(root, "reports"))
	time.Sleep(10 * time.Millisecond)
	cmdCheck([]string{"--vault", root, "--offline"})
	for path, was := range mtimes(t, filepath.Join(root, "reports")) {
		if !was.Equal(before[path]) {
			t.Errorf("%s rewritten on an unchanged vault", filepath.Base(path))
		}
	}
}

func mtimes(t *testing.T, dir string) map[string]time.Time {
	t.Helper()
	ents, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	out := map[string]time.Time{}
	for _, e := range ents {
		fi, err := e.Info()
		if err != nil {
			t.Fatal(err)
		}
		out[filepath.Join(dir, e.Name())] = fi.ModTime()
	}
	return out
}

// TestOneBadRendererCostsOneFile is the whole point of rendering each report under its own
// recover: a nil map in one collector must not cost the eight reports that collected fine.
func TestOneBadRendererCostsOneFile(t *testing.T) {
	root := t.TempDir()
	ok := func() ([]byte, error) { return []byte("fine\n"), nil }
	js := []job{
		{"reports/a.md", ok},
		{"reports/b.md", func() ([]byte, error) { var m map[string]int; m["x"] = 1; return nil, nil }},
		{"reports/c.md", ok},
	}
	if code := writeAll(root, js); code != 0 {
		t.Fatalf("writeAll exit %d; a failed report is not a failed run", code)
	}
	for _, name := range []string{"a.md", "c.md"} {
		if _, err := os.Stat(filepath.Join(root, "reports", name)); err != nil {
			t.Errorf("%s: %v", name, err)
		}
	}
	if _, err := os.Stat(filepath.Join(root, "reports", "b.md")); err == nil {
		t.Error("reports/b.md written by a panicking renderer")
	}
}

// TestCollectorErrorIsSkippedNotPanicked: a vault that is not a git repo is an expected
// condition, and churn.md must say "skipped" rather than report a runtime panic.
func TestCollectorErrorIsSkippedNotPanicked(t *testing.T) {
	root := fixtureCopy(t) // no git init on purpose
	if code := cmdCheck([]string{"--vault", root, "--offline"}); code != 0 {
		t.Fatalf("forge check exit %d on a non-repo vault", code)
	}
	if _, err := os.Stat(filepath.Join(root, "reports", "churn.md")); err == nil {
		t.Error("churn.md written from a vault with no history")
	}
	if _, err := os.Stat(filepath.Join(root, "reports", "coverage.md")); err != nil {
		t.Errorf("coverage.md lost to churn's failure: %v", err)
	}
}

func TestTypeOf(t *testing.T) {
	cases := []struct{ rel, want string }{
		{"notes/concept/foo.md", "concept"},
		{"notes/howto/deep/bar.md", "howto"},
		{"moc/java.md", "moc"},
		{"notes/loose.md", ""},
		{"top.md", ""},
	}
	for _, c := range cases {
		if got := typeOf(c.rel); got != c.want {
			t.Errorf("typeOf(%q) = %q, want %q", c.rel, got, c.want)
		}
	}
}

// TestNotesAndEntriesDifferByExactlyTheNonContractNotes is the crisp check on the two
// populations: everything the graph counts, minus the maps and hubs the schema does not
// judge, is what coverage.md and staleness.md are allowed to divide by.
func TestNotesAndEntriesDifferByExactlyTheNonContractNotes(t *testing.T) {
	root := gitVault(t)
	d, err := collectVault(checkCfg{vault: root, offline: true}, root)
	if err != nil {
		t.Fatal(err)
	}
	want := 0
	for _, n := range d.notes {
		if vault.IsContractNote(n.Rel) {
			want++
		}
	}
	if len(d.entries) != want {
		t.Errorf("entries = %d, contract notes = %d", len(d.entries), want)
	}
	if len(d.entries) > len(d.notes) {
		t.Errorf("entries (%d) exceeds notes (%d)", len(d.entries), len(d.notes))
	}
}

func TestSourceURLs(t *testing.T) {
	cases := []struct {
		name, fm string
		want     []string
	}{
		// The schema's key is plural. Reading only the singular is how the first real run
		// reported 0 cited URLs across 91 notes that all carry sources.
		{"schema shape: sources, list of mappings",
			"sources:\n  - url: https://a.example/x\n    kind: official\n",
			[]string{"https://a.example/x"}},
		{"pre-migration singular, list of mappings",
			"source:\n  - url: https://a.example/x\n    title: A\n",
			[]string{"https://a.example/x"}},
		{"bare scalar", "source: https://b.example/y\n", []string{"https://b.example/y"}},
		{"vault-relative path is not an HTTP source",
			"sources:\n  - url: sources/daily/2026-04-13.md\n    kind: session\n", nil},
		{"empty list", "sources: []\n", nil},
		{"absent", "type: concept\n", nil},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			fm, _, _ := vault.SplitFrontmatter([]byte("---\n" + c.fm + "---\n\nbody\n"))
			p, err := vault.ParseFrontmatter(fm)
			if err != nil {
				t.Fatal(err)
			}
			got := sourceURLs(p)
			if strings.Join(got, ",") != strings.Join(c.want, ",") {
				t.Errorf("sourceURLs = %v, want %v", got, c.want)
			}
		})
	}
}

// TestCachedOnlyKeepsTheDenominator: an offline run learned nothing about an uncached URL,
// and dropping it would let deadlinks.md claim a cleaner vault than it actually checked.
func TestCachedOnlyKeepsTheDenominator(t *testing.T) {
	dir := t.TempDir()
	c := linkcheck.LoadCache(dir)
	c.Put(linkcheck.Status{URL: "https://known.example", Verdict: linkcheck.Alive,
		Code: 200, Checked: time.Now()})
	if err := c.Save(); err != nil {
		t.Fatal(err)
	}
	got := cachedOnly(dir, []string{"https://known.example", "https://unknown.example"})
	if len(got) != 2 {
		t.Fatalf("got %d statuses for 2 urls", len(got))
	}
	if got[0].Verdict != linkcheck.Alive {
		t.Errorf("cached url: verdict %v", got[0].Verdict)
	}
	if got[1].Verdict != linkcheck.Unreachable || got[1].Detail != "offline, not cached" {
		t.Errorf("uncached url: %+v, want unreachable/offline", got[1])
	}
}

// TestUncoveredOfThresholds: section B.5 asks for high churn and real size. A file touched
// once is not churning, a short symbol is not a documentation gap, and a cited file is
// covered whatever its symbols look like.
func TestUncoveredOfThresholds(t *testing.T) {
	big := codeindex.Symbol{Name: "Big", Start: 1, End: minSymbolLOC}
	small := codeindex.Symbol{Name: "Small", Start: 1, End: 3}
	ix := codeindex.Index{Files: map[string]codeindex.File{
		"a.java": {Path: "a.java", Symbols: []codeindex.Symbol{big, small}},
		"b.java": {Path: "b.java", Symbols: []codeindex.Symbol{big}}, // churn below the floor
		"c.java": {Path: "c.java", Symbols: []codeindex.Symbol{big}}, // cited
	}}
	st := &gitsig.Stats{Churn: map[string]int{"a.java": 5, "b.java": 1, "c.java": 9}}
	got := uncoveredOf(ix, st, map[string][]string{"c.java": {"some-note"}})
	if len(got) != 1 {
		t.Fatalf("uncovered = %+v, want only a.java's Big", got)
	}
	if got[0].Path != "a.java" || got[0].Symbol != "Big" || got[0].Commits != 5 {
		t.Errorf("uncovered[0] = %+v", got[0])
	}
}

// symbolSource is drift's symbol table reduced to the one question citedPaths asks it:
// which file declares this name.
type symbolSource map[string]string

func (s symbolSource) Find(name, asOf string) (string, string, codeindex.Symbol, bool) {
	p, ok := s[name]
	return "repo", p, codeindex.Symbol{Name: name}, ok
}

// TestSymbolCitationCoversItsFile is the defect the first real run produced: leprecoin's
// map listed SignUpPage as "0 notes" in the same run where drift.md named two notes citing
// it. Most of the vault cites a class and no path, and coderef gives those no RepoPath, so
// coverage has to reach the symbol table the way drift does.
func TestSymbolCitationCoversItsFile(t *testing.T) {
	rg := coderef.NewRegistry(nil)
	src := symbolSource{"SignUpPage": "src/app/SignUpPage.tsx"}
	refs := coderef.FromBody("notes/decision/a.md", []byte("the `SignUpPage` wrapper\n"))
	if len(refs) != 1 || refs[0].Kind != coderef.KindSymbol {
		t.Fatalf("FromBody gave %+v, want one symbol ref", refs)
	}
	repo, p, found := locate(refs[0], rg, src)
	if !found || repo != "repo" || p != "src/app/SignUpPage.tsx" {
		t.Fatalf("locate = %q %q %v, want the file the symbol table names", repo, p, found)
	}
	cited := map[string][]string{p: {"a-note"}}
	ix := codeindex.Index{Files: map[string]codeindex.File{p: {Path: p,
		Symbols: []codeindex.Symbol{{Name: "SignUpPage", Start: 1, End: 40}}}}}
	st := &gitsig.Stats{Churn: map[string]int{p: 13}}
	if got := uncoveredOf(ix, st, cited); len(got) != 0 {
		t.Errorf("uncovered = %+v, want none: the file is cited by name", got)
	}
}
