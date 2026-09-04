package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"runtime"
)

// version and commit are stamped at link time by the Makefile and .goreleaser.yml via
// -X main.version / -X main.buildSHA.
var (
	version  = "dev"
	buildSHA = "none"
)

const versionUsage = `usage: forge version [--json]

Prints this binary's version, the commit it was built from, and the Go toolchain and
platform it was built for. The values come from -X main.version / -X main.buildSHA at
link time; an unstamped build (a plain "go build") reports "dev" and "none".
`

func cmdVersion(args []string) int {
	fs := flag.NewFlagSet("forge version", flag.ContinueOnError)
	asJSON := fs.Bool("json", false, "print as JSON")
	fs.SetOutput(io.Discard)
	fs.Usage = func() { fmt.Fprint(os.Stderr, versionUsage); fs.PrintDefaults() }
	if err := fs.Parse(args); err != nil {
		fmt.Fprint(os.Stderr, versionUsage)
		return 2
	}
	if *asJSON {
		fmt.Printf("{\"version\":%q,\"commit\":%q,\"go\":%q,\"platform\":\"%s/%s\"}\n",
			version, buildSHA, runtime.Version(), runtime.GOOS, runtime.GOARCH)
		return 0
	}
	fmt.Printf("forge %s (commit %s, %s, %s/%s)\n",
		version, buildSHA, runtime.Version(), runtime.GOOS, runtime.GOARCH)
	return 0
}
