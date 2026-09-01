package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/mimir45/Knowledge-Forge/pkg/config"
	"github.com/mimir45/Knowledge-Forge/pkg/dataset"
)

const exportDatasetUsage = `usage: forge export-dataset --set d1|d2|d3|d4|d5|d6 --out DIR
                            [--format jsonl-sft|jsonl-dpo|csv] [--since YYYY-MM-DD]
                            [--anonymize | --no-anonymize] [--vault DIR]

Exports one dataset (ADDENDUM D.1) as a training corpus, with a datasheet beside it.
Zero model calls. Everything stays on this machine: the files land in --out and nothing
transmits them, which is why the export itself is recorded in .forge/exports.jsonl
rather than merely happening.

Not every (set, format) pair exists. jsonl-dpo needs a chosen/rejected pair in the data
and only d3 and d4 have one; csv is d1's alone. An undefined combination is refused with
a message naming what is valid — the alternative is inventing an output shape the data
does not support.

d6 is different from d1-d5: it is derived, not captured — recomputed on
every export from the vault's code citations against whatever
.forge/code-index-<repo>.json caches this machine holds, not accumulated over time by a
capture path. --since is refused for it (no per-record timestamp), and so is
--anonymize (repo/path/symbol names are the whole feature and cannot be redacted without
destroying it) — since dataset.anonymize_on_export defaults true, exporting d6 at all
means passing --no-anonymize explicitly.

Anonymization is on by default, from dataset.anonymize_on_export, for every other set.
--anonymize forces it on; --no-anonymize turns it off and says so loudly, because the
result is a file of raw captured text. That is a separate path from a scrubber failure:
if redaction itself fails the whole export aborts and --out is never created.

Exit 0 = success. Exit 2 = usage error. Exit 3 = export failed, --out untouched.
`

func cmdExportDataset(args []string) int {
	fs := flag.NewFlagSet("forge export-dataset", flag.ContinueOnError)
	vaultDir := fs.String("vault", "", "vault root; defaults to config vault_path, then .")
	set := fs.String("set", "", "dataset: d1 d2 d3 d4 d5 (captured) or d6 (derived)")
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
//
// It does not print the raw-export warning itself — runExportDataset does, and only
// once dataset.Export has actually written something. Printing it here fired even when
// the request went on to be refused for --since or an undefined format (a d6 --since
// request that also passed --no-anonymize shouted "will contain captured text" about a
// file that was never created), which is the same defect writeDistribution's own doc
// comment names one layer down: describing content a file does not hold.
func anonymizeChoice(cfg *config.Config, anon, noAnon bool) (bool, int) {
	if anon && noAnon {
		fmt.Fprintln(os.Stderr,
			"forge export-dataset: --anonymize and --no-anonymize are mutually exclusive")
		return false, 2
	}
	if noAnon {
		return false, 0
	}
	return anon || cfg.Dataset.AnonymizeOnExport, 0
}

// warnRawExport describes what an unanonymized export of set actually holds. d6 gets its
// own enumeration: a record is {repo, path, symbol, note}, never a note body, draft,
// critique or URL, and unlike d1-d5 this is not a choice the flag made — --no-anonymize
// is d6's only working path, so this fires on every successful d6 export.
func warnRawExport(set string) {
	fmt.Fprintln(os.Stderr, "forge export-dataset: WARNING --no-anonymize was given.")
	if set == dataset.D6Tag {
		fmt.Fprintln(os.Stderr, "  The export will contain repo, path and symbol names, and")
		fmt.Fprintln(os.Stderr, "  the vault-relative path of the note citing each, exactly as")
		fmt.Fprintln(os.Stderr, "  they are — d6 cannot be anonymized.")
	} else {
		fmt.Fprintln(os.Stderr, "  The export will contain captured text exactly as recorded:")
		fmt.Fprintln(os.Stderr, "  note bodies, drafts, critiques, source URLs and file paths.")
	}
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
	if !rep.Anonymized {
		warnRawExport(rep.Set) // only once something was actually written, see its doc comment
	}
	b, _ := json.MarshalIndent(rep, "", "  ")
	fmt.Println(string(b))
	if rep.Records == 0 {
		fmt.Fprintln(os.Stderr, zeroRecordsMsg(o.Set))
	}
	return 0
}

// zeroRecordsMsg distinguishes d6's empty case from every other tier's. "Has captured
// nothing yet" is the right read for a write path nobody has fired; d6 has no write
// path, so an empty export more often means no repo has been indexed on this machine at
// all — pointing the user at forge logback is the actionable version of that.
func zeroRecordsMsg(set string) string {
	if set == dataset.D6Tag {
		return "forge export-dataset: d6 resolved no citations against a cached code index; " +
			"run forge logback (or forge check/drift) against the repos this vault cites first"
	}
	return fmt.Sprintf("forge export-dataset: %s has captured nothing yet", set)
}
