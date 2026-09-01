package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mimir45/Knowledge-Forge/pkg/coderef"
	"github.com/mimir45/Knowledge-Forge/pkg/drift"
	"github.com/mimir45/Knowledge-Forge/pkg/vault"
)

type repoList []drift.Repo

func (r *repoList) String() string { return "" }

func (r *repoList) Set(v string) error {
	name, path, ok := strings.Cut(v, "=")
	if !ok || name == "" || path == "" {
		return fmt.Errorf("want name=path, got %q", v)
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	*r = append(*r, drift.Repo{Name: name, Root: abs})
	return nil
}

type driftCfg struct {
	vault  string
	repos  repoList
	since  string
	deep   bool
	apply  bool
	asJSON bool
}

func cmdDrift(args []string) int {
	var cfg driftCfg
	fs := flag.NewFlagSet("forge drift", flag.ContinueOnError)
	fs.StringVar(&cfg.vault, "vault", "", "vault root; defaults to config vault_path, then .")
	fs.Var(&cfg.repos, "repo", "code repository as name=path (repeatable)")
	fs.StringVar(&cfg.since, "since-commit", "",
		"evaluate only files changed since this sha; empty checks every citation")
	fs.BoolVar(&cfg.deep, "deep", false,
		"re-index each repo at the note's verified-era revision to adjudicate missing symbols")
	fs.BoolVar(&cfg.apply, "apply", false,
		"move confidence and stamp drift_checked_at; without it the run is read-only")
	fs.BoolVar(&cfg.asJSON, "json", false, "emit findings as JSON")
	fs.Usage = func() { fmt.Fprint(os.Stderr, driftUsage); fs.PrintDefaults() }
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if len(cfg.repos) == 0 {
		fmt.Fprint(os.Stderr, "forge drift: at least one --repo is required\n")
		return 2
	}
	// resolveVault short-circuits on a non-empty flag, so the git hooks — which always
	// pass --vault — never touch the chain. That keeps the 100ms hook-path budget a
	// question about drift itself rather than about config I/O.
	root, code := vaultOrExit("drift", cfg.vault)
	if code != 0 {
		return code
	}
	cfg.vault = root
	return runDrift(cfg)
}

const driftUsage = `usage: forge drift --repo NAME=PATH [--vault DIR] [--since-commit SHA]
                   [--deep] [--apply] [--json]

Asks, per citation, whether the code a note cites still says what the note says it
says. Git-anchored: it reads the object store, never the working tree, so a
half-finished edit is not drift and a revert restores a demoted note on its own.

--since-commit is the cheap gate. Without it every citation is evaluated, which is
what the weekly check does; the git hooks always pass it.

`

func runDrift(cfg driftCfg) int {
	root, err := filepath.Abs(cfg.vault)
	if err != nil {
		fmt.Fprintf(os.Stderr, "forge drift: %v\n", err)
		return 2
	}
	notes, err := loadNotes(root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "forge drift: %v\n", err)
		return 1
	}
	rg, err := registryOf(cfg.repos)
	if err != nil {
		fmt.Fprintf(os.Stderr, "forge drift: %v\n", err)
		return 1
	}
	return runCheck(cfg, root, notes, rg)
}

func runCheck(cfg driftCfg, root string, notes []*vault.Note, rg *coderef.Registry) int {
	gate, err := gateOf(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "forge drift: %v\n", err)
		return 1
	}
	src := drift.NewGitSource(cfg.repos, filepath.Join(root, ".forge"))
	findings := drift.Check(driftNotes(notes), rg, src, gate, drift.Opts{Deep: cfg.deep})
	moved := applyIfAsked(cfg, root, notes, findings, src)
	if cfg.asJSON {
		return emitJSON(findings, moved)
	}
	emitText(findings, moved)
	return 0
}

func applyIfAsked(cfg driftCfg, root string, notes []*vault.Note, findings []drift.Finding,
	src drift.Source) []drift.Result {

	if !cfg.apply {
		return nil
	}
	sch, err := vault.LoadSchema()
	if err != nil {
		fmt.Fprintf(os.Stderr, "forge drift: %v\n", err)
		return nil
	}
	st := drift.OpenStore(filepath.Join(root, ".forge"))
	moved := drift.Apply(byRel(notes), findings, st, sch, src)
	if err := st.Save(); err != nil {
		fmt.Fprintf(os.Stderr, "forge drift: .forge: %v\n", err)
	}
	return moved
}

