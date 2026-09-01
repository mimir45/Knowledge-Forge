package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mimir45/Knowledge-Forge/pkg/qualitygate"
)

const verifyCodeUsage = `usage: forge verify-code --lang <java|ts|bash|auto> [--file PATH|--stdin] [--timeout 15s]

A syntax/compile check only — never a dependency resolver. It never runs npm install,
never resolves Maven/Gradle coordinates, and never touches the network; it shells out to
whatever javac/tsc/bash is already on this machine. Exit 0 on pass or skipped, 1 on fail,
2 on a usage error. Emits one CompileResult as JSON on stdout.

--lang auto only works with --file: it infers java/ts/bash from the file extension.
`

func cmdVerifyCode(args []string) int {
	fs := flag.NewFlagSet("forge verify-code", flag.ContinueOnError)
	lang := fs.String("lang", "", "java, ts, bash, or auto (requires --file)")
	file := fs.String("file", "", "path to the snippet; omit to read --stdin")
	stdin := fs.Bool("stdin", false, "read the snippet from stdin")
	timeout := fs.Duration("timeout", 15*time.Second, "per-invocation timeout")
	fs.Usage = func() { fmt.Fprint(os.Stderr, verifyCodeUsage); fs.PrintDefaults() }
	if err := fs.Parse(args); err != nil {
		return 2
	}
	src, resolvedLang, code := readSnippet(*file, *stdin, *lang)
	if code != 0 {
		fs.Usage()
		return code
	}
	r := qualitygate.CompileCheck(resolvedLang, src, *timeout)
	b, _ := json.Marshal(r)
	fmt.Println(string(b))
	if r.Verdict == qualitygate.Fail {
		return 1
	}
	return 0
}

// readSnippet resolves the source bytes and the language, applying --lang auto only
// when --file names a recognized extension.
func readSnippet(file string, stdin bool, lang string) ([]byte, string, int) {
	if file == "" && !stdin {
		fmt.Fprintln(os.Stderr, "forge verify-code: give --file PATH or --stdin")
		return nil, "", 2
	}
	if lang == "auto" {
		if file == "" {
			fmt.Fprintln(os.Stderr, "forge verify-code: --lang auto requires --file")
			return nil, "", 2
		}
		lang = langFromExt(file)
	}
	if lang == "" {
		fmt.Fprintln(os.Stderr, "forge verify-code: --lang is required")
		return nil, "", 2
	}
	src, err := readSource(file, stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "forge verify-code: %v\n", err)
		return nil, "", 1
	}
	return src, lang, 0
}

func readSource(file string, stdin bool) ([]byte, error) {
	if stdin || file == "" {
		return io.ReadAll(os.Stdin)
	}
	return os.ReadFile(file)
}

func langFromExt(file string) string {
	switch strings.ToLower(filepath.Ext(file)) {
	case ".java":
		return "java"
	case ".ts":
		return "ts"
	case ".sh", ".bash":
		return "bash"
	default:
		return ""
	}
}
