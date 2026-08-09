package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// cmdConfig prints what the chain actually resolved to. It exists because a four-layer
// merge is not something a user can evaluate in their head: the answer to "why is my
// vault path wrong" is which layer set it, and without this command that answer requires
// reading four files and knowing the merge rule.
func cmdConfig(args []string) int {
	fs := flag.NewFlagSet("forge config", flag.ContinueOnError)
	asJSON := fs.Bool("json", false, "emit the merged config as JSON")
	layersOnly := fs.Bool("layers", false, "list the contributing files only")
	fs.Usage = func() { fmt.Fprint(os.Stderr, configUsage); fs.PrintDefaults() }
	if err := fs.Parse(args); err != nil {
		return 2
	}
	cfg, code := configOrExit("config")
	if code != 0 {
		return code
	}
	if *layersOnly {
		return printLayers(cfg.Layers)
	}
	return printConfig(cfg, *asJSON)
}

const configUsage = `usage: forge config [--json] [--layers]

Resolves the configuration chain and prints the result. Layers are listed lowest
precedence first, which is the order they were merged in — the last one listed wins
on any key it sets.

Maps merge key by key; scalars and lists replace wholesale.

flags:
`

func printLayers(layers []string) int {
	for i, l := range layers {
		fmt.Printf("%d. %s\n", i+1, l)
	}
	return 0
}

func printConfig(cfg any, asJSON bool) int {
	var out []byte
	var err error
	if asJSON {
		out, err = json.MarshalIndent(cfg, "", "  ")
	} else {
		out, err = yaml.Marshal(cfg)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "forge config: %v\n", err)
		return 1
	}
	fmt.Printf("%s", out)
	return 0
}
