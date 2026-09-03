package qualitygate

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

var javaPublicClassRe = regexp.MustCompile(`public\s+(?:final\s+|abstract\s+)?(?:class|interface|enum)\s+(\w+)`)

// compileJava runs javac in a throwaway temp dir with no classpath beyond the JDK
// itself: no Maven/Gradle resolution, no network, ever (compile.go's doc comment).
func compileJava(ctx context.Context, src []byte) CompileResult {
	if _, err := exec.LookPath("javac"); err != nil {
		return toolchainMissing("javac")
	}
	dir, err := os.MkdirTemp("", "forge-verify-java-")
	if err != nil {
		return CompileResult{Lang: "java", Verdict: Skipped, Detail: "temp dir: " + err.Error()}
	}
	defer os.RemoveAll(dir)
	name := "Snippet"
	if m := javaPublicClassRe.FindSubmatch(src); m != nil {
		name = string(m[1])
	}
	file := filepath.Join(dir, name+".java")
	if err := os.WriteFile(file, src, 0o644); err != nil {
		return CompileResult{Lang: "java", Verdict: Skipped, Detail: "write source: " + err.Error()}
	}
	return runJavac(ctx, dir, file)
}

func runJavac(ctx context.Context, dir, file string) CompileResult {
	cmd := exec.CommandContext(ctx, "javac", "-d", dir, "-cp", "", file)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err == nil {
		return CompileResult{Lang: "java", Verdict: Pass}
	}
	diags := stripDir(splitNonEmpty(stderr.String()), dir)
	kinds := classifyJavac(diags)
	return CompileResult{
		Lang: "java", Verdict: verdictFromDiagnostics(kinds),
		Diagnostics: diags, Detail: "javac reported diagnostics",
	}
}

// classifyJavac only classifies the "error:" lines — javac's summary line ("N errors")
// and caret/context lines carry no diagnostic content of their own.
func classifyJavac(diags []string) []diagKind {
	var kinds []diagKind
	for _, d := range diags {
		if !strings.Contains(d, "error:") {
			continue
		}
		kinds = append(kinds, classifyJavacLine(d))
	}
	return kinds
}

func classifyJavacLine(d string) diagKind {
	switch {
	case strings.Contains(d, "cannot find symbol"),
		strings.Contains(d, "does not exist"),
		strings.Contains(d, "cannot access"):
		return kindUnresolved
	case strings.Contains(d, "expected"),
		strings.Contains(d, "illegal start of"),
		strings.Contains(d, "reached end of file while parsing"),
		strings.Contains(d, "not a statement"):
		return kindSyntax
	default:
		return kindOther
	}
}
