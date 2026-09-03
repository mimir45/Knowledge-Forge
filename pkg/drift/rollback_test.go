package drift

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mimir45/Knowledge-Forge/pkg/codeindex"
	"github.com/mimir45/Knowledge-Forge/pkg/coderef"
	"github.com/mimir45/Knowledge-Forge/pkg/vault"
)

const orderV1 = `package com.food.order;

public class Order {
    public void place(String id) {
        repo.save(id);
    }
}
`

// v2 renames the cited method. To drift that is a removal, which is the one verdict that
// costs a note its confidence.
const orderV2 = `package com.food.order;

public class Order {
    public void submit(String id) {
        repo.save(id);
    }
}
`

const noteSrc = `---
title: Placing an order
slug: placing-an-order
type: concept
confidence: high
created: 2026-01-01
updated: 2026-01-01
verified: 2026-01-01
code_refs: [app:src/main/java/Order.java:4#Order.place]
---

Order placement goes through Order.place.
`

// TestRollbackSymmetry is the brief's mandated test: break a symbol, assert demotion;
// revert, assert restoration.
func TestRollbackSymmetry(t *testing.T) {
	if !codeindex.Available() {
		t.Skip("built without cgo: no symbol table to break")
	}
	repo, vaultDir := t.TempDir(), t.TempDir()
	writeRepo(t, repo, orderV1)
	base := commit(t, repo, "add Order")
	notePath := filepath.Join(vaultDir, "note.md")
	if err := os.WriteFile(notePath, []byte(noteSrc), 0o644); err != nil {
		t.Fatal(err)
	}

	writeRepo(t, repo, orderV2)
	broke := commit(t, repo, "rename place -> submit")
	if got := run(t, repo, vaultDir, base, broke); got != "low" {
		t.Fatalf("confidence after rename = %q, want low", got)
	}
	assertStored(t, vaultDir, "high")

	git(t, repo, "revert", "--no-edit", broke)
	restored := head(t, repo)
	if got := run(t, repo, vaultDir, broke, restored); got != "high" {
		t.Fatalf("confidence after revert = %q, want high — the symbol is back", got)
	}
	assertCleared(t, vaultDir)
	assertLogCitesBoth(t, vaultDir, broke, restored)
}

