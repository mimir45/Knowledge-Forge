package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"knowledge-forge/pkg/vault"
)

func cmdSlug(args []string) int {
	fs := flag.NewFlagSet("forge slug", flag.ContinueOnError)
	vaultDir := fs.String("vault", "", "vault root; when set, existing slugs are avoided")
	asJSON := fs.Bool("json", false, "emit {\"title\":…,\"slug\":…}")
	fs.Usage = func() {
		fmt.Fprint(os.Stderr, "usage: forge slug [--vault DIR] [--json] <title>...\n\n"+
			"Deterministic: the same title always yields the same slug. With --vault,\n"+
			"a collision with an existing note's slug gets a numeric suffix.\n\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if fs.NArg() == 0 {
		fs.Usage()
		return 2
	}
	return emitSlug(strings.Join(fs.Args(), " "), *vaultDir, *asJSON)
}

func emitSlug(title, vaultDir string, asJSON bool) int {
	taken := map[string]bool{}
	if vaultDir != "" {
		var err error
		if taken, err = existingSlugs(vaultDir); err != nil {
			fmt.Fprintf(os.Stderr, "forge slug: %v\n", err)
			return 1
		}
	}
	s := vault.SlugUnique(title, taken)
	if asJSON {
		b, _ := json.Marshal(map[string]string{"title": title, "slug": s})
		fmt.Println(string(b))
		return 0
	}
	fmt.Println(s)
	return 0
}

// existingSlugs collects every slug already claimed, falling back to the filename for
// notes the migration has not reached yet — otherwise a pre-contract note's name could
// be handed out twice.
func existingSlugs(root string) (map[string]bool, error) {
	rels, err := vault.Walk(root)
	if err != nil {
		return nil, err
	}
	taken := make(map[string]bool, len(rels))
	for _, rel := range rels {
		n, err := vault.Load(root+"/"+rel, rel)
		if err != nil {
			continue
		}
		taken[slugOf(n)] = true
	}
	return taken, nil
}

func slugOf(n *vault.Note) string {
	if n.FM != nil {
		if s := n.FM.Str("slug"); s != "" {
			return s
		}
	}
	return vault.Slug(strings.TrimSuffix(n.Rel[strings.LastIndex(n.Rel, "/")+1:], ".md"))
}
