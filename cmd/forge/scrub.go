package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/mimir45/Knowledge-Forge/pkg/scrub"
)

const scrubUsage = `usage: forge scrub --src DIR --dst DIR

Walks --src as a vault, redacts secret/PII-shaped content from every note's
frontmatter and body (emails, absolute home paths, API-key-shaped tokens), and writes
the result under --dst with the same relative layout. Used to build examples/vault/
from a real vault, and by Phase 6b's --anonymize export path.

Fails closed: if any note cannot be scrubbed and re-validated against
references/schema.yaml, nothing is written to --dst at all — not even the notes that
scrubbed cleanly. --dst is never touched on error.

Prints a JSON Report to stdout on success: NotesTotal, NotesWritten, Redactions, and
NoFrontmatter (vault-relative paths written body-only, for the pre-migration notes
that still have none).

Exit 0 = success. Exit 2 = usage error. Exit 3 = scrub failed, --dst untouched.
`

func cmdScrub(args []string) int {
	fs := flag.NewFlagSet("forge scrub", flag.ContinueOnError)
	src := fs.String("src", "", "source vault directory")
	dst := fs.String("dst", "", "destination directory for the scrubbed copy")
	fs.Usage = func() { fmt.Fprint(os.Stderr, scrubUsage); fs.PrintDefaults() }
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if *src == "" || *dst == "" {
		fmt.Fprintln(os.Stderr, "forge scrub: --src and --dst are both required")
		return 2
	}
	return runScrub(*src, *dst)
}

func runScrub(src, dst string) int {
	rep, err := scrub.Scrub(src, dst)
	if err != nil {
		fmt.Fprintf(os.Stderr, "forge scrub: %v\n", err)
		return 3
	}
	b, _ := json.MarshalIndent(rep, "", "  ")
	fmt.Println(string(b))
	return 0
}
