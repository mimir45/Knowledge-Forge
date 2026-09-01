package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mimir45/Knowledge-Forge/pkg/codeindex"
	"github.com/mimir45/Knowledge-Forge/pkg/coderef"
	"github.com/mimir45/Knowledge-Forge/pkg/gitsig"
	"github.com/mimir45/Knowledge-Forge/pkg/report"
)

// TestDependsOnResolvesJavaClassImport covers the ordinary case: one file imports a
// class declared in a sibling package, and the two packages' directories are the modules
// CodeGroup groups by.
func TestDependsOnResolvesJavaClassImport(t *testing.T) {
	ix := codeindex.Index{Files: map[string]codeindex.File{
		"src/main/java/com/food/order/Consumer.java": {Lang: "java",
			Imports: []string{"com.food.repo.Repo"}},
	}}
	files := []string{
		"src/main/java/com/food/order/Consumer.java",
		"src/main/java/com/food/repo/Repo.java",
	}
	got := dependsOn(ix, files)
	want := []string{"src/main/java/com/food/repo"}
	assertDeps(t, got, "src/main/java/com/food/order", want)
}

// TestDependsOnResolvesJavaWildcardAndStaticImports covers the two shapes the extractor
// hands the resolver besides a plain class import: a wildcard (already star-stripped by
// codeindex, arrives as a bare package name) and a static member import (one segment
// past the class).
func TestDependsOnResolvesJavaWildcardAndStaticImports(t *testing.T) {
	ix := codeindex.Index{Files: map[string]codeindex.File{
		"src/main/java/com/food/order/Consumer.java": {Lang: "java",
			Imports: []string{"com.food.repo", "com.food.util.Const.MAX"}},
	}}
	files := []string{
		"src/main/java/com/food/order/Consumer.java",
		"src/main/java/com/food/repo/Repo.java",
		"src/main/java/com/food/util/Const.java",
	}
	got := dependsOn(ix, files)
	want := []string{"src/main/java/com/food/repo", "src/main/java/com/food/util"}
	assertDeps(t, got, "src/main/java/com/food/order", want)
}

// TestDependsOnResolvesRelativeTSImports covers both TypeScript shapes: a specific
// sibling-directory file and a subdirectory's barrel index.
func TestDependsOnResolvesRelativeTSImports(t *testing.T) {
	ix := codeindex.Index{Files: map[string]codeindex.File{
		"src/pages/LoginPage.tsx": {Lang: "typescript",
			Imports: []string{"../widgets/Button", "../hooks"}},
	}}
	files := []string{
		"src/pages/LoginPage.tsx", "src/widgets/Button.tsx", "src/hooks/index.ts",
	}
	got := dependsOn(ix, files)
	want := []string{"src/hooks", "src/widgets"}
	assertDeps(t, got, "src/pages", want)
}

// TestDependsOnDropsUnresolvableAndSelfImports is the plan's explicit "do not invent a
// node" case: a third-party Java package, a bare TS specifier, and a same-directory
// import must all leave DependsOn empty rather than guessing.
func TestDependsOnDropsUnresolvableAndSelfImports(t *testing.T) {
	ix := codeindex.Index{Files: map[string]codeindex.File{
		"src/main/java/com/food/order/Consumer.java": {Lang: "java",
			Imports: []string{"org.springframework.stereotype.Service"}},
		"src/pages/LoginPage.tsx": {Lang: "typescript",
			Imports: []string{"react", "./LoginForm"}},
	}}
	files := []string{
		"src/main/java/com/food/order/Consumer.java", "src/pages/LoginPage.tsx",
		"src/pages/LoginForm.tsx",
	}
	got := dependsOn(ix, files)
	if len(got["src/main/java/com/food/order"]) != 0 {
		t.Errorf("java deps = %v, want none (third-party import)", got["src/main/java/com/food/order"])
	}
	if len(got["src/pages"]) != 0 {
		t.Errorf("ts deps = %v, want none (bare specifier + same-directory import)", got["src/pages"])
	}
}

// TestDependsOnIsDedupedAndSorted: two files in the same group importing the same or
// different dependencies must fold to one deduplicated, sorted slice — moc/codebase.md
// has to render identically across runs (2b's determinism standard).
func TestDependsOnIsDedupedAndSorted(t *testing.T) {
	ix := codeindex.Index{Files: map[string]codeindex.File{
		"src/pages/A.tsx": {Lang: "typescript", Imports: []string{"../widgets/Button"}},
		"src/pages/B.tsx": {Lang: "typescript", Imports: []string{"../widgets/Button", "../hooks"}},
	}}
	files := []string{
		"src/pages/A.tsx", "src/pages/B.tsx", "src/widgets/Button.tsx", "src/hooks/index.ts",
	}
	got := dependsOn(ix, files)
	want := []string{"src/hooks", "src/widgets"}
	assertDeps(t, got, "src/pages", want)
}

