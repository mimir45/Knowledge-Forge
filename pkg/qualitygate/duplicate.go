package qualitygate

import (
	"fmt"
	"path/filepath"
	"strings"

	"knowledge-forge/pkg/config"
	"knowledge-forge/pkg/similarity"
	"knowledge-forge/pkg/vault"
)

// duplicateGate scores the draft's body against every existing note in the same
// directory-group, mirroring cmd/forge/check_collect.go's typeOf grouping exactly — a
// pair only ever compares within one group, so this gate has to agree with how the
// weekly checker groups or it would flag pairs the report never would (or miss ones it
// does). Threshold is cfg.Verify.DuplicateThreshold, deliberately not cfg.Check's: see
// references/duplicate-spec.md §6 — a write-time trigger and a passive report threshold
// are asymmetric uses of the same number, not one config value in two places.
//
// SwitchToUpdate is a routing recommendation, not a vault mutation: it never sets
// Quarantine's Fail path to something the CLI is forced to honour, because the skill may
// have a stated reason (DESIGN §12) to publish two notes on the same topic anyway.
func duplicateGate(cfg *config.Config, vaultRoot string, draft *vault.Note) Outcome {
	rels, err := vault.Walk(vaultRoot)
	if err != nil {
		return Outcome{Gate: "duplicate", Verdict: Skipped, Detail: "walk vault: " + err.Error()}
	}
	ix := similarity.NewIndex()
	for _, rel := range rels {
		// _inbox/ holds quarantine drafts, not published content — without this a
		// draft that gets re-quarantined on retry would score as a near-duplicate of
		// its own previous _inbox/ copy (different path, same body) every time.
		if rel == draft.Rel || !vault.IsContentNote(rel) || strings.HasPrefix(rel, "_inbox/") {
			continue
		}
		n, err := vault.Load(filepath.Join(vaultRoot, rel), rel)
		if err != nil {
			continue // unreadable note is the vault's problem, not this draft's
		}
		ix.Add(n.Rel, gateTypeOf(n.Rel), string(n.Body))
	}
	ix.Add(draft.Rel, gateTypeOf(draft.Rel), string(draft.Body))

	threshold := cfg.Verify.DuplicateThreshold
	if threshold <= 0 {
		threshold = similarity.DuplicateThreshold
	}
	var hits []string
	for _, p := range ix.Pairs(threshold) {
		if p.A == draft.Rel {
			hits = append(hits, fmt.Sprintf("%s (%.2f)", p.B, p.Score))
		} else if p.B == draft.Rel {
			hits = append(hits, fmt.Sprintf("%s (%.2f)", p.A, p.Score))
		}
	}
	if len(hits) == 0 {
		return Outcome{Gate: "duplicate", Verdict: Pass}
	}
	return Outcome{Gate: "duplicate", Verdict: Fail, Remedy: SwitchToUpdate,
		Detail: fmt.Sprintf("near-duplicate of %s", strings.Join(hits, ", "))}
}

// gateTypeOf mirrors cmd/forge/check_collect.go's typeOf: directory, not frontmatter,
// because that is what duplicate scoping has to agree with on disk. Duplicated rather
// than imported — cmd/forge depends on pkg/qualitygate, not the other way around.
func gateTypeOf(rel string) string {
	rest, ok := strings.CutPrefix(rel, "notes/")
	if !ok {
		if i := strings.Index(rel, "/"); i > 0 {
			return rel[:i]
		}
		return ""
	}
	if i := strings.Index(rest, "/"); i > 0 {
		return rest[:i]
	}
	return ""
}
