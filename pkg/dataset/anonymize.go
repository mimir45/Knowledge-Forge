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

func redactList(ss []string) int {
	total := 0
	for i := range ss {
		s, n := redactText(ss[i])
		ss[i], total = s, total+n
	}
	return total
}

func redactMap(m map[string]string) int {
	total := 0
	for k, v := range m {
		s, n := redactText(v)
		m[k], total = s, total+n
	}
	return total
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
// Topic is redacted but not hashed, in every tier that has one. That is a deliberate,
// stated limit rather than an oversight: a topic is the slug of the question and is the
// only semantic feature D1 and D5 carry, so hashing it makes those two corpora
// untrainable. redactText catches token- and address-shaped content inside it; it does
// not catch a product name spelled in plain kebab-case. Every datasheet says so under
// Limitations, and it is the one field to read before sharing an export.
func anonymizeRecord(rec any) (any, int) {
	switch p := rec.(type) {
	case D1Pair:
		return p, redactEach(&p.Topic) + redactList(p.Stack)
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
	return p, redactEach(&p.Topic, &p.Generated, &p.Preferred) + redactList(p.Stack)
}

func anonymizeD5(p D5Pair) (D5Pair, int) {
	p.Rel = hashRel(p.Rel)
	n := redactEach(&p.Topic, &p.Note)
	return p, n + redactList(p.Stack) + redactList(p.Sources) + redactMap(p.Profile)
}