// TestDependsOnIsDeterministicAcrossRuns guards against a recurring class of bug
// elsewhere in this codebase: Go's map iteration order is randomized per run, so a slice built
// from map keys without a sort is nondeterministic even though every individual value is
// "correct". Enough files share a directory here that an unsorted build would show it.
func TestDependsOnIsDeterministicAcrossRuns(t *testing.T) {
	ix := codeindex.Index{Files: map[string]codeindex.File{
		"src/pages/A.tsx": {Lang: "typescript", Imports: []string{
			"../widgets/Button", "../widgets/Card", "../hooks", "../hooks/useAuth"}},
	}}
	files := []string{
		"src/pages/A.tsx", "src/widgets/Button.tsx", "src/widgets/Card.tsx",
		"src/hooks/index.ts", "src/hooks/useAuth.ts",
	}
	first := dependsOn(ix, files)["src/pages"]
	for i := 0; i < 20; i++ {
		if got := dependsOn(ix, files)["src/pages"]; !slicesEq(got, first) {
			t.Fatalf("run %d: %v, want %v (same as run 0)", i, got, first)
		}
	}
}

func assertDeps(t *testing.T, got map[string][]string, dir string, want []string) {
	t.Helper()
	if !slicesEq(got[dir], want) {
		t.Fatalf("dependsOn[%q] = %v, want %v", dir, got[dir], want)
	}
}

func slicesEq(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// TestGroupsOfPopulatesDependsOnEndToEnd builds a real temp git repo — two Java packages
// and a TypeScript pages/widgets pair — and drives the actual Parse -> Build -> dependsOn
// -> groupsOf pipeline, not a hand-built Index. This is the one test in the suite that
// would fail if the tree-sitter grammar's import node shapes ever changed underneath the
// hand-built-Index tests above.
func TestGroupsOfPopulatesDependsOnEndToEnd(t *testing.T) {
	if !codeindex.Available() {
		t.Skip("built without cgo: no symbol table to build")
	}
	root := t.TempDir()
	writeSrc(t, root, "src/main/java/com/food/repo/Repo.java",
		"package com.food.repo;\npublic class Repo {\n    public void save(String id) {}\n}\n")
	writeSrc(t, root, "src/main/java/com/food/order/Consumer.java",
		"package com.food.order;\nimport com.food.repo.Repo;\npublic class Consumer {\n"+
			"    public void receive(String id) {}\n}\n")
	writeSrc(t, root, "src/widgets/Button.tsx", "export function Button() { return null; }\n")
	writeSrc(t, root, "src/pages/LoginPage.tsx",
		"import { Button } from '../widgets/Button';\nexport function LoginPage() { return Button(); }\n")
	commitAll(t, root, "seed")

	scanned, err := coderef.ScanRepo("app", root, "HEAD")
	if err != nil {
		t.Fatal(err)
	}
	ix, err := codeindex.Build("app", root, "HEAD", scanned.Files)
	if err != nil {
		t.Fatal(err)
	}
	groups := groupsOf(scanned.Files, gitsig.Analyze(nil), nil, dependsOn(ix, scanned.Files))
	order := findGroup(t, groups, "src/main/java/com/food/order")
	if !slicesEq(order.DependsOn, []string{"src/main/java/com/food/repo"}) {
		t.Errorf("order.DependsOn = %v, want [src/main/java/com/food/repo]", order.DependsOn)
	}
	pages := findGroup(t, groups, "src/pages")
	if !slicesEq(pages.DependsOn, []string{"src/widgets"}) {
		t.Errorf("pages.DependsOn = %v, want [src/widgets]", pages.DependsOn)
	}
}

func writeSrc(t *testing.T, root, rel, src string) {
	t.Helper()
	full := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(full, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
}

func findGroup(t *testing.T, groups []report.CodeGroup, name string) report.CodeGroup {
	t.Helper()
	for _, g := range groups {
		if g.Name == name {
			return g
		}
	}
	t.Fatalf("no group named %q in %+v", name, groups)
	return report.CodeGroup{}
}
