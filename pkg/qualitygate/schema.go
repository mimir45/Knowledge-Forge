package qualitygate

import (
	"strings"

	"github.com/mimir45/Knowledge-Forge/pkg/vault"
)

// schemaGate is a thin wrapper around vault.Validate — the same check `forge validate`
// runs, so a draft that would fail `forge validate --all` fails here too, before it
// ever reaches disk. Any issue fails the gate: both missing and unknown keys block the
// write, with no distinction drawn between them.
func schemaGate(draft *vault.Note, s *vault.Schema) Outcome {
	issues := vault.Validate(draft, s)
	if len(issues) == 0 {
		return Outcome{Gate: "schema", Verdict: Pass}
	}
	return Outcome{
		Gate: "schema", Verdict: Fail, Remedy: RetryOnce,
		Detail: strings.Join(issueStrings(issues), "; "),
	}
}

func issueStrings(issues []vault.Issue) []string {
	out := make([]string, len(issues))
	for i, iss := range issues {
		out[i] = iss.String()
	}
	return out
}
