package qualitygate

import (
	"fmt"

	"github.com/mimir45/Knowledge-Forge/pkg/config"
	"github.com/mimir45/Knowledge-Forge/pkg/vault"
)

// citationGate is deliberately narrower than DESIGN §12's prose ("every claim tagged
// version-specific/performance-claim/security-claim has a source"). No doc or
// references/schema.yaml defines how a body claim gets tagged with one of those three
// categories — require_citation_for names the categories a strict pipeline cares about,
// not the mechanism a T0 tool could detect one with, and inventing body markup here
// would be a note-contract change, not a gate. That gap is a known, open limitation.
//
// What is checkable without a model call: whether this note's type requires a source at
// all (schema.yaml's sources.min_items_by_type — decision/incident may cite nothing,
// everything else needs ≥1) and whether the draft meets that floor. schema.go's
// checkSourcesArity already blocks on the same fact with RetryOnce; this gate reports
// it a second time with MarkUnverified so a caller that treats citation failures as
// soft (publish, but flagged) can act on that distinction instead of the hard block.
func citationGate(cfg *config.Config, draft *vault.Note, s *vault.Schema) Outcome {
	if len(cfg.Verify.RequireCitationFor) == 0 {
		return Outcome{Gate: "citation", Verdict: Skipped,
			Detail: "verify.require_citation_for is empty; citation strictness is off"}
	}
	if draft.FM == nil {
		return Outcome{Gate: "citation", Verdict: Skipped, Detail: "no frontmatter"}
	}
	min := sourceFloor(s, draft.FM.Str("type"))
	if min == 0 {
		return Outcome{Gate: "citation", Verdict: Skipped,
			Detail: "per-claim tagging has no defined convention; note type requires no source"}
	}
	if got := len(draft.FM.List("sources")); got < min {
		return Outcome{Gate: "citation", Verdict: Fail, Remedy: MarkUnverified,
			Detail: fmt.Sprintf("%d source(s), type requires at least %d; per-claim "+
				"category tagging not checked — no convention defines it", got, min)}
	}
	return Outcome{Gate: "citation", Verdict: Pass,
		Detail: "source floor met; per-claim category tagging not checked — no convention defines it"}
}

func sourceFloor(s *vault.Schema, typ string) int {
	f, ok := s.Fields["sources"]
	if !ok || f.MinItemsByType == nil {
		return 0
	}
	return f.MinItemsByType[typ]
}
