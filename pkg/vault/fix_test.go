package vault

import (
	"strings"
	"testing"
)

// legacyNote is the shape the real vault is actually in: v1 keys (date, status), tags in
// alias and mixed case, no contract keys at all.
const legacyNote = `---
title: Testcontainers Docker socket
date: 2026-04-13
status: active
tags: [Docker, testcontainers, configuration]
source: daily/2026-04-13.md
---

# Testcontainers Docker socket

Body text that must survive byte-for-byte.

- bullet
-  bullet with odd spacing
`

func fixed(t *testing.T, src string) (string, []string) {
	t.Helper()
	out, changes, err := Fix(noteFrom(t, src), schema(t))
	if err != nil {
		t.Fatalf("Fix: %v", err)
	}
	return string(out), changes
}

// TestFixNeverTouchesBody is the safety property the migration rests on.
func TestFixNeverTouchesBody(t *testing.T) {
	for _, src := range []string{legacyNote, goodNote, "# No frontmatter\n\nbody\n"} {
		_, wantBody, _ := SplitFrontmatter([]byte(src))
		out, _ := fixed(t, src)
		_, gotBody, err := SplitFrontmatter([]byte(out))
		if err != nil {
			t.Fatalf("fixed output has no frontmatter: %v", err)
		}
		if string(gotBody) != string(wantBody) {
			t.Errorf("body changed:\n got %q\nwant %q", gotBody, wantBody)
		}
	}
}

func TestFixNormalizesTags(t *testing.T) {
	out, _ := fixed(t, legacyNote)
	if !strings.Contains(out, "tags: [docker, testcontainers, config]") {
		t.Errorf("tags not lowercased and de-aliased:\n%s", out)
	}
}

// TestFixDropsRetiredKeys: `status` and `date` are not in the contract. render() emits
// only schema keys, so they disappear rather than lingering as unknown-key issues.
func TestFixDropsRetiredKeys(t *testing.T) {
	out, _ := fixed(t, legacyNote)
	fmSrc, _, err := SplitFrontmatter([]byte(out))
	if err != nil {
		t.Fatal(err)
	}
	for _, k := range []string{"status:", "date:", "source:"} {
		if strings.Contains(string(fmSrc), k) {
			t.Errorf("retired key %q survived:\n%s", k, fmSrc)
		}
	}
}

// TestFixCarriesLegacySourceIntoSources: `source:` is retired, but on 63 of the real
// vault's 93 notes it is the only provenance there is. Dropping it unconverted would
// destroy the citation and fail the note on the §12 gate, so the value must survive the
// retirement rather than be deleted with the key.
func TestFixCarriesLegacySourceIntoSources(t *testing.T) {
	out, changes := fixed(t, legacyNote)
	for _, want := range []string{
		"url: daily/2026-04-13.md",
		"accessed: 2026-04-13", // the note's own date, not today
		"kind: session",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in converted sources:\n%s", want, out)
		}
	}
	if strings.Contains(out, "sources: []") {
		t.Errorf("sources was emitted empty despite a source: value:\n%s", out)
	}
	if !strings.Contains(strings.Join(changes, "\n"), "sources: <- source:") {
		t.Errorf("conversion not reported in changes: %v", changes)
	}
}

// TestFixDoesNotOverwriteRealSources: a note that already cites something keeps it; the
// legacy key is only a fallback, never an override.
func TestFixDoesNotOverwriteRealSources(t *testing.T) {
	src := "---\ntitle: T\ncreated: 2026-01-01\nsource: legacy.md\n" +
		"sources:\n  - url: https://example.test/a\n    accessed: 2026-01-01\n    kind: official\n---\n\nbody\n"
	out, _ := fixed(t, src)
	if strings.Contains(out, "legacy.md") {
		t.Errorf("legacy source overwrote a real citation:\n%s", out)
	}
	if !strings.Contains(out, "https://example.test/a") {
		t.Errorf("existing citation lost:\n%s", out)
	}
}

// TestFixBackfillsDatesFromLegacyDate: the v1 `date:` key is the honest origin date, so
// it seeds created/updated/verified rather than the file's mtime.
func TestFixBackfillsDatesFromLegacyDate(t *testing.T) {
	out, _ := fixed(t, legacyNote)
	for _, k := range []string{"created", "updated", "verified"} {
		if !strings.Contains(out, k+": 2026-04-13") {
			t.Errorf("%s not seeded from date: 2026-04-13:\n%s", k, out)
		}
	}
}

// TestFixEmitsUnquotedScalars: yaml.v3 would quote everything if the encoder were told
// the tag; dates and ints must read as dates and ints in a file a human edits.
func TestFixEmitsUnquotedScalars(t *testing.T) {
	out, _ := fixed(t, legacyNote)
	for _, want := range []string{"depth: 3", "created: 2026-04-13", "forge_version: 2.0.0"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in:\n%s", want, out)
		}
	}
}

// TestFixDoesNotInventMeaning: type, stack and tags are judgement calls. --fix leaves
// them absent so Validate keeps reporting them.
func TestFixDoesNotInventMeaning(t *testing.T) {
	out, _ := fixed(t, "# Orphan note\n\nbody\n")
	for _, k := range []string{"type:", "stack:", "tags:"} {
		if strings.Contains(out, k) {
			t.Errorf("--fix invented %q:\n%s", k, out)
		}
	}
	// Title and slug are derived from the H1, not invented, so they must appear.
	if !strings.Contains(out, "title: Orphan note") || !strings.Contains(out, "slug: orphan-note") {
		t.Errorf("title/slug not derived from the H1:\n%s", out)
	}
}

func TestFixNormalizesKeyOrder(t *testing.T) {
	scrambled := "---\norigin: import\ntype: concept\ntitle: X note\n---\n\nbody\n"
	out, _ := fixed(t, scrambled)
	if strings.Index(out, "title:") > strings.Index(out, "type:") ||
		strings.Index(out, "type:") > strings.Index(out, "origin:") {
		t.Errorf("keys not in schema order:\n%s", out)
	}
}

// TestFixIsIdempotent: running --fix twice must produce the identical file, or every
// migration rerun would churn mtimes and invalidate the index cache.
func TestFixIsIdempotent(t *testing.T) {
	for _, src := range []string{legacyNote, goodNote, "# Bare\n\nbody\n"} {
		once, _ := fixed(t, src)
		twice, changes := fixed(t, once)
		if once != twice {
			t.Errorf("second pass differs:\n--- first ---\n%s\n--- second ---\n%s", once, twice)
		}
		if len(changes) != 0 {
			t.Errorf("second pass still reported changes: %v", changes)
		}
	}
}

// TestFixResolvesFixableIssues: everything Validate marks [--fix] must actually be gone
// after a fix pass. What remains must be judgement calls only.
func TestFixResolvesFixableIssues(t *testing.T) {
	out, _ := fixed(t, legacyNote)
	for _, is := range Validate(noteFrom(t, out), schema(t)) {
		if is.Fixable {
			t.Errorf("fixable issue survived --fix: %v", is)
		}
	}
}
