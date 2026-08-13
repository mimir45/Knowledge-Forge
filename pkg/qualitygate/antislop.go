package qualitygate

import (
	"bytes"
	"fmt"
	"strings"

	"knowledge-forge/pkg/config"
	"knowledge-forge/pkg/vault"
	"knowledge-forge/references"
)

// antislopGate parses references/writing-rules.md at run time rather than hardcoding the
// phrase list — see that file's own doc comment: the list grows from _inbox/ reject
// history without a recompile. cfg is accepted for symmetry with the other six gates and
// because a future config toggle (e.g. verify.antislop: off) is the obvious next knob;
// nothing in pkg/config defines one yet, so this gate always runs.
func antislopGate(cfg *config.Config, draft *vault.Note) Outcome {
	_ = cfg
	// Fences stripped first: "leverage" inside a Java snippet is someone's variable
	// name, not filler prose. Same rationale as vault.Wikilinks stripping fences
	// before it scans for [[links]] — code is not prose.
	body := strings.ToLower(string(vault.StripCode(draft.Body)))
	for _, phrase := range bannedPhrases() {
		if strings.Contains(body, phrase) {
			return Outcome{Gate: "antislop", Verdict: Fail, Remedy: RewritePass,
				Detail: fmt.Sprintf("banned phrase: %q", phrase)}
		}
	}
	if o, fail := structuralFail(draft); fail {
		return o
	}
	return Outcome{Gate: "antislop", Verdict: Pass}
}

// structuralFail checks writing-rules.md's one Go-enforced structural rule: howto/api
// notes need a demonstration, not just a claim. See that file's Structural requirements
// section for why the other five types are exempt.
func structuralFail(draft *vault.Note) (Outcome, bool) {
	if draft.FM == nil {
		return Outcome{}, false
	}
	typ := draft.FM.Str("type")
	if typ != "howto" && typ != "api" {
		return Outcome{}, false
	}
	if len(vault.FencedBlocks(draft.Body)) > 0 {
		return Outcome{}, false
	}
	return Outcome{Gate: "antislop", Verdict: Fail, Remedy: RewritePass,
		Detail: fmt.Sprintf("type %q requires at least one fenced code block", typ)}, true
}

// bannedPhrases parses the "## Banned phrases" bullet list out of the embedded
// writing-rules.md, stopping at the next "## " heading.
func bannedPhrases() []string {
	inSection := false
	var out []string
	for _, l := range bytes.Split(references.WritingRulesMD, []byte("\n")) {
		s := strings.TrimSpace(string(l))
		switch {
		case s == "## Banned phrases":
			inSection = true
		case strings.HasPrefix(s, "## ") && inSection:
			return out
		case inSection && strings.HasPrefix(s, "- "):
			out = append(out, strings.TrimSpace(strings.TrimPrefix(s, "- ")))
		}
	}
	return out
}
