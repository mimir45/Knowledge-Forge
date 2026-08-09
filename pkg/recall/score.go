package recall

import "strings"

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
}

func newScope(q Query, docs []Doc) scope {
	terms := Terms(q.Question)
	tagVocab, stackVocab := vocab(docs)
	s := scope{terms: terms, tagTerms: intersect(terms, tagVocab)}
	s.stackTerms = dedupe(append(Terms(strings.Join(q.Stack, " ")),
		intersect(terms, stackVocab)...))
	return s
}

func vocab(docs []Doc) (tags, stack map[string]bool) {
	tags, stack = map[string]bool{}, map[string]bool{}
	for _, d := range docs {
		for t := range setOf(d.Tags) {
			tags[t] = true
		}
		for t := range setOf(d.Stack) {
			stack[t] = true
		}
	}
	return tags, stack
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
// not by the note's tag count. Dividing by the note's count would score a note tagged
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
	c := Channel{Name: "tags", Weight: wTags, Hits: hits,
		Active: len(s.tagTerms) > 0 && len(tags) > 0}
	if c.Active {
		c.Value = float64(len(hits)) / float64(len(s.tagTerms))
	}
	return c
}

// stackChannel is containment over the query's hints: a note listing extra technologies
// is not thereby less relevant, so a superset scores full marks.
func (s scope) stackChannel(d Doc) Channel {
	stack := setOf(d.Stack)
	hits := intersect(s.stackTerms, stack)
	c := Channel{Name: "stack", Weight: wStack, Hits: hits,
		Active: len(s.stackTerms) > 0 && len(stack) > 0}
	if c.Active {
		c.Value = float64(len(hits)) / float64(len(s.stackTerms))
	}
	return c
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
