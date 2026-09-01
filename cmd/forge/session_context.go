package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/mimir45/Knowledge-Forge/pkg/report"
)

// cmdSessionContext is Phase 5's SessionStart hook: print the vault index and the
// developer profile into the session's context, budget-capped, and never fail the
// session — every problem goes to .forge/session-context.log, exit is always 0. Claude
// Code replays a saved transcript on --continue/--resume rather than re-running hooks
// mid-session, so nothing here needs to be time-sensitive.
func cmdSessionContext(args []string) int {
	fs := flag.NewFlagSet("forge session-context", flag.ContinueOnError)
	vaultDir := fs.String("vault", "", "vault root; defaults to config vault_path, then .")
	budget := fs.Int("max-bytes", 4096, "byte budget applied to the index and, separately, the profile")
	fs.SetOutput(io.Discard)
	if err := fs.Parse(args); err != nil {
		return 0
	}
	root, err := resolveVault(*vaultDir)
	if err != nil {
		logSessionContext(root, err)
		return 0
	}
	printSessionContext(root, *budget)
	return 0
}

// printSessionContext writes whatever it can build to stdout — the index alone still
// helps if the profile is missing, and vice versa — logging only the read failures.
func printSessionContext(root string, budget int) {
	if idx, err := readTrimmed(indexPath(root), budget); err != nil {
		logSessionContext(root, err)
	} else {
		fmt.Print(idx)
	}
	if prof, err := readTrimmed(filepath.Join(root, "profiles", "me.md"), budget); err != nil {
		logSessionContext(root, err)
	} else {
		fmt.Print("\n---\n\n")
		fmt.Print(prof)
	}
}

// readTrimmed shares pkg/report's exact 4KB-budget cut-at-a-line-boundary logic rather
// than reimplementing a second truncator for the profile.
func readTrimmed(path string, budget int) (string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return report.Trim(string(b), budget), nil
}

// indexPath honours cfg.Paths.Index when the config loads cleanly; a config error here
// is not fatal to the hook, it just falls back to the packaged default filename.
func indexPath(root string) string {
	name := "_index.md"
	if cfg, err := loadConfig(); err == nil && cfg.Paths.Index != "" {
		name = cfg.Paths.Index
	}
	return filepath.Join(root, name)
}

// logSessionContext is best-effort by design: a hook that fails loudly defeats its own
// fail-silent contract, so a logging failure here is swallowed too.
func logSessionContext(root string, err error) {
	dir := root
	if dir == "" {
		dir = "."
	}
	logDir := filepath.Join(dir, ".forge")
	if mkErr := os.MkdirAll(logDir, 0o755); mkErr != nil {
		return
	}
	f, openErr := os.OpenFile(filepath.Join(logDir, "session-context.log"),
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if openErr != nil {
		return
	}
	defer f.Close()
	fmt.Fprintf(f, "%s %v\n", time.Now().Format(time.RFC3339), err)
}
