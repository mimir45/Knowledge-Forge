package qualitygate

import (
	"fmt"
	"time"

	"github.com/mimir45/Knowledge-Forge/pkg/config"
	"github.com/mimir45/Knowledge-Forge/pkg/vault"
)

// codeGate extracts every fenced code block from the draft body and runs CompileCheck
// on each. verify.run_code governs whether it runs at all: "never" and "ask" both
// produce Skipped (the binary never prompts — the CLI usage text says so already), only
// "auto" actually shells out. See compile.go's doc comment for the sandboxing guarantee
// this delegates to: no dependency resolution, no network, ever.
func codeGate(cfg *config.Config, draft *vault.Note) Outcome {
	if cfg.Verify.RunCode != "auto" {
		return Outcome{Gate: "code", Verdict: Skipped,
			Detail: "verify.run_code=" + cfg.Verify.RunCode + "; not run"}
	}
	blocks := vault.FencedBlocks(draft.Body)
	if len(blocks) == 0 {
		return Outcome{Gate: "code", Verdict: Skipped, Detail: "no fenced code blocks"}
	}
	langs := vault.FencedBlockLangs(draft.Body)
	return worstCodeResult(blocks, langs)
}

// worstCodeResult runs every block and keeps fail-dominates-skipped-dominates-pass, the
// same ordering compile.go uses within one snippet's diagnostics — here across blocks.
// langs[i] is blocks[i]'s fence info-string, e.g. "java" — DetectLang wants that tag, not
// the block's code-body content, so the two slices (both from links.go, same fenceRe
// walk, same order) are zipped by index rather than passing block content to DetectLang.
func worstCodeResult(blocks [][]byte, langs []string) Outcome {
	best := Outcome{Gate: "code", Verdict: Pass}
	for i, b := range blocks {
		lang := ""
		if i < len(langs) { // defensive only — same fenceRe walk guarantees equal length today
			lang = langs[i]
		}
		r := CompileCheck(DetectLang(lang), b, 15*time.Second)
		o := Outcome{Gate: "code", Verdict: r.Verdict,
			Detail: fmt.Sprintf("block %d (%s): %s", i, r.Lang, r.Detail)}
		if worseCode(o.Verdict, best.Verdict) {
			best = o
		}
	}
	if best.Verdict == Fail {
		best.Remedy = DropConfidence
	}
	return best
}

func worseCode(a, b Verdict) bool {
	rank := map[Verdict]int{Pass: 0, Skipped: 1, Fail: 2}
	return rank[a] > rank[b]
}
