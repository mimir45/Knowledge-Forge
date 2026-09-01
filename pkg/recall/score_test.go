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

// The question is phrased in terms the vault's stack vocabulary carries, which is
// load-bearing: an absent question term sits in the denominator and would
// pull a superset match below 1.000 for a reason that has nothing to do with supersets.
// That effect has its own test below; this one is about activation and containment.
func TestStackChannelActivation(t *testing.T) {
	docs := []Doc{{Slug: "a", Stack: []string{"spring-boot", "kafka"}}, {Slug: "b"}}
	q := Query{Question: "how does kafka work", Stack: []string{"spring-boot"}}
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

// Before IDF weighting, "Redis caching in Spring Boot" scored 0.740 against a Spring CLI
// note: every query term counted the same, so two tags half the vault carries stood in
// for the one that carried the meaning. The shape is reproduced here in miniature.
func TestTagsChannelWeightsRareTermsHigher(t *testing.T) {
	var docs []Doc
	for i := 0; i < 9; i++ {
		docs = append(docs, Doc{Slug: "common", Tags: []string{"spring"}})
	}
	docs = append(docs, Doc{Slug: "rare", Tags: []string{"spring", "redis"}})
	s := newScope(Query{Question: "redis caching in spring"}, docs)

	// df(spring) = 10 of 10 -> log(2) = 0.693; df(redis) = 1 -> log(11) = 2.398, and
	// "caching" is in the denominator too, at the mean of those two
	// (1.546), which is why both values sit below their pre-admission figures — the note
	// is being asked to account for a term the vault tags nowhere, and cannot.
	common := s.tagsChannel(Doc{Tags: []string{"spring"}})
	rare := s.tagsChannel(Doc{Tags: []string{"redis"}})
	near(t, "vault-wide tag only", common.Value, 0.149) // 0.500 flat, then 0.224 weighted
	near(t, "discriminating tag only", rare.Value, 0.517)
	if common.Value >= rare.Value {
		t.Errorf("common tag %.3f >= rare tag %.3f", common.Value, rare.Value)
	}
}

// --stack accepts anything; the vault may never have seen it. A hint is a user filter and
// not evidence, so an unknown hint term must not move the score — narrowing a search by
// "kotlin" in a vault with no Kotlin cannot make every note match less well.
//
// Stated as a comparison, deliberately, rather than as an absolute (value == 1.000) —
// that is no longer what "undiluted" means under IDF weighting: the question's own
// absent terms do count against a note now. Only the hint side is asserted here.
//
// This is why the vocabulary filter changed sides rather than being deleted. Question
// terms are evidence and go in unfiltered; hints are filtered, because once absent terms
// carry weight an unfiltered hint would be a real regression.
func TestStackChannelIgnoresTermsNoNoteCarries(t *testing.T) {
	docs := []Doc{{Slug: "a", Stack: []string{"java"}}}
	base := newScope(Query{Question: "how does retry work",
		Stack: []string{"java"}}, docs).stackChannel(docs[0])
	withUnknown := newScope(Query{Question: "how does retry work",
		Stack: []string{"java", "kotlin"}}, docs).stackChannel(docs[0])
	if !base.Active || withUnknown.Active != base.Active || withUnknown.Value != base.Value {
		t.Errorf("unknown hint moved the score: %.4f -> %.4f", base.Value, withUnknown.Value)
	}
	unknown := Query{Question: "how does retry work", Stack: []string{"kotlin"}}
	if c := newScope(unknown, docs).stackChannel(docs[0]); c.Active {
		t.Error("stack hint outside the vault: channel active, want inactive")
	}
}

// Every note carries the tag, so log(N/df) would be exactly zero and the denominator
// would vanish on an active channel. The smoothed form bottoms out at log(2).
func TestTagsChannelUniversalTermStillScores(t *testing.T) {
	docs := []Doc{{Tags: []string{"java"}}, {Tags: []string{"java"}}}
	c := newScope(Query{Question: "how does java work"}, docs).tagsChannel(docs[0])
	if !c.Active || c.Value != 1 {
		t.Errorf("universal tag: active=%v value=%v, want true/1", c.Active, c.Value)
	}
}

// The cap is a guard, not the fix: it holds the spread between the rarest and the
// commonest term at about 5:1 however large the vault grows.
func TestIDFCapAndDegenerateCases(t *testing.T) {
	near(t, "hapax in a large vault", idf(1, 10000), idfCap)
	near(t, "universal term", idf(500, 500), math.Log(2))
	if idf(0, 91) != 0 {
		t.Error("a term no note carries must weigh nothing")
	}
	if r := idf(1, 10000) / idf(10000, 10000); r > 5.1 {
		t.Errorf("weight spread %.2f, want <= 5.1", r)
	}
}

// A question term no note carries stays in the channel's denominator:
// the vault holding nothing about "redis" is evidence about the vault, not an absence of
// evidence. Before, inVocab dropped it before any weight was computed, so the tags channel
// read 1.000 off the single term the note happened to share and a Spring CLI note answered
// a Redis question at 0.740.
func TestTagsChannelCountsTermsNoNoteCarries(t *testing.T) {
	docs := []Doc{{Slug: "cli", Tags: []string{"spring-cli"}}}
	// Terms are {caching, redis, spring}; setOf folds the tag to {spring, cli}, so only
	// "spring" is present. Every weight is then equal — the present set has one member and
	// the absent terms take its mean — and the ratio is a plain one-of-three.
	c := newScope(Query{Question: "redis caching in spring"}, docs).tagsChannel(docs[0])
	near(t, "one term of three", c.Value, 1.0/3.0)
}

// The absent term's weight is the mean of the present ones. The alternative considered
// — flooring document frequency at 1 — would hand it the largest weight any term can have
// and invert idfCap's purpose, letting a term the vault has never seen outweigh every term
// it has. A mean of capped values is also still capped, so the guard survives.
func TestAbsentTermWeighsThePresentMean(t *testing.T) {
	df := map[string]int{"java": 2, "spring": 1}
	w := weightsOver([]string{"java", "spring", "redis"}, df, 3)
	near(t, "present mean", w["redis"], (w["java"]+w["spring"])/2)
	if w["redis"] >= w["spring"] {
		t.Errorf("absent term %.3f outweighs the rarest present term %.3f",
			w["redis"], w["spring"])
	}
	if w["redis"] > idfCap {
		t.Errorf("absent term %.3f exceeds the cap %.3f", w["redis"], idfCap)
	}
}

// Before this fix, activation asked only whether the note carries the field
// (len(tags) > 0), so a tagged note with zero overlap paid the absent-term penalty in
// full while an untagged note skipped the channel entirely — the note that carries
// nothing relevant was worse off than the note that carries nothing at all. weighted's
// own comment already argues the fix: "scoring such a query 0.0 on an active channel
// would drag every note down uniformly, the same mistake spec §2.5 rejects for untagged
// notes." Both notes here have zero hits and must be treated alike.
func TestTagsChannelZeroOverlapMatchesUntagged(t *testing.T) {
	docs := []Doc{
		{Slug: "untagged"},
		{Slug: "irrelevant-tag", Tags: []string{"issue"}},
		// Gives the query vocabulary a real hit somewhere in the corpus, so ok=true and
		// the other two notes' inactivity is a note-side decision, not a global one —
		// without this the query has nothing any tag channel could answer at all, and
		// both would be inactive regardless of this fix.
		{Slug: "relevant-tag", Tags: []string{"spring"}},
	}
	s := newScope(Query{Question: "redis caching in spring boot"}, docs)
	if c := s.tagsChannel(docs[0]); c.Active {
		t.Error("untagged note: channel active, want inactive")
	}
	if c := s.tagsChannel(docs[1]); c.Active {
		t.Error("tagged-but-zero-overlap note: channel active, want inactive (same as untagged)")
	}
	if c := s.tagsChannel(docs[2]); !c.Active {
		t.Error("tagged note with a real hit: channel inactive, want active")
	}
}

// The degenerate case falls out of the mean rule rather than being special-cased: with no
// present term there is nothing to average, every weight stays 0, and weighted's empty
// denominator leaves the channel inactive. Pinned deliberately, because the alternative is
// the mistake spec §2.5 rejects — a query the vault's vocabulary cannot speak to at all
// activating the channel and scoring every note 0.0 on it.
func TestChannelInactiveWhenNoQueryTermIsCarried(t *testing.T) {
	if w := weightsOver([]string{"goroutines"}, map[string]int{"pagination": 1}, 1); w["goroutines"] != 0 {
		t.Errorf("weight without a present term = %v, want 0", w["goroutines"])
	}
	docs := []Doc{{Slug: "a", Tags: []string{"pagination"}}}
	if c := newScope(Query{Question: "how do goroutines work"}, docs).tagsChannel(docs[0]); c.Active {
		t.Error("query outside the tag vocabulary: channel active, want inactive")
	}
}
