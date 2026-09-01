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
session, from an initial design spec through to a packaged, installable plugin. The
roadmap (`docs/ROADMAP.md`) is complete.

## Read the docs in this order

1. **`docs/ROADMAP.md`** — condensed index over everything else. Always start here.
2. **`docs/KNOWLEDGE-FORGE-STACK.md`** (ADR-001) — **wins on every stack question.** It
   supersedes ADDENDUM §B (which specified Python — the doc itself says "that was
   wrong") and B2B §8 (which assumed Spring Boot — now an open decision, ADR-002).
3. **`docs/KNOWLEDGE-FORGE-DESIGN.md`** — the master spec (schema, pipeline, gates,
   vault topology, subagents). Its rev-2 note means every `scripts/*.py` reference reads
   as a `forge` subcommand.
4. **`docs/KNOWLEDGE-FORGE-ADDENDUM.md`** — engine tiers (§A), no-AI capability boundary
   and the reports (§B), drift detection (§B.6), weekly checker (§C), datasets (§D),
   full config YAML + presets (§E).
5. **`docs/CLAUDE-CODE-PROMPT.md`** — the phase-by-phase prompts this project was
   actually built from.
6. `docs/KNOWLEDGE-FORGE-B2B.md` — describes a **separate project**, not a phase of this
   one. Kept in this repo only for reference/history.
7. `docs/adr/` — two standalone ADRs: why lexical recall over embeddings, why Go for
   the static core.

Only surviving Python: the one-time `migrate_vault.py` and the offline dataset /
fine-tuning tooling. Neither ships in the binary.

## Things that live outside this repo

- **Vault:** wherever `--vault` / the config chain points (see `forge config
  --layers`), a separate git repo laid out per DESIGN §7's `notes/<type>/ moc/ _inbox/
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

## Agent crew (`.claude/agents/`)

The top-level session manages; three project-scoped Sonnet subagents do the work. Route
by verb: **find → `finder`** (read-only search, reports `file:line` hits, also searches
the configured vault), **do → `executor`** (Read/Write/Edit/Bash; stays in scope,
verifies with real command output), **explain → `explainer`** (read-only; writes
nothing, so TIL notes stay with a dedicated note-writing skill).

Two more for audit work: **`vault-analyst`** (read-only vault metrics — counts,
frontmatter key frequency, inbound links, orphans, near-dupes) and **`doc-auditor`**
(finds contradictions between the design docs that they don't self-flag).

And one competing run: **`cross-checker`** — independently re-derives another agent's
numbers or findings and returns strict JSON, one verdict per claim, each `links`-ed to
the primary's finding ID. Spawn it **in parallel with** the primary, not after: a
checker that has already seen the answer anchors to it. `vault-analyst` and
`doc-auditor` therefore end their reports with a JSON block whose IDs match their prose,
so the two runs join mechanically. Use it whenever a number is going into a document
later work will re-measure against.

Prefer delegating over doing it inline. Independent tasks go out in parallel.

These are **workflow** agents for building the project. They are not the four
**product** agents (`forge-researcher`, `forge-codebase-scout`, `forge-verifier`,
`forge-librarian`) that DESIGN §11 specifies and that live as spec files under
`agents/*.md` — see the packaging note below.

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
pkg/drift/        the key package — ADDENDUM §B.6, AST comparison not line diffs
pkg/linkcheck/    HTTP HEAD on sources, cached, rate-limited
pkg/report/       renders analyses to markdown; must not import pkg/codeindex
pkg/store/        SQLite via modernc.org/sqlite, derived cache only except the budget table
pkg/engine/       none/host/api/advisor backends, per-stage select+fallback, engine_trail
pkg/config/       the four-layer config chain
pkg/sentinel/     id-based begin/end managed comment blocks; Upsert/UpsertBefore/Remove
pkg/scrub/        redacts secret/PII-shaped content from a vault copy; fails closed
pkg/dataset/      D1-D6 training-data capture and export; D6 is derived, not captured
pkg/qualitygate/  the seven DESIGN §12 gates + orchestration + `_inbox/` quarantine
pkg/telemetry/    the `ask` event, sha256 topic hashing, never raw question text
```

Latency budgets and measured actuals on an Apple M4: `forge drift` <100ms → **60–70ms**
(the binding constraint — it runs on the git-hook path), `forge index` <200ms → **20ms**,
`forge check` <10s warm → **390ms** (930ms cold). `pkg/qualitygate.Run`'s six in-process
gates minus `code` (schema, citation, freshness, antislop, link, duplicate) →
**~0.13ms** per run; `forge verify-code` per invocation, dominated by toolchain startup,
not gate logic → bash **~10ms warm** (~470ms cold), java **~170ms warm** (~370ms cold).
The two Claude Code hook commands measured the same warm/cold way: `forge
session-context` <200ms budget and `forge intent` <50ms budget both land **well under
budget warm** — the intent budget's plausibility comes from reusing `forge recall`'s
already-warm SQLite cache.

## Config chain

Four layers, highest precedence first: `FORGE_CONFIG` env var > `.forge.config.md` (repo
root) > `~/.forge/forge.config.md` (user) > the packaged `config/forge.config.example.md`
template. The schema is the *union* of ADDENDUM §E and the DESIGN §10 keys §E never
restates. `forge init` is the **only** writer of `~/.forge/forge.config.md` and
`<vault>/profiles/me.md` (rendered from `profiles/me.template.md`) — never
`config/forge.config.md`, which stays a packaged template. `on_exhausted` defaults to
`queue`; accepted values are `queue | degrade | stop` — `queue` stamps
`pending_advisor: true` and falls through to `none`, `degrade` is today's silent
`none`-fallback (the honest reading of the word), `stop` exits non-zero without calling
the tier. Budget counters live in SQLite and must survive `forge reindex`.

## Commands

`CGO_ENABLED=0 go build ./...` and `go test ./...` both work; `make build test bench
dist install-hook` covers the rest. Two build lanes, because `pkg/codeindex` is the one
cgo package and is build-tag gated — the default lane is pure Go and cross-compiles; the
codeindex lane needs cgo and a host toolchain (`make full`, `CGO_ENABLED=1 -tags
codeindex`).

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

## Packaging

`agents/*.md` (`forge-researcher`, `forge-codebase-scout`, `forge-verifier`,
`forge-librarian`) are correct spec for the four product subagents DESIGN §11
describes, but nothing in this repo auto-discovers agents from a root-level `agents/`
directory — Claude Code loads `.claude/agents/`. `skills/forge/SKILL.md`'s dispatch to
them goes through the generic Agent tool with an explicit tool allowlist, not live
agent auto-discovery, until a plugin-level mechanism exists for it.