// TestRollbackSymmetryOnDeletion covers same-commit deletion end to end: a file deleted
// in the very commit the hook checks demotes the citing note immediately.
func TestRollbackSymmetryOnDeletion(t *testing.T) {
	repo, vaultDir := t.TempDir(), t.TempDir()
	writeRepo(t, repo, orderV1)
	base := commit(t, repo, "add Order")
	notePath := filepath.Join(vaultDir, "note.md")
	if err := os.WriteFile(notePath, []byte(noteSrc), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := os.Remove(filepath.Join(repo, "src/main/java/Order.java")); err != nil {
		t.Fatal(err)
	}
	deleted := commit(t, repo, "delete Order.java")
	if got := run(t, repo, vaultDir, base, deleted); got != "low" {
		t.Fatalf("confidence after deletion = %q, want low", got)
	}
	assertStored(t, vaultDir, "high")

	if err := os.WriteFile(filepath.Join(repo, "unrelated.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	unrelated := commit(t, repo, "unrelated change")
	if got := run(t, repo, vaultDir, deleted, unrelated); got != "low" {
		t.Fatalf("confidence after unrelated commit = %q, want still low: the flip-flop bug", got)
	}

	if !codeindex.Available() {
		return // the deletion and flip-flop legs above are pure path matching; only the
		// restore leg below needs a symbol table to resolve the citation back to OK.
	}
	git(t, repo, "revert", "--no-edit", deleted)
	restored := head(t, repo)
	if got := run(t, repo, vaultDir, unrelated, restored); got != "high" {
		t.Fatalf("confidence after revert = %q, want high", got)
	}
	assertCleared(t, vaultDir)
}

// run is one `forge drift --since-commit` invocation: gate, check, apply.
func run(t *testing.T, repo, vaultDir, since, to string) string {
	t.Helper()
	src := NewGitSource([]Repo{{Name: "app", Root: repo}}, filepath.Join(vaultDir, ".forge"))
	n, err := vault.Load(filepath.Join(vaultDir, "note.md"), "note.md")
	if err != nil {
		t.Fatal(err)
	}
	note := Note{Rel: n.Rel, Verified: n.FM.Str("verified"),
		Refs: coderef.FromFrontmatter(n.Rel, n.FM.List("code_refs"))}
	st := OpenStore(filepath.Join(vaultDir, ".forge"))
	findings := Check([]Note{note}, registryFor(t, repo, to), src, gate(t, repo, since, to), Opts{})
	Apply(map[string]*vault.Note{n.Rel: n}, findings, st, schema(t), src)
	if err := st.Save(); err != nil {
		t.Fatal(err)
	}
	return reload(t, vaultDir).FM.Str("confidence")
}

func gate(t *testing.T, repo, since, to string) *Changed {
	t.Helper()
	files, err := coderef.ChangedFilesStatus(repo, since, to)
	if err != nil {
		t.Fatal(err)
	}
	g := &Changed{Touched: map[string]bool{}, Deleted: map[string]string{}}
	for _, f := range files {
		g.Touched[f.Path] = true
		if f.Deleted {
			g.Deleted[f.Path] = "app"
		}
	}
	return g
}

func registryFor(t *testing.T, repo, rev string) *coderef.Registry {
	t.Helper()
	r, err := coderef.ScanRepo("app", repo, rev)
	if err != nil {
		t.Fatal(err)
	}
	return coderef.NewRegistry([]coderef.Repo{r})
}

func reload(t *testing.T, vaultDir string) *vault.Note {
	t.Helper()
	n, err := vault.Load(filepath.Join(vaultDir, "note.md"), "note.md")
	if err != nil {
		t.Fatal(err)
	}
	return n
}

func schema(t *testing.T) *vault.Schema {
	t.Helper()
	s, err := vault.LoadSchema()
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func assertStored(t *testing.T, vaultDir, want string) {
	t.Helper()
	d, ok := OpenStore(filepath.Join(vaultDir, ".forge")).Take("placing-an-order")
	if !ok || d.Confidence != want {
		t.Fatalf("stored demotion = %+v (found=%v), want confidence %q", d, ok, want)
	}
}

// After a restore the record is gone: .forge/ holds restore targets, and a restored note
// has nothing left to be owed.
func assertCleared(t *testing.T, vaultDir string) {
	t.Helper()
	if _, ok := OpenStore(filepath.Join(vaultDir, ".forge")).Take("placing-an-order"); ok {
		t.Error("demotion record survived the restore")
	}
}

func assertLogCitesBoth(t *testing.T, vaultDir, broke, restored string) {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(vaultDir, ".forge", "drift.log"))
	if err != nil {
		t.Fatal(err)
	}
	for _, sha := range []string{short(broke), short(restored)} {
		if !strings.Contains(string(b), sha) {
			t.Errorf("drift.log does not cite %s:\n%s", sha, b)
		}
	}
}

func writeRepo(t *testing.T, root, src string) {
	t.Helper()
	dir := filepath.Join(root, "src", "main", "java")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "Order.java"), []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
}

func commit(t *testing.T, root, msg string) string {
	t.Helper()
	ensureGitRepo(t, root)
	git(t, root, "add", "-A")
	git(t, root, "commit", "-q", "-m", msg)
	return head(t, root)
}

// ensureGitRepo is commit and commitDated's (gitindex_test.go) shared one-time init.
func ensureGitRepo(t *testing.T, root string) {
	t.Helper()
	if _, err := os.Stat(filepath.Join(root, ".git")); os.IsNotExist(err) {
		git(t, root, "init", "-q")
		git(t, root, "config", "user.email", "t@example.com")
		git(t, root, "config", "user.name", "t")
	}
}

func head(t *testing.T, root string) string {
	t.Helper()
	sha, err := coderef.HeadSHA(root, "HEAD")
	if err != nil {
		t.Fatal(err)
	}
	return sha
}

func git(t *testing.T, root string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", root}, args...)...)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}
