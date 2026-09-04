package main

import (
	"fmt"
	"os"
)

// cmdEngine dispatches Phase 3b's three subcommands.
func cmdEngine(args []string) int {
	if len(args) == 0 {
		fmt.Fprint(os.Stderr, engineUsage)
		return 2
	}
	switch args[0] {
	case "-h", "--help", "help":
		fmt.Print(engineUsage)
		return 0
	case "select":
		return cmdEngineSelect(args[1:])
	case "run":
		return cmdEngineRun(args[1:])
	case "record":
		return cmdEngineRecord(args[1:])
	default:
		fmt.Fprintf(os.Stderr, "forge engine: unknown subcommand %q\n\n%s", args[0], engineUsage)
		return 2
	}
}

const engineUsage = `usage: forge engine <select|run|record> [flags]

Phase 3b's execution layer over the four tiers: none, host, api, advisor. Run
"forge engine <subcommand> --help" for that subcommand's flags.

  select   resolve a pipeline stage to an engine — no HTTP, no spend
  run      call the resolved engine and book its cost against today's budget
  record   stamp engine_trail onto a note after a host-tier step completes
`
