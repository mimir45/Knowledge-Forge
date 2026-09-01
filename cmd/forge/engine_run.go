package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/mimir45/Knowledge-Forge/pkg/config"
	"github.com/mimir45/Knowledge-Forge/pkg/dataset"
	"github.com/mimir45/Knowledge-Forge/pkg/engine"
	"github.com/mimir45/Knowledge-Forge/pkg/store"
	"github.com/mimir45/Knowledge-Forge/pkg/vault"
)

func cmdEngineRun(args []string) int {
	fs := flag.NewFlagSet("forge engine run", flag.ContinueOnError)
	vaultDir := fs.String("vault", "", "vault root; defaults to config vault_path, then .")
	stage := fs.String("stage", "", "pipeline stage to run")
	promptFile := fs.String("prompt-file", "", "file containing the prompt to send")
	rel := fs.String("rel", "", "note path, relative to the vault root — only needed so "+
		"on_exhausted:queue has somewhere to stamp pending_advisor")
	fs.Usage = func() { fmt.Fprint(os.Stderr, engineRunUsage); fs.PrintDefaults() }
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if *stage == "" || *promptFile == "" {
		fmt.Fprint(os.Stderr, "forge engine run: --stage and --prompt-file are required\n\n"+engineRunUsage)
		return 2
	}
	root, code := vaultOrExit("engine run", *vaultDir)
	if code != 0 {
		return code
	}
	cfg, code := configOrExit("engine run")
	if code != 0 {
		return code
	}
	prompt, err := os.ReadFile(*promptFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "forge engine run: %v\n", err)
		return 2
	}
	return runEngineRun(root, cfg, *stage, string(prompt), *rel)
}

const engineRunUsage = `usage: forge engine run --stage NAME --prompt-file FILE [--rel PATH] [--vault DIR]

Resolves the stage, calls the winning tier (real HTTP for api/advisor/local, an
in-session instruction for host, a typed refusal for none), and books any spend
against today's budget in SQLite before printing the result as JSON. If the winning
tier degraded to none because today's budget is spent and on_exhausted is "queue",
--rel stamps pending_advisor: true on that note instead of silently doing nothing.
If on_exhausted is "stop" instead, the same degradation exits non-zero without ever
calling the tier; "degrade" (or any other configured value) falls through silently,
same as today.

`

func runEngineRun(vaultDir string, cfg *config.Config, stage, prompt, rel string) int {
	root, err := filepath.Abs(vaultDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "forge engine run: %v\n", err)
		return 2
	}
	st, err := store.Open(root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "forge engine run: cache: %v\n", err)
		return 1
	}
	defer st.Close()
	name, _, err := engine.Resolve(cfg, st, time.Now, stage)
	if err != nil {
		fmt.Fprintf(os.Stderr, "forge engine run: %v\n", err)
		return 2
	}
	if name == "none" && engine.Exhausted(cfg, st, time.Now, stage) {
		if code, halt := onExhausted(cfg, root, stage, rel); halt {
			return code
		}
	}
	return callAndSpend(cfg, st, root, name, stage, prompt)
}

// onExhausted applies on_exhausted's configured meaning once Resolve has already
// degraded stage to "none" for lack of budget (B-023's behavior half): "queue" stamps
// pending_advisor and lets the run fall through to none as before; "stop" halts with a
// real non-zero exit instead of none's usual quiet refusal; "degrade" (or anything else
// the validator accepts) is today's silent fallthrough, deliberately unchanged.
func onExhausted(cfg *config.Config, root, stage, rel string) (code int, halt bool) {
	switch cfg.Engines.Budget.OnExhausted {
	case "queue":
		if rel == "" {
			return 0, false
		}
		if err := queueNote(root, rel); err != nil {
			fmt.Fprintf(os.Stderr, "forge engine run: queue: %v\n", err)
			return 1, true
		}
	case "stop":
		fmt.Fprintf(os.Stderr, "forge engine run: stage %q: budget exhausted and "+
			"on_exhausted is \"stop\"\n", stage)
		return 1, true
	}
	return 0, false
}

