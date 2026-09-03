// Package qualitygate implements seven pre-write gates and the compile check from
// docs/ARCHITECTURE.md §8.
package qualitygate

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"
)

// Verdict is forge verify-code's three-valued outcome.
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

// MarshalJSON serializes the name, not the iota ordinal.
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

// verdictFromDiagnostics applies the ordering invariant this package is built around.
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
// already be resolved.
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
// ```bash) to a CompileCheck lang value.
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
// of diagnostic lines.
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
