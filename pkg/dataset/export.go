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

// formatsFor is the export matrix.
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
	// D1Joined is D1-only: how many of Records joined a gate outcome by run_id.
	D1Joined  int    `json:"d1_joined,omitempty"`
	OutFile   string `json:"out_file"`
	Datasheet string `json:"datasheet"`
}

// Export reads one tier, filters, anonymizes, renders, and writes — in that order, with
// every byte held in memory until the whole run has succeeded.
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

// ExportLogPath records every export; in the same principle, "nothing leaves the
// machine without an explicit command".
const ExportLogPath = ".forge/exports.jsonl"

// UsageError marks a request that was malformed rather than an export that failed.
type UsageError struct{ error }

func usageErr(format string, a ...any) error { return UsageError{fmt.Errorf(format, a...)} }

// resolve turns --set and --format into a tier.
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

// refuseDerivedOptions rejects --since and --anonymize on a derived tier before a
// record is read.
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

// commit writes the export log, the export, and its datasheet, in that order.
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
