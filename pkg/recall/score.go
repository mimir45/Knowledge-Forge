package recall

import (
	"math"
	"strings"
)

// Channel weights, DESIGN §8's blend verbatim. Their ratio is what the design fixes;
// the denominator they are divided by is what §2.5 of the spec decides.
const (
	wTitle = 0.4
	wTags  = 0.3
	wStack = 0.2
	wBody  = 0.1
)

// scope is the query-side state shared by every candidate in one run: the terms, and
// which channels the query gave any input for. Activation is decided once here rather
// than per candidate, so renormalization sets the scale without reordering results.
type scope struct {
	terms      []string
	tagTerms   []string // query terms that are a tag somewhere in the vault
	stackTerms []string // --stack values plus query terms in the stack vocabulary
	tagIDF     map[string]float64
	stackIDF   map[string]float64
	tagDF      map[string]int // raw counts behind the weights, carried for --explain
	stackDF    map[string]int
}

func newScope(q Query, docs []Doc) scope {
	terms := Terms(q.Question)
	tagDF, stackDF := docFreq(docs)
	s := scope{terms: terms, tagTerms: inVocab(terms, tagDF)}
	s.stackTerms = dedupe(append(Terms(strings.Join(q.Stack, " ")),
		inVocab(terms, stackDF)...))
	s.tagIDF = idfOver(s.tagTerms, tagDF, len(docs))
	s.stackIDF = idfOver(s.stackTerms, stackDF, len(docs))
	s.tagDF, s.stackDF = dfOver(s.tagTerms, tagDF), dfOver(s.stackTerms, stackDF)
	return s
}

// docFreq counts how many notes carry each tag and stack token. setOf deduplicates
// within a note, so this is document frequency, not term frequency. It replaced a
// presence-only vocabulary and costs nothing extra: the pass already walked every note
// once, which is why B-008 could be a channel change rather than a new scan.
func docFreq(docs []Doc) (tags, stack map[string]int) {
	tags, stack = map[string]int{}, map[string]int{}
	for _, d := range docs {
		for t := range setOf(d.Tags) {
			tags[t]++
		}
		for t := range setOf(d.Stack) {
			stack[t]++
		}
	}
	return tags, stack
}

// inVocab keeps the terms some note actually carries, in the order given.
func inVocab(terms []string, df map[string]int) []string {
	var out []string
	for _, t := range terms {
		if df[t] > 0 {
			out = append(out, t)
		}
	}
	return out
}

// dfOver narrows the vault-wide counts to the query's own terms, mirroring idfOver. The
// full maps are not kept: nothing reads a count for a term nobody asked about, and a
// query-scope map is shared read-only by every candidate.
func dfOver(terms []string, df map[string]int) map[string]int {
	out := make(map[string]int, len(terms))
	for _, t := range terms {
		out[t] = df[t]
	}
	return out
}

// idfOver weighs only the query's own terms. The vault-wide vocabulary is never
// materialised as weights, because nothing reads a weight for a term nobody asked about.
func idfOver(terms []string, df map[string]int, n int) map[string]float64 {
	w := make(map[string]float64, len(terms))
	for _, t := range terms {
		w[t] = idf(df[t], n)
	}
	return w
}

// idfCap bounds one term's weight. Because a universal term always weighs log(2), the
// cap fixes the widest spread between the rarest and the commonest term at about 5:1
// whatever the vault's size — a guard against a hapax tag deciding a verdict on its own,
// not the fix itself. The fix is that terms are weighted at all.
const idfCap = 3.5

// idf is a term's inverse document frequency over n notes. Unweighted, the tags and
// stack channels scored "Redis caching in Spring Boot" at 0.740 against a Spring CLI
// note: "spring" and "boot", carried by much of the vault, counted for exactly as much
// as "redis" (B-008).
//
// log(1 + n/df) rather than log(n/df) for a mechanical reason: the unsmoothed form is
// exactly zero when a term is on every note, so a query whose terms are all universal
// would divide by a zero weight sum on an active channel. This form bottoms out at
// log(2) and cannot vanish. A term no note carries weighs 0 — inside this vault it
// separates nothing, so it neither helps nor penalises.
func idf(df, n int) float64 {
	if df <= 0 || n <= 0 {
		return 0
	}
	return math.Min(math.Log(1+float64(n)/float64(df)), idfCap)
}

func dedupe(in []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, v := range in {
		if !seen[v] {
			seen[v] = true
			out = append(out, v)
		}
	}
	return out
}

// titleChannel compares the query terms against the note's title and slug, both sides
// stopword-filtered. Filtering the note side matters as much as the query side: a slug
// like "hexagonal-architecture-ports-and-adapters" carries an "and" that no question
// will ever match, and every such token dilutes the precision half of f2.
func (s scope) titleChannel(d Doc) Channel {
	t := termSetOf(d.Title, strings.ReplaceAll(d.Slug, "-", " "))
	hits := intersect(s.terms, t)
	c := Channel{Name: "title", Weight: wTitle, Active: len(t) > 0, Hits: hits}
	c.Value = f2(len(hits), len(s.terms), len(t))
	return c
}

