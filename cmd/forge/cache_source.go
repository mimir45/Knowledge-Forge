package main

import (
	"encoding/json"
	"errors"
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

const cacheSourceUsage = `usage: forge cache-source [--vault DIR]

PostToolUse hook, for WebFetch only. Caches the fetched source under
<vault>/.forge/cache/<url-hash>.md with a TTL (static.cache_ttl_days, default 30).

stdin: a PostToolUse JSON payload. This command reads three fields:

    tool_name       string   must be "WebFetch"; anything else is ignored
    tool_input      object   the tool call; its "url" field is the cache key
    tool_response   object   the fetched content

Note this writes into the *configured* vault regardless of which project the session is
working in — pass --vault to scope it.
Fail-silent: the exit code is always 0.
`

// cmdCacheSource is Phase 5's PostToolUse hook for WebFetch: caches the fetched result
// under .forge/cache/<hash-of-url>.md so a repeated fetch of the same URL within the TTL
// window can later be short-circuited (nothing reads/enforces the TTL yet in Phase 5 —
// this command only writes the cache).
//
// PostToolUse's tool_response shape for WebFetch was confirmed by capturing a live hook
// payload, since it was never documented officially: it is an
// object {result, url, code, codeText, bytes, durationMs}, and result carries the actual
// fetched/summarized text. cacheBody extracts result when present; see its doc comment
// for the fallback chain that still applies to any other shape.
func cmdCacheSource(args []string) int {
	fs := flag.NewFlagSet("forge cache-source", flag.ContinueOnError)
	vaultDir := fs.String("vault", "", "vault root; defaults to config vault_path, then .")
	fs.SetOutput(io.Discard)
	// flag's own error path must stay silent (fail-silent contract), so Usage is stubbed
	// and an explicit -h/--help is handled below instead — in any flag position.
	fs.Usage = func() {
		{
		}
	}
	if err := fs.Parse(args); err != nil {
		{
			if errors.Is(err, flag.ErrHelp) {
				{
					// stderr, not stdout: stdout is this hook's JSON output channel.
					fmt.Fprint(os.Stderr, cacheSourceUsage)
					fs.SetOutput(os.Stderr)
					fs.PrintDefaults()
				}
			}
			return 0
		}
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

// cacheBody extracts the fetched text from tool_response. The confirmed shape (captured
// from a live hook payload) is an object carrying the text under a "result" key; that
// key is looked up by name via a map decode, not a fixed struct, so a present-but-empty result
// ("" — a real, if unlikely, fetch outcome) is still returned rather than falling through.
// Two fallbacks cover anything else without asserting a schema this can't confirm: a bare
// JSON string (kept in case the shape ever simplifies) unmarshals cleanly and is used
// as-is; any other shape (object with no "result" key, array, non-JSON) is cached as the
// raw bytes Claude Code sent, verbatim — never silently dropped, never guessed.
func cacheBody(raw json.RawMessage) string {
	var obj map[string]json.RawMessage
	if json.Unmarshal(raw, &obj) == nil {
		if result, ok := obj["result"]; ok {
			var s string
			if json.Unmarshal(result, &s) == nil {
				return s
			}
		}
	}
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
