package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"knowledge-forge/pkg/config"
	"knowledge-forge/pkg/coderef"
	"knowledge-forge/pkg/drift"
)

// logbackCfg is forge logback's inputs — the same --vault/--repo shape drift and check
// use, plus the two flags specific to this command.
type logbackCfg struct {
	vault         string
	repos         repoList
	dryRun        bool
	removeMarkers bool
}

func cmdLogback(args []string) int {
	var cfg logbackCfg
	fs := flag.NewFlagSet("forge logback", flag.ContinueOnError)
	fs.StringVar(&cfg.vault, "vault", "", "vault root; defaults to config vault_path, then .")
	fs.Var(&cfg.repos, "repo", "code repository as name=path (repeatable, required)")
	fs.BoolVar(&cfg.dryRun, "dry-run", false, "print what would change; write nothing")
	fs.BoolVar(&cfg.removeMarkers, "remove-markers", false,
		"strip inline // forge: markers this command wrote; ignores the logback config gates")
	fs.Usage = func() { fmt.Fprint(os.Stderr, logbackUsage); fs.PrintDefaults() }
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if len(cfg.repos) == 0 {
		fmt.Fprint(os.Stderr, "forge logback: at least one --repo is required\n")
		return 2
	}
	root, code := vaultOrExit("logback", cfg.vault)
	if code != 0 {
		return code
	}
	abs, err := filepath.Abs(root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "forge logback: %v\n", err)
		return 2
	}
	cfg.vault = abs
	gcfg, code := configOrExit("logback")
	if code != 0 {
		return code
	}
	return runLogback(cfg, gcfg)
}

const logbackUsage = `usage: forge logback --repo NAME=PATH [--vault DIR] [--dry-run] [--remove-markers]

Makes the vault's knowledge discoverable from the code repo itself: docs/knowledge-map.md,
a CLAUDE.md fragment per module (a managed begin/end block — hand-written prose above and
below it is left alone), and — opt-in via static.logback.inline_markers — a
"// forge: [[slug]]" comment above the symbols a note documents. T0, deterministic,
idempotent: a second run with nothing changed writes nothing. Never touches code
semantics — comments and separate files only.

--dry-run prints what would change without writing. --remove-markers strips inline
markers only, leaving docs/knowledge-map.md and any CLAUDE.md fragments in place.

`

func runLogback(cfg logbackCfg, gcfg *config.Config) int {
	rg, err := registryOf(cfg.repos)
	if err != nil {
		fmt.Fprintf(os.Stderr, "forge logback: %v\n", err)
		return 1
	}
	src := drift.NewGitSource(cfg.repos, filepath.Join(cfg.vault, ".forge"))
	ok := true
	for _, r := range cfg.repos {
		if !runOneRepo(cfg, gcfg, r, rg, src) {
			ok = false
		}
	}
	if !ok {
		return 1
	}
	return 0
}

func runOneRepo(cfg logbackCfg, gcfg *config.Config, r drift.Repo,
	rg *coderef.Registry, src symbolFinder) bool {

	groups, err := buildGroups(cfg.vault, r, rg, src)
	if err != nil {
		fmt.Fprintf(os.Stderr, "forge logback: %s: %v\n", r.Name, err)
		return false
	}
	if cfg.removeMarkers {
		return removeMarkers(cfg.vault, r, rg, src, cfg.dryRun)
	}
	ok := true
	if gcfg.Static.LogBack.KnowledgeMap {
		ok = writeKnowledgeMap(r, groups, cfg.dryRun) && ok
	}
	if gcfg.Static.LogBack.ClaudeMDFragment {
		ok = writeClaudeFragments(r, groups, cfg.dryRun) && ok
	}
	if gcfg.Static.LogBack.InlineMarkers {
		ok = writeInlineMarkers(cfg.vault, r, rg, src, cfg.dryRun) && ok
	}
	return ok
}
