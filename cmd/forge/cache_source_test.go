package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestCacheSourceWritesCacheFile is the happy path: a WebFetch PostToolUse payload with
// a string tool_response is cached verbatim under .forge/cache/<hash>.md.
func TestCacheSourceWritesCacheFile(t *testing.T) {
	root := t.TempDir()
	p := postToolUsePayload{
		ToolName:     "WebFetch",
		ToolInput:    []byte(`{"url":"https://example.com/x","prompt":"summarize"}`),
		ToolResponse: []byte(`"fetched markdown content"`),
	}

	cacheFetch(root, p)

	files := cacheFiles(t, root)
	if len(files) != 1 {
		t.Fatalf("want 1 cache file, got %d: %v", len(files), files)
	}
	body, _ := os.ReadFile(files[0])
	if !strings.Contains(string(body), "fetched markdown content") ||
		!strings.Contains(string(body), "https://example.com/x") {
		t.Errorf("cache file missing url/body: %s", body)
	}
}

// TestCacheSourceSkipsNonWebFetch pins the tool_name gate: only WebFetch is cached.
func TestCacheSourceSkipsNonWebFetch(t *testing.T) {
	oldStdin := setStdin(t, `{"tool_name":"Bash","tool_input":{},"tool_response":{}}`)
	defer oldStdin()
	root := t.TempDir()

	if code := cmdCacheSource([]string{"--vault", root}); code != 0 {
		t.Errorf("exit = %d, want 0", code)
	}
	if got := len(cacheFiles(t, root)); got != 0 {
		t.Errorf("want 0 cache files for non-WebFetch tool, got %d", got)
	}
}

// TestCmdCacheSourceAlwaysExitsZero pins the fail-silent contract shared by every hook
// subcommand: malformed stdin must never fail the session.
func TestCmdCacheSourceAlwaysExitsZero(t *testing.T) {
	oldStdin := setStdin(t, "not json")
	defer oldStdin()
	if code := cmdCacheSource([]string{"--vault", t.TempDir()}); code != 0 {
		t.Errorf("exit = %d, want 0 (fail-silent contract)", code)
	}
}

// TestCacheSourceObjectResponseCachedVerbatim covers the unconfirmed-schema fallback: a
// non-string tool_response (object shape) is cached as raw JSON rather than dropped.
func TestCacheSourceObjectResponseCachedVerbatim(t *testing.T) {
	root := t.TempDir()
	p := postToolUsePayload{
		ToolName:     "WebFetch",
		ToolInput:    []byte(`{"url":"https://example.com/y"}`),
		ToolResponse: []byte(`{"content":"nested","ok":true}`),
	}

	cacheFetch(root, p)

	files := cacheFiles(t, root)
	if len(files) != 1 {
		t.Fatalf("want 1 cache file, got %d", len(files))
	}
	body, _ := os.ReadFile(files[0])
	if !strings.Contains(string(body), `"content":"nested"`) {
		t.Errorf("cache file missing raw object body: %s", body)
	}
}

func cacheFiles(t *testing.T, root string) []string {
	t.Helper()
	matches, _ := filepath.Glob(filepath.Join(root, ".forge", "cache", "*.md"))
	return matches
}
