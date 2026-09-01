package dataset

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Format is an export's output shape.
type Format string

const (
	FormatSFT Format = "jsonl-sft"
	FormatDPO Format = "jsonl-dpo"
	FormatCSV Format = "csv"
)

// formatsFor is the export matrix, and the reason it is a table rather than a fallthrough
// is that the alternative to refusing a combination is inventing one.
//
// DPO needs a (chosen, rejected) pair that the data actually contains. D3 has one — the
// generated note was rejected, the human's edit was preferred — and so does D4, where the
// gate itself rejected the failing draft. D1 and D5 have no rejected side at all.
//
// D2 is the interesting exclusion, because ADDENDUM §D.1's own table marks it "DPO / SFT"
// and this phase's plan repeated that. The table's D2 pair is written as
// "draft → advisor critique → accepted patch", three parts, and the patch is the chosen
// side. What Phase 3b actually captures is D2Pair{Draft, Critique} — two parts. Nothing
// records whether a patch was accepted or what it contained, so the chosen side does not
// exist in the data, and emitting DPO here would mean nominating the critique as the
// preferred continuation of the draft: a fabricated preference. If capture ever grows the
// patch field, add FormatDPO to this row and nothing else has to change.
//
// CSV is D1's alone: it is the one tier whose record is a fixed-width feature row over a
// categorical label. Flattening a note body into a CSV cell is not a format, it is a
// quoting problem.
// D6 gets FormatSFT only: it has no chosen/rejected pair for DPO (there is nothing to
// prefer between — every record is one resolved citation), and it is not a fixed-width
// feature row the way D1 is, so CSV would be the same quoting problem D1's own comment
// above rules out.
var formatsFor = map[string][]Format{
	D1Tag: {FormatSFT, FormatCSV},
	D2Tag: {FormatSFT},
	D3Tag: {FormatSFT, FormatDPO},
	D4Tag: {FormatSFT, FormatDPO},
	D5Tag: {FormatSFT},
	D6Tag: {FormatSFT},
}

// ExportOptions is one export request. Since zero means no lower bound.
type ExportOptions struct {
	Set       string
	Format    Format
	Since     time.Time
	Anonymize bool
	Out       string
}

// ExportReport is what the caller prints and what the datasheet is rendered from. Every
// count describes the file that was written, not the file that was read.
type ExportReport struct {
	Set         string         `json:"set"`
	Kind        string         `json:"kind"`
	Format      string         `json:"format"`
	Records     int            `json:"records"`
	Available   int            `json:"available"`
	DroppedBy   string         `json:"dropped_by,omitempty"`
	Anonymized  bool           `json:"anonymized"`
	Redactions  int            `json:"redactions"`
	From        time.Time      `json:"from"`
	To          time.Time      `json:"to"`
	Stacks      map[string]int `json:"stacks,omitempty"`
	EngineTrail map[string]int `json:"engine_trail,omitempty"`
	// D1Joined is D1-only: how many of Records joined a gate outcome by
	// run_id, so the datasheet can say what fraction of the corpus is genuinely
	// outcome-labelled rather than implying the whole export is.
	D1Joined  int    `json:"d1_joined,omitempty"`
	OutFile   string `json:"out_file"`
	Datasheet string `json:"datasheet"`
}

// Export reads one tier, filters, anonymizes, renders, and writes — in that order, with
// every byte held in memory until the whole run has succeeded. That buffer-then-commit
// shape is pkg/scrub's (scrub.go's `file` slice): it means
// on any failure --out is left exactly as it was, rather than holding a partial export
// whose unwritten tail is the part nobody scrubbed.
func Export(vaultRoot string, o ExportOptions) (ExportReport, error) {
	tier, err := resolve(o.Set, o.Format)
	if err != nil {
		return ExportReport{}, err
	}
	if err := refuseDerivedOptions(tier, o); err != nil {
		return ExportReport{}, err
	}
	recs, err := loadTier(vaultRoot, tier)
	if err != nil {
		return ExportReport{}, err
	}
	rep, body, err := prepare(tier, recs, o)
	if err != nil {
		return ExportReport{}, err
	}
	return rep, commit(vaultRoot, o.Out, rep, body)
}

// ExportLogPath records every export, which ADDENDUM §D.4 requires in the same breath as
// "nothing leaves the machine without an explicit command" — the log is what makes that
// claim checkable afterwards. It holds the report, never the exported records.
const ExportLogPath = ".forge/exports.jsonl"

// UsageError marks a request that was malformed rather than an export that failed. The
// CLI exits 2 on it rather than 3, and the distinction is not cosmetic: exit 3 promises
// "a real attempt was made and --out is untouched", which would be a confusing thing to
// say about a combination that was rejected before a single record was read.
type UsageError struct{ error }

