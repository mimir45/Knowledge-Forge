package main

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestSessionContextPrintsIndexAndProfile is the happy path: both files exist, both
// appear in the output, separated by the "---" the SessionStart transcript needs to
// tell the two sections apart.
func TestSessionContextPrintsIndexAndProfile(t *testing.T) {
	root := fixtureCopy(t)
	runIndex(root, "_index.md", 4096, false)
	writeProfile(t, root, "# Me\n\nprimary_language: go\n")

	out := captureStdout(t, func() { printSessionContext(root, 4096) })

	if !strings.Contains(out, "# Vault index") {
		t.Errorf("missing index content: %.80s", out)
	}
	if !strings.Contains(out, "primary_language: go") {
		t.Errorf("missing profile content: %.80s", out)
	}
}

// TestSessionContextSkipsMissingProfileSilently: no profile ever exists until `forge
// init` runs (plan's known gap — no per-project profile format yet). The index must
// still print, and the miss goes to the log, never to stdout/stderr.
func TestSessionContextSkipsMissingProfileSilently(t *testing.T) {
	root := fixtureCopy(t)
	runIndex(root, "_index.md", 4096, false)

	out := captureStdout(t, func() { printSessionContext(root, 4096) })

	if !strings.Contains(out, "# Vault index") {
		t.Errorf("index missing even though profile is absent: %.80s", out)
	}
	if _, err := os.Stat(filepath.Join(root, ".forge", "session-context.log")); err != nil {
		t.Errorf("expected a log entry for the missing profile: %v", err)
	}
}

// TestCmdSessionContextAlwaysExitsZero pins the fail-silent contract itself: even a
// vault that does not exist at all must not fail the session.
func TestCmdSessionContextAlwaysExitsZero(t *testing.T) {
	root := filepath.Join(t.TempDir(), "missing-vault")
	if code := cmdSessionContext([]string{"--vault", root}); code != 0 {
		t.Errorf("exit = %d, want 0 (fail-silent contract)", code)
	}
}

func writeProfile(t *testing.T, root, body string) {
	t.Helper()
	dir := filepath.Join(root, "profiles")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "me.md"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

// captureStdout redirects os.Stdout for the duration of fn — printSessionContext writes
// straight to it, matching what a real SessionStart hook does.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w
	fn()
	w.Close()
	os.Stdout = old
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatal(err)
	}
	return buf.String()
}
