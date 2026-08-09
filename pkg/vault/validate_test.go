package vault

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const goodNote = `---
title: "Kafka consumer group rebalancing"
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
`

func schema(t *testing.T) *Schema {
	t.Helper()
	s, err := LoadSchema()
	if err != nil {
		t.Fatalf("LoadSchema: %v", err)
	}
	return s
}

func noteFrom(t *testing.T, src string) *Note {
	t.Helper()
	p := filepath.Join(t.TempDir(), "n.md")
	if err := os.WriteFile(p, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	n, err := Load(p, "n.md")
	if err != nil {
		t.Fatal(err)
	}
	return n
}

func TestValidateAcceptsContractNote(t *testing.T) {
	if got := Validate(noteFrom(t, goodNote), schema(t)); len(got) != 0 {
		t.Errorf("a contract-shaped note produced %d issues: %v", len(got), got)
	}
}

func TestValidateRejections(t *testing.T) {
	cases := []struct {
		name, replace, with, code string
	}{
		{"bad type", "type: concept", "type: essay", "not-in-enum"},
		{"bad slug", "slug: kafka-consumer-group-rebalancing", "slug: Kafka_Group", "bad-format"},
		{"bad date", "created: 2026-08-07", "created: 07/08/2026", "bad-date"},
		{"depth out of range", "depth: 3", "depth: 9", "out-of-range"},
		{"unknown stack value", "stack: [java, spring-boot]", "stack: [cobol]", "not-in-vocabulary"},
		{"stack alias", "stack: [java, spring-boot]", "stack: [golang]", "alias"},
		{"tag alias", "tags: [messaging, consumer-group]", "tags: [configuration]", "alias"},
		{"tag uppercase", "tags: [messaging, consumer-group]", "tags: [Messaging]", "not-lowercase"},
		{"unknown key", "origin: ask", "origin: ask\nstatus: active", "unknown-key"},
		{"bad source kind", "kind: official", "kind: tweet", "not-in-enum"},
		{"bad forge_version", "forge_version: 2.0.0", "forge_version: v2", "bad-format"},
		{"bad origin", "origin: ask", "origin: dream", "not-in-enum"},
		{"updated before created", "updated: 2026-08-07", "updated: 2020-01-01", "date-order"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			src := strings.Replace(goodNote, c.replace, c.with, 1)
			if !hasCode(Validate(noteFrom(t, src), schema(t)), c.code) {
				t.Errorf("expected code %q; got %v", c.code, Validate(noteFrom(t, src), schema(t)))
			}
		})
	}
}

// TestValidateEngineTrailInvariant guards the T0 rule from CLAUDE.md: recall, write and
// index are static-core stages and may never record a model-backed engine.
func TestValidateEngineTrailInvariant(t *testing.T) {
	cases := []struct {
		trail string
		want  bool
	}{
		{`engine_trail: ["recall=none", "research=api"]`, false},
		{`engine_trail: ["recall=api"]`, true},
		{`engine_trail: ["write=host"]`, true},
		{`engine_trail: ["index=advisor"]`, true},
		{`engine_trail: ["critique=advisor"]`, false},
	}
	for _, c := range cases {
		src := strings.Replace(goodNote, "origin: ask", "origin: ask\n"+c.trail, 1)
		got := hasCode(Validate(noteFrom(t, src), schema(t)), "engine-invariant")
		if got != c.want {
			t.Errorf("%s: engine-invariant = %v, want %v", c.trail, got, c.want)
		}
	}
}

// TestValidateCitationFloor: decisions and incidents are first-party records and may
// cite nothing; every other type must cite at least one source.
func TestValidateCitationFloor(t *testing.T) {
	stripped := strings.Replace(goodNote, `sources:
  - url: https://kafka.apache.org/documentation/
    accessed: 2026-08-07
    kind: official`, "sources: []", 1)
	for _, tc := range []struct {
		noteType string
		want     bool
	}{
		{"concept", true}, {"howto", true}, {"decision", false}, {"incident", false},
	} {
		src := strings.Replace(stripped, "type: concept", "type: "+tc.noteType, 1)
		if got := hasCode(Validate(noteFrom(t, src), schema(t)), "uncited"); got != tc.want {
			t.Errorf("type %s: uncited = %v, want %v", tc.noteType, got, tc.want)
		}
	}
}

func TestValidateNoFrontmatter(t *testing.T) {
	got := Validate(noteFrom(t, "# Just a heading\n\nbody\n"), schema(t))
	if len(got) != 1 || got[0].Code != "no-frontmatter" || !got[0].Fixable {
		t.Errorf("got %v; want a single fixable no-frontmatter issue", got)
	}
}

func TestValidateKeyOrder(t *testing.T) {
	src := strings.Replace(goodNote, "title: \"Kafka consumer group rebalancing\"\nslug:",
		"slug:", 1)
	src = strings.Replace(src, "origin: ask", "origin: ask\ntitle: \"Kafka\"", 1)
	if !hasCode(Validate(noteFrom(t, src), schema(t)), "key-order") {
		t.Error("out-of-order frontmatter was accepted")
	}
}

func hasCode(issues []Issue, code string) bool {
	for _, i := range issues {
		if i.Code == code {
			return true
		}
	}
	return false
}
