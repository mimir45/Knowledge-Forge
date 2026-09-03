package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/mimir45/Knowledge-Forge/pkg/vault"
)

// maxStubs caps how many stub notes one SessionEnd firing can write — the plan's own
// wording ("a small max, e.g. 3"), so a chatty session can't flood _inbox/.
const maxStubs = 3

const sessionCaptureUsage = `usage: forge session-capture [--vault DIR]

SessionEnd hook. Regex-scans the session transcript for conclusion sentences and writes
up to three low-confidence stub notes to _inbox/, deduped by session id + content hash.

stdin: a SessionEnd JSON payload. This command reads three fields:

    session_id        string   dedupe key, also stamped on each stub
    transcript_path   string   path to the transcript to scan; required
    reason            string   why the session ended

Writes nothing when transcript_path is absent or unreadable.
Fail-silent: the exit code is always 0.
`

// cmdSessionCapture is Phase 5's SessionEnd hook: a cheap, model-free scan of the
// session's transcript for "we established/decided/concluded that..." moments, written
// as capped, low-confidence stub notes via the same _inbox/ convention forge gate uses.
// Like every other hook subcommand here, it must never fail the session: any error just
// means no stubs get written, never a nonzero exit.
func cmdSessionCapture(args []string) int {
	if len(args) > 0 && (args[0] == "-h" || args[0] == "--help") {
		// stderr, not stdout: stdout is this hook's JSON output channel.
		fmt.Fprint(os.Stderr, sessionCaptureUsage)
		return 0
	}
	fs := flag.NewFlagSet("forge session-capture", flag.ContinueOnError)
	vaultDir := fs.String("vault", "", "vault root; defaults to config vault_path, then .")
	fs.SetOutput(io.Discard)
	if err := fs.Parse(args); err != nil {
		return 0
	}
	payload, err := readSessionEnd(os.Stdin)
	if err != nil || payload.TranscriptPath == "" {
		return 0
	}
	if root, err := resolveVault(*vaultDir); err == nil {
		captureSession(root, payload)
	}
	return 0
}

type sessionEndPayload struct {
	SessionID      string `json:"session_id"`
	TranscriptPath string `json:"transcript_path"`
	Reason         string `json:"reason"`
}

func readSessionEnd(r io.Reader) (sessionEndPayload, error) {
	var p sessionEndPayload
	err := json.NewDecoder(r).Decode(&p)
	return p, err
}

// captureSession finds conclusion sentences, dedupes against the session-id+content-hash
// store (so a resumed session that re-fires SessionEnd doesn't duplicate stubs — but the
// same conclusion restated in a *different* session is treated as new, since the key
// includes session id), and writes up to maxStubs.
func captureSession(root string, p sessionEndPayload) {
	texts, err := extractAssistantText(p.TranscriptPath)
	if err != nil {
		return
	}
	found := scanConclusions(texts)
	seen := loadSeen(root)
	written := 0
	for _, text := range found {
		if written >= maxStubs {
			break
		}
		key := p.SessionID + ":" + hashText(text)
		if seen[key] {
			continue
		}
		if writeStub(root, text) == nil {
			seen[key] = true
			written++
		}
	}
	if written > 0 {
		saveSeen(root, seen)
	}
}

// extractAssistantText reads a transcript JSONL file and returns every assistant text
// block, in order — the only role worth scanning for conclusions the assistant reached.
func extractAssistantText(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var texts []string
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 1<<20), 10<<20)
	for sc.Scan() {
		texts = append(texts, parseAssistantLine(sc.Bytes())...)
	}
	return texts, sc.Err()
}

type transcriptEntry struct {
	Message struct {
		Role    string `json:"role"`
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	} `json:"message"`
}

// parseAssistantLine extracts text blocks from one transcript JSONL line, nil for
// anything that isn't an assistant turn (user turns, tool results, summaries) — this is
// deliberate: raw transcript text can hold tool output and file contents, and the
// telemetry invariant against ever exporting those extends to what session-capture writes.
func parseAssistantLine(line []byte) []string {
	var e transcriptEntry
	if json.Unmarshal(line, &e) != nil || e.Message.Role != "assistant" {
		return nil
	}
	var out []string
	for _, c := range e.Message.Content {
		if c.Type == "text" {
			out = append(out, c.Text)
		}
	}
	return out
}

var conclusionRe = regexp.MustCompile(`(?i)\bwe (?:established|decided|concluded|agreed) that\b[^.\n]*[.\n]?`)

