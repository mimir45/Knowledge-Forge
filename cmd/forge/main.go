// Command forge is the Knowledge Forge static core. Every subcommand here runs with
// zero model calls except `forge engine`, Phase 3b's execution layer over the four
// tiers pkg/config names; the LLM tiers otherwise sit above this binary, never inside it.
package main

import (
	"fmt"
	"os"
)

var commands = map[string]func([]string) int{
	"init":            cmdInit,
	"config":          cmdConfig,
	"slug":            cmdSlug,
	"validate":        cmdValidate,
	"recall":          cmdRecall,
	"index":           cmdIndex,
	"reindex":         cmdReindex,
	"capture":         cmdCapture,
	"drift":           cmdDrift,
	"check":           cmdCheck,
	"engine":          cmdEngine,
	"verify-code":     cmdVerifyCode,
	"gate":            cmdGate,
	"session-context": cmdSessionContext,
	"intent":          cmdIntent,
	"session-capture": cmdSessionCapture,
	"cache-source":    cmdCacheSource,
	"stats":           cmdStats,
	"logback":         cmdLogback,
}

const usage = `forge — Knowledge Forge static core (no model calls)

usage: forge <command> [flags]

commands:
  init       write ~/.forge/forge.config.md and <vault>/profiles/me.md
  config     print the resolved configuration chain
  slug       generate the canonical kebab-case slug for a title
  validate   check notes against references/schema.yaml
  recall     score the vault against a question before any research runs
  index      rebuild <vault>/_index.md from the markdown
  reindex    discard the derived SQLite cache and rebuild it from markdown
  capture    harvest human-correction training pairs from a vault commit
  drift      check note code citations against a code repo's git history
  check      the weekly pass: render every report into <vault>/reports/
  engine       select/run/record against the four engine tiers (the one exception
               to zero model calls — see forge engine --help)
  verify-code  compile-check a code snippet against the system toolchain — never a
               dependency resolver, see forge verify-code --help
  gate         run the seven DESIGN §12 quality gates against one draft note and
               quarantine it to _inbox/ on a blocking failure — see forge gate --help
  session-context  SessionStart hook: print the vault index + developer profile into
                   context, budget-capped, fail-silent, exit 0 always
  intent           UserPromptSubmit hook: cheap regex-free recall check on stdin's
                   prompt, emits the top vault hit as additionalContext above 0.7
  session-capture  SessionEnd hook: regex-scans stdin's transcript for conclusion
                   sentences, writes up to 3 low-confidence stubs to _inbox/, deduped
                   by session-id+content hash, fail-silent, exit 0 always
  cache-source     PostToolUse (WebFetch) hook: writes .forge/cache/<url-hash>.md with a
                   TTL (static.cache_ttl_days, default 30), fail-silent, exit 0 always
  stats            hit rate, most-asked topics, gaps, an approximate time-saved
                   estimate, and the weekly vault-stats trend — see forge stats --help
  logback          make the vault's knowledge discoverable from the code repo itself:
                   docs/knowledge-map.md, per-module CLAUDE.md fragments, opt-in inline
                   markers — see forge logback --help

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
