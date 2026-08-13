package main

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

// TestIntentEmitsHitAboveThreshold pins the plan's >0.7 contract: a prompt that nearly
// echoes an existing note's title scores 0.99 against the fixture (measured via
// `forge recall --explain`), so it must surface as additionalContext.
func TestIntentEmitsHitAboveThreshold(t *testing.T) {
	root := fixtureCopy(t)
	runIndex(root, "_index.md", 4096, false)

	out := captureStdout(t, func() {
		printIntent(root, "Hibernate ORM Patterns and Gotchas")
	})

	var got struct {
		AdditionalContext string `json:"additionalContext"`
		Continue          bool   `json:"continue"`
	}
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("output not valid JSON: %v (%q)", err, out)
	}
	if !strings.Contains(got.AdditionalContext, "Hibernate") || !got.Continue {
		t.Errorf("unexpected hit: %+v", got)
	}
}

// TestIntentSilentBelowThreshold: a topic the fixture has nothing on must print nothing
// — no additionalContext, no accidental prompt interference.
func TestIntentSilentBelowThreshold(t *testing.T) {
	root := fixtureCopy(t)
	runIndex(root, "_index.md", 4096, false)

	out := captureStdout(t, func() {
		printIntent(root, "quantum entanglement in distributed rate limiters")
	})
	if out != "" {
		t.Errorf("expected silence below 0.7, got %q", out)
	}
}

// TestCmdIntentAlwaysExitsZero pins the fail-silent contract: malformed stdin, a
// missing vault, anything at all — never a nonzero exit, never a blocked prompt.
func TestCmdIntentAlwaysExitsZero(t *testing.T) {
	oldStdin := setStdin(t, "not json")
	defer oldStdin()
	if code := cmdIntent([]string{"--vault", t.TempDir()}); code != 0 {
		t.Errorf("exit = %d, want 0 (fail-silent contract)", code)
	}
}

// setStdin redirects os.Stdin to body for the duration of a test, returning a restore
// func the caller defers.
func setStdin(t *testing.T, body string) func() {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	go func() { w.WriteString(body); w.Close() }()
	old := os.Stdin
	os.Stdin = r
	return func() { os.Stdin = old }
}
