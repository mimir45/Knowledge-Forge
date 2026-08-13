package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestSessionCaptureWritesStub is the happy path: a transcript with one conclusion
// sentence yields exactly one _inbox/ stub carrying that sentence in its body.
func TestSessionCaptureWritesStub(t *testing.T) {
	root := fixtureCopy(t)
	tp := writeTranscript(t, root, "We decided that MinHash beats embeddings here.")

	captureSession(root, sessionEndPayload{SessionID: "s1", TranscriptPath: tp})

	files := inboxFiles(t, root)
	if len(files) != 1 {
		t.Fatalf("want 1 stub, got %d: %v", len(files), files)
	}
	body, _ := os.ReadFile(files[0])
	if !strings.Contains(string(body), "MinHash beats embeddings") {
		t.Errorf("stub missing conclusion text: %s", body)
	}
}

// TestSessionCaptureDedupesOnRefire pins the resume-safety contract: SessionEnd can
// re-fire for the same session (--continue/--resume), so a second call against the same
// transcript and session id must not duplicate the stub.
func TestSessionCaptureDedupesOnRefire(t *testing.T) {
	root := fixtureCopy(t)
	tp := writeTranscript(t, root, "We concluded that drift must stay git-anchored.")
	p := sessionEndPayload{SessionID: "s1", TranscriptPath: tp}

	captureSession(root, p)
	captureSession(root, p)

	if got := len(inboxFiles(t, root)); got != 1 {
		t.Errorf("want 1 stub after two fires, got %d", got)
	}
}

// TestCmdSessionCaptureAlwaysExitsZero pins the fail-silent contract shared by every
// hook subcommand: malformed stdin must never fail the session.
func TestCmdSessionCaptureAlwaysExitsZero(t *testing.T) {
	oldStdin := setStdin(t, "not json")
	defer oldStdin()
	if code := cmdSessionCapture([]string{"--vault", t.TempDir()}); code != 0 {
		t.Errorf("exit = %d, want 0 (fail-silent contract)", code)
	}
}

// writeTranscript writes a one-line JSONL transcript with a single assistant text turn.
func writeTranscript(t *testing.T, root, text string) string {
	t.Helper()
	entry := map[string]any{
		"message": map[string]any{
			"role":    "assistant",
			"content": []map[string]string{{"type": "text", "text": text}},
		},
	}
	line, err := json.Marshal(entry)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "transcript.jsonl")
	if err := os.WriteFile(path, append(line, '\n'), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func inboxFiles(t *testing.T, root string) []string {
	t.Helper()
	matches, _ := filepath.Glob(filepath.Join(root, "_inbox", "*.md"))
	return matches
}
