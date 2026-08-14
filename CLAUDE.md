# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Status

**Phases 0, 1, 2, 2b, 3, 3b, 4, 5 and 5b are done** (2026-08-09 through 2026-08-14; Phase 1
merged as `1c9df95`, Phase 2 as `3619b72`, Phase 2b committed straight to `main`,
`cb12a08`…`15a795f`; Phase 3 committed straight to `main` in one commit; Phase 3b
likewise, on top of `847098a`; Phase 4 likewise, on top of `884e42e`; Phase 5 likewise,
on top of Phase 4's commit; Phase 5b likewise, on top of Phase 5's commit). The repo is
a git repo with a Go source tree: `cmd/forge`
(`slug validate index reindex capture recall drift check config init engine verify-code
gate session-context intent session-capture cache-source stats logback`) over `pkg/vault`,
`pkg/graph`, `pkg/report` (now including `weekly.go`'s rollup renderer,
`weekly_store.go`'s week-over-week `.forge/weekly-stats.json` persistence, and, new in
Phase 5b, `knowledgemap.go`'s `RenderKnowledgeMap`), `pkg/store`,
`pkg/dataset`, `pkg/recall`, `pkg/similarity`, `pkg/codeindex`, `pkg/coderef`,
`pkg/gitsig`, `pkg/drift`, `pkg/linkcheck`, `pkg/config` (the four-layer config chain),
`pkg/engine` (the four-backend engine abstraction — `none/host/api/advisor`, per-stage
routing with fallback chains, SQLite-backed budget accounting), `pkg/qualitygate` (the
seven DESIGN §12 gates + `Run`/`Report` orchestration + `_inbox/` quarantine),
`pkg/telemetry` (DESIGN §14's `ask` event, sha256 topic hashing, never
raw question text; gated fully behind `cfg.Telemetry.Enabled` and wired into `forge
recall`), `pkg/sentinel` (new in Phase 5b — the id-based begin/end managed-block
primitive `forge logback` uses for CLAUDE.md fragments and inline markers; `Upsert`/
`UpsertBefore`/`Remove`, idempotent, position-independent), plus seven note templates in
`templates/`, `skills/forge/SKILL.md`,
`skills/forge-init/SKILL.md`, `skills/forge-check/SKILL.md` and `skills/forge-stats/
SKILL.md`, `references/recall-spec.md`, `references/writing-rules.md`,
eight packaged presets in `config/presets/`, a Makefile with a six-target cross-compile
matrix, a hash-verifying `bin/forge` shim, a `hooks/` + `scripts/` pair that now installs
both the vault's D3 capture hook (`vault-post-commit`) and four Claude
Code lifecycle shims (`session-context`, `user-prompt-intent`, `session-end-capture`,
`post-tool-cache-source`, declared in `hooks/hooks.json`) plus three git-anchored
drift shims for code repos (`code-post-commit`, `code-post-merge`, `code-post-checkout`,
installed via `scripts/install_drift_hook.sh`), and four `agents/*.md` spec
files (`forge-researcher`, `forge-codebase-scout`, `forge-verifier`, `forge-librarian` —
spec only, see the packaging-gap note below). Build and test with `CGO_ENABLED=0 go
build ./...` / `go test ./...` — 18 packages report `ok` (`config`, `profiles`,
`references` are data-only, no test files), all green under both `CGO_ENABLED=0` and
`CGO_ENABLED=1`.
Phase 5's own `forge session-context` / `forge intent` warm-latency check ran 20
iterations against synthetic stdin; both landed well under budget (<200ms / <50ms).
`forge check` against the real vault confirmed the new `moc/weekly/YYYY-WW.md` rollup
renders, is byte-identical on a second immediate run, and `.forge/weekly-stats.json`
persists across runs without zeroing or duplicating deltas within the same ISO week. Two
new backlog items recorded rather than fixed this phase: **B-025** (`forge cache-source`'s
`PostToolUse`/WebFetch `tool_response` JSON shape was never confirmed from official docs,
so `cacheBody` deliberately caches the raw bytes rather than guessing a field name) and
**B-026** (a citation to a fully deleted file can never verdict BROKEN, because
`registryOf` always builds `pkg/coderef`'s registry from the current `HEAD` tree — found
smoke-testing Phase 5's drift hooks, pre-existing in Phase 2b's `pkg/drift`, not this
phase's to fix). The packaging gap already on file for root-level `agents/` now also
covers `hooks/hooks.json`: nothing in this repo auto-installs it into
`~/.claude/settings.json` or a project's `.claude/settings.json` — closed only when
Phase 6's plugin manifest lands.
Phase 5b (ADDENDUM §B.7 / DESIGN §15, "Log-back into the codebase") built `forge logback`:
T0, deterministic, idempotent, generates `docs/knowledge-map.md` and per-module `CLAUDE.md`
fragments in the target code repo (both gated independently by
`static.logback.{knowledge_map,claude_md_fragment}`), plus opt-in inline
`// forge:logback:<symbol>` markers (`static.logback.inline_markers`, default off,
revertible via `--remove-markers`) — new package `pkg/sentinel` is the id-based
begin/end managed-block primitive both the CLAUDE.md fragments and the inline markers
share, and it never touches anything outside its own marker pair. `.forge/code-index-
<repo>.json` freshness was already solved by Phase 2b's `pkg/drift`/`pkg/codeindex` and
is reused as-is, not rebuilt. Verified via `logback_test.go` (dispatch, full pipeline,
idempotent rerun, `--remove-markers` round-trip, per-flag config gating) passing under
both `CGO_ENABLED=0` and `CGO_ENABLED=1`, plus a hand-built smoke test against a real
temp git repo confirming byte-identical reruns (`diff`, no output) and a byte-for-byte
`--remove-markers` revert. One correctness fix worth naming: inline-marker resolution
must key off `coderef.Ref.Symbol != ""`, not `Ref.Kind == KindSymbol` — the canonical
`code_refs:` citation form (`repo:path#Symbol`) parses to `KindPath` with `Symbol` set,
so filtering on `Kind` alone would silently skip nearly every real citation. New backlog
item recorded rather than fixed this phase: **B-027** (`pkg/drift/gitindex.go` caches
per-repo as `.forge/code-index-<repo>.json`, while ADDENDUM/DESIGN both describe the
singular `.forge/code-index.json` — correct behavior, since one shared name would
collide across repos, but undocumented; found during Phase 5b's explore pass,
pre-existing since Phase 2b, not this phase's to fix).
Everything else below is still design spec; **Phase 6 is next.**
One item Phase 3 explicitly did not touch: **B-008's §3.1 recalibration** — see BACKLOG, it
needs its own session because honest verification means re-deriving the whole calibration
table, not re-running two queries. **B-023** is still open, recorded rather than fixed:
code's `on_exhausted: stop` vs. every doc's `fail`, and `stop`/`degrade` are behaviorally
identical to each other today — nothing reads either value. **B-022 closed in Phase 4**
(the schema pattern now covers all nine `cfg.Pipeline` stages minus `critique`); **B-007
closed in Phase 4** (`agents/forge-librarian.md`'s prompt stamps `Forge-Write: true` on
every commit it authors, and `pkg/dataset/d3_forge_write_test.go` pins the guard both
ways). One new item Phase 4 found and recorded rather than fixed: **B-024**
(`pkg/dataset/d2.go`'s `D2Tag = "d2_advisor"` never matches the packaged config's `"d2"`
list entry, so D2 capture is silently inert under the shipped config).
**Packaging gap, recorded rather than implied fixed:** nothing in this repo loads agents
from a root-level `agents/` directory — Claude Code loads `.claude/agents/`, and no
plugin manifest exists yet (Phase 0's finding, still true). The four `agents/*.md` files
are correct spec for when packaging exists but are not live, dispatchable agents today;
`skills/forge/SKILL.md`'s dispatch to them is verified today via the generic Agent tool
with an explicit tool allowlist, not live agent auto-discovery.
`testdata/vault/` is a markdown fixture, described below.

Phase 2b's measured actuals, so no later phase re-derives them: `forge index` 0.02s,
`forge drift --since-commit` 0.06–0.07s (budget 100ms, the binding one), `forge check`
0.93s cold / 0.39s warm. Nine reports render deterministically — six consecutive runs,
md5-identical. Against the real vault: drift finds **9 notes referencing changed code**
(2 broken, 7 suspect) over 140 citations; 21 of 94 orphans; 23 graph components; 3
duplicate pairs ≥0.40; 39 of 41 stacks covered. Two knowing deviations: `pkg/gitsig`
shells out to the `git` CLI rather than go-git (**B-009**), and **B-008 is still open** —
the IDF weighting it prescribes shipped and fixed neither named case, for a reason the
backlog entry now records. Do not respond to that by moving the thresholds.

Two Phase 2 decisions that later phases must not undo without reading
`references/recall-spec.md` first: the score is a weighted **mean over active
channels**, not DESIGN §8's literal weighted sum (§2.5), and the title measure is **F₂,
not Dice** (§2.2). Both are argued from measured vault behaviour. The verdict ships
inside `forge recall`'s JSON envelope so nothing downstream restates the threshold tree
— AUDIT §8.4 D-7 moves those thresholds into Phase 3's config chain. Thresholds stay at
DESIGN §5.3's 0.85 / 0.55; the calibration sweep is spec §3.1 and its one open defect is
BACKLOG **B-008**.

The project ("Knowledge Forge") is a Claude Code plugin that turns "explain X" moments
into permanent, linked, verified markdown notes in an Obsidian vault. Its defensible
core is a Go static-analysis engine that runs with **zero model calls**; an optional
four-tier LLM layer (none / host / API / advisor-critique) sits on top, configurable
per pipeline stage.

## Read the docs in this order

0. **`docs/AUDIT.md` §8.4** — the binding decision record (D-1 … D-8). Read it *first*,
   because the design docs below were deliberately **not** edited: where §8.4 marks a line
   stale, the doc still says the old thing and §8.4 is what you follow. Details under
   "Phase workflow".
1. **`docs/ROADMAP.md`** — condensed index over everything else. Always start here.
2. **`docs/KNOWLEDGE-FORGE-STACK.md`** (ADR-001) — **wins on every stack question.** It
   supersedes ADDENDUM §B (which specified Python — the doc itself says "that was
   wrong") and B2B §8 (which assumed Spring Boot — now an open decision, ADR-002).
3. **`docs/KNOWLEDGE-FORGE-DESIGN.md`** — the master spec (schema, pipeline, gates,
   vault topology, subagents). Its rev-2 note means every `scripts/*.py` reference reads
   as a `forge` subcommand.
4. **`docs/KNOWLEDGE-FORGE-ADDENDUM.md`** — engine tiers (§A), no-AI capability boundary
   and the 10 reports (§B), drift detection (§B.6), weekly checker (§C), datasets (§D),
   full config YAML + presets (§E).
5. **`docs/CLAUDE-CODE-PROMPT.md`** — the actual execution mechanism: a ready-to-paste
   prompt per phase.
6. `docs/KNOWLEDGE-FORGE-B2B.md` — describes a **separate project**, not a phase of this
   one (BACKLOG B-021). Kept in this repo only for reference/history.

Only surviving Python: the one-time `migrate_vault.py` and the offline dataset /
fine-tuning tooling. Neither ships in the binary.

## Things that live outside this repo

- **Vault:** `/Users/mimir45/Documents/Base`, a git repo. **Migrated by Phase 1** on
  2026-08-09: 91 notes moved to DESIGN §7's `notes/<type>/ moc/ _inbox/ _archive/
  profiles/` topology, 345 wikilinks rewritten, 0 broken, 60/91 schema-valid (the 31
  failures are 47 issues needing human judgment — see `lint-report.md` in the vault).
  All seven `notes/<type>/` subdirs exist per B-005. Rollback: backup at
  `/Users/mimir45/Documents/Base-backup-2026-08-09`, or vault commit `b3168f0`.
  `raw/` (5) and `sources/` (9) stay live and outside the note contract; the other old
  topology dirs survive as empty `.gitkeep` shells.
- **v1 skill:** `/Users/mimir45/.claude/skills/til-writer/` — this is the system Phase 0
  audits and this project replaces. The user's global `~/.claude/CLAUDE.md` already
  routes "explain X" prompts into the same vault through it. It contains **only
  `SKILL.md`** — no scripts, no agent definitions, no hooks, no plugin manifest. Phase 0's
  file-map step should expect ABSENT for most rows.
- **The D3 hook is installed and live** in the vault: `.git/hooks/post-commit` runs
  `forge capture` from **`~/.forge/bin/forge`** (the absolute path is pinned in
  `<vault>/.forge/forge-bin`; `$FORGE_BIN` overrides it). That binary is a **copy**, not
  the repo's build output — **rebuild it after any change to `pkg/dataset` or
  `cmd/forge/capture.go`**: `CGO_ENABLED=0 go build -o ~/.forge/bin/forge ./cmd/forge`.
  By design the hook can never fail a commit and never prints, so a stale or broken
  binary is silent: if pairs stop appearing, read `<vault>/.forge/capture.log`. It
  captures nothing today — every migrated note is `origin: import` — and starts paying
  off in Phase 4. Uninstall is `rm .git/hooks/post-commit`.

## Fixture vault (`testdata/vault/`)

A 13-note fixture reproducing the real vault's **pre-migration** topology plus twelve
deliberate defects (F1–F12) — mixed frontmatter shapes, a dangling wikilink, a dangling
`source:` path, an orphan, a near-duplicate pair, notes with no frontmatter at all,
status carried as body prose. Catalogue: `testdata/README.md`.

- Rehearse anything that mutates a vault here first. Phase 1's migration is irreversible
  and the real vault has no backups.
- It has **no `.git`, deliberately** — a nested repo would become a stray gitlink once
  this repo is `git init`-ed. The harness copies the fixture into a temp dir and
  `git init`s the copy; that is how the migration's "refuses a dirty tree" precondition
  and drift's `--since-commit` get exercised. Never `git init` it in place.
- The defects are the test surface. **Do not fix them.**
- It is **not** `examples/vault/`, which is a separate, unbuilt Phase 6 deliverable.

## Phase workflow

`0 → 1 → 2 → 2b → 3 → 3b → 4 → 5 → 5b → 6 → 6b`

This repo's roadmap ends at 6b — B2B (`docs/KNOWLEDGE-FORGE-B2B.md`) is a fully separate
project, not a phase gated inside this one; see BACKLOG B-021. One phase per session. Do not start phase N+1 with phase N unmerged. Never cut 2b; if
time runs out the cut order is `6b → 5b → advisor tier`. If work comes up outside the
current phase's scope, write it to `docs/BACKLOG.md` rather than building it.

**Read `docs/BACKLOG.md` at the start of a phase** — B-002…B-004, **B-007**, **B-008**,
**B-009** and the twelve findings 2b recorded are open; B-001 (doc coherence), B-005
(seven note types) and B-006 (link rewrite) closed on 2026-08-09. B-007 is Phase 4's:
`forge-librarian` must stamp `Forge-Write: true` on every commit it authors, or
`pkg/dataset` records its output as human corrections. **B-008 is now Phase 3's** and its
entry has a second half worth reading before touching `pkg/recall`: the weighting the first
half prescribes is already implemented and did not fix either case, because the terms that
carry a question's meaning are filtered out of the denominator when no note carries them.
The next attempt owns the §3.1 recalibration.

**Then read `docs/AUDIT.md` §8.** It is the output of that pass: thirteen contradictions
the docs do *not* self-flag, eight resolved by the precedence rule above. **§8.4 is a
binding decision record** (D-1 … D-8) covering the six precedence could not settle. The
design docs were deliberately **not** edited, so where §8.4 marks a line stale the doc
still says the old thing — §8.4 is what you follow. It changes Phase 3, 3b, 6 and 6b:

- **3** — config is a four-layer chain (`FORGE_CONFIG` > `.forge.config.md` > `~/.forge/forge.config.md` > packaged `config/forge.config.example.md`); the schema is the *union* of ADDENDUM §E and the DESIGN §10 keys §E never restates; `forge init` is the **only** writer of `~/.forge/forge.config.md` and `<vault>/profiles/me.md` (rendered from `profiles/me.template.md`) — never `config/forge.config.md`, which stays a packaged template; `skills/forge-init/` asks the questions and shells out.
- **3b** — `on_exhausted` defaults to `queue`; `cost.md` is built here, not in 2b; budget counters live in SQLite and must survive `forge reindex`.
- **2b** — ships **nine** reports, not ten.
- **6** — build `pkg/scrub` / `forge scrub` and use it to generate `examples/vault/`; it needs a fixture test before the phase passes. Ship exactly two ADR files: `0001-lexical-recall-vs-embeddings` (from DESIGN §8) and `0002-go-for-static-core` (from STACK §1).
- **6b** — `--anonymize` calls `pkg/scrub` and **fails closed**; it never exports raw on scrubber error.

## Agent crew (`.claude/agents/`)

The top-level session manages; three project-scoped Sonnet subagents do the work. Route
by verb: **find → `finder`** (read-only search, reports `file:line` hits, also searches
the vault at `/Users/mimir45/Documents/Base`), **do → `executor`** (Read/Write/Edit/Bash;
stays in scope, verifies with real command output), **explain → `explainer`**
(read-only; writes nothing, so TIL notes stay with the `til-writer` skill).

Two more for audit work: **`vault-analyst`** (read-only vault metrics — counts,
frontmatter key frequency, inbound links, orphans, near-dupes) and **`doc-auditor`**
(finds contradictions between the design docs that they don't self-flag — Backlog B-001).

And one competing run: **`cross-checker`** — independently re-derives another agent's
numbers or findings and returns strict JSON, one verdict per claim, each `links`-ed to the
primary's finding ID. Spawn it **in parallel with** the primary, not after: a checker that
has already seen the answer anchors to it. `vault-analyst` and `doc-auditor` therefore end
their reports with a JSON block whose IDs match their prose, so the two runs join
mechanically. Use it when a number is going into a document later phases re-measure
against — the Phase 0 baseline table is the case that motivated it.

Prefer delegating over doing it inline. Independent tasks go out in parallel.

These are **workflow** agents for building the project. They are not the four **product**
agents (`forge-researcher`, `forge-codebase-scout`, `forge-verifier`, `forge-librarian`)
that DESIGN §11 specifies and Phase 4 builds into `agents/`. Deferred to Phase 1, when
there is code to point them at: a `go-reviewer` and a `test-writer`.

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
- CLI only for v1. Do not build the daemon on speculation — measure first.

## Layout and budgets — all built as of 3b

```
cmd/forge/        CLI
pkg/vault/        frontmatter + markdown AST (goldmark), mtime-cached
pkg/recall/       deterministic question -> note scoring; zero model calls
pkg/similarity/   MinHash + LSH banding
pkg/graph/        note link graph: components, hubs, orphans, centrality
pkg/codeindex/    go-tree-sitter (Java + TypeScript) — the only cgo package, tag-gated
pkg/coderef/      extracts code citations from note bodies and frontmatter
pkg/gitsig/       churn, ownership, co-change coupling — via the git CLI, not go-git (B-009)
pkg/drift/        the key package — ADDENDUM §B.6, AST comparison not line diffs
pkg/linkcheck/    HTTP HEAD on sources, cached, rate-limited
pkg/report/       renders analyses to markdown; must not import pkg/codeindex
pkg/store/        SQLite via modernc.org/sqlite, derived cache only except the budget table
pkg/engine/       none/host/api/advisor backends, per-stage select+fallback, engine_trail
pkg/config/       the four-layer config chain
pkg/sentinel/     id-based begin/end managed comment blocks; Upsert/UpsertBefore/Remove
```

Latency budgets and the **measured** actuals on an Apple M4: `forge drift` <100ms → **60–70ms**
(the binding constraint — it runs on the git-hook path), `forge index` <200ms → **20ms**,
`forge check` <10s warm → **390ms** (930ms cold). Phase 4 adds two more, measured, since
DESIGN sets no combined gate-pipeline budget: `pkg/qualitygate.Run`'s six in-process
gates minus `code` (schema, citation, freshness, antislop, link, duplicate) →
**~0.13ms** per run, far under the informal sub-100ms target set against `forge check`'s
existing warm figure above; `forge verify-code` per invocation, dominated by toolchain
startup, not gate logic → bash **~10ms warm** (~470ms cold, one-time OS page-cache
effect), java **~170ms warm** (~370ms cold). `tsc` is not installed in this environment,
so the TypeScript lane is untested here — `TestCompileTSSkippedWhenToolchainAbsent`
covers the absent-toolchain path instead. Phase 5 measured its two Claude Code hook
commands the same warm/cold way: `forge session-context` <200ms budget → measured **well
under budget warm** over 20 iterations against synthetic stdin; `forge intent` <50ms
budget → likewise **well under budget warm**, the reuse of `forge recall`'s already-warm
SQLite cache being what makes that budget plausible at all. `hooks/hooks.json` declares
the bindings but nothing in this repo installs it into a live `settings.json` yet (see
the packaging-gap note in Status), so these are direct-invocation measurements, not a
measurement of a live session.

## Commands

`CGO_ENABLED=0 go build ./...` and `go test ./...` both work, and 2b added a `Makefile`:
`make build test bench dist install-hook`. There is still no lint target. Two build lanes,
because `pkg/codeindex` is the one cgo package and is build-tag gated — the default lane is
pure Go and cross-compiles; the codeindex lane needs cgo and a host toolchain. Phases 1, 2,
2b, 3, 3b, 4, 5 and 5b's commands ship; the rest is the intended surface, by the phase that creates it:

| Command | Phase |
|---|---|
| `forge slug`, `forge validate`, `forge index`, `forge reindex`, `forge capture` | 1 — **built** |
| `forge recall` (deterministic scoring, JSON, `--explain`) | 2 — **built** |
| `forge drift`, `forge check`, cross-compile + goreleaser | 2b — **built** |
| `forge config` (`--layers`, `--json`), `forge init`, `skills/forge-init/` wizard | 3 — **built** |
| `forge engine select/run/record` — the zero-model-call binary's one named exception | 3b — **built** |
| `forge verify-code` (sandboxed compile check, bash/ts/java), `forge gate` (seven-gate `_inbox/` quarantine) | 4 — **built** |
| `forge session-context`, `forge intent`, `forge session-capture`, `forge cache-source`, `forge stats`, `/forge-check`, `/forge-stats`, git-anchored drift hooks (`scripts/install_drift_hook.sh`) | 5 — **built** |
| `forge logback` (`docs/knowledge-map.md`, per-module `CLAUDE.md` fragments, opt-in inline markers, `--remove-markers`, `--dry-run`) | 5b — **built** |
| `/forge-export-dataset`, `/forge-dataset-stats` | 6b |

## Known discrepancies (record, don't fix)

- The Go module was renamed `TIL` → `knowledge-forge` on 2026-08-08 (bare path, no VCS
  host prefix — deliberately deferred, see BACKLOG B-004). Imports will read
  `knowledge-forge/pkg/vault`. The **directory is still `/Users/mimir45/TIL`** and the
  docs still call the artifact `knowledge-forge/`; that mismatch is cosmetic and stays
  (B-003). Don't rename the directory unasked.
- `docs/CLAUDE-CODE-PROMPT.md` says to put the docs in the repo root; they live in
  `docs/`. Don't shuffle files to match the prompt text.
- BACKLOG **B-005** decided seven note types against DESIGN §7's five-directory tree; all
  seven `notes/<type>/` subdirs now exist in the vault, three of them empty `.gitkeep`
  shells. Don't prune them to match §7.
