// Package qualitygate implements DESIGN §12's seven pre-write gates and the compile
// check ADDENDUM §B.2 calls forge verify-code.
//
// forge verify-code is a syntax/compile check, never a dependency resolver: it never
// runs `npm install`, never resolves Maven/Gradle coordinates, and never touches the
// network. It shells out to the system toolchain already on the machine (javac, tsc,
// bash -n), following the same precedent as pkg/gitsig (BACKLOG B-009: shell to the
// tool, don't embed it). A snippet that references a real dependency the sandbox does
// not have on its classpath is not a defect in the snippet — see verdictFromDiagnostics.
package qualitygate

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"
)

// Verdict is forge verify-code's three-valued outcome. ADDENDUM §B.2's honest
// capability boundary is the reason there are three, not two: T0 can prove "this does
// not parse" but can never prove "this is semantically correct against a classpath it
// was never given," so that case is Skipped, not Pass and not Fail.
type Verdict int

const (
	Pass Verdict = iota
	Fail
	Skipped
)

func (v Verdict) String() string {
	switch v {
	case Pass:
		return "pass"
	case Fail:
		return "fail"
	case Skipped:
		return "skipped"
	default:
		return "unknown"
	}
}

// MarshalJSON serializes the name, not the iota ordinal. Without this, forge
// verify-code's JSON output silently reorders itself every time this const block
// gains a value — a wire-format break no test would catch, since Go's default int
// encoding never errors, just changes meaning.
func (v Verdict) MarshalJSON() ([]byte, error) { return []byte(`"` + v.String() + `"`), nil }

// UnmarshalJSON is the counterpart, so gate_test.go and any golden-file test can
// round-trip a CompileResult/Outcome through JSON instead of only ever writing it.
func (v *Verdict) UnmarshalJSON(b []byte) error {
	switch strings.Trim(string(b), `"`) {
	case "pass":
		*v = Pass
	case "fail":
		*v = Fail
	case "skipped":
		*v = Skipped
	default:
		return fmt.Errorf("qualitygate: unknown Verdict %s", b)
	}
	return nil
}

// CompileResult is one forge verify-code outcome for a single snippet.
type CompileResult struct {
	Lang        string   `json:"lang"`
	Verdict     Verdict  `json:"verdict"`
	Diagnostics []string `json:"diagnostics,omitempty"`
	Detail      string   `json:"detail,omitempty"`
}

// diagKind classifies one compiler diagnostic for verdictFromDiagnostics.
type diagKind int

const (
	kindOther diagKind = iota
	kindSyntax
	kindUnresolved
)

// verdictFromDiagnostics applies the ordering invariant this package is built around:
// any syntax (or otherwise unclassified) diagnostic forces Fail, regardless of any
// unresolved-dependency diagnostic present in the same run. Skipped is reserved for the
// narrow case where every diagnostic is unresolved-dependency in nature — "fail
// dominates skipped," never the reverse.
func verdictFromDiagnostics(kinds []diagKind) Verdict {
	if len(kinds) == 0 {
		return Pass
	}
	sawUnresolved := false
	for _, k := range kinds {
		if k != kindUnresolved {
			return Fail
		}
		sawUnresolved = true
	}
	if sawUnresolved {
		return Skipped
	}
	return Pass
}

// CompileCheck dispatches to the language-specific checker under a timeout. lang must
// already be resolved — DetectLang below turns a fenced code block's info-string into
// one of the supported values, or "" when nothing recognized applies.
func CompileCheck(lang string, src []byte, timeout time.Duration) CompileResult {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	switch lang {
	case "bash", "sh":
		return compileBash(ctx, src)
	case "ts", "typescript":
		return compileTS(ctx, src)
	case "java":
		return compileJava(ctx, src)
	default:
		return CompileResult{Lang: lang, Verdict: Skipped, Detail: "unsupported language: " + lang}
	}
}

// DetectLang maps a fenced block's info-string (```java, ```ts, ```typescript, ```sh,
// ```bash) to a CompileCheck lang value. Kotlin and Python are recognized but
// unsupported — reported, not silently dropped from the note.
func DetectLang(infoString string) string {
	tag := strings.ToLower(strings.TrimSpace(strings.SplitN(infoString, " ", 2)[0]))
	switch tag {
	case "java", "ts", "typescript", "bash", "sh":
		return tag
	default:
		return tag // caller decides supported vs unsupported via CompileCheck's default
	}
}

func splitNonEmpty(s string) []string {
	var out []string
	for _, l := range strings.Split(s, "\n") {
		if t := strings.TrimSpace(l); t != "" {
			out = append(out, t)
		}
	}
	return out
}

// stripDir removes every occurrence of a throwaway temp dir's absolute path from a set
// of diagnostic lines. Without this, two identical runs of the same snippet produce
// different os.MkdirTemp paths embedded in Diagnostics, which breaks the B-020
// determinism convention (gate_test.go hashes Report.Outcomes across two runs). The
// snippet's basename survives — only the random directory prefix is noise.
func stripDir(lines []string, dir string) []string {
	out := make([]string, len(lines))
	for i, l := range lines {
		out[i] = strings.ReplaceAll(l, dir+string(filepath.Separator), "")
	}
	return out
}

func toolchainMissing(name string) CompileResult {
	return CompileResult{
		Lang: name, Verdict: Skipped,
		Detail: "toolchain not installed: " + name + " — syntax-checked only, not run",
	}
}