// scanConclusions finds "we established/decided/concluded/agreed that..." sentences
// across assistant turns — the cheap, regex-only signal the plan calls for, no model
// call and no attempt at real NLU.
func scanConclusions(texts []string) []string {
	var found []string
	for _, t := range texts {
		for _, m := range conclusionRe.FindAllString(t, -1) {
			found = append(found, strings.TrimSpace(m))
		}
	}
	return found
}

func hashText(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])[:16]
}

func seenPath(root string) string {
	return filepath.Join(root, ".forge", "session-capture-seen.json")
}

// loadSeen reads the dedup store; a missing or corrupt file just means an empty set —
// this is derived cache, not source of truth, so losing it costs duplicate stubs, not data.
func loadSeen(root string) map[string]bool {
	data, err := os.ReadFile(seenPath(root))
	if err != nil {
		return map[string]bool{}
	}
	var seen map[string]bool
	if json.Unmarshal(data, &seen) != nil || seen == nil {
		return map[string]bool{}
	}
	return seen
}

func saveSeen(root string, seen map[string]bool) {
	data, err := json.Marshal(seen)
	if err != nil {
		return
	}
	_ = os.MkdirAll(filepath.Join(root, ".forge"), 0o755)
	_ = os.WriteFile(seenPath(root), data, 0o644)
}

// writeStub builds a low-confidence draft note from one captured conclusion and hands it
// to WriteToInbox, which stamps confidence: low and writes it under _inbox/.
func writeStub(root, text string) error {
	s, err := vault.LoadSchema()
	if err != nil {
		return err
	}
	title := stubTitle(text)
	slug := vault.Slug(title) + "-" + hashText(text)[:6]
	fm, err := stubFrontmatter(title, slug)
	if err != nil {
		return err
	}
	n := &vault.Note{FM: fm, Body: []byte("\n" + text + "\n")}
	return vault.WriteToInbox(root, n, s, stubOpenQuestions())
}

// stubTitle turns a captured sentence into a schema-valid title (3-120 chars): trimmed,
// truncated, with a trailing period dropped.
func stubTitle(text string) string {
	t := strings.TrimSuffix(strings.TrimSpace(text), ".")
	if len(t) > 120 {
		t = truncateValidUTF8(t, 120)
	}
	if len(t) < 3 {
		t += "..."
	}
	return t
}

// truncateValidUTF8 cuts s to at most n bytes without splitting a multi-byte rune.
// validate.go's MaxLength check counts bytes (len(v)), so this stays under budget while
// never emitting invalid UTF-8 — a real risk here since captured conclusions can be
// non-English.
func truncateValidUTF8(s string, n int) string {
	for n > 0 && !utf8.RuneStart(s[n]) {
		n--
	}
	return strings.TrimSpace(s[:n])
}

// stubFrontmatter emits only the keys session-capture can honestly derive: title/slug
// from the captured text, type/depth/freshness_days at schema defaults (not a real
// classification), origin/dates/forge_version from the run itself. stack and tags are
// deliberately absent — inventing them would be worse than leaving them for a human,
// per stubOpenQuestions.
func stubFrontmatter(title, slug string) (*vault.Frontmatter, error) {
	now := time.Now().UTC().Format("2006-01-02")
	src := fmt.Sprintf(
		"title: %q\nslug: %s\ntype: concept\norigin: session-capture\n"+
			"created: %s\nupdated: %s\nverified: %s\ndepth: 3\nfreshness_days: 365\n"+
			"forge_version: %s\nsources: []\nrelated: []\nsupersedes: []\n",
		title, slug, now, now, now, vault.ForgeVersion)
	return vault.ParseFrontmatter([]byte(src))
}

// stubOpenQuestions names exactly what a session-capture stub can't honestly claim to
// know, so a reviewer sees what to fill in before this leaves _inbox/ — the same
// convention forge gate's own quarantine path uses for its own failing drafts.
func stubOpenQuestions() []string {
	return []string{
		"stack: not inferred from the transcript; fill in before publishing",
		"tags: not inferred from the transcript; fill in before publishing",
		"type: defaulted to concept, not classified; reclassify if this is really a pattern/howto/pitfall/decision/api/incident",
		"verified: stamped at capture time, not actually human-reviewed — treat as unverified",
		"sources: none captured; add a citation or confirm this is truly first-party",
	}
}
