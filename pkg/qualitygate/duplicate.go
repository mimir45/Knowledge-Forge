package qualitygate

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/mimir45/Knowledge-Forge/pkg/config"
	"github.com/mimir45/Knowledge-Forge/pkg/similarity"
	"github.com/mimir45/Knowledge-Forge/pkg/vault"
)

// duplicateGate scores the draft's body against every existing note in the same
// directory-group, mirroring cmd/forge/check_collect.go's typeOf grouping exactly.
func duplicateGate(cfg *config.Config, vaultRoot string, draft *vault.Note) Outcome {
	rels, err := vault.Walk(vaultRoot)
	if err != nil {
		return Outcome{Gate: "duplicate", Verdict: Skipped, Detail: "walk vault: " + err.Error()}
	}
	ix := similarity.NewIndex()
	for _, rel := range rels {
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

// gateTypeOf mirrors cmd/forge/check_collect.go's typeOf: directory, not frontmatter.
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
