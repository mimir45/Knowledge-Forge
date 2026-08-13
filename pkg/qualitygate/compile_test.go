package qualitygate

import (
	"os/exec"
	"testing"
	"time"
)

func TestCompileBashPass(t *testing.T) {
	r := CompileCheck("bash", []byte("echo hi\n"), 5*time.Second)
	if r.Verdict != Pass {
		t.Errorf("verdict = %v, want Pass; diagnostics: %v", r.Verdict, r.Diagnostics)
	}
}

func TestCompileBashSyntaxErrorFails(t *testing.T) {
	r := CompileCheck("bash", []byte("if true; then\n  echo hi\n"), 5*time.Second)
	if r.Verdict != Fail {
		t.Errorf("verdict = %v, want Fail; diagnostics: %v", r.Verdict, r.Diagnostics)
	}
}

func TestCompileUnsupportedLangIsSkipped(t *testing.T) {
	r := CompileCheck("python", []byte("print(1)"), 5*time.Second)
	if r.Verdict != Skipped {
		t.Errorf("verdict = %v, want Skipped", r.Verdict)
	}
}

// TestCompileJavaSyntaxDominatesUnresolvedImport pins the ordering invariant this
// package is built around: a genuine syntax error (missing closing brace) must fail
// the gate even when the same file also has an unresolved import — fail dominates
// skipped, never the reverse.
func TestCompileJavaSyntaxDominatesUnresolvedImport(t *testing.T) {
	if _, err := exec.LookPath("javac"); err != nil {
		t.Skip("javac not installed")
	}
	src := `import org.springframework.stereotype.Service;

public class Snippet {
    void broken() {
        if (true) {
`
	r := CompileCheck("java", []byte(src), 15*time.Second)
	if r.Verdict != Fail {
		t.Errorf("verdict = %v, want Fail; diagnostics: %v", r.Verdict, r.Diagnostics)
	}
}

func TestCompileJavaUnresolvedOnlyIsSkipped(t *testing.T) {
	if _, err := exec.LookPath("javac"); err != nil {
		t.Skip("javac not installed")
	}
	src := `import org.springframework.stereotype.Service;

public class Snippet {
    Service s;
}
`
	r := CompileCheck("java", []byte(src), 15*time.Second)
	if r.Verdict != Skipped {
		t.Errorf("verdict = %v, want Skipped; diagnostics: %v", r.Verdict, r.Diagnostics)
	}
}

func TestCompileJavaPass(t *testing.T) {
	if _, err := exec.LookPath("javac"); err != nil {
		t.Skip("javac not installed")
	}
	src := "public class Snippet {\n    int x = 1;\n}\n"
	r := CompileCheck("java", []byte(src), 15*time.Second)
	if r.Verdict != Pass {
		t.Errorf("verdict = %v, want Pass; diagnostics: %v", r.Verdict, r.Diagnostics)
	}
}

func TestCompileTSSkippedWhenToolchainAbsent(t *testing.T) {
	if _, err := exec.LookPath("tsc"); err == nil {
		t.Skip("tsc is installed; this test only covers the absent case")
	}
	r := CompileCheck("ts", []byte("const x: number = 1;\n"), 5*time.Second)
	if r.Verdict != Skipped {
		t.Errorf("verdict = %v, want Skipped", r.Verdict)
	}
}
