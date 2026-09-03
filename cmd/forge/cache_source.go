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

// cmdCacheSource is Phase 5's PostToolUse hook for WebFetch.
func cmdCacheSource(args []string) int {
	fs := flag.NewFlagSet("forge cache-source", flag.ContinueOnError)
	vaultDir := fs.String("vault", "", "vault root; defaults to config vault_path, then .")
	fs.SetOutput(io.Discard)
	// flag's own error path must stay silent (fail-silent contract), so Usage is stubbed
	// and an explicit -h/--help is handled below instead — in any flag position.
	fs.Usage = func() {}
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			// stderr, not stdout: stdout is this hook's JSON output channel.
			fmt.Fprint(os.Stderr, cacheSourceUsage)
			fs.SetOutput(os.Stderr)
			fs.PrintDefaults()
		}
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

// cacheFetch writes .forge/cache/<hash-of-url>.md.
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

// cacheBody extracts the fetched text from tool_response.
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

// cacheTTLDays reads Static.CacheTTLDays.
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
