package qualitygate

import (
	"fmt"

	"github.com/mimir45/Knowledge-Forge/pkg/vault"
)

// OpenQuestions turns a Report's failing outcomes into vault.WriteToInbox's
// openQuestions bullets, one per Fail, gate order.
func OpenQuestions(rep Report) []string {
	var qs []string
	for _, o := range rep.Outcomes {
		if o.Verdict == Fail {
			qs = append(qs, fmt.Sprintf("%s: %s", o.Gate, o.Detail))
		}
	}
	return qs
}

// Quarantine writes a failing draft to _inbox/ via vault.WriteToInbox, applying the
// CREATE/UPDATE split this plan settled on: CREATE has no published note to protect.
func Quarantine(vaultRoot string, draft *vault.Note, s *vault.Schema, rep Report, mode Mode, targetSlug string) error {
	if mode == ModeUpdate && targetSlug != "" && draft.FM != nil {
		draft.FM.SetScalar("supersedes", targetSlug)
	}
	return vault.WriteToInbox(vaultRoot, draft, s, OpenQuestions(rep))
}
