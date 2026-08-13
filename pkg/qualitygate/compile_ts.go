package qualitygate

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
)

// tsUnresolvedCodes are the TS diagnostic codes a missing import or missing @types
// package produces — a sandbox-classpath limitation, not a defect in the snippet.
// Everything else (in particular the TS1xxx parser-error range) is a real problem.
var tsUnresolvedCodes = map[string]bool{"TS2307": true, "TS2304": true}

var tsDiagCodeRe = regexp.MustCompile(`error (TS\d+):`)

// compileTS runs `tsc --noEmit --skipLibCheck` against one file in a throwaway temp
// dir with no node_modules present, so it never resolves a real npm dependency — only
// the language's own syntax and (via the code map above) which failures are ours to
// call a defect versus an artifact of having no package installed.
func compileTS(ctx context.Context, src []byte) CompileResult {
	if _, err := exec.LookPath("tsc"); err != nil {
		return toolchainMissing("tsc")
	}
	dir, err := os.MkdirTemp("", "forge-verify-ts-")
	if err != nil {
		return CompileResult{Lang: "ts", Verdict: Skipped, Detail: "temp dir: " + err.Error()}
	}
	defer os.RemoveAll(dir)
	file := filepath.Join(dir, "snippet.ts")
	if err := os.WriteFile(file, src, 0o644); err != nil {
		return CompileResult{Lang: "ts", Verdict: Skipped, Detail: "write source: " + err.Error()}
	}
	return runTSC(ctx, file)
}

func runTSC(ctx context.Context, file string) CompileResult {
	cmd := exec.CommandContext(ctx, "tsc", "--noEmit", "--skipLibCheck", "--strict", file)
	var out bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &out
	if err := cmd.Run(); err == nil {
		return CompileResult{Lang: "ts", Verdict: Pass}
	}
	diags := stripDir(splitNonEmpty(out.String()), filepath.Dir(file))
	kinds := make([]diagKind, 0, len(diags))
	for _, d := range diags {
		kinds = append(kinds, classifyTSLine(d))
	}
	return CompileResult{
		Lang: "ts", Verdict: verdictFromDiagnostics(kinds),
		Diagnostics: diags, Detail: "tsc reported diagnostics",
	}
}

func classifyTSLine(d string) diagKind {
	m := tsDiagCodeRe.FindStringSubmatch(d)
	if m == nil {
		return kindOther
	}
	switch {
	case tsUnresolvedCodes[m[1]]:
		return kindUnresolved
	case m[1][:3] == "TS1": // the parser's own error range
		return kindSyntax
	default:
		return kindOther
	}
}
