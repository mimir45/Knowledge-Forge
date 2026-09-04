package dataset

import (
	"crypto/sha256"
	"encoding/hex"
	"path"
	"regexp"

	"github.com/mimir45/Knowledge-Forge/pkg/scrub"
)

// reInternalURL is the one pattern export adds on top of pkg/scrub's vault-tuned set.
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

// redactList returns a redacted copy rather than editing in place.
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

// hashRel replaces a vault-relative note path with notes/<type>/<hash>.md.
func hashRel(rel string) string {
	if rel == "" {
		return ""
	}
	sum := sha256.Sum256([]byte(rel))
	return path.Join(path.Dir(rel), hex.EncodeToString(sum[:8])+path.Ext(rel))
}

// anonymizeRecord redacts one record and returns it with the replacement count.
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
