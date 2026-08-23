package drift

import (
	"fmt"
	"path"
	"sort"
	"strings"

	"knowledge-forge/pkg/vault"
)

// Result is one note's confidence movement, for the CLI's output and for drift.md.
type Result struct {
	Note   string `json:"note"`
	Slug   string `json:"slug"`
	Action string `json:"action"` // demoted | restored
	From   string `json:"from"`
	To     string `json:"to"`
	SHA    string `json:"sha"`
}

// Apply is the write half of drift, and the half that makes rollback symmetric. Check
// produces verdicts from tree state alone; Apply moves confidence to match them, using
// .forge/ only to remember what a demoted note is owed on the way back up.
//
// Only notes that were actually evaluated are touched. A note whose citations all sat
// outside the cheap gate produced no findings, and "not looked at" must never read as
// "not broken" — that would restore a note on an unrelated commit.
func Apply(notes map[string]*vault.Note, findings []Finding, st *Store, sch *vault.Schema,
	src Source) []Result {

	byNote, rels := group(findings)
	var out []Result
	for _, rel := range rels {
		n, ok := notes[rel]
		if !ok || n.FM == nil {
			continue
		}
		if r, moved := applyNote(n, byNote[rel], st, sch, src); moved {
			out = append(out, r)
		}
	}
	return out
}

func group(findings []Finding) (map[string][]Finding, []string) {
	byNote := map[string][]Finding{}
	for _, f := range findings {
		byNote[f.Note] = append(byNote[f.Note], f)
	}
	rels := make([]string, 0, len(byNote))
	for rel := range byNote {
		rels = append(rels, rel)
	}
	sort.Strings(rels)
	return byNote, rels
}

func applyNote(n *vault.Note, fs []Finding, st *Store, sch *vault.Schema,
	src Source) (Result, bool) {

	slug, head := slugOf(n), src.Head(repoOf(fs))
	if d, broken := firstBroken(fs); broken {
		return demote(n, slug, d, head, st, sch)
	}
	return restore(n, slug, head, st, sch)
}

func demote(n *vault.Note, slug string, d Finding, head string, st *Store,
	sch *vault.Schema) (Result, bool) {

	prev := n.FM.Str("confidence")
	if prev == "" || prev == "low" || !st.Record(rec(n, slug, d, head)) {
		stamp(n, sch, head) // already down, or already remembered: nothing to move
		return Result{}, false
	}
	if write(n, sch, map[string]string{"confidence": "low", "drift_checked_at": head}) != nil {
		return Result{}, false
	}
	st.Log(fmt.Sprintf("demote %s at %s ref=%q confidence=%s->low", slug, short(head), d.Ref, prev))
	return Result{n.Rel, slug, "demoted", prev, "low", head}, true
}

// restore is rollback symmetry in six lines: the verdict came back clean, so whatever the
// note was worth before it broke is what it is worth again. The log line cites both shas —
// the one that demoted it and the one that cleared it — because that pair is the only
// audit trail, .forge/ having just forgotten the record.
func restore(n *vault.Note, slug, head string, st *Store, sch *vault.Schema) (Result, bool) {
	d, ok := st.Take(slug)
	if !ok {
		stamp(n, sch, head)
		return Result{}, false
	}
	if write(n, sch, map[string]string{"confidence": d.Confidence, "drift_checked_at": head}) != nil {
		return Result{}, false
	}
	st.Log(fmt.Sprintf("restore %s %s -> %s confidence=low->%s",
		slug, short(d.SHA), short(head), d.Confidence))
	return Result{n.Rel, slug, "restored", "low", d.Confidence, head}, true
}

// stamp records that this note was looked at, and does nothing when the sha has not
// moved — a vault that rewrites every evaluated note on every commit is churn, not
// history.
func stamp(n *vault.Note, sch *vault.Schema, head string) {
	if head == "" || n.FM.Str("drift_checked_at") == head {
		return
	}
	write(n, sch, map[string]string{"drift_checked_at": head}) //nolint:errcheck // a failed stamp leaves drift_checked_at stale, so the next run re-evaluates the note: the same non-event demote's own write failure is
}

func write(n *vault.Note, sch *vault.Schema, kv map[string]string) error {
	return vault.SetScalars(n, sch, kv)
}

func rec(n *vault.Note, slug string, d Finding, head string) Demotion {
	return Demotion{Note: n.Rel, Slug: slug, Confidence: n.FM.Str("confidence"),
		Repo: d.Repo, SHA: head, Ref: d.Ref}
}

func firstBroken(fs []Finding) (Finding, bool) {
	for _, f := range fs {
		if f.Demoting() {
			return f, true
		}
	}
	return Finding{}, false
}

func repoOf(fs []Finding) string {
	for _, f := range fs {
		if f.Repo != "" {
			return f.Repo
		}
	}
	return ""
}

func slugOf(n *vault.Note) string {
	if s := n.FM.Str("slug"); s != "" {
		return s
	}
	return strings.TrimSuffix(path.Base(n.Rel), ".md")
}

func short(sha string) string {
	if len(sha) > 8 {
		return sha[:8]
	}
	return sha
}
