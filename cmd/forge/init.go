package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

// initOpts is the wizard's five answers plus the two preset choices derived from them.
type initOpts struct {
	vault        string
	language     string
	frameworks   string
	infra        string
	seniority    string
	depth        int
	noteLanguage string
	explainStyle string
	trigger      string
	enginePreset string
	stackPreset  string
	force        bool
	dryRun       bool
}

func cmdInit(args []string) int {
	var o initOpts
	fs := flag.NewFlagSet("forge init", flag.ContinueOnError)
	bindInitFlags(fs, &o)
	fs.Usage = func() { fmt.Fprint(os.Stderr, initUsage); fs.PrintDefaults() }
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if err := o.normalize(); err != nil {
		fmt.Fprintf(os.Stderr, "forge init: %v\n", err)
		return 2
	}
	return runInit(o)
}

func bindInitFlags(fs *flag.FlagSet, o *initOpts) {
	fs.StringVar(&o.vault, "vault", "", "vault root (required)")
	fs.StringVar(&o.language, "language", "", "primary language, e.g. java")
	fs.StringVar(&o.frameworks, "frameworks", "", "comma-separated, e.g. spring-boot,hibernate")
	fs.StringVar(&o.infra, "infra", "", "comma-separated, e.g. docker,postgres,kafka")
	fs.StringVar(&o.seniority, "seniority", "mid", "junior | mid | senior")
	fs.IntVar(&o.depth, "depth", 0, "1-5; 0 derives it from --seniority")
	fs.StringVar(&o.noteLanguage, "note-language", "en", "prose language of note bodies")
	fs.StringVar(&o.explainStyle, "explain-style", "mechanism-first",
		"mechanism-first | example-first | analogy-first")
	fs.StringVar(&o.trigger, "trigger", "ask", "ask | auto | manual")
	fs.StringVar(&o.enginePreset, "engine-preset", "claude-only",
		"offline | claude-only | byo-api | max")
	fs.StringVar(&o.stackPreset, "stack-preset", "",
		"java-backend | frontend | devops | minimal")
	fs.BoolVar(&o.force, "force", false, "overwrite files that already exist")
	fs.BoolVar(&o.dryRun, "dry-run", false, "print what would be written, write nothing")
}

const initUsage = `usage: forge init --vault DIR [--language L] [--frameworks a,b] [--infra a,b]
                  [--seniority junior|mid|senior] [--depth 1-5] [--note-language en]
                  [--explain-style mechanism-first] [--trigger ask|auto|manual]
                  [--engine-preset claude-only] [--stack-preset java-backend]
                  [--force] [--dry-run]

Writes two files and nothing else:

  ~/.forge/forge.config.md    your settings — only the keys that differ from the
                              packaged defaults, so a binary upgrade still brings you
                              new defaults for everything you did not decide
  <vault>/profiles/me.md      the developer profile, rendered from profiles/me.template.md

Both are yours to edit afterwards. This command refuses to overwrite either without
--force. It does not prompt: skills/forge-init/ asks the questions and calls this.

flags:
`

// normalize fills what the wizard did not ask and rejects what it cannot fix.
func (o *initOpts) normalize() error {
	if o.vault == "" {
		return fmt.Errorf("--vault is required")
	}
	o.seniority = strings.ToLower(strings.TrimSpace(o.seniority))
	if o.depth == 0 {
		o.depth = depthFor(o.seniority)
	}
	if o.depth < 1 || o.depth > 5 {
		return fmt.Errorf("--depth %d is outside 1-5", o.depth)
	}
	if o.language == "" {
		o.language = "en-agnostic"
	}
	return checkVault(o.vault)
}

// depthFor maps seniority onto the profile's default_depth.
func depthFor(seniority string) int {
	switch seniority {
	case "junior":
		return 2
	case "senior":
		return 4
	}
	return 3
}

func checkVault(dir string) error {
	st, err := os.Stat(dir)
	if err != nil {
		return fmt.Errorf("vault %s: %w", dir, err)
	}
	if !st.IsDir() {
		return fmt.Errorf("vault %s is not a directory", dir)
	}
	probe := dir + "/.forge-init-probe"
	if err := os.WriteFile(probe, nil, 0o644); err != nil {
		return fmt.Errorf("vault %s is not writable: %w", dir, err)
	}
	return os.Remove(probe)
}
