package codeindex

import (
	"path/filepath"
	"testing"
)

// TestLoadRejectsOldExtractor pins the cache-versioning guarantee, re-checked here
// after a later change bumped Extractor for the first time since it was written.
func TestLoadRejectsOldExtractor(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "code-index-app.json")
	stale := Index{Repo: "app", Commit: "abc123", Extractor: Extractor - 1,
		Files: map[string]File{"A.java": {Path: "A.java"}}}
	if err := Save(path, stale); err != nil {
		t.Fatal(err)
	}
	if _, ok := Load(path); ok {
		t.Fatal("Load accepted a cache stamped by an older extractor")
	}
}

func TestLoadAcceptsCurrentExtractor(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "code-index-app.json")
	fresh := Index{Repo: "app", Commit: "abc123", Extractor: Extractor,
		Files: map[string]File{"A.java": {Path: "A.java"}}}
	if err := Save(path, fresh); err != nil {
		t.Fatal(err)
	}
	ix, ok := Load(path)
	if !ok {
		t.Fatal("Load rejected a cache stamped by the current extractor")
	}
	if ix.Commit != "abc123" {
		t.Errorf("Commit = %q, want abc123", ix.Commit)
	}
}
