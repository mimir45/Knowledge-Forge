package recall

import (
	"math"
	"testing"
)

func near(t *testing.T, name string, got, want float64) {
	t.Helper()
	if math.Abs(got-want) > 0.001 {
		t.Errorf("%s = %.4f, want %.4f", name, got, want)
	}
}

// The two cases f2's doc comment cites as the reason it is not pure coverage and not
// symmetric Dice. Pinning them means a future tweak to the measure has to explain
// itself against the vault behaviour that motivated this one.
func TestF2AnchorCases(t *testing.T) {
	// "how does spring boot work" -> Terms{boot, spring}; slug
	// spring-boot-4-breaking-changes-artifact-renames-and-test-module-split -> 9
	// tokens ("4" and "and" drop out). Over-broad: coverage alone calls this 1.000.
	near(t, "over-broad title", f2(2, 2, 9), 0.588)
	// "how does keyset pagination work" -> Terms{keyset, pagination}; slug
	// keyset-pagination-compound-or-predicate -> 4 tokens. More specific than the
	// question, and must not be punished for it — symmetric Dice scored it 0.67.
	near(t, "more-specific title", f2(2, 2, 4), 0.833)
}

func TestF2Degenerate(t *testing.T) {
	for _, c := range [][3]int{{0, 2, 3}, {2, 0, 3}, {2, 3, 0}} {
		if got := f2(c[0], c[1], c[2]); got != 0 {
			t.Errorf("f2%v = %v, want 0", c, got)
		}
	}
}

// blend divides by the active weight, not by the constant 1.0. This is spec §2.5: a
// perfect title match with no tag or stack input must reach the answer threshold.
func TestBlendRenormalizesOverActiveChannels(t *testing.T) {
	chs := []Channel{
		{Name: "title", Weight: wTitle, Value: 1.0, Active: true},
		{Name: "tags", Weight: wTags, Active: false},
		{Name: "stack", Weight: wStack, Active: false},
		{Name: "body", Weight: wBody, Value: 0.9, Active: true},
	}
	score, matched := blend(chs)
	near(t, "blend", score, 0.98) // (0.4 + 0.09) / 0.5 — a raw sum gives 0.49
	if len(matched) != 2 || matched[0] != "title" || matched[1] != "body" {
		t.Errorf("matched_on = %v, want [title body]", matched)
	}
}

// An active channel that scored zero stays out of matched_on but stays in the
// denominator: the query asked and the note failed to answer.
func TestBlendActiveZeroChannelCountsAgainst(t *testing.T) {
	chs := []Channel{
		{Name: "title", Weight: wTitle, Value: 1.0, Active: true},
		{Name: "tags", Weight: wTags, Value: 0, Active: true},
	}
	score, matched := blend(chs)
	near(t, "blend", score, 0.571) // 0.4 / 0.7
	if len(matched) != 1 || matched[0] != "title" {
		t.Errorf("matched_on = %v, want [title]", matched)
	}
}

func TestBlendNoActiveChannels(t *testing.T) {
	if score, _ := blend([]Channel{{Name: "tags", Weight: wTags}}); score != 0 {
		t.Errorf("blend = %v, want 0", score)
	}
}

// Two-sided activation: a channel needs input from the query *and* the field on the
// note. 31 of the real vault's 91 notes have no tags after the Phase 1 migration, and
// scoring them zero ranked them below well-tagged irrelevant notes.
func TestTagsChannelTwoSidedActivation(t *testing.T) {
	docs := []Doc{{Slug: "a", Tags: []string{"pagination"}}, {Slug: "b"}}
	s := newScope(Query{Question: "how does pagination work"}, docs)

	if c := s.tagsChannel(docs[0]); !c.Active || c.Value != 1 {
		t.Errorf("tagged note: active=%v value=%v, want true/1", c.Active, c.Value)
	}
	if c := s.tagsChannel(docs[1]); c.Active {
		t.Error("untagged note: channel active, want inactive")
	}
	// Query term matches no tag anywhere in the vault: nothing to compare against.
	s2 := newScope(Query{Question: "how do goroutines work"}, docs)
	if c := s2.tagsChannel(docs[0]); c.Active {
		t.Error("query outside the tag vocabulary: channel active, want inactive")
	}
}

func TestStackChannelActivation(t *testing.T) {
	docs := []Doc{{Slug: "a", Stack: []string{"spring-boot", "kafka"}}, {Slug: "b"}}
	q := Query{Question: "how does retry work", Stack: []string{"spring-boot"}}
	s := newScope(q, docs)

	if c := s.stackChannel(docs[0]); !c.Active || c.Value != 1 {
		t.Errorf("superset stack: active=%v value=%v, want true/1", c.Active, c.Value)
	}
	if c := s.stackChannel(docs[1]); c.Active {
		t.Error("note without stack: channel active, want inactive")
	}
	if c := newScope(Query{Question: "x"}, docs).stackChannel(docs[0]); c.Active {
		t.Error("no stack hints and no stack terms: channel active, want inactive")
	}
}

// The title channel folds the note side through the stopword filter too. Without it
// "hexagonal-architecture-ports-and-adapters" carries an "and" that no question can
// match, diluting f2's precision half — it ranked 3rd at 0.411 on the real vault.
func TestTitleChannelDropsNoteSideStopwords(t *testing.T) {
	d := Doc{Title: "Hexagonal Architecture", Slug: "hexagonal-architecture-ports-and-adapters"}
	c := newScope(Query{Question: "what is hexagonal architecture"}, nil).titleChannel(d)
	if len(c.Hits) != 2 {
		t.Fatalf("hits = %v, want [architecture hexagonal]", c.Hits)
	}
	// Note-side tokens: hexagonal, architecture, ports, adapters — "and" is gone.
	near(t, "title", c.Value, f2(2, 2, 4))
}

// Body term hits saturate at three occurrences so one word repeated cannot stand in
// for coverage of the whole question.
func TestBodyChannelSaturates(t *testing.T) {
	s := newScope(Query{Question: "what is pagination and keyset"}, nil)
	flooded := s.bodyChannel([]byte("pagination pagination pagination pagination pagination"))
	near(t, "one term flooded", flooded.Value, 0.5)
	both := s.bodyChannel([]byte("pagination pagination pagination keyset keyset keyset"))
	near(t, "both terms", both.Value, 1.0)
}

func TestTermsDropsStopwordsAndSorts(t *testing.T) {
	got := Terms("How does the Transactional Outbox pattern work?")
	want := []string{"outbox", "pattern", "transactional"}
	if len(got) != len(want) {
		t.Fatalf("Terms = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("Terms = %v, want %v", got, want)
		}
	}
}

// A tag like "spring-boot" has to be reachable from the query terms "spring" and
// "boot", so the note side goes through the same tokenizer as the question.
func TestSetOfSplitsHyphenatedValues(t *testing.T) {
	s := setOf([]string{"spring-boot"})
	if !s["spring"] || !s["boot"] {
		t.Errorf("setOf = %v, want spring and boot", s)
	}
}
