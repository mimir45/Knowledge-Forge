package qualitygate

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mimir45/Knowledge-Forge/pkg/config"
	"github.com/mimir45/Knowledge-Forge/pkg/vault"
)

// goodNote mirrors pkg/vault/validate_test.go's fixture.
const goodNote = `---
title: "Kafka consumer group rebalancing"
slug: kafka-consumer-group-rebalancing
type: concept
stack: [java, spring-boot]
tags: [messaging, consumer-group]
depth: 3
confidence: high
created: 2026-08-07
updated: 2026-08-07
verified: 2026-08-07
freshness_days: 365
sources:
  - url: https://kafka.apache.org/documentation/
    accessed: 2026-08-07
    kind: official
related: ["[[kafka-partitions]]"]
supersedes: []
forge_version: 2.0.0
origin: ask
---

# Kafka consumer group rebalancing
`

// noteFrom writes src under a fresh temp dir and loads it as a draft at the given
// vault-relative path.
func noteFrom(t *testing.T, src, rel string) *vault.Note {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, "draft.md")
	if err := os.WriteFile(p, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	n, err := vault.Load(p, rel)
	if err != nil {
		t.Fatal(err)
	}
	return n
}

func testSchema(t *testing.T) *vault.Schema {
	t.Helper()
	s, err := vault.LoadSchema()
	if err != nil {
		t.Fatalf("LoadSchema: %v", err)
	}
	return s
}

// testConfig loads the packaged defaults in isolation from the real environment.
func testConfig(t *testing.T) *config.Config {
	t.Helper()
	c, err := config.Load(config.Options{ProjectDir: t.TempDir(), HomeDir: t.TempDir()})
	if err != nil {
		t.Fatalf("config.Load: %v", err)
	}
	return c
}

// emptyVault makes a temp vault root with the directories the gates walk, so
// vault.Walk succeeds with zero notes instead of erroring on a missing directory.
func emptyVault(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "notes", "concept"), 0o755); err != nil {
		t.Fatal(err)
	}
	return root
}
