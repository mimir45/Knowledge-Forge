package scrub

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mimir45/Knowledge-Forge/pkg/vault"
)

// A small in-repo fixture, deliberately separate from testdata/vault/ (that fixture's
// F1-F12 defects are the surface for other packages' tests, not this one's — CLAUDE.md
// says not to touch it). Plants one of each secret shape D-6 requires: an email, an
// absolute home path, and an API-key-shaped token — one in frontmatter, one in body.
const secretNote = `---
title: "Debugging samir.alizade@example.com's Kafka setup"
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

Config lives at /Users/samir/dev/kafka/config.yaml. The broker's API key is
sk-testtoken1234567890ABCDEFghijkl, planted here only for this fixture.
`

const noFMNote = "Config at /Users/samir/notes/scratch.md, contact samir@example.com.\n"

func writeFixture(t *testing.T, dir string) {
	t.Helper()
	notesDir := filepath.Join(dir, "notes", "concept")
	if err := os.MkdirAll(notesDir, 0o755); err != nil {
		t.Fatal(err)
	}
	write := func(rel, body string) {
		p := filepath.Join(dir, filepath.FromSlash(rel))
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("notes/concept/kafka-consumer-group-rebalancing.md", secretNote)
	write("notes/concept/no-frontmatter.md", noFMNote)
}

func TestScrubRedactsPlantedSecrets(t *testing.T) {
	src, dst := t.TempDir(), filepath.Join(t.TempDir(), "out")
	writeFixture(t, src)

	rep, err := Scrub(src, dst)
	if err != nil {
		t.Fatalf("Scrub: %v", err)
	}
	if rep.NotesTotal != 2 || rep.NotesWritten != 2 {
		t.Fatalf("report = %+v, want 2 notes total/written", rep)
	}
	if rep.Redactions == 0 {
		t.Fatalf("report.Redactions = 0, want > 0")
	}

	assertScrubbed(t, dst)
}

// assertScrubbed checks every planted secret is gone from the written output and
// nothing else was silently dropped alongside it.
func assertScrubbed(t *testing.T, dst string) {
	t.Helper()
	out := readAll(t, filepath.Join(dst, "notes/concept/kafka-consumer-group-rebalancing.md"))
	for _, secret := range []string{
		"samir.alizade@example.com", "/Users/samir/dev/kafka/config.yaml",
		"sk-testtoken1234567890ABCDEFghijkl",
	} {
		if strings.Contains(out, secret) {
			t.Errorf("scrubbed note still contains secret %q", secret)
		}
	}
	for _, marker := range []string{"[REDACTED-EMAIL]", "[REDACTED-PATH]", "[REDACTED-KEY]"} {
		if !strings.Contains(out, marker) {
			t.Errorf("scrubbed note missing expected marker %q:\n%s", marker, out)
		}
	}

	noFM := readAll(t, filepath.Join(dst, "notes/concept/no-frontmatter.md"))
	if strings.Contains(noFM, "samir@example.com") || strings.Contains(noFM, "/Users/samir/notes") {
		t.Errorf("no-frontmatter note still contains a secret:\n%s", noFM)
	}
}

func TestScrubOutputStillValidatesAgainstSchema(t *testing.T) {
	src, dst := t.TempDir(), filepath.Join(t.TempDir(), "out")
	writeFixture(t, src)
	if _, err := Scrub(src, dst); err != nil {
		t.Fatalf("Scrub: %v", err)
	}

	rel := "notes/concept/kafka-consumer-group-rebalancing.md"
	n, err := vault.Load(filepath.Join(dst, rel), rel)
	if err != nil {
		t.Fatalf("Load scrubbed note: %v", err)
	}
	s, err := vault.LoadSchema()
	if err != nil {
		t.Fatalf("LoadSchema: %v", err)
	}
	if issues := vault.Validate(n, s); len(issues) != 0 {
		t.Errorf("scrubbed note fails schema: %v", issues)
	}
}

func TestScrubIsDeterministic(t *testing.T) {
	src := t.TempDir()
	writeFixture(t, src)
	dst1, dst2 := filepath.Join(t.TempDir(), "a"), filepath.Join(t.TempDir(), "b")

	if _, err := Scrub(src, dst1); err != nil {
		t.Fatalf("Scrub (run 1): %v", err)
	}
	if _, err := Scrub(src, dst2); err != nil {
		t.Fatalf("Scrub (run 2): %v", err)
	}

	rel := "notes/concept/kafka-consumer-group-rebalancing.md"
	a, b := readAll(t, filepath.Join(dst1, rel)), readAll(t, filepath.Join(dst2, rel))
	if a != b {
		t.Errorf("scrub output not byte-identical across runs:\n--- run 1 ---\n%s\n--- run 2 ---\n%s", a, b)
	}
}

// TestScrubFailsClosed checks that a note whose frontmatter cannot be parsed aborts the
// whole run with nothing written to dst — the D-6 requirement, checked directly rather
// than only implied by scrubOne's doc comment.
func TestScrubFailsClosed(t *testing.T) {
	src, dst := t.TempDir(), filepath.Join(t.TempDir(), "out")
	writeFixture(t, src)
	bad := "---\ntitle: [unterminated\n---\nbody\n"
	p := filepath.Join(src, "notes/concept/bad.md")
	if err := os.WriteFile(p, []byte(bad), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := Scrub(src, dst); err == nil {
		t.Fatal("Scrub with an unparseable note: got nil error, want failure")
	}
	if _, err := os.Stat(dst); !os.IsNotExist(err) {
		t.Errorf("Scrub failed but dst = %s exists (want untouched)", dst)
	}
}

// slugNote is D-6's missing negative case: reLongToken must not treat an ordinary
// kebab-case slug or a dated filename as a secret. Found via a real-vault dry run —
// "2026-04-13-local-ai-continue-rag-spring" (40 chars, all hyphens/lowercase) was
// getting redacted, corrupting a legitimate sources: citation. One real secret stays
// planted alongside so this test also proves the fix didn't disable redaction.
const slugNote = `---
title: "Debugging a rebalance"
slug: kafka-consumer-group-rebalancing
type: concept
stack: [java]
tags: [messaging]
depth: 3
confidence: high
created: 2026-08-07
updated: 2026-08-07
verified: 2026-08-07
freshness_days: 365
sources:
  - url: sources/daily/2026-04-13-local-ai-continue-rag-spring.md
    accessed: 2026-08-07
    kind: personal
related: ["[[a-really-long-descriptive-slug-that-exceeds-thirty-two-characters]]"]
supersedes: []
forge_version: 2.0.0
origin: ask
---

# Rebalance

See the-consumer-group-rebalancing-strategy-for-multi-tenant-kafka-clusters for
background. Contact samir@example.com about it.
`

// TestScrubDoesNotRedactSlugsOrFilenames is the negative case D-6's fixture test
// lacked: real secrets still get caught, but hyphenated slugs/filenames pass through.
func TestScrubDoesNotRedactSlugsOrFilenames(t *testing.T) {
	src, dst := t.TempDir(), filepath.Join(t.TempDir(), "out")
	notesDir := filepath.Join(src, "notes", "concept")
	if err := os.MkdirAll(notesDir, 0o755); err != nil {
		t.Fatal(err)
	}
	p := filepath.Join(notesDir, "kafka-consumer-group-rebalancing.md")
	if err := os.WriteFile(p, []byte(slugNote), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := Scrub(src, dst); err != nil {
		t.Fatalf("Scrub: %v", err)
	}
	out := readAll(t, filepath.Join(dst, "notes/concept/kafka-consumer-group-rebalancing.md"))
	for _, untouched := range []string{
		"sources/daily/2026-04-13-local-ai-continue-rag-spring.md",
		"a-really-long-descriptive-slug-that-exceeds-thirty-two-characters",
		"the-consumer-group-rebalancing-strategy-for-multi-tenant-kafka-clusters",
	} {
		if !strings.Contains(out, untouched) {
			t.Errorf("scrubbed note lost non-secret text %q:\n%s", untouched, out)
		}
	}
	if strings.Contains(out, "samir@example.com") || !strings.Contains(out, "[REDACTED-EMAIL]") {
		t.Errorf("planted email was not redacted:\n%s", out)
	}
}

// codeNote is the second false-positive class reLongToken had, found spot-checking
// the slug fix's own output: a camelCase Java method name in a code sample is a long
// unbroken alphanumeric run with no digits, so it matched even after the dash/underscore
// exclusion. A digit-bearing JWT-shaped token stays planted alongside so this test also
// proves the digit requirement didn't disable real detection.
const codeNote = `---
title: "Saga outbox lookup"
slug: saga-outbox-lookup
type: howto
stack: [java, spring-boot]
tags: [messaging]
depth: 3
confidence: high
created: 2026-08-07
updated: 2026-08-07
verified: 2026-08-07
freshness_days: 365
sources: []
related: []
supersedes: []
forge_version: 2.0.0
origin: ask
---

# Saga outbox lookup

` + "```java" + `
paymentOutboxHelper.getPaymentOutboxMessageBySagaIdAndSagaStatus(sagaId, status);
` + "```" + `

The demo JWT header is eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9, planted for this fixture.
`

// TestScrubDoesNotRedactCamelCaseCodeIdentifiers is the negative case for the second
// reLongToken false-positive class: pure-alphabetic camelCase identifiers in code
// samples pass through untouched, while a digit-bearing token-shaped string is still
// caught.
func TestScrubDoesNotRedactCamelCaseCodeIdentifiers(t *testing.T) {
	src, dst := t.TempDir(), filepath.Join(t.TempDir(), "out")
	notesDir := filepath.Join(src, "notes", "howto")
	if err := os.MkdirAll(notesDir, 0o755); err != nil {
		t.Fatal(err)
	}
	p := filepath.Join(notesDir, "saga-outbox-lookup.md")
	if err := os.WriteFile(p, []byte(codeNote), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := Scrub(src, dst); err != nil {
		t.Fatalf("Scrub: %v", err)
	}
	out := readAll(t, filepath.Join(dst, "notes/howto/saga-outbox-lookup.md"))
	if !strings.Contains(out, "getPaymentOutboxMessageBySagaIdAndSagaStatus") {
		t.Errorf("scrubbed note lost a camelCase code identifier:\n%s", out)
	}
	if strings.Contains(out, "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9") || !strings.Contains(out, "[REDACTED-KEY]") {
		t.Errorf("planted digit-bearing token was not redacted:\n%s", out)
	}
}

func readAll(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(b)
}
