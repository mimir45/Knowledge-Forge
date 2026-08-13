package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

// defaultCacheTTLDays is applied when Static.CacheTTLDays is unset (zero) — matches the
// config field's own doc comment: zero means unset, not "expire immediately".
const defaultCacheTTLDays = 30

// cmdCacheSource is Phase 5's PostToolUse hook for WebFetch: caches the fetched result
// under .forge/cache/<hash-of-url>.md so a repeated fetch of the same URL within the TTL
// window can later be short-circuited (nothing reads/enforces the TTL yet in Phase 5 —
// this command only writes the cache).
//
// PostToolUse's exact tool_response JSON shape for WebFetch is not documented in Claude
// Code's hooks/tools-reference docs as of 2026-08-13 (checked via WebFetch itself,
// inconclusive both times — see docs/BACKLOG.md). tool_input.url is the one field known
// with confidence, since it's WebFetch's own published parameter schema. tool_response is
// therefore cached verbatim rather than field-extracted: see cacheBody's doc comment.
func cmdCacheSource(args []string) int {
	fs := flag.NewFlagSet("forge cache-source", flag.ContinueOnError)
	vaultDir := fs.String("vault", "", "vault root; defaults to config vault_path, then .")
	fs.SetOutput(io.Discard)
	if err := fs.Parse(args); err != nil {
		return 0
	}
	payload, err := readPostToolUse(os.Stdin)
	if err != nil || payload.ToolName != "WebFetch" {
		return 0
	}
	if root, err := resolveVault(*vaultDir); err == nil {
		cacheFetch(root, payload)
	}
	return 0
}

type postToolUsePayload struct {
	ToolName     string          `json:"tool_name"`
	ToolInput    json.RawMessage `json:"tool_input"`
	ToolResponse json.RawMessage `json:"tool_response"`
}

func readPostToolUse(r io.Reader) (postToolUsePayload, error) {
	var p postToolUsePayload
	err := json.NewDecoder(r).Decode(&p)
	return p, err
}

// cacheFetch writes .forge/cache/<hash-of-url>.md. Any failure (bad tool_input, no url,
// write error) is a silent no-op — a missed cache write just costs a re-fetch later,
// never a broken session. .forge/ is in pkg/vault's skipDirs, so this file is never seen
// by Walk/validate/index — it cannot leak into the note contract.
func cacheFetch(root string, p postToolUsePayload) {
	var in struct {
		URL string `json:"url"`
	}
	if json.Unmarshal(p.ToolInput, &in) != nil || in.URL == "" {
		return
	}
	content := renderCacheEntry(in.URL, cacheTTLDays(), cacheBody(p.ToolResponse))
	path := filepath.Join(root, ".forge", "cache", hashText(in.URL)+".md")
	_ = os.MkdirAll(filepath.Dir(path), 0o755)
	_ = os.WriteFile(path, []byte(content), 0o644)
}

// cacheBody extracts a cacheable string from tool_response without asserting a schema it
// cannot confirm: a plain JSON string unmarshals cleanly and is used as-is; anything else
// (object, array, non-JSON) is cached as the raw bytes Claude Code sent, verbatim.
func cacheBody(raw json.RawMessage) string {
	var s string
	if json.Unmarshal(raw, &s) == nil {
		return s
	}
	return string(raw)
}

// cacheTTLDays reads Static.CacheTTLDays, falling back to defaultCacheTTLDays on an
// unset (zero) value or a config load error — a hook must never fail the session over a
// bad config chain.
func cacheTTLDays() int {
	cfg, err := loadConfig()
	if err != nil || cfg.Static.CacheTTLDays == 0 {
		return defaultCacheTTLDays
	}
	return cfg.Static.CacheTTLDays
}

func renderCacheEntry(url string, ttlDays int, body string) string {
	fetched := time.Now().UTC().Format("2006-01-02")
	return fmt.Sprintf("url: %s\nfetched: %s\nttl_days: %d\n\n---\n\n%s\n",
		url, fetched, ttlDays, body)
}
