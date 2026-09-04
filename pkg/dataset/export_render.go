package dataset

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// sftLine and dpoLine are the two JSONL shapes. Kind rides along so a file that has been
// moved away from its datasheet still says which tier produced it.
type (
	sftLine struct {
		Kind       string `json:"kind"`
		ID         string `json:"id,omitempty"`
		Prompt     string `json:"prompt"`
		Completion string `json:"completion"`
		// Outcome is D1-only: nil for every other tier, omitted from their JSON.
		Outcome *bool `json:"outcome,omitempty"`
	}
	dpoLine struct {
		Kind     string `json:"kind"`
		ID       string `json:"id,omitempty"`
		Prompt   string `json:"prompt"`
		Chosen   string `json:"chosen"`
		Rejected string `json:"rejected"`
	}
)

// idOf gives a rendered line a stable handle back to the record it came from.
func idOf(rec any) string {
	switch p := rec.(type) {
	case D1Pair:
		return p.QHash
	case Pair:
		return p.Note
	case D5Pair:
		return p.Rel
	case D6Pair:
		return d6ID(p)
	}
	return ""
}

// d6ID has no single natural key the way the other tiers do — Note alone is not unique.
func d6ID(p D6Pair) string {
	if p.Symbol == "" {
		return p.Repo + ":" + p.Path
	}
	return p.Repo + ":" + p.Path + "#" + p.Symbol
}

func render(t Tier, recs []any, f Format) ([]byte, error) {
	switch f {
	case FormatSFT:
		return renderJSONL(t, recs, sftOf)
	case FormatDPO:
		return renderJSONL(t, recs, dpoOf)
	case FormatCSV:
		return renderCSV(recs)
	}
	return nil, fmt.Errorf("unknown --format %q", f)
}

func renderJSONL(t Tier, recs []any, line func(Tier, any) (any, error)) ([]byte, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	for i, r := range recs {
		l, err := line(t, r)
		if err != nil {
			return nil, fmt.Errorf("record %d: %w", i+1, err)
		}
		if err := enc.Encode(l); err != nil {
			return nil, err
		}
	}
	return buf.Bytes(), nil
}

func sftOf(t Tier, rec any) (any, error) {
	l := sftLine{Kind: t.Kind, ID: idOf(rec)}
	switch p := rec.(type) {
	case D1Pair:
		l.Prompt, l.Completion, l.Outcome = d1Prompt(p), p.Decision, p.Outcome
	case D2Pair:
		l.Prompt, l.Completion = p.Draft, p.Critique
	case Pair:
		l.Prompt, l.Completion = p.Generated, p.Preferred
	case D4Pair:
		l.Prompt, l.Completion = repairPrompt(p), p.FixedDraft
	case D5Pair:
		l.Prompt, l.Completion = d5Prompt(p), p.Note
	case D6Pair:
		l.Prompt, l.Completion = d6Prompt(p), p.Note
	default:
		return nil, fmt.Errorf("unhandled record type %T", rec)
	}
	return l, nil
}

// dpoOf is only reached for D3 and D4; formatsFor refuses every other tier before a
// record is read. See its doc comment for why D2 is not on that list.
func dpoOf(t Tier, rec any) (any, error) {
	l := dpoLine{Kind: t.Kind, ID: idOf(rec)}
	switch p := rec.(type) {
	case Pair:
		l.Prompt, l.Chosen, l.Rejected = p.Topic, p.Preferred, p.Generated
	case D4Pair:
		l.Prompt, l.Chosen, l.Rejected = p.GateError, p.FixedDraft, p.FailingDraft
	default:
		return nil, fmt.Errorf("--format %s has no shape for %T", FormatDPO, rec)
	}
	return l, nil
}

// d1Prompt renders the routing features as the lines a small model reads.
func d1Prompt(p D1Pair) string {
	return fields("topic", p.Topic, "stack", strings.Join(p.Stack, ", "),
		"top_score", strconv.FormatFloat(p.RecallTopScore, 'f', 3, 64),
		"candidates", strconv.Itoa(p.Candidates))
}

