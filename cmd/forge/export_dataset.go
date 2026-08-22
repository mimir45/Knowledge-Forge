package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"time"

	"knowledge-forge/pkg/config"
	"knowledge-forge/pkg/dataset"
)

const exportDatasetUsage = `usage: forge export-dataset --set d1|d2|d3|d4|d5 --out DIR
                            [--format jsonl-sft|jsonl-dpo|csv] [--since YYYY-MM-DD]
                            [--anonymize | --no-anonymize] [--vault DIR]

Exports one capture tier (ADDENDUM D.1) as a training corpus, with a datasheet beside
it. Zero model calls. Everything stays on this machine: the files land in --out and
nothing transmits them, which is why the export itself is recorded in
.forge/exports.jsonl rather than merely happening.

Not every (set, format) pair exists. jsonl-dpo needs a chosen/rejected pair in the data
and only d3 and d4 have one; csv is d1's alone. An undefined combination is refused with
a message naming what is valid — the alternative is inventing an output shape the data
does not support.

Anonymization is on by default, from dataset.anonymize_on_export. --anonymize forces it
on; --no-anonymize turns it off and says so loudly, because the result is a file of raw
captured text. That is a separate path from a scrubber failure: if redaction itself
fails the whole export aborts and --out is never created.

Exit 0 = success. Exit 2 = usage error. Exit 3 = export failed, --out untouched.
`

func cmdExportDataset(args []string) int {
	fs := flag.NewFlagSet("forge export-dataset", flag.ContinueOnError)
	vaultDir := fs.String("vault", "", "vault root; defaults to config vault_path, then .")
	set := fs.String("set", "", "capture tier: d1 d2 d3 d4 or d5")
	format := fs.String("format", string(dataset.FormatSFT), "jsonl-sft, jsonl-dpo or csv")
	since := fs.String("since", "", "keep records on or after this date (YYYY-MM-DD)")
	out := fs.String("out", "", "directory to write the export and its datasheet into")
	anon := fs.Bool("anonymize", false, "force anonymization on")
	noAnon := fs.Bool("no-anonymize", false, "export raw captured text (prints a warning)")
	fs.Usage = func() { fmt.Fprint(os.Stderr, exportDatasetUsage); fs.PrintDefaults() }
	if err := fs.Parse(args); err != nil {
		return 2
	}
	o, code := exportOptions(*set, *format, *since, *out, *anon, *noAnon)
	if code != 0 {
		return code
	}
	root, code := vaultOrExit("export-dataset", *vaultDir)
	if code != 0 {
		return code
	}
	return runExportDataset(root, o)
}

// exportOptions turns flags into an ExportOptions, resolving the anonymize default from
// the config chain. Reading config here rather than in runExportDataset keeps the usage
// errors above the vault resolution, so a typo'd --set fails on the typo.
func exportOptions(set, format, since, out string,
	anon, noAnon bool) (dataset.ExportOptions, int) {
	if set == "" || out == "" {
		fmt.Fprintln(os.Stderr, "forge export-dataset: --set and --out are both required")
		return dataset.ExportOptions{}, 2
	}
	cut, err := parseSince(since)
	if err != nil {
		fmt.Fprintf(os.Stderr, "forge export-dataset: %v\n", err)
		return dataset.ExportOptions{}, 2
	}
	cfg, code := configOrExit("export-dataset")
	if code != 0 {
		return dataset.ExportOptions{}, code
	}
	on, code := anonymizeChoice(cfg, anon, noAnon)
	return dataset.ExportOptions{Set: set, Format: dataset.Format(format), Since: cut,
		Anonymize: on, Out: out}, code
}

func parseSince(s string) (time.Time, error) {
	if s == "" {
		return time.Time{}, nil
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return time.Time{}, fmt.Errorf("--since %q is not a YYYY-MM-DD date", s)
	}
	return t, nil
}

// anonymizeChoice resolves the three inputs — the config default, --anonymize,
// --no-anonymize — into one boolean, and is where this phase's most consequential
// decision is spelled out. Turning anonymization off is a choice the user makes; a
// scrubber that fails is not, and the two must never share a code path, because that is
// exactly how a warning quietly becomes a raw export (AUDIT 8.4 D-6).
func anonymizeChoice(cfg *config.Config, anon, noAnon bool) (bool, int) {
	if anon && noAnon {
		fmt.Fprintln(os.Stderr,
			"forge export-dataset: --anonymize and --no-anonymize are mutually exclusive")
		return false, 2
	}
	if noAnon {
		warnRawExport()
		return false, 0
	}
	return anon || cfg.Dataset.AnonymizeOnExport, 0
}

func warnRawExport() {
	fmt.Fprintln(os.Stderr, "forge export-dataset: WARNING --no-anonymize was given.")
	fmt.Fprintln(os.Stderr, "  The export will contain captured text exactly as recorded:")
	fmt.Fprintln(os.Stderr, "  note bodies, drafts, critiques, source URLs and file paths.")
	fmt.Fprintln(os.Stderr, "  Do not share the result. The datasheet records that this run")
	fmt.Fprintln(os.Stderr, "  was not anonymized.")
}

func runExportDataset(root string, o dataset.ExportOptions) int {
	rep, err := dataset.Export(root, o)
	if err != nil {
		fmt.Fprintf(os.Stderr, "forge export-dataset: %v\n", err)
		var bad dataset.UsageError
		if errors.As(err, &bad) {
			return 2 // nothing was attempted, so the exit-3 promise would misdescribe it
		}
		fmt.Fprintf(os.Stderr, "forge export-dataset: no export file was written to %s\n", o.Out)
		return 3
	}
	b, _ := json.MarshalIndent(rep, "", "  ")
	fmt.Println(string(b))
	if rep.Records == 0 {
		fmt.Fprintf(os.Stderr, "forge export-dataset: %s has captured nothing yet\n", o.Set)
	}
	return 0
}