// driftNotes reads citations from both shapes: the canonical code_refs block new notes
// write, and inline code spans, which is all the migrated vault has.
func driftNotes(notes []*vault.Note) []drift.Note {
	out := make([]drift.Note, 0, len(notes))
	for _, n := range notes {
		d := drift.Note{Rel: n.Rel, Refs: coderef.FromBody(n.Rel, n.Body)}
		if n.FM != nil {
			d.Verified = n.FM.Str("verified")
			d.Refs = append(coderef.FromFrontmatter(n.Rel, n.FM.List("code_refs")), d.Refs...)
		}
		if len(d.Refs) > 0 {
			out = append(out, d)
		}
	}
	return out
}

func byRel(notes []*vault.Note) map[string]*vault.Note {
	m := make(map[string]*vault.Note, len(notes))
	for _, n := range notes {
		m[n.Rel] = n
	}
	return m
}

// scanRepos runs coderef.ScanRepo once per repository, keyed by name. registryOf and
// forge check's codebase collector both need every repo's file list — the registry to
// resolve citations, the collector to list files per module — so scanning once here and
// sharing the result avoids a second `git ls-tree` per repo on every run.
func scanRepos(repos []drift.Repo) (map[string]coderef.Repo, error) {
	out := make(map[string]coderef.Repo, len(repos))
	for _, r := range repos {
		scanned, err := coderef.ScanRepo(r.Name, r.Root, "HEAD")
		if err != nil {
			return nil, fmt.Errorf("%s: %w", r.Name, err)
		}
		out[r.Name] = scanned
	}
	return out, nil
}

// newRegistryFrom builds the registry in repos' order rather than the map's — coderef's
// registry breaks same-base-name ties by (repo, path) order, so scan order is part of its
// verdict and a map iteration would make that verdict non-deterministic.
func newRegistryFrom(repos []drift.Repo, scans map[string]coderef.Repo) *coderef.Registry {
	out := make([]coderef.Repo, 0, len(repos))
	for _, r := range repos {
		out = append(out, scans[r.Name])
	}
	return coderef.NewRegistry(out)
}

func registryOf(repos []drift.Repo) (*coderef.Registry, error) {
	scans, err := scanRepos(repos)
	if err != nil {
		return nil, err
	}
	return newRegistryFrom(repos, scans), nil
}

// gateOf unions the changed sets of every repository. Touched paths are not
// repo-qualified, so a path present in two repos over-includes rather than
// under-includes — the safe direction for a gate whose only job is to keep work off the
// hook path. Deleted paths are repo-qualified, because a BROKEN finding built from one
// needs a real repo to resolve against (B-028).
//
// It fails closed. An empty gate and an unresolvable anchor produce the same findings —
// none — but they mean opposite things, and the second is drift silently never running on
// a hook that by design never prints. So a sha no repository can resolve is an error, not
// a clean run.
func gateOf(cfg driftCfg) (*drift.Changed, error) {
	if cfg.since == "" {
		return nil, nil
	}
	gate := &drift.Changed{Touched: map[string]bool{}, Deleted: map[string]string{}}
	resolved, last := 0, error(nil)
	for _, r := range cfg.repos {
		files, err := coderef.ChangedFilesStatus(r.Root, cfg.since, "HEAD")
		if err != nil {
			last = fmt.Errorf("%s: %w", r.Name, err) // one bad anchor must not silence the others
			continue
		}
		resolved++
		for _, f := range files {
			gate.Touched[f.Path] = true
			if f.Deleted {
				gate.Deleted[f.Path] = r.Name
			}
		}
	}
	if resolved == 0 {
		return nil, fmt.Errorf("no repository could resolve --since-commit %s: %w", cfg.since, last)
	}
	return gate, nil
}

func emitJSON(findings []drift.Finding, moved []drift.Result) int {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(struct {
		Findings []drift.Finding `json:"findings"`
		Moved    []drift.Result  `json:"moved"`
	}{findings, moved}); err != nil {
		return 1
	}
	return 0
}

func emitText(findings []drift.Finding, moved []drift.Result) {
	counts := map[drift.Verdict]int{}
	for _, f := range findings {
		counts[f.Verdict]++
		if f.Verdict != drift.OK && f.Verdict != drift.Skipped {
			fmt.Printf("%-8s %s  %s\n", f.Verdict, f.Note, f.Reason)
		}
	}
	fmt.Printf("\n%d citations: %d ok, %d repaired, %d suspect, %d broken, %d skipped\n",
		len(findings), counts[drift.OK], counts[drift.Repaired], counts[drift.Suspect],
		counts[drift.Broken], counts[drift.Skipped])
	for _, m := range moved {
		fmt.Printf("%s %s: %s -> %s\n", m.Action, m.Slug, m.From, m.To)
	}
}
