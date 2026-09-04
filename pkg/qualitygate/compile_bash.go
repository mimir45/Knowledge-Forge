package qualitygate

import (
	"bytes"
	"context"
	"os/exec"
)

// compileBash runs `bash -n` — a syntax check that never executes a single line of the
// script. bash -n has no concept of an unresolved dependency.
func compileBash(ctx context.Context, src []byte) CompileResult {
	if _, err := exec.LookPath("bash"); err != nil {
		return toolchainMissing("bash")
	}
	cmd := exec.CommandContext(ctx, "bash", "-n", "/dev/stdin")
	cmd.Stdin = bytes.NewReader(src)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err == nil {
		return CompileResult{Lang: "bash", Verdict: Pass}
	}
	diags := splitNonEmpty(stderr.String())
	kinds := make([]diagKind, len(diags))
	for i := range kinds {
		kinds[i] = kindSyntax
	}
	return CompileResult{
		Lang: "bash", Verdict: verdictFromDiagnostics(kinds),
		Diagnostics: diags, Detail: "bash -n reported a syntax error",
	}
}
