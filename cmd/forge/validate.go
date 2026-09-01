package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/mimir45/Knowledge-Forge/pkg/vault"
)

func cmdValidate(args []string) int {
	fs := flag.NewFlagSet("forge validate", flag.ContinueOnError)
	all := fs.Bool("all", false, "validate every note under --vault")
	vaultDir := fs.String("vault", "", "vault root; defaults to config vault_path, then .")
	fix := fs.Bool("fix", false, "repair mechanically fixable problems in place")
	quiet := fs.Bool("quiet", false, "print only the summary line")
	fs.Usage = func() {
		fmt.Fprint(os.Stderr, "usage: forge validate [--all] [--vault DIR] [--fix] <path>...\n\n"+
			"Exit 0 = every note conforms. Exit 1 = at least one does not.\n"+
			"--fix repairs only mechanical problems: missing dates, key order, tag case,\n"+
			"alias rewriting, schema-constant defaults. It never invents type, stack or tags.\n\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return 2
	}
	root, code := vaultOrExit("validate", *vaultDir)
	if code != 0 {
		return code
	}
	return runValidate(fs.Args(), *all, root, *fix, *quiet)
}

func runValidate(paths []string, all bool, vaultDir string, fix, quiet bool) int {
	s, err := vault.LoadSchema()
	if err != nil {
		fmt.Fprintf(os.Stderr, "forge validate: %v\n", err)
		return 2
	}
	targets, root, err := resolveTargets(paths, all, vaultDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "forge validate: %v\n", err)
		return 2
	}
	return printResult(validateAll(targets, root, s, fix), len(targets), quiet)
}

func plural(paths []string) string {
	if len(paths) == 1 {
		return ""
	}
	return "s"
}

func resolveTargets(paths []string, all bool, vaultDir string) ([]string, string, error) {
	root, err := filepath.Abs(vaultDir)
	if err != nil {
		return nil, "", err
	}
	if !all {
		if len(paths) == 0 {
			return nil, "", fmt.Errorf("give one or more paths, or --all")
		}
		return paths, "", nil
	}
	// --all walks --vault, so a positional path is not merely redundant: someone who
	// wrote `forge validate --all --fix /some/vault` meant that directory and would get
	// the working directory rewritten instead. Refuse rather than guess.
	if len(paths) > 0 {
		return nil, "", fmt.Errorf("--all walks --vault; remove the path argument%s"+
			" or pass it as --vault %s", plural(paths), paths[0])
	}
	rels, err := vault.Walk(root)
	if err != nil {
		return nil, "", err
	}
	var out []string
	for _, rel := range rels {
		if vault.IsContractNote(rel) {
			out = append(out, filepath.Join(root, rel))
		}
	}
	return out, root, nil
}

type result struct {
	issues  []vault.Issue
	fixed   int
	badFile int
}

func validateAll(targets []string, root string, s *vault.Schema, fix bool) result {
	var r result
	for _, abs := range targets {
		n, err := loadTarget(abs, root)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  %s: %v\n", abs, err)
			r.badFile++
			continue
		}
		if fix {
			r.fixed += applyFix(n, s)
			n, _ = loadTarget(abs, root)
		}
		r.issues = append(r.issues, vault.Validate(n, s)...)
	}
	return r
}

func loadTarget(abs, root string) (*vault.Note, error) {
	rel := abs
	if root != "" {
		if x, err := filepath.Rel(root, abs); err == nil {
			rel = x
		}
	}
	return vault.Load(abs, rel)
}

// applyFix rewrites the file only when Fix actually changed something, so a clean run
// leaves every mtime untouched and the index cache stays warm.
func applyFix(n *vault.Note, s *vault.Schema) int {
	out, changes, err := vault.Fix(n, s)
	if err != nil || len(changes) == 0 {
		return 0
	}
	if err := os.WriteFile(n.Path, out, 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "  %s: write failed: %v\n", n.Rel, err)
		return 0
	}
	return len(changes)
}

func printResult(r result, total int, quiet bool) int {
	sort.SliceStable(r.issues, func(i, j int) bool { return r.issues[i].Rel < r.issues[j].Rel })
	if !quiet {
		for _, is := range r.issues {
			fmt.Println(is)
		}
	}
	fmt.Printf("\n%d notes checked · %d issues · %d unreadable · %d fixes applied\n",
		total, len(r.issues), r.badFile, r.fixed)
	if len(r.issues) > 0 || r.badFile > 0 {
		return 1
	}
	return 0
}
