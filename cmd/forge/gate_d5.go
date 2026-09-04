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

// d5ProfileKeys are the profiles/me.md fields D5 carries.
var d5ProfileKeys = []string{"primary_language", "frameworks", "infra", "seniority",
	"default_depth", "note_language", "explain_style"}

// captureAccepted records the D5 style pair.
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

// readProfile reads the conditioning fields out of <vault>/profiles/me.md.
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
