package qualitygate

import (
	"fmt"

	"github.com/mimir45/Knowledge-Forge/pkg/config"
	"github.com/mimir45/Knowledge-Forge/pkg/vault"
)

// Remedy is what a failing gate recommends.
type Remedy int

const (
	None                Remedy = iota
	RetryOnce                  // schema: block write, fix, retry once
	MarkUnverified             // citation: publish flagged, never silently trusted
	DropConfidence             // code / freshness: publish, but confidence drops
	RewritePass                // anti-slop: needs a rewrite pass, not a fact fix
	DelegateToLibrarian        // link: librarian adds the missing links post-write
	SwitchToUpdate             // duplicate: a routing decision, not a vault mutation
)

var remedyNames = [...]string{"none", "retry_once", "mark_unverified", "drop_confidence",
	"rewrite_pass", "delegate_to_librarian", "switch_to_update"}

func (r Remedy) String() string {
	if int(r) < 0 || int(r) >= len(remedyNames) {
		return "unknown"
	}
	return remedyNames[r]
}

// MarshalJSON serializes the name, not the iota ordinal — same rationale as
// compile.go's Verdict.MarshalJSON.
func (r Remedy) MarshalJSON() ([]byte, error) { return []byte(`"` + r.String() + `"`), nil }

func (r *Remedy) UnmarshalJSON(b []byte) error {
	name := string(b)
	for i, n := range remedyNames {
		if `"`+n+`"` == name {
			*r = Remedy(i)
			return nil
		}
	}
	return fmt.Errorf("qualitygate: unknown Remedy %s", b)
}

// Mode is CREATE vs UPDATE — see Run's doc comment for why a typo here must not be
// allowed to silently fall through to CREATE's behaviour.
type Mode int

const (
	ModeCreate Mode = iota
	ModeUpdate
)

// Outcome is one gate's verdict.
type Outcome struct {
	Gate    string  `json:"gate"`
	Verdict Verdict `json:"verdict"`
	Remedy  Remedy  `json:"remedy"`
	Detail  string  `json:"detail,omitempty"`
}

// Report is every gate's outcome plus the one bit the CLI and skill act on directly.
type Report struct {
	Outcomes   []Outcome `json:"outcomes"`
	Quarantine bool      `json:"quarantine"`
}

// Run executes all seven gates against one draft note and returns their combined
// report. draft.Rel must already be set to the note's intended vault-relative path.
func Run(cfg *config.Config, vaultRoot string, draft *vault.Note, mode Mode) (Report, error) {
	s, err := vault.LoadSchema()
	if err != nil {
		return Report{}, err
	}
	rep := Report{Outcomes: []Outcome{
		schemaGate(draft, s),
		citationGate(cfg, draft, s),
		codeGate(cfg, draft),
		freshnessGate(draft, s, mode),
		antislopGate(cfg, draft),
		linkGate(vaultRoot, draft),
		duplicateGate(cfg, vaultRoot, draft),
	}}
	for _, o := range rep.Outcomes {
		if o.Verdict == Fail && blocksWrite(o.Remedy) {
			rep.Quarantine = true
		}
	}
	return rep, nil
}

// blocksWrite reports whether a Fail with this remedy should hold the note back.
func blocksWrite(r Remedy) bool {
	switch r {
	case None, DelegateToLibrarian, SwitchToUpdate:
		return false
	}
	return true
}
