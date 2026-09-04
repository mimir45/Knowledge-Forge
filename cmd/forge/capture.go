package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/mimir45/Knowledge-Forge/pkg/dataset"
)

func cmdCapture(args []string) int {
	fs := flag.NewFlagSet("forge capture", flag.ContinueOnError)
	vaultDir := fs.String("vault", "", "vault git repository; defaults to config vault_path, then .")
	commit := fs.String("commit", "HEAD", "commit to harvest")
	days := fs.Int("window-days", 7, "max days between generation and edit")
	out := fs.String("out", dataset.D3Path, "dataset file, relative to the vault root")
	dry := fs.Bool("dry-run", false, "report the pairs without writing them")
	quiet := fs.Bool("quiet", false, "print nothing when no pair was captured")
	fs.Usage = func() { fmt.Fprint(os.Stderr, captureUsage); fs.PrintDefaults() }
	if err := fs.Parse(args); err != nil {
		return 2
	}
	root, code := vaultOrExit("capture", *vaultDir)
	if code != 0 {
		return code
	}
	if !captureConsented(*quiet) {
		return 0
	}
	return runCapture(root, *commit, *out, *days, *dry, *quiet)
}

// captureConsented makes dataset.capture's list a real gate.
func captureConsented(quiet bool) bool {
	cfg, err := loadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "forge capture: config: %v — capture skipped\n", err)
		return false
	}
	if !dataset.D3.Enabled(cfg.Dataset) {
		if !quiet {
			fmt.Fprintln(os.Stderr,
				"forge capture: d3 capture is off (dataset.enabled / dataset.capture)")
		}
		return false
	}
	return true
}

const captureUsage = `usage: forge capture [--vault DIR] [--commit REV] [--window-days N]
                     [--out FILE] [--dry-run] [--quiet]

Harvests ADDENDUM D.1 human-correction pairs from one vault commit: notes forge
generated that a human edited within the window. Zero model calls; the verdict is a
pure function of the commit and the tree, so a rerun is a no-op rather than a
duplicate. Meant to run from the vault's post-commit hook.

`

func runCapture(vaultDir, commit, out string, days int, dry, quiet bool) int {
	root, err := filepath.Abs(vaultDir)
	if err != nil {
		return fail(err)
	}
	opt := dataset.Options{Commit: commit, Window: time.Duration(days) * 24 * time.Hour}
	pairs, err := dataset.Capture(dataset.Git{Dir: root}, opt)
	if err != nil {
		return fail(err)
	}
	if len(pairs) == 0 {
		if !quiet {
			fmt.Printf("no human-correction pairs in %s\n", commit)
		}
		return 0
	}
	return record(filepath.Join(root, out), pairs, dry)
}

func record(path string, pairs []dataset.Pair, dry bool) int {
	if dry {
		for _, p := range pairs {
			fmt.Printf("would capture %s (generated %s, origin %s)\n",
				p.Note, p.GeneratedAt.Format("2006-01-02"), p.Origin)
		}
		return 0
	}
	n, err := dataset.Append(path, pairs)
	if err != nil {
		return fail(err)
	}
	fmt.Printf("%d pair(s) captured, %d already present → %s\n",
		n, len(pairs)-n, filepath.Base(path))
	return 0
}

func fail(err error) int {
	fmt.Fprintf(os.Stderr, "forge capture: %v\n", err)
	return 1
}
