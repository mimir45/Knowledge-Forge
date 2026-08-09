// Command forge is the Knowledge Forge static core. Every subcommand here runs with
// zero model calls; the LLM tiers sit above this binary, never inside it.
package main

import (
	"fmt"
	"os"
)

var commands = map[string]func([]string) int{
	"slug":     cmdSlug,
	"validate": cmdValidate,
	"recall":   cmdRecall,
	"index":    cmdIndex,
	"reindex":  cmdReindex,
	"capture":  cmdCapture,
}

const usage = `forge — Knowledge Forge static core (no model calls)

usage: forge <command> [flags]

commands:
  slug       generate the canonical kebab-case slug for a title
  validate   check notes against references/schema.yaml
  recall     score the vault against a question before any research runs
  index      rebuild <vault>/_index.md from the markdown
  reindex    discard the derived SQLite cache and rebuild it from markdown
  capture    harvest human-correction training pairs from a vault commit

run "forge <command> --help" for that command's flags.
`

func main() { os.Exit(run(os.Args[1:])) }

func run(args []string) int {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" || args[0] == "help" {
		fmt.Print(usage)
		return 0
	}
	cmd, ok := commands[args[0]]
	if !ok {
		fmt.Fprintf(os.Stderr, "forge: unknown command %q\n\n%s", args[0], usage)
		return 2
	}
	return cmd(args[1:])
}
