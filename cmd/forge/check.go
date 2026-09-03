package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/mimir45/Knowledge-Forge/pkg/config"
	"github.com/mimir45/Knowledge-Forge/pkg/report"
)

// checkCfg is the weekly pass's inputs.
type checkCfg struct {
	vault        string
	repos        repoList
	months       int
	days         int
	offline      bool
	dupThreshold float64
	config       *config.Config // nil in tests that build a checkCfg by hand; cost.md degrades then
}

func cmdCheck(args []string) int {
	var cfg checkCfg
	fs := flag.NewFlagSet("forge check", flag.ContinueOnError)
	fs.StringVar(&cfg.vault, "vault", "", "vault root; defaults to config vault_path, then .")
	fs.Var(&cfg.repos, "repo", "code repository as name=path (repeatable)")
	fs.IntVar(&cfg.months, "months", 0, "vault history window for churn.md; 0 reads all of it")
	fs.IntVar(&cfg.days, "days", 0, "code churn window in days; 0 uses config check.churn_days")
	fs.BoolVar(&cfg.offline, "offline", false,
		"skip the network; deadlinks.md then reports cached verdicts only")
	fs.Usage = func() { fmt.Fprint(os.Stderr, checkUsage); fs.PrintDefaults() }
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if code := cfg.applyConfig(); code != 0 {
		return code
	}
	return runWeekly(cfg)
}

// applyConfig fills what the flags left at zero. check is the one command that reads
// more than the vault path out of the chain.
func (c *checkCfg) applyConfig() int {
	root, code := vaultOrExit("check", c.vault)
	if code != 0 {
		return code
	}
	c.vault = root
	cfg, code := configOrExit("check")
	if code != 0 {
		return code
	}
	if c.days == 0 {
		c.days = orDefaultInt(cfg.Check.ChurnDays, 90)
	}
	c.dupThreshold = cfg.Check.DuplicateThreshold
	c.config = cfg
	return 0
}

func orDefaultInt(v, def int) int {
	if v > 0 {
		return v
	}
	return def
}

const checkUsage = `usage: forge check [--vault DIR] [--repo NAME=PATH] [--months N] [--days N]
                   [--offline]

The weekly pass. Collects the vault once and renders the nine vault-health reports
plus cost.md (Phase 3b) into <vault>/reports/, the codebase map into
<vault>/moc/codebase.md, and the weekly rollup into <vault>/moc/weekly/<ISO-week>.md.
Zero model calls, like everything else in this binary.

Every report is rendered independently: a renderer that fails costs its own file and
nothing else, and the run says which files it wrote and which it skipped. Headers carry
a date rather than a timestamp, so a second run on the same day rewrites nothing and
the vault's git diff stays readable.

Without --repo, drift.md and moc/codebase.md have nothing to check against and are
skipped rather than written empty. --offline is for a run on a bad network: an
unreachable URL is not a dead one, and deadlinks.md counts the two apart.

`

func runWeekly(cfg checkCfg) int {
	root, err := filepath.Abs(cfg.vault)
	if err != nil {
		fmt.Fprintf(os.Stderr, "forge check: %v\n", err)
		return 2
	}
	data, err := collectVault(cfg, root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "forge check: %v\n", err)
		return 1
	}
	code := writeAll(root, jobs(cfg, data))
	aiPass(data)
	drainAdvisorQueue(data)
	return code
}

// job is one report: where it goes and how to build it.
type job struct {
	rel    string // vault-relative destination
	render func() ([]byte, error)
}

func writeAll(root string, js []job) int {
	written, skipped := 0, 0
	for _, j := range js {
		md, err := safeRender(j.render)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  skipped %s: %v\n", j.rel, err)
			skipped++
			continue
		}
		changed, err := writeReport(filepath.Join(root, j.rel), md)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  skipped %s: %v\n", j.rel, err)
			skipped++
			continue
		}
		fmt.Printf("  %-24s %d bytes%s\n", j.rel, len(md), unchangedNote(changed))
		written++
	}
	fmt.Printf("\n%d written, %d skipped\n", written, skipped)
	return 0
}

// safeRender turns a panicking renderer into one failed file.
func safeRender(f func() ([]byte, error)) (md []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("renderer panicked: %v", r)
		}
	}()
	return f()
}

// writeReport creates the parent directory and skips the write when the bytes already
// match, for the same reason forge index does: an identical rewrite still bumps mtime.
func writeReport(path string, md []byte) (changed bool, err error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return false, err
	}
	if old, err := os.ReadFile(path); err == nil && string(old) == string(md) {
		return false, nil
	}
	return true, os.WriteFile(path, md, 0o644)
}

func unchangedNote(changed bool) string {
	if changed {
		return ""
	}
	return " (unchanged)"
}

// jobs lists the nine vault-health reports, cost.md, the codebase map.
func jobs(cfg checkCfg, d *checkData) []job {
	js := []job{
		{"reports/coverage.md", d.coverage},
		{"reports/staleness.md", d.staleness},
		{"reports/duplicates.md", d.duplicates},
		{"reports/orphans.md", d.orphans},
		{"reports/gaps.md", d.gaps},
		{"reports/graph-health.md", d.graphHealth},
		{"reports/churn.md", d.churn},
		{"reports/deadlinks.md", d.deadlinks},
		{"reports/cost.md", d.cost},
	}
	if len(cfg.repos) > 0 {
		js = append(js, job{"reports/drift.md", d.drift}, job{"moc/codebase.md", d.codebase})
	}
	// weekly runs unconditionally, unlike drift/codebase — it has no --repo dependency
	// and degrades gracefully when there is nothing to act on.
	return append(js, job{"moc/weekly/" + report.WeekKey(d.now) + ".md", d.weekly})
}
