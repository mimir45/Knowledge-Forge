package qualitygate

import (
	"fmt"

	"github.com/mimir45/Knowledge-Forge/pkg/config"
	"github.com/mimir45/Knowledge-Forge/pkg/vault"
)

// citationGate is deliberately narrower than the original spec's prose.
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
