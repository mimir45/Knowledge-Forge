# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in
this repository.

## What it is

Knowledge Forge is a Claude Code plugin that turns "explain X" moments into permanent,
linked, verified markdown notes in an Obsidian vault. Its defensible core is a Go
static-analysis engine that runs with **zero model calls**; an optional four-tier LLM
layer (none / host / API / advisor-critique) sits on top, configurable per pipeline
stage.

The project was built end to end with Claude Code itself, one development phase per
session, from an initial design spec through to a packaged, installable plugin. It is
feature-complete.

## Read the docs in this order

1. **`docs/ARCHITECTURE.md`** — the canonical reference. Package map and import DAG,
   the two build lanes, the four-layer config chain, the engine tiers, the recall
   engine's scoring, drift, the quality gates, the three data flows, the Claude Code
   integration layer, the invariant table, and the latency budgets.
2. **`docs/USAGE.md`** — installation, `forge init`, and a per-command reference
   including the four hook commands and hook installation.
3. **`docs/datasets.md`** — the D1–D6 capture tiers, what each records, and what a
   given volume is honestly enough for.
4. `references/` — the specs the code is actually checked against:
   `schema.yaml` (the note contract `forge validate` enforces), `recall-spec.md`
   (scoring, **read before touching `pkg/recall`**), `duplicate-spec.md`,
   `taxonomy.md`, `writing-rules.md` (parsed by the anti-slop gate).

The original design documents (`KNOWLEDGE-FORGE-DESIGN.md`, `-ADDENDUM.md`, `-STACK.md`,
`ROADMAP.md`, `CLAUDE-CODE-PROMPT.md`, `-B2B.md`, `docs/adr/`, `docs/tr/`) were removed
in `df5ccea`. They are **not coming back** — `docs/ARCHITECTURE.md` supersedes them.
Where a comment still says "the original spec (since removed)", that is deliberate: it
records a place where this code knowingly deviates from what was first specified.

Only surviving Python: the one-time `migrate_vault.py` and the offline dataset /
fine-tuning tooling. Neither ships in the binary.

## Things that live outside this repo

- **Vault:** wherever `--vault` / the config chain points (see `forge config
  --layers`), a separate git repo laid out with the `notes/<type>/ moc/ _inbox/
  _archive/ profiles/` topology.
- **The D3 capture hook**, once installed in a vault: `.git/hooks/post-commit` runs
  `forge capture` from a pinned binary copy (path recorded in `<vault>/.forge/forge-bin`;
  `$FORGE_BIN` overrides it). That binary is a **copy**, not the repo's build output —
  rebuild it after any change to `pkg/dataset` or `cmd/forge/capture.go`:
  `CGO_ENABLED=0 go build -o <path> ./cmd/forge`. By design the hook can never fail a
  commit and never prints, so a stale or broken binary is silent: if pairs stop
  appearing, read `<vault>/.forge/capture.log`. Uninstall is `rm .git/hooks/post-commit`.

## Fixture vault (`testdata/vault/`)

A 13-note fixture reproducing a pre-migration vault topology plus twelve deliberate
defects (F1–F12) — mixed frontmatter shapes, a dangling wikilink, a dangling `source:`
path, an orphan, a near-duplicate pair, notes with no frontmatter at all, status carried
as body prose. Catalogue: `testdata/README.md`.

- Rehearse anything that mutates a vault here first. A real vault's own topology
  migration is irreversible and real vaults have no guaranteed backups.
- It has **no `.git`, deliberately** — a nested repo would become a stray gitlink once
  this repo is `git init`-ed. The harness copies the fixture into a temp dir and
  `git init`s the copy; that is how the migration's "refuses a dirty tree" precondition
  and drift's `--since-commit` get exercised. Never `git init` it in place.
- The defects are the test surface. **Do not fix them.**
- It is **not** `examples/vault/` — 93 files generated from a real vault via `forge
  scrub`, scoped to `notes/`+`moc/` only, meant as a clean, exemplary reference vault.

## Product agents (`agents/`)

Four product subagents ship with the plugin and Claude Code **does** discover them —
they appear as `forge:forge-researcher`, `forge:forge-codebase-scout`,
`forge:forge-verifier` and `forge:forge-librarian`. Their spec files live under
`agents/*.md`; `docs/ARCHITECTURE.md` §11.5 describes what each is for.

Prefer delegating over doing it inline. Independent tasks go out in parallel.

## Invariants

Each is stated in a different doc and each is easy to violate by accident:

- The T0 static core makes **zero model calls**. If a design seems to require one, stop
  and ask.
