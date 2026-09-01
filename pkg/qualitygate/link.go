package qualitygate

import (
	"fmt"

	"github.com/mimir45/Knowledge-Forge/pkg/vault"
)

// linkGate checks a draft's own [[wikilinks]] resolve against the rest of the vault. It
// never checks inbound links — a brand-new CREATE note has none yet, and that is
// expected, not a defect. A dangling outbound link does not block the write: DESIGN §12
// gives this gate DelegateToLibrarian, because the librarian agent (Phase 4's fourth
// subagent) is the one that goes on to create or link the missing target after write —
// see references/duplicate-spec.md's write-time-gate section for the same pattern
// applied to duplicates. So this Fail is reported honestly but does not set
// Report.Quarantine (gate.go's blocksWrite excludes DelegateToLibrarian) — the note
// still lands in notes/, and the librarian's job is the follow-up, not a hold.
func linkGate(vaultRoot string, draft *vault.Note) Outcome {
	links := vault.Wikilinks(draft.Body)
	if len(links) == 0 {
		return Outcome{Gate: "link", Verdict: Skipped, Detail: "no wikilinks in body"}
	}
	rels, err := vault.Walk(vaultRoot)
	if err != nil {
		return Outcome{Gate: "link", Verdict: Skipped, Detail: "walk vault: " + err.Error()}
	}
	ix := vault.NewIndex(rels)
	var dangling []string
	for _, l := range links {
		if _, ok := ix.Resolve(l); !ok {
			dangling = append(dangling, l)
		}
	}
	if len(dangling) == 0 {
		return Outcome{Gate: "link", Verdict: Pass}
	}
	return Outcome{Gate: "link", Verdict: Fail, Remedy: DelegateToLibrarian,
		Detail: fmt.Sprintf("%d dangling link(s): %v", len(dangling), dangling)}
}