// d6Prompt renders the code side of the pair as a retrieval query.
func d6Prompt(p D6Pair) string {
	return fields("repo", p.Repo, "path", p.Path, "symbol", p.Symbol)
}

func d5Prompt(p D5Pair) string {
	kv := []string{"topic", p.Topic, "stack", strings.Join(p.Stack, ", "),
		"sources", strings.Join(p.Sources, " ")}
	for _, k := range sortedKeys(p.Profile) {
		kv = append(kv, k, p.Profile[k])
	}
	return fields(kv...)
}

func repairPrompt(p D4Pair) string {
	return p.FailingDraft + "\n\n--- gate error ---\n" + p.GateError
}

// fields joins alternating key/value pairs, dropping empty values so an absent profile
// field is absent rather than a blank line the model has to learn to ignore.
func fields(kv ...string) string {
	var b strings.Builder
	for i := 0; i+1 < len(kv); i += 2 {
		if kv[i+1] == "" {
			continue
		}
		fmt.Fprintf(&b, "%s: %s\n", kv[i], kv[i+1])
	}
	return b.String()
}

// renderCSV is D1's alone (see formatsFor). encoding/csv rather than hand-joining: a topic
// slug will not contain a comma today, but a redaction marker in a stack list could.
func renderCSV(recs []any) ([]byte, error) {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	if err := w.Write([]string{"topic", "stack", "top_score", "candidates", "decision", "outcome"}); err != nil {
		return nil, err
	}
	for i, r := range recs {
		p, ok := r.(D1Pair)
		if !ok {
			return nil, fmt.Errorf("record %d: --format %s has no shape for %T", i+1, FormatCSV, r)
		}
		if err := w.Write(csvRow(p)); err != nil {
			return nil, err
		}
	}
	w.Flush()
	return buf.Bytes(), w.Error()
}

func csvRow(p D1Pair) []string {
	return []string{p.Topic, strings.Join(p.Stack, " "),
		strconv.FormatFloat(p.RecallTopScore, 'f', 3, 64),
		strconv.Itoa(p.Candidates), p.Decision, outcomeStr(p.Outcome)}
}

// outcomeStr spells out the three states a CSV cell can't leave to a null.
func outcomeStr(o *bool) string {
	switch {
	case o == nil:
		return ""
	case *o:
		return "published"
	default:
		return "quarantined"
	}
}

// summarize fills the report's descriptive half from the records that will actually be
// written — after --since and after redaction.
func summarize(rep ExportReport, recs []any, o ExportOptions) ExportReport {
	rep.Records = len(recs)
	rep.From, rep.To = span(recs)
	rep.Stacks, rep.EngineTrail = distribution(recs, stacksOf), distribution(recs, trailOf)
	if rep.Set == D1Tag {
		rep.D1Joined = countD1Joined(recs)
	}
	base := rep.Set + "-" + string(o.Format)
	rep.OutFile, rep.Datasheet = base+extFor(o.Format), base+"-datasheet.md"
	return rep
}

func countD1Joined(recs []any) int {
	n := 0
	for _, r := range recs {
		if p, ok := r.(D1Pair); ok && p.Outcome != nil {
			n++
		}
	}
	return n
}

func extFor(f Format) string {
	if f == FormatCSV {
		return ".csv"
	}
	return ".jsonl"
}

func span(recs []any) (from, to time.Time) {
	for _, r := range recs {
		t := stampOf(r)
		if t.IsZero() {
			continue
		}
		if from.IsZero() || t.Before(from) {
			from = t
		}
		if t.After(to) {
			to = t
		}
	}
	return from, to
}

func distribution(recs []any, of func(any) []string) map[string]int {
	out := map[string]int{}
	for _, r := range recs {
		for _, v := range of(r) {
			if v != "" {
				out[v]++
			}
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func stacksOf(rec any) []string {
	switch p := rec.(type) {
	case D1Pair:
		return p.Stack
	case Pair:
		return p.Stack
	case D5Pair:
		return p.Stack
	}
	return nil
}

// trailOf reports which engine tiers produced a record.
func trailOf(rec any) []string {
	if p, ok := rec.(Pair); ok {
		return p.EngineTrail
	}
	return nil
}