- Stages `recall`, `write`, and `index` accept engine `none` only. On a config that says
  otherwise, **refuse to start with a clear error** — never silently override.
- Drift is git-anchored: post-commit / post-merge / post-checkout, `--since-commit <sha>`.
  Never on file save, never against the uncommitted working tree. Verdicts are a pure
  function of (note refs, tree state) so a revert restores demoted notes symmetrically.
  Demotion history lives in `.forge/`, never in note bodies.
- `CGO_ENABLED=0` for every package except `pkg/codeindex` (go-tree-sitter needs cgo).
  The gate is `//go:build cgo` / `//go:build !cgo` on `pkg/codeindex/parse_*.go` — there
  is no `codeindex` build tag; `make full` selects the lane with `CGO_ENABLED=1` alone.
- Markdown is the only source of truth. SQLite (`modernc.org/sqlite`, pure Go) is a
  derived cache; `forge reindex` must rebuild it entirely from markdown.
- `pkg/similarity` is hand-rolled MinHash + LSH. **No embeddings.**
- Never auto-mutate the vault on a schedule. Quality-gate failures go to `_inbox/` with
  `confidence: low`, never a silent publish.
- Code verification compiles in a throwaway directory, never in the user's project.
- The advisor tier (T3) is critique-only: it returns disputed claims and a patch, never
  a rewrite.
- Telemetry logs the topic and a hash. Never raw question text, code, or file contents.
- CLI only for v1. Do not build a daemon on speculation — measure first.
- Recall scoring is a weighted **mean over active channels**, not a literal weighted
  sum, and the title measure is **F₂, not Dice** — both argued from measured vault
  behavior, not the naive reading of the spec text. See `references/recall-spec.md`
  before touching `pkg/recall`'s scoring; a change there isn't reviewable unless it
  updates `cmd/forge/testdata/calibration.golden` (`go test ./cmd/forge -run
  TestCalibration -update`) and the sweep goldens under the same directory.

## Layout and budgets

```
cmd/forge/        CLI
pkg/vault/        frontmatter + markdown AST (goldmark), mtime-cached
pkg/recall/       deterministic question -> note scoring; zero model calls
pkg/similarity/   MinHash + LSH banding
pkg/graph/        note link graph: components, hubs, orphans, centrality
pkg/codeindex/    go-tree-sitter (Java + TypeScript) — the only cgo package, tag-gated
pkg/coderef/      extracts code citations from note bodies and frontmatter
pkg/gitsig/       churn, ownership, co-change coupling — via the git CLI, not go-git
pkg/drift/        the key package — AST comparison, not line diffs
pkg/linkcheck/    HTTP HEAD on sources, cached, rate-limited
pkg/report/       renders analyses to markdown; must not import pkg/codeindex
pkg/store/        SQLite via modernc.org/sqlite, derived cache only except the budget table
pkg/engine/       none/host/api/advisor backends, per-stage select+fallback, engine_trail
pkg/config/       the four-layer config chain
pkg/sentinel/     id-based begin/end managed comment blocks; Upsert/UpsertBefore/Remove
pkg/scrub/        redacts secret/PII-shaped content from a vault copy; fails closed
pkg/dataset/      D1-D6 training-data capture and export; D6 is derived, not captured
pkg/qualitygate/  the seven gates + orchestration + `_inbox/` quarantine
pkg/telemetry/    the `ask` event, sha256 topic hashing, never raw question text
```

Latency **budgets** — these are targets, not measurements: `forge drift` <100ms (the
binding constraint — it runs on the git-hook path), `forge index` <200ms, `forge check`
<10s warm, `forge session-context` <200ms, `forge intent` <50ms. The intent budget's
plausibility comes from reusing `forge recall`'s already-warm SQLite cache.

There is no committed harness that measures any of these end to end — the nine
`Benchmark` functions are library micro-benchmarks and `pkg/qualitygate` has none at
all. Earlier revisions of this file quoted specific actuals (drift 60–70ms, index 20ms,
`qualitygate.Run` ~0.13ms, `verify-code` bash ~10ms / java ~170ms warm); they could not
be reproduced from anything in the repo and an external campaign measured drift at
**151ms median / 208ms p95**, over its own budget. Per `MANIFESTO.md`: performance
claims that can't be reproduced aren't claims. Restore a number here only alongside the
harness that produces it.

## Config chain

