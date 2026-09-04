package qualitygate

import (
	"fmt"

	"github.com/mimir45/Knowledge-Forge/pkg/vault"
)

// linkGate checks a draft's own [[wikilinks]] resolve against the rest of the vault.
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
