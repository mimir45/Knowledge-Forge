package recall

import (
	"testing"
	"time"
)

var now = time.Date(2026, 8, 9, 0, 0, 0, 0, time.UTC)

func TestIsStale(t *testing.T) {
	cases := []struct {
		name          string
		verified, upd string
		days          int
		want          bool
	}{
		{"fresh", "2026-08-01", "2026-08-01", 180, false},
		{"aged past window", "2025-01-01", "2025-01-01", 180, true},
		// A typo fix bumps `updated` without anyone re-checking the claims. Reading
		// `updated` first is how a vault quietly starts lying.
		{"recent edit does not refresh an old verification", "2025-01-01", "2026-08-08", 180, true},
		{"falls back to updated when unverified", "", "2026-08-01", 180, false},
		{"undatable is stale", "", "", 180, true},
		{"freshness 0 never goes stale", "2015-01-01", "2015-01-01", 0, false},
		{"freshness 0 outranks undatable", "", "", 0, false},
	}
	for _, c := range cases {
		d := Doc{Verified: c.verified, Updated: c.upd, FreshnessDays: c.days}
		if got := IsStale(d, now); got != c.want {
			t.Errorf("%s: IsStale = %v, want %v", c.name, got, c.want)
		}
	}
}

func TestDecideAtThresholdBoundaries(t *testing.T) {
	cases := []struct {
		score float64
		stale bool
		want  Decision
	}{
		{0.85, false, AnswerFromVault}, // inclusive lower bound
		{0.85, true, UpdateRefresh},    // same score, stale -> re-verify instead
		{1.0, false, AnswerFromVault},
		{0.849, false, UpdateExtend},
		{0.55, false, UpdateExtend}, // inclusive lower bound
		{0.549, false, Create},
		{0, false, Create},
	}
	for _, c := range cases {
		top := Candidate{Score: c.score, Stale: c.stale}
		if got := DefaultThresholds.Decide(&top); got != c.want {
			t.Errorf("score %.3f stale=%v: %s, want %s", c.score, c.stale, got, c.want)
		}
	}
	if got := DefaultThresholds.Decide(nil); got != Create {
		t.Errorf("no candidates: %s, want %s", got, Create)
	}
}

// TestNeighbourBandEdges pins the band's inclusivity — closed below, open above — and
// deliberately does not pin the floor's value. It used to spell 0.30 into the fixture,
// which meant B-033's re-derivation to 0.125 failed here as if the band had broken. The
// number is argued in doc.go and measured by cmd/forge's sweep; what belongs in a unit
// test is which side of each edge is included.
func TestNeighbourBandEdges(t *testing.T) {
	floor := DefaultThresholds.Neighbour
	cands := []Candidate{
		{Slug: "too-high", Score: DefaultThresholds.Update}, // update band starts here
		{Slug: "in-upper", Score: DefaultThresholds.Update - 0.001},
		{Slug: "in-lower", Score: floor}, // inclusive
		{Slug: "too-low", Score: floor - 0.001},
	}
	got := DefaultThresholds.Neighbours(cands)
	if len(got) != 2 || got[0].Slug != "in-upper" || got[1].Slug != "in-lower" {
		t.Errorf("Neighbours = %v, want [in-upper in-lower]", slugs(got))
	}
}

func slugs(cs []Candidate) []string {
	out := make([]string, len(cs))
	for i, c := range cs {
		out[i] = c.Slug
	}
	return out
}

// Candidates no channel matched are truncated, not padded to TopN. Emitting ten
// 0.000 rows put `index.md` and `log.md` at the top of a CREATE verdict.
func TestRankDropsZeroScores(t *testing.T) {
	docs := []Doc{
		{Rel: "a.md", Slug: "keyset-pagination", Title: "Keyset Pagination"},
		{Rel: "b.md", Slug: "index", Title: "Index"},
		{Rel: "c.md", Slug: "log", Title: "Log"},
	}
	got := Rank(Query{Question: "how does keyset pagination work"}, docs, now)
	if len(got) != 1 || got[0].Slug != "keyset-pagination" {
		t.Fatalf("Rank = %v, want [keyset-pagination]", slugs(got))
	}
	if empty := Rank(Query{Question: "how do goroutines work"}, docs, now); len(empty) != 0 {
		t.Errorf("disjoint question: Rank = %v, want empty", slugs(empty))
	}
}

// Ties break on path so two runs over the same tree emit byte-identical JSON; Phase 2b
// re-measures against these numbers.
func TestRankTieBreakIsDeterministic(t *testing.T) {
	docs := []Doc{
		{Rel: "z.md", Slug: "pagination", Title: "Pagination"},
		{Rel: "a.md", Slug: "pagination", Title: "Pagination"},
	}
	got := Rank(Query{Question: "what is pagination"}, docs, now)
	if len(got) != 2 || got[0].Path != "a.md" {
		t.Errorf("order = %v, want a.md first", []string{got[0].Path, got[1].Path})
	}
}

func TestRankRoundsToThreeDecimals(t *testing.T) {
	docs := []Doc{{Rel: "a.md", Slug: "spring-boot-retry", Title: "Spring Boot Retry"}}
	got := Rank(Query{Question: "how does spring retry work"}, docs, now)[0].Score
	if got*1000 != float64(int(got*1000+0.5)) {
		t.Errorf("score %v carries more than three decimals", got)
	}
}

// The body channel is worth 0.1 of the blend and only the leaders are opened, so a
// note outside the window keeps its frontmatter score rather than being penalised.
func TestRankBodyPassRunsOnLoadableDocs(t *testing.T) {
	body := []byte("keyset pagination keyset pagination keyset pagination")
	docs := []Doc{{Rel: "a.md", Slug: "keyset-pagination", Title: "Keyset Pagination",
		LoadBody: func() []byte { return body }}}
	got := Rank(Query{Question: "how does keyset pagination work"}, docs, now)[0]
	if len(got.Channels) != 4 || got.Channels[3].Name != "body" {
		t.Fatalf("channels = %d, want a body channel appended", len(got.Channels))
	}
	if got.Score != 1 {
		t.Errorf("score = %v, want 1 (title and body both perfect)", got.Score)
	}
}
