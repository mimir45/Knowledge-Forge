package main

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"knowledge-forge/pkg/vault"
)

const fixtureSrc = "../../testdata/vault"

// fixtureCopy stages the fixture vault in a temp dir. Everything that mutates a vault is
// rehearsed on a copy: testdata/vault carries twelve deliberate defects that are the test
// surface, and it must never be written to or git-init-ed in place (BACKLOG B-002).
func fixtureCopy(t *testing.T) string {
	t.Helper()
	dst := filepath.Join(t.TempDir(), "vault")
	err := filepath.WalkDir(fixtureSrc, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(fixtureSrc, p)
		if d.IsDir() {
			return os.MkdirAll(filepath.Join(dst, rel), 0o755)
		}
		return copyFile(p, filepath.Join(dst, rel))
	})
	if err != nil {
		t.Fatalf("staging fixture: %v", err)
	}
	return dst
}

func copyFile(src, dst string) error {
	b, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, b, 0o644)
}

// bodies returns every note's body, keyed by relative path, so a mutation pass can be
// checked for having left all prose untouched.
func bodies(t *testing.T, root string) map[string]string {
	t.Helper()
	rels, err := vault.Walk(root)
	if err != nil {
		t.Fatal(err)
	}
	out := map[string]string{}
	for _, rel := range rels {
		b, err := os.ReadFile(filepath.Join(root, rel))
		if err != nil {
			t.Fatal(err)
		}
		_, body, _ := vault.SplitFrontmatter(b)
		out[rel] = string(body)
	}
	return out
}

// TestE2EValidateFailsOnTheFixture: the fixture is deliberately broken, so a clean exit
// code here would mean the validator is not looking.
func TestE2EValidateFailsOnTheFixture(t *testing.T) {
	if code := runValidate(nil, true, fixtureCopy(t), false, true); code != 1 {
		t.Errorf("exit = %d, want 1 on a vault with known defects", code)
	}
}

// TestE2EFixReducesIssuesWithoutTouchingBodies is the migration's core safety claim,
// exercised end to end rather than per function.
func TestE2EFixReducesIssuesWithoutTouchingBodies(t *testing.T) {
	root := fixtureCopy(t)
	before := bodies(t, root)
	if code := runValidate(nil, true, root, true, true); code != 1 {
		t.Errorf("exit = %d; the fixture cannot become fully valid by --fix alone", code)
	}
	for rel, body := range bodies(t, root) {
		if before[rel] != body {
			t.Errorf("%s: body changed by --fix:\n got %q\nwant %q", rel, body, before[rel])
		}
	}
}

// TestE2EFixIsIdempotent: a second --fix pass must report zero further changes, or every
// migration rerun would churn the vault.
func TestE2EFixIsIdempotent(t *testing.T) {
	root := fixtureCopy(t)
	runValidate(nil, true, root, true, true)
	after := snapshot(t, root)
	runValidate(nil, true, root, true, true)
	for rel, want := range after {
		if got := snapshot(t, root)[rel]; got != want {
			t.Errorf("%s changed on the second --fix pass", rel)
		}
	}
}

func snapshot(t *testing.T, root string) map[string]string {
	t.Helper()
	rels, err := vault.Walk(root)
	if err != nil {
		t.Fatal(err)
	}
	out := map[string]string{}
	for _, rel := range rels {
		b, err := os.ReadFile(filepath.Join(root, rel))
		if err != nil {
			t.Fatal(err)
		}
		out[rel] = string(b)
	}
	return out
}

// TestE2EIndexIsIdempotentOnDisk: the second run must leave the file's mtime alone, not
// merely write the same bytes — downstream mtime caches depend on it.
func TestE2EIndexIsIdempotentOnDisk(t *testing.T) {
	root := fixtureCopy(t)
	if code := runIndex(root, "_index.md", 4096, false); code != 0 {
		t.Fatalf("first index exit = %d", code)
	}
	out := filepath.Join(root, "_index.md")
	first, err := os.Stat(out)
	if err != nil {
		t.Fatal(err)
	}
	if code := runIndex(root, "_index.md", 4096, false); code != 0 {
		t.Fatalf("second index exit = %d", code)
	}
	second, err := os.Stat(out)
	if err != nil {
		t.Fatal(err)
	}
	if !first.ModTime().Equal(second.ModTime()) {
		t.Error("_index.md was rewritten on an unchanged vault")
	}
}

// TestE2EIndexRespectsTheBudget and covers the DESIGN §7.1 4KB SessionStart budget.
func TestE2EIndexRespectsTheBudget(t *testing.T) {
	root := fixtureCopy(t)
	runIndex(root, "_index.md", 4096, false)
	b, err := os.ReadFile(filepath.Join(root, "_index.md"))
	if err != nil {
		t.Fatal(err)
	}
	if len(b) > 4096 {
		t.Errorf("index is %d bytes, over the 4096 budget", len(b))
	}
	if !strings.HasPrefix(string(b), "# Vault index — ") {
		t.Errorf("unexpected header: %.60s", b)
	}
}

// TestE2EReindexRebuildsFromMarkdown: deleting the derived cache must lose nothing.
// Markdown is the only source of truth.
func TestE2EReindexRebuildsFromMarkdown(t *testing.T) {
	root := fixtureCopy(t)
	runIndex(root, "_index.md", 4096, false)
	want, err := os.ReadFile(filepath.Join(root, "_index.md"))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.RemoveAll(filepath.Join(root, ".forge")); err != nil {
		t.Fatal(err)
	}
	if code := runIndex(root, "_index.md", 4096, true); code != 0 {
		t.Fatalf("reindex exit = %d", code)
	}
	got, err := os.ReadFile(filepath.Join(root, "_index.md"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(want) {
		t.Error("reindex from markdown produced a different index than the cached run")
	}
}

// TestE2EExcludesNonNotes: raw/, sources/ and archive/ are inputs and history, not vault
// notes. Counting them would inflate every metric Phase 0's baseline is measured against.
func TestE2EExcludesNonNotes(t *testing.T) {
	root := fixtureCopy(t)
	rels, err := vault.Walk(root)
	if err != nil {
		t.Fatal(err)
	}
	counted := 0
	for _, rel := range rels {
		if vault.IsContentNote(rel) {
			counted++
		}
	}
	if len(rels) != 15 || counted != 11 {
		t.Errorf("walked %d files, counted %d notes; want 15 and 11", len(rels), counted)
	}
}

// TestFixtureIsNeverMutated guards the rule directly: the fixture on disk must still be
// dirty and must still have no .git after the whole suite has run against copies of it.
func TestFixtureIsNeverMutated(t *testing.T) {
	if _, err := os.Stat(filepath.Join(fixtureSrc, ".git")); !os.IsNotExist(err) {
		t.Error("testdata/vault has a .git; it must never be git-init-ed in place")
	}
	if _, err := os.Stat(filepath.Join(fixtureSrc, "_index.md")); !os.IsNotExist(err) {
		t.Error("a generated _index.md landed in testdata/vault")
	}
	if _, err := os.Stat(filepath.Join(fixtureSrc, ".forge")); !os.IsNotExist(err) {
		t.Error("a derived .forge cache landed in testdata/vault")
	}
}
