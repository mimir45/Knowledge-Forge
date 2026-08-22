package dataset

import (
	"crypto/sha256"
	"encoding/hex"
	"path"
	"regexp"

	"knowledge-forge/pkg/scrub"
)

// reInternalURL is the one pattern export adds on top of pkg/scrub's vault-tuned set.
// It is not in pkg/scrub because the vault path cannot afford it: a note body legitimately
// cites http://localhost:8080 in a worked example, and turning that into [REDACTED-URL] in
// examples/vault/ would corrupt a howto for no gain. A training corpus meant to leave the
// machine is the opposite trade — an intranet hostname is exactly the employer-identifying
// content AUDIT §8.4 D-6 is about, and losing a localhost example costs nothing.
//
// The TLD list is the set that cannot resolve publicly: RFC 6762's .local, RFC 6761's
// .test and .localhost, and the four conventional private suffixes. Private-range IPv4
// literals are matched directly.
var reInternalURL = regexp.MustCompile(`\bhttps?://(?:localhost|127\.\d+\.\d+\.\d+|` +
	`10\.\d+\.\d+\.\d+|192\.168\.\d+\.\d+|172\.(?:1[6-9]|2\d|3[01])\.\d+\.\d+|` +
	`[A-Za-z0-9.-]+\.(?:internal|local|localhost|test|corp|lan|intranet))(?::\d+)?\S*`)

// redactText is every free-text field's path out of an export: pkg/scrub's patterns, then
// the internal-URL rule above.
func redactText(s string) (string, int) {
	s, n := scrub.Redact(s)
	s = reInternalURL.ReplaceAllStringFunc(s, func(string) string {
		n++
		return "[REDACTED-URL]"
	})
	return s, n
}

// redactEach redacts every field in place and returns the total replacement count.
func redactEach(fields ...*string) int {
	total := 0
	for _, f := range fields {
		s, n := redactText(*f)
		*f, total = s, total+n
	}
	return total
}

// redactList returns a redacted copy rather than editing in place. The copy is the point:
// anonymizeRecord's type switch gives it a copy of the struct but not of the slice's
// backing array, so an in-place version would reach back into the record read off disk.
// Nothing depends on that today — anonymizeAll discards the originals — but the safety is
// a call-ordering fact no test pins, and the first caller to want a before-and-after
// comparison would get two identical tables with no clue why.
func redactList(ss []string) ([]string, int) {
	if ss == nil {
		return nil, 0
	}
	out, total := make([]string, len(ss)), 0
	for i, v := range ss {
		s, n := redactText(v)
		out[i], total = s, total+n
	}
	return out, total
}

// redactMap copies for the same reason redactList does: a map value is shared, not copied,
// when the struct holding it is.
func redactMap(m map[string]string) (map[string]string, int) {
	if m == nil {
		return nil, 0
	}
	out, total := make(map[string]string, len(m)), 0
	for k, v := range m {
		s, n := redactText(v)
		out[k], total = s, total+n
	}
	return out, total
}

// hashRel replaces a vault-relative note path with notes/<type>/<hash>.md. The directory
// is kept because the note type is a training feature and carries nothing identifying;
// the base name is a slug, and a slug can carry a product or repository name.
func hashRel(rel string) string {
	if rel == "" {
		return ""
	}
	sum := sha256.Sum256([]byte(rel))
	return path.Join(path.Dir(rel), hex.EncodeToString(sum[:8])+path.Ext(rel))
}

// anonymizeRecord redacts one record and returns it with the replacement count. The
// structural rules are per-tier and stated in each function: paths hash, commit SHAs
// empty out, free text goes through redactText.
//
// Two field classes are redacted but not hashed, and this is a stated limit rather than
// an oversight. Topic slugs, in every tier that has one, are the only semantic feature D1
// and D5 carry — hashing them makes those corpora untrainable. Profile values are D5's
// entire conditioning half, so the same argument applies. redactText catches token- and
// address-shaped content inside both; it does not catch a product or employer name
// spelled in plain kebab-case, which is exactly the shape a slug and a framework name
// take. Every datasheet names both under Limitations, and they are the two things to read
// before sharing an export.
func anonymizeRecord(rec any) (any, int) {
	switch p := rec.(type) {
	case D1Pair:
		n := redactEach(&p.Topic)
		stack, sn := redactList(p.Stack)
		p.Stack = stack
		return p, n + sn
	case D2Pair:
		return p, redactEach(&p.Draft, &p.Critique)
	case Pair:
		return anonymizeD3(p)
	case D4Pair:
		return p, redactEach(&p.FailingDraft, &p.GateError, &p.FixedDraft)
	case D5Pair:
		return anonymizeD5(p)
	}
	return rec, 0
}

func anonymizeD3(p Pair) (Pair, int) {
	p.Note = hashRel(p.Note)
	p.GenCommit, p.EditCommit = "", "" // a SHA identifies a private repository's history
	p.ContentHash = ""                 // derived from the pre-redaction text, so now a lie
	n := redactEach(&p.Topic, &p.Generated, &p.Preferred)
	stack, sn := redactList(p.Stack)
	p.Stack = stack
	return p, n + sn
}

func anonymizeD5(p D5Pair) (D5Pair, int) {
	p.Rel = hashRel(p.Rel)
	n := redactEach(&p.Topic, &p.Note)
	stack, sn := redactList(p.Stack)
	src, cn := redactList(p.Sources)
	prof, pn := redactMap(p.Profile)
	p.Stack, p.Sources, p.Profile = stack, src, prof
	return p, n + sn + cn + pn
}
