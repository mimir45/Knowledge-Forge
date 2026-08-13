package main

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"knowledge-forge/pkg/report"
)

// TestLoadAskLogMissingFileIsEmpty pins the tolerant-of-absence contract loadAskLog's
// doc comment promises: every vault before its first telemetry-enabled run has no
// .forge/log.jsonl at all, and that must not be an error.
func TestLoadAskLogMissingFileIsEmpty(t *testing.T) {
	bySlug, asks := loadAskLog(filepath.Join(t.TempDir(), "log.jsonl"), map[string]string{})
	if len(bySlug) != 0 || len(asks) != 0 {
		t.Fatalf("missing file: got bySlug=%v asks=%v, want empty", bySlug, asks)
	}
}

// TestLoadAskLogSplitsWrittenFromGaps: a topic matching a real note's slug counts toward
// staleness.md's Asks map; a topic matching nothing is a gaps.md candidate instead.
func TestLoadAskLogSplitsWrittenFromGaps(t *testing.T) {
	root := fixtureCopy(t)
	notes, err := loadNotes(root)
	if err != nil {
		t.Fatalf("loadNotes: %v", err)
	}
	slugs := slugMap(notes)
	var known string
	for _, s := range slugs {
		known = s
		break
	}
	path := writeAskLog(t, root, known, "totally-unknown-topic")
	bySlug, asks := loadAskLog(path, slugs)
	if bySlug[known] != 3 {
		t.Errorf("bySlug[%q] = %d, want 3", known, bySlug[known])
	}
	assertAsk(t, asks, known, 3, true)
	assertAsk(t, asks, "totally-unknown-topic", 2, false)
}

// writeAskLog writes known asked 3x and unknown asked 2x as raw "ask" event lines.
func writeAskLog(t *testing.T, root, known, unknown string) string {
	t.Helper()
	dir := filepath.Join(root, ".forge")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	var body string
	for i := 0; i < 3; i++ {
		body += fmt.Sprintf(`{"event":"ask","topic":%q}`+"\n", known)
	}
	for i := 0; i < 2; i++ {
		body += fmt.Sprintf(`{"event":"ask","topic":%q}`+"\n", unknown)
	}
	path := filepath.Join(dir, "log.jsonl")
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatalf("write log: %v", err)
	}
	return path
}

func assertAsk(t *testing.T, asks []report.Ask, topic string, count int, written bool) {
	t.Helper()
	for _, a := range asks {
		if a.Topic != topic {
			continue
		}
		if a.Count != count || a.Written != written {
			t.Errorf("ask %q = {%d %v}, want {%d %v}", topic, a.Count, a.Written, count, written)
		}
		return
	}
	t.Errorf("no ask entry for topic %q", topic)
}