// queueNote stamps pending_advisor: true (ADDENDUM §A.4's `queue` behavior) via the same
// frontmatter writer engine record uses. The run still falls through to none below —
// queuing records that today's advisor call was deferred, it does not retry it inline.
func queueNote(root, rel string) error {
	n, s, err := loadNoteAndSchema(root, rel)
	if err != nil {
		return err
	}
	if n.FM == nil {
		return vault.ErrNoFM
	}
	return vault.SetScalars(n, s, map[string]string{"pending_advisor": "true"})
}

func callAndSpend(cfg *config.Config, st *store.Store, root, name, stage, prompt string) int {
	res, err := buildEngine(cfg, name).Run(engine.Request{Stage: stage, Prompt: prompt})
	if err != nil {
		if _, ok := err.(*engine.NoGenerationError); ok {
			fmt.Println(err)
			return 0
		}
		fmt.Fprintf(os.Stderr, "forge engine run: %v\n", err)
		return 1
	}
	if tier := spendTier(name); tier != "" {
		if err := st.Spend(tier, res.CostUSD, time.Now); err != nil {
			fmt.Fprintf(os.Stderr, "forge engine run: spend: %v\n", err)
			return 1
		}
	}
	if name == "advisor" {
		captureD2(cfg, root, stage, prompt, res.Output)
	}
	return emitResult(res)
}

// captureD2 logs the critique verbatim (ADDENDUM §D.1's D2) when the config chain has
// opted in. It never fails the run — a dataset write error is a side channel, not the
// command's job, the same posture the D3 post-commit hook takes toward its own writes.
//
// D2.Enabled now checks dataset.enabled as well as the capture list, which this call site
// never did: `{enabled: false, capture: [d2]}` used to capture anyway. The packaged layer
// sets enabled: true, so no default behaviour changes.
func captureD2(cfg *config.Config, root, stage, draft, critique string) {
	if !dataset.D2.Enabled(cfg.Dataset) {
		return
	}
	p := dataset.D2Pair{Kind: dataset.D2Kind, Stage: stage, Draft: draft,
		Critique: critique, CapturedAt: time.Now()}
	if err := dataset.AppendD2(root, p); err != nil {
		fmt.Fprintf(os.Stderr, "forge engine run: d2 capture: %v\n", err)
	}
}

// buildEngine constructs the real Engine value for name. "local" still returns an API
// (api.go's doc-comment): Provider "ollama", pointed at engines.local.base_url — it is a
// routing alias, not a fifth Engine implementation.
func buildEngine(cfg *config.Config, name string) engine.Engine {
	switch name {
	case "host":
		return engine.Host{}
	case "api":
		return apiFor(cfg.Engines.API, cfg.Engines.API.Model)
	case "local":
		return localFor(cfg.Engines.Local)
	case "advisor":
		return engine.Advisor{API: apiFor(cfg.Engines.API, cfg.Engines.Advisor.Model)}
	default:
		return engine.None{}
	}
}

// apiFor shares engines.api's provider/base_url/key_env between the api tier and the
// advisor tier, which the config schema gives no connection block of its own — advisor is
// a critique wrapper over the same backend, with its own model (see pkg/engine/advisor.go).
func apiFor(a config.API, model string) engine.API {
	return engine.API{
		RoundTripper: http.DefaultTransport,
		Provider:     a.Provider, Model: model,
		BaseURL: a.BaseURL, APIKey: os.Getenv(a.KeyEnv),
	}
}

func localFor(l config.Local) engine.API {
	return engine.API{
		RoundTripper: http.DefaultTransport,
		Provider:     "ollama", Model: l.Model, BaseURL: l.BaseURL,
	}
}

// spendTier maps a winning engine name to the budget bucket it draws from. "host" and
// "local" have no cap in cfg.Engines.Budget, so neither books anything.
func spendTier(name string) string {
	switch name {
	case "api":
		return "api"
	case "advisor":
		return "advisor"
	default:
		return ""
	}
}

func emitResult(res engine.Result) int {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(res); err != nil {
		fmt.Fprintf(os.Stderr, "forge engine run: %v\n", err)
		return 1
	}
	return 0
}