func usageErr(format string, a ...any) error { return UsageError{fmt.Errorf(format, a...)} }

// resolve turns --set and --format into a tier, refusing an unknown set and a combination
// the matrix does not define. Both messages name what is valid, since a user who reached
// here guessed once already.
func resolve(set string, f Format) (Tier, error) {
	for _, t := range Tiers() {
		if t.Tag == set {
			return t, checkFormat(t, f)
		}
	}
	return Tier{}, usageErr("unknown --set %q: want one of %s", set, tagList())
}

// tagList reads Tiers() rather than restating it, so a set added there does not also
// need updating in this message by hand.
func tagList() string {
	tags := make([]string, len(Tiers()))
	for i, t := range Tiers() {
		tags[i] = t.Tag
	}
	return strings.Join(tags, " ")
}

// refuseDerivedOptions rejects --since and --anonymize on a derived tier before a record
// is read, keyed off Derived rather than D6Tag so a future derived tier inherits both
// refusals automatically. Neither flag has a safe silent behaviour here: --since has no
// per-record timestamp to filter on (a filtered-looking export that never filtered), and
// no redaction of a derived set's whole reason for existing — its repo/path/symbol
// content — has been found that leaves it useful.
func refuseDerivedOptions(t Tier, o ExportOptions) error {
	if !t.Derived {
		return nil
	}
	if !o.Since.IsZero() {
		return usageErr("--since has no meaning for --set %s: it is a derived set with no "+
			"per-record timestamp, so filtering it would silently do nothing", t.Tag)
	}
	if o.Anonymize {
		return usageErr("--anonymize is refused for --set %s: repo, path and symbol names "+
			"are the whole feature, and no redaction of them has been found that leaves the "+
			"export useful; pass --no-anonymize and treat the result as raw", t.Tag)
	}
	return nil
}

func checkFormat(t Tier, f Format) error {
	for _, ok := range formatsFor[t.Tag] {
		if ok == f {
			return nil
		}
	}
	return usageErr("--format %s is not defined for --set %s (valid: %v); refusing rather "+
		"than inventing an output shape the data does not support", f, t.Tag, formatsFor[t.Tag])
}

// prepare does everything that can fail before anything is written.
func prepare(t Tier, recs []any, o ExportOptions) (ExportReport, []byte, error) {
	rep := ExportReport{Set: t.Tag, Kind: t.Kind, Format: string(o.Format),
		Available: len(recs), Anonymized: o.Anonymize}
	kept := since(recs, o.Since)
	if !o.Since.IsZero() && len(kept) < len(recs) {
		rep.DroppedBy = fmt.Sprintf("--since %s dropped %d of %d records",
			o.Since.Format("2006-01-02"), len(recs)-len(kept), len(recs))
	}
	if o.Anonymize {
		kept, rep.Redactions = anonymizeAll(kept)
	}
	if err := roundTripAll(kept); err != nil {
		return ExportReport{}, nil, err
	}
	body, err := render(t, kept, o.Format)
	return summarize(rep, kept, o), body, err
}

func since(recs []any, cut time.Time) []any {
	if cut.IsZero() {
		return recs
	}
	out := make([]any, 0, len(recs))
	for _, r := range recs {
		if !stampOf(r).Before(cut) {
			out = append(out, r)
		}
	}
	return out
}

func anonymizeAll(recs []any) ([]any, int) {
	out, total := make([]any, len(recs)), 0
	for i, r := range recs {
		red, n := anonymizeRecord(r)
		out[i], total = red, total+n
	}
	return out, total
}

// commit writes the export log, the export, and its datasheet, in that order, and is the
// only function here that touches the filesystem. Reached only once every record has been
// redacted and re-validated.
//
// The log goes first on purpose: if it cannot be written the export aborts with --out
// untouched, so an export that happened without being recorded is not reachable. The
// converse — a logged export whose file write then failed — over-records, which is the
// harmless direction.
func commit(vaultRoot, out string, rep ExportReport, body []byte) error {
	if err := appendJSONL(filepath.Join(vaultRoot, ExportLogPath), rep); err != nil {
		return fmt.Errorf("export log: %w", err)
	}
	if err := os.MkdirAll(out, 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(out, filepath.Base(rep.OutFile)), body, 0o644); err != nil {
		return err
	}
	sheet := []byte(renderDatasheet(rep))
	return os.WriteFile(filepath.Join(out, filepath.Base(rep.Datasheet)), sheet, 0o644)
}
