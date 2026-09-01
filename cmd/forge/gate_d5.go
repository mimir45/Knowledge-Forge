package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mimir45/Knowledge-Forge/pkg/config"
	"github.com/mimir45/Knowledge-Forge/pkg/dataset"
	"github.com/mimir45/Knowledge-Forge/pkg/vault"
)

// d5ProfileKeys are the profiles/me.md fields D5 carries. Every one has a fixed shape the
// wizard writes (an enum, a language name, a 1–5 depth, a list of framework names), which
// is what makes capturing them safe without a scrubber in the path. The four omitted
// fields — assume_known, never_assume, code_style, avoid — are free text the user writes
// by hand and can carry an employer's vocabulary; profiles/me.template.md invites exactly
// that ("as specific as you like"). Not capturing them beats scrubbing them on export.
var d5ProfileKeys = []string{"primary_language", "frameworks", "infra", "seniority",
	"default_depth", "note_language", "explain_style"}

// captureAccepted records the D5 style pair. The hook is the branch that just decided this
// draft may be published, which is the only acceptance signal in the tree — see
// pkg/dataset/d5.go for why that makes D5 a subset of accepted notes rather than a census.
// A capture error reaches stderr only: the gate verdict is already on stdout and a
// side-channel write must not turn a passing gate into a failing command.
func captureAccepted(cfg *config.Config, root string, draft *vault.Note) {
	if cfg == nil || !dataset.D5.Enabled(cfg.Dataset) {
		return
	}
	p := dataset.D5Pair{Kind: dataset.D5Kind, Topic: draftSlug(draft), Rel: draft.Rel,
		Note: string(draft.Body), Profile: readProfile(root), CapturedAt: time.Now()}
	if draft.FM != nil {
		p.Stack, p.Sources = draft.FM.List("stack"), sourceURLs(draft.FM)
	}
	if err := dataset.AppendD5(root, p); err != nil {
		fmt.Fprintf(os.Stderr, "forge gate: d5 capture: %v\n", err)
	}
}

// readProfile reads the conditioning fields out of <vault>/profiles/me.md. A missing file
// is the normal case on a vault where `forge init` has not run and returns nil rather than
// an error: an absent profile costs D5 a feature column, it does not invalidate the pair.
// Str and List are tried in that order because the wizard writes both shapes — frameworks
// and infra are sequences, the rest are scalars.
func readProfile(root string) map[string]string {
	n, err := vault.Load(filepath.Join(root, "profiles", "me.md"), "profiles/me.md")
	if err != nil || n.FM == nil {
		return nil
	}
	out := map[string]string{}
	for _, k := range d5ProfileKeys {
		if v := n.FM.Str(k); v != "" {
			out[k] = v
		} else if l := n.FM.List(k); len(l) > 0 {
			out[k] = strings.Join(l, ",")
		}
	}
	return out
}
