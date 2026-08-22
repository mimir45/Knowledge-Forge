package dataset

import "time"

// D5 is ADDENDUM §D.1's style dataset: (question, profile, sources) → the note that was
// accepted. The acceptance signal is `forge gate`'s non-quarantine branch, which is the
// only place in the tree that decides a draft may be published.
//
// The limitation to state in the datasheet: nothing in Go enforces that gate runs. It is
// an invariant of skills/forge/SKILL.md ("forge gate runs before every write, in every
// branch — no branch skips it"), so a note written around the skill is a note D5 never
// sees. D5 is therefore a subset of accepted notes, not a census of them.
const (
	D5Kind = "d5-style"
	D5Path = ".forge/datasets/d5.jsonl"
	D5Tag  = "d5"
)

// D5Pair is one JSONL record. The question is not stored even in hashed form — by the
// time gate runs, the draft is what carries the signal and Topic (the note's slug) is the
// question's surviving trace.
//
// Profile holds only the conditioning fields of profiles/me.md that have a fixed shape:
// primary_language, frameworks, infra, seniority, default_depth, note_language,
// explain_style. The four free-text fields (assume_known, never_assume, code_style,
// avoid) are deliberately excluded — the user writes them by hand and they can carry an
// employer's vocabulary, so not capturing them is cheaper than scrubbing them on export.
type D5Pair struct {
	Kind       string            `json:"kind"`
	Topic      string            `json:"topic"`
	Rel        string            `json:"rel"`  // vault-relative path; hashed on export
	Note       string            `json:"note"` // the accepted body — the training target
	Profile    map[string]string `json:"profile,omitempty"`
	Stack      []string          `json:"stack,omitempty"`
	Sources    []string          `json:"sources,omitempty"`
	CapturedAt time.Time         `json:"captured_at"`
}

// AppendD5 writes one pair as a JSONL line.
func AppendD5(vaultRoot string, p D5Pair) error { return D5.Append(vaultRoot, p) }