// f2 is the recall-weighted F-measure over the query terms covered by the title
// (recall) and the title tokens the query accounts for (precision).
//
// Neither half works alone, measured on the real vault:
//   - Pure coverage rates "Spring Boot 4 Breaking Changes" a perfect match for "how does
//     spring boot work", because a long title contains everything.
//   - Symmetric Dice punishes a title for being *more specific* than the question:
//     "Keyset Pagination — Compound OR Predicate" is exactly the note you want for "how
//     does keyset pagination work" and scored 0.67, below the answer threshold.
//
// β=2 leans on coverage — the question is what is being looked up — while precision
// still pulls an over-broad title down. The two cases above become 0.59 and 0.83.
func f2(hits, queryTerms, titleTokens int) float64 {
	if hits == 0 || queryTerms == 0 || titleTokens == 0 {
		return 0
	}
	p := float64(hits) / float64(titleTokens)
	r := float64(hits) / float64(queryTerms)
	return 5 * p * r / (4*p + r)
}

// tagsChannel divides by the query terms that could have matched *some* note's tags,
// not by the note's tag count — each side weighted by IDF, see weighted below. Dividing by the note's count would score a note tagged
// [goroutines] at 1.0 and one tagged [goroutines, concurrency, runtime] at 0.33 on the
// same match — ranking the better-curated note lower. See spec §2.3.
// A note with no tags at all leaves the channel inactive rather than scoring zero. Tag
// *mismatch* is evidence against relevance; tag *absence* is no evidence either way, and
// 31 of this vault's 91 notes are missing tags or stack after the Phase 1 migration. See
// recall-spec.md §2.5 for why zeroing them ranked an under-curated note below a
// well-tagged irrelevant one.
func (s scope) tagsChannel(d Doc) Channel {
	tags := setOf(d.Tags)
	hits := intersect(s.tagTerms, tags)
	c := Channel{Name: "tags", Weight: wTags, Hits: hits, Terms: s.tagIDF, DF: s.tagDF}
	value, ok := weighted(s.tagTerms, hits, s.tagIDF)
	if c.Active = ok && len(tags) > 0; c.Active {
		c.Value = value
	}
	return c
}

// stackChannel is containment over the query's hints: a note listing extra technologies
// is not thereby less relevant, so a superset scores full marks.
func (s scope) stackChannel(d Doc) Channel {
	stack := setOf(d.Stack)
	hits := intersect(s.stackTerms, stack)
	c := Channel{Name: "stack", Weight: wStack, Hits: hits, Terms: s.stackIDF, DF: s.stackDF}
	value, ok := weighted(s.stackTerms, hits, s.stackIDF)
	if c.Active = ok && len(stack) > 0; c.Active {
		c.Value = value
	}
	return c
}

// weighted is the shared body of both: summed IDF of the terms that hit over summed IDF
// of the terms that could have. It replaces a plain hit ratio, in which every term
// counted the same and a note tagged [spring, boot] matched a Redis question perfectly.
//
// An empty denominator reports the channel as having nothing to say, and the caller
// deactivates it. That is not a corner case: --stack accepts hints the vault has never
// seen, and those terms weigh zero. Scoring them 0.0 on an *active* channel would drag
// every note down uniformly — the same mistake spec §2.5 rejects for untagged notes.
//
// Both slices are iterated in order rather than ranged over a map, so the sums are
// bit-identical between runs and --explain's hit list stays byte-stable.
func weighted(terms, hits []string, w map[string]float64) (value float64, ok bool) {
	den := 0.0
	for _, t := range terms {
		den += w[t]
	}
	if den == 0 {
		return 0, false
	}
	num := 0.0
	for _, t := range hits {
		num += w[t]
	}
	return num / den, true
}

// bodyChannel counts query terms in the body, saturating each at three occurrences so
// one word repeated forty times cannot stand in for coverage.
func (s scope) bodyChannel(body []byte) Channel {
	c := Channel{Name: "body", Weight: wBody, Active: true}
	if len(s.terms) == 0 {
		return c
	}
	counts := map[string]int{}
	for _, t := range Tokens(string(body)) {
		counts[t]++
	}
	sum := 0.0
	for _, t := range s.terms {
		if n := counts[t]; n > 0 {
			c.Hits = append(c.Hits, t)
			sum += min(float64(n), 3) / 3
		}
	}
	c.Value = sum / float64(len(s.terms))
	return c
}

// blend is the weighted mean over active channels. An inactive channel leaves the
// denominator as well as the numerator, which is the whole point: a channel the query
// gave no input for is absence of evidence, not evidence of absence.
func blend(chs []Channel) (score float64, matched []string) {
	num, den := 0.0, 0.0
	for _, c := range chs {
		if !c.Active {
			continue
		}
		num += c.Weight * c.Value
		den += c.Weight
		if c.Value > 0 {
			matched = append(matched, c.Name)
		}
	}
	if den == 0 {
		return 0, matched
	}
	return num / den, matched
}
