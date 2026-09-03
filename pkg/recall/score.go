package recall

import (
	"math"
	"strings"
)

// Channel weights, the original blend verbatim. Their ratio is what the design fixes;
// the denominator they are divided by is what §2.5 of the spec decides.
const (
	wTitle = 0.4
	wTags  = 0.3
	wStack = 0.2
	wBody  = 0.1
)

// scope is the query-side state shared by every candidate in one run: the terms.
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
// within a note, so this is document frequency, not term frequency.
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

// dfOver narrows the vault-wide counts to the query's own terms, mirroring idfOver.
func dfOver(terms []string, df map[string]int) map[string]int {
	out := make(map[string]int, len(terms))
	for _, t := range terms {
		out[t] = df[t]
	}
	return out
}

// weightsOver weighs only the query's own terms.
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

// idfCap bounds one term's weight.
const idfCap = 3.5

// idf is a term's inverse document frequency over n notes.
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
// stopword-filtered.
func (s scope) titleChannel(d Doc) Channel {
	t := termSetOf(d.Title, strings.ReplaceAll(d.Slug, "-", " "))
	hits := intersect(s.terms, t)
	c := Channel{Name: "title", Weight: wTitle, Active: len(t) > 0, Hits: hits}
	c.Value = f2(len(hits), len(s.terms), len(t))
	return c
}

// f2 is the recall-weighted F-measure over the query terms covered by the title.
func f2(hits, queryTerms, titleTokens int) float64 {
	if hits == 0 || queryTerms == 0 || titleTokens == 0 {
		return 0
	}
	p := float64(hits) / float64(titleTokens)
	r := float64(hits) / float64(queryTerms)
	return 5 * p * r / (4*p + r)
}

// tagsChannel divides by all of the query's terms — each side weighted, see weightsOver
// and weighted below — and not by the note's tag count.
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

// weighted is the shared body of both.
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

// blend is the weighted mean over active channels.
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
