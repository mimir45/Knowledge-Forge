package qualitygate

import (
	"fmt"

	"github.com/mimir45/Knowledge-Forge/pkg/config"
	"github.com/mimir45/Knowledge-Forge/pkg/vault"
)

// Remedy is what a failing gate recommends. It is advisory — Run never acts on it,
// pkg/qualitygate/quarantine.go and the forge gate CLI do — because DESIGN §12 gives
// each gate a different failure response and collapsing them to one Fail/Pass bit would
// lose exactly the distinction the skill needs to act correctly.
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
// compile.go's Verdict.MarshalJSON: forge gate's JSON output must not silently
// change meaning every time this const block gains a value.
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

// Outcome is one gate's verdict. Verdict reuses compile.go's three-valued type: a gate
// that cannot judge a note (no toolchain, no defined convention) is Skipped, not Pass —
// see citation.go and freshness.go for the two gates this applies to today.
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

// Run executes all seven DESIGN §12 gates against one draft note and returns their
// combined report. draft.Rel must already be set to the note's intended vault-relative
// path — including for CREATE, where the file does not exist on disk yet — because the
// link and duplicate gates both need it to know which directory-group (== note type) to
// score the draft against.
//
// Quarantine is true when any gate Fails with a remedy that blocks the write.
// SwitchToUpdate and DelegateToLibrarian are the two remedies that do not imply
// Quarantine on their own: the first is a routing decision the caller may still choose
// to honour as a Fail, the second is the librarian's post-write job, not a reason to
// hold the note back.
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
// SwitchToUpdate and DelegateToLibrarian are routing/follow-up decisions, not defects
// the note needs fixed before it can land — see Run's doc comment.
func blocksWrite(r Remedy) bool {
	switch r {
	case None, DelegateToLibrarian, SwitchToUpdate:
		return false
	}
	return true
}
