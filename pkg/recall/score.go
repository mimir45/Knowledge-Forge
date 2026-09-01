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
	tagTerms   []string // every query term — a term no note tags is evidence too
	stackTerms []string // --stack values the vault knows, plus every query term
	tagIDF     map[string]float64
	stackIDF   map[string]float64
	tagDF      map[string]int // raw counts behind the weights, carried for --explain
	stackDF    map[string]int
}

// newScope decides, once per run, which terms each channel is answerable for.
//
// The vocabulary filter applies to --stack hints and not to question terms — the
// reverse of the naive reading, and a bug once shipped that way. The asymmetry is the point: a hint is a user
// filter, and narrowing a search by "kotlin" in a vault that has never seen Kotlin should
// not thereby make every note match less well. A question term is evidence — the vault
// carrying no note about "redis" is exactly what the caller needs the score to reflect,
// and filtering it out is what let a Spring CLI note answer a Redis question at 0.740.
func newScope(q Query, docs []Doc) scope {
	terms := Terms(q.Question)
	tagDF, stackDF := docFreq(docs)
	s := scope{terms: terms, tagTerms: terms}
	s.stackTerms = dedupe(append(inVocab(Terms(strings.Join(q.Stack, " ")), stackDF),
		terms...))
	s.tagIDF = weightsOver(s.tagTerms, tagDF, len(docs))
	s.stackIDF = weightsOver(s.stackTerms, stackDF, len(docs))
	s.tagDF, s.stackDF = dfOver(s.tagTerms, tagDF), dfOver(s.stackTerms, stackDF)
	return s
}

// docFreq counts how many notes carry each tag and stack token. setOf deduplicates
// within a note, so this is document frequency, not term frequency. It replaced a
// presence-only vocabulary and costs nothing extra: the pass already walked every note
// once, so IDF weighting could be a channel change rather than a new scan.
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

// inVocab keeps the terms some note actually carries, in the order given. It
// is applied to --stack hints only; see newScope for why question terms do not go
// through it.
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

// weightsOver weighs only the query's own terms. The vault-wide vocabulary is never
// materialised as weights, because nothing reads a weight for a term nobody asked about.
//
// A term some note carries weighs its IDF. A term no note carries weighs the mean of the
// ones that do — a deliberate design choice among several considered. The mean is parameterless
// and calibrates itself against whatever the query's present terms happen to weigh, so it
// introduces no constant to tune. It also preserves the
// invariant the ratio is supposed to express: a query whose m channel terms the vault
// carries k of scores about k/m, exactly when the present weights are equal. The
// alternative considered — flooring document frequency at 1 — hands an absent
// term the *maximum* weight and inverts idfCap's purpose, letting absence outweigh
// presence. A mean of capped values is capped, so the guard still holds here.
//
// With no present terms the mean is undefined and every weight stays 0, so weighted's
// empty-denominator contract leaves the channel inactive. That falls out rather than
// being special-cased, and it is the behaviour spec §2.5 argues for: a query the vault's
// vocabulary cannot speak to at all must not activate a channel and score every note 0.0.
//
// Iterated over the terms slice, never over the map, so two runs agree bit for bit.
func weightsOver(terms []string, df map[string]int, n int) map[string]float64 {
	w := make(map[string]float64, len(terms))
	sum, present := 0.0, 0
	for _, t := range terms {
		if w[t] = idf(df[t], n); df[t] > 0 {
			sum, present = sum+w[t], present+1
		}
	}
	if present == 0 {
		return w
	}
	mean := sum / float64(present)
	for _, t := range terms {
		if df[t] <= 0 {
			w[t] = mean
		}
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
// as "redis".
//
// log(1 + n/df) rather than log(n/df) for a mechanical reason: the unsmoothed form is
// exactly zero when a term is on every note, so a query whose terms are all universal
// would divide by a zero weight sum on an active channel. This form bottoms out at
// log(2) and cannot vanish. A term no note carries has a raw IDF of 0, which is the
// honest measure of it: inverse document frequency is defined inside the corpus, and this
// term is outside it. What such a term is worth as *evidence* is a separate question, and
// it is answered a layer up in weightsOver — deliberately, so this function stays the
// measure it claims to be.
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

// tagsChannel divides by all of the query's terms — each side weighted, see weightsOver
// and weighted below — and not by the note's tag count. The denominator holds
// terms no note tags at all: a question about Redis that the vault answers nowhere must
// not read as a perfect tag match on the strength of the one term it shares with half the
// vault. Dividing by the note's count would score a note tagged
// [goroutines] at 1.0 and one tagged [goroutines, concurrency, runtime] at 0.33 on the
// same match — ranking the better-curated note lower. See spec §2.3.
//
// Activation requires an actual hit, not merely a non-empty tag list. Before,
// "the note carries the field" was tested as len(tags) > 0, so a note tagged [issue]
// against a Redis question paid the absent-term penalty in full — active, scoring
// 0.000 — while an untagged note skipped the channel entirely. That is the same mistake
// weighted's own comment argues against for a query outside the vault's vocabulary
// altogether: scoring 0.0 on an active channel drags a note down for a reason that has
// nothing to do with relevance. len(hits) > 0 makes "carries the field" mean "carries
// something the query could match," so a note with no relevant tags is inactive whether
// it has irrelevant ones or none — 31 of this vault's 91 notes are missing tags or stack
// after the Phase 1 migration. See recall-spec.md §2.5.
func (s scope) tagsChannel(d Doc) Channel {
	tags := setOf(d.Tags)
	hits := intersect(s.tagTerms, tags)
	c := Channel{Name: "tags", Weight: wTags, Hits: hits, Terms: s.tagIDF, DF: s.tagDF}
	value, ok := weighted(s.tagTerms, hits, s.tagIDF)
	if c.Active = ok && len(hits) > 0; c.Active {
		c.Value = value
	}
	return c
}

// stackChannel is containment over the query's hints: a note listing extra technologies
// is not thereby less relevant, so a superset scores full marks.
//
// Activation requires an actual hit, not merely a non-empty stack list — the same
// fix as tagsChannel, for the same reason: a note listing stack entries none of which the
// query touches must be inactive, not active-and-zero.
func (s scope) stackChannel(d Doc) Channel {
	stack := setOf(d.Stack)
	hits := intersect(s.stackTerms, stack)
	c := Channel{Name: "stack", Weight: wStack, Hits: hits, Terms: s.stackIDF, DF: s.stackDF}
	value, ok := weighted(s.stackTerms, hits, s.stackIDF)
	if c.Active = ok && len(hits) > 0; c.Active {
		c.Value = value
	}
	return c
}

// weighted is the shared body of both: summed IDF of the terms that hit over summed IDF
// of the terms that could have. It replaces a plain hit ratio, in which every term
// counted the same and a note tagged [spring, boot] matched a Redis question perfectly.
//
// An empty denominator reports the channel as having nothing to say, and the caller
// deactivates it. That is not a corner case: it is how a query lands whose
// terms the vault's tag or stack vocabulary knows none of — weightsOver leaves every
// weight at 0 because there is no present term to take a mean of. Scoring such a query
// 0.0 on an *active* channel would drag every note down uniformly, the same mistake spec
// §2.5 rejects for untagged notes.
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
