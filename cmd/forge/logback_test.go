package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mimir45/Knowledge-Forge/pkg/codeindex"
	"github.com/mimir45/Knowledge-Forge/pkg/config"
)

const orderJava = `package com.food.order;

public class Order {
    public void place(String id) {
        repo.save(id);
    }
}
`

const orderNote = `---
title: Placing an order
slug: placing-an-order
type: concept
confidence: high
created: 2026-01-01
updated: 2026-01-01
verified: 2026-01-01
code_refs: [app:src/main/java/com/food/order/Order.java#Order.place]
---

Order placement goes through Order.place.
`

// fullLogback turns knowledge_map, claude_md_fragment and inline_markers all on.
func fullLogback() *config.Config {
	return &config.Config{Static: config.Static{LogBack: config.LogBack{
		KnowledgeMap: true, ClaudeMDFragment: true, InlineMarkers: true,
	}}}
}

func TestLogbackWritesKnowledgeMapAndClaudeFragmentAndIsIdempotent(t *testing.T) {
	repo, vaultDir := t.TempDir(), t.TempDir()
	writeJavaRepo(t, repo, orderJava)
	commitAll(t, repo, "add Order")
	writeNote(t, vaultDir, "note.md", orderNote)

	cfg := logbackCfg{vault: vaultDir, repos: repoList{{Name: "app", Root: repo}}}
	if code := runLogback(cfg, fullLogback()); code != 0 {
		t.Fatalf("runLogback = %d, want 0", code)
	}

	mapPath := filepath.Join(repo, "docs", "knowledge-map.md")
	mapBefore := readFile(t, mapPath)
	if !strings.Contains(mapBefore, "[[placing-an-order]]") {
		t.Fatalf("knowledge-map.md missing note link:\n%s", mapBefore)
	}

	claudePath := filepath.Join(repo, "src/main/java/com/food/order", "CLAUDE.md")
	claudeBefore := readFile(t, claudePath)
	if !strings.Contains(claudeBefore, "[[placing-an-order]]") {
		t.Fatalf("CLAUDE.md fragment missing note link:\n%s", claudeBefore)
	}

	if code := runLogback(cfg, fullLogback()); code != 0 {
		t.Fatalf("second runLogback = %d, want 0", code)
	}
	if got := readFile(t, mapPath); got != mapBefore {
		t.Errorf("knowledge-map.md changed on a no-op rerun:\nbefore:\n%s\nafter:\n%s", mapBefore, got)
	}
	if got := readFile(t, claudePath); got != claudeBefore {
		t.Errorf("CLAUDE.md changed on a no-op rerun:\nbefore:\n%s\nafter:\n%s", claudeBefore, got)
	}

	if codeindex.Available() {
		javaBefore := readFile(t, filepath.Join(repo, "src/main/java/com/food/order/Order.java"))
		if !strings.Contains(javaBefore, "// forge:logback:") {
			t.Fatalf("Order.java missing an inline marker:\n%s", javaBefore)
		}
		if strings.Contains(javaBefore, "\n\n\n") {
			t.Errorf("inline marker left a blank uncommented gap:\n%s", javaBefore)
		}
	}
}

func TestLogbackRemoveMarkersRevertsTheSourceFile(t *testing.T) {
	if !codeindex.Available() {
		t.Skip("built without cgo: no symbol table to anchor a marker to")
	}
	repo, vaultDir := t.TempDir(), t.TempDir()
	writeJavaRepo(t, repo, orderJava)
	commitAll(t, repo, "add Order")
	writeNote(t, vaultDir, "note.md", orderNote)

	javaPath := filepath.Join(repo, "src/main/java/com/food/order/Order.java")
	before := readFile(t, javaPath)

	cfg := logbackCfg{vault: vaultDir, repos: repoList{{Name: "app", Root: repo}}}
	if code := runLogback(cfg, fullLogback()); code != 0 {
		t.Fatalf("runLogback = %d, want 0", code)
	}
	if readFile(t, javaPath) == before {
		t.Fatal("marker run left Order.java unchanged; nothing to revert")
	}

	cfg.removeMarkers = true
	if code := runLogback(cfg, fullLogback()); code != 0 {
		t.Fatalf("runLogback --remove-markers = %d, want 0", code)
	}
	if got := readFile(t, javaPath); got != before {
		t.Errorf("Order.java after --remove-markers = %q, want the original %q", got, before)
	}
}

func TestLogbackConfigGatesEachStepIndependently(t *testing.T) {
	repo, vaultDir := t.TempDir(), t.TempDir()
	writeJavaRepo(t, repo, orderJava)
	commitAll(t, repo, "add Order")
	writeNote(t, vaultDir, "note.md", orderNote)

	cfg := logbackCfg{vault: vaultDir, repos: repoList{{Name: "app", Root: repo}}}
	off := &config.Config{} // minimal.md's shape: every logback key false
	if code := runLogback(cfg, off); code != 0 {
		t.Fatalf("runLogback = %d, want 0", code)
	}
	if _, err := os.Stat(filepath.Join(repo, "docs", "knowledge-map.md")); !os.IsNotExist(err) {
		t.Error("knowledge-map.md written despite knowledge_map: false")
	}
	if _, err := os.Stat(filepath.Join(repo, "src/main/java/com/food/order", "CLAUDE.md")); !os.IsNotExist(err) {
		t.Error("CLAUDE.md written despite claude_md_fragment: false")
	}
}

func writeJavaRepo(t *testing.T, root, src string) {
	t.Helper()
	dir := filepath.Join(root, "src", "main", "java", "com", "food", "order")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "Order.java"), []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
}

func writeNote(t *testing.T, vaultDir, name, src string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(vaultDir, name), []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
}

func commitAll(t *testing.T, root, msg string) {
	t.Helper()
	if _, err := os.Stat(filepath.Join(root, ".git")); os.IsNotExist(err) {
		gitLB(t, root, "init", "-q")
		gitLB(t, root, "config", "user.email", "t@example.com")
		gitLB(t, root, "config", "user.name", "t")
	}
	gitLB(t, root, "add", "-A")
	gitLB(t, root, "commit", "-q", "-m", msg)
}

func gitLB(t *testing.T, root string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", root}, args...)...)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}