Four layers, highest precedence first: `FORGE_CONFIG` env var > `.forge.config.md` (repo
root) > `~/.forge/forge.config.md` (user) > the packaged `config/forge.config.example.md`
template. The schema is the *union* of the engine/config blocks and the pipeline keys
that block never restated — see `docs/ARCHITECTURE.md` §4. `forge init` is the **only** writer of `~/.forge/forge.config.md` and
`<vault>/profiles/me.md` (rendered from `profiles/me.template.md`) — never
`config/forge.config.md`, which stays a packaged template. `on_exhausted` defaults to
`queue`; accepted values are `queue | degrade | stop` — `queue` stamps
`pending_advisor: true` and falls through to `none`, `degrade` is today's silent
`none`-fallback (the honest reading of the word), `stop` exits non-zero without calling
the tier. Budget counters live in SQLite and must survive `forge reindex`.

## Commands

`CGO_ENABLED=0 go build ./...` and `go test ./...` both work; `make build test bench
dist install-hook` covers the rest. Two build lanes, because `pkg/codeindex` is the one
cgo package — the default lane is pure Go and cross-compiles; the codeindex lane needs
cgo and a host toolchain (`make full`, i.e. `CGO_ENABLED=1`). The selection is
`//go:build cgo` / `!cgo` on `pkg/codeindex/parse_*.go`; there is no build tag to pass.

| Command | Purpose |
|---|---|
| `forge slug`, `forge validate`, `forge index`, `forge reindex`, `forge capture` | note contract, indexing, D3 capture |
| `forge recall` (deterministic scoring, JSON, `--explain`) | question → existing-note lookup |
| `forge drift`, `forge check` | git-anchored drift detection, full vault report |
| `forge config` (`--layers`, `--json`), `forge init` | config chain inspection, onboarding wizard |
| `forge engine select/run/record` | the zero-model-call binary's one named exception |
| `forge verify-code` (sandboxed compile check, bash/ts/java), `forge gate` (seven-gate `_inbox/` quarantine) | code verification, quality gate |
| `forge session-context`, `forge intent`, `forge session-capture`, `forge cache-source`, `forge stats` | Claude Code lifecycle hooks + stats |
| `forge logback` (`docs/knowledge-map.md`, per-module `CLAUDE.md` fragments, opt-in inline markers, `--remove-markers`, `--dry-run`) | log knowledge back into a code repo |
| `forge scrub <src> <dst>` | redacts secret/PII-shaped content, fails closed |
| `forge export-dataset`, `forge dataset-stats` | training-data export (D1-D6), stats |

## Testing

Two build lanes: `CGO_ENABLED=1 go test ./...` then `CGO_ENABLED=0 go build ./...` —
testing only the cgo lane can silently break the pure lane. Before touching
`pkg/recall`'s scoring, read `references/recall-spec.md` §2.3.1 for the IDF weighting
rationale: the vocabulary filter applies to `--stack` hints and **not** to question
terms (the reverse looks like the obvious reading and is wrong), `idf(0, n) == 0` is
correct and its test must not be inverted (the absent-term policy lives one layer up in
`weightsOver`), and §3.1's calibration table is **generated, not transcribed** — a
scoring change that doesn't update `cmd/forge/testdata/calibration.golden` is unreviewed.

## Comments

One or two lines, and only where a reader needs them. A doc comment is its first
sentence: two lines at most for an exported identifier, one for an unexported one, and
comment blocks longer than two lines do not belong inside a function body at all. The
tree was swept to this rule in one pass; do not re-inflate it.

Argument goes in `references/` and `docs/`, not in a comment block — that is where the
recall scoring channels, the duplicate threshold and the intent gate's 0.50 are already
derived, each against a golden the test suite checks. When a constraint has no home
there and breaking it does real damage, say it in one line at the thing it constrains:
`Extractor`, `reLongToken`'s character class and `excludedPrefixes` are the three that
earned it back after the sweep.

## Packaging

Version comes from two places and they are allowed to differ: `.claude-plugin/plugin.json`
declares the plugin version being developed toward, while the binary is stamped at link
time from `git describe` (`-X main.version` / `-X main.buildSHA`, see `Makefile`). Ask a
binary which it is with `forge version` — those vars have to exist in package `main` for
the stamp to land, because the Go linker silently ignores an `-X` naming a symbol that
is absent, which is what happened before `cmd/forge/version.go` existed.


`agents/*.md` (`forge-researcher`, `forge-codebase-scout`, `forge-verifier`,
`forge-librarian`) are the spec for the four product subagents, and they **are**
discovered when the plugin is loaded — as `forge:<name>`. An earlier note here claimed
the opposite and prescribed a manual Agent-tool dispatch as a workaround; that was
wrong, and it steered callers away from a mechanism that works. `skills/forge/SKILL.md`
dispatches to them by name.
