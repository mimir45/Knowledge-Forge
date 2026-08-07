# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Status

This is a **design-spec repository with no code**. `go.mod` (`module knowledge-forge`,
`go 1.26`) is the only non-doc artifact; there is no source tree, no `.git`, no README,
no Makefile. Phase 0 has not run. The one exception to "no code" is
`testdata/vault/` — a fixture vault of markdown, described below.

The project ("Knowledge Forge") is a Claude Code plugin that turns "explain X" moments
into permanent, linked, verified markdown notes in an Obsidian vault. Its defensible
core is a Go static-analysis engine that runs with **zero model calls**; an optional
four-tier LLM layer (none / host / API / advisor-critique) sits on top, configurable
per pipeline stage.

## Read the docs in this order

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
6. `docs/KNOWLEDGE-FORGE-B2B.md` — out of scope until Phase 7's gate ("OSS v2.0 shipped,
   30 days of real usage, ≥3 outside users reporting value").

Only surviving Python: the one-time `migrate_vault.py` and the offline dataset /
fine-tuning tooling. Neither ships in the binary.

## Things that live outside this repo

- **Vault:** `/Users/mimir45/Documents/Base` — 108 markdown notes. **Not a git repo.**
- Its current topology is `concepts/ decisions/ entities/ issues/ raw/ sources/
  syntheses/ archive/ TIL/`, which does **not** match DESIGN §7's prescribed
  `notes/{concept,howto,…}/ moc/ _inbox/ _archive/ profiles/`. Phase 1's migration is
  therefore a *topology change*, not just a frontmatter backfill — plan it that way.
- **v1 skill:** `/Users/mimir45/.claude/skills/til-writer/` — this is the system Phase 0
  audits and this project replaces. The user's global `~/.claude/CLAUDE.md` already
  routes "explain X" prompts into the same vault through it.
- Phase 0/1 preconditions are currently **unmet**: `git init` + commit the vault (the
  migration refuses a dirty tree, and D3 human-edit capture needs a post-commit hook on
  the vault repo).

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

`0 → 1 → 2 → 2b → 3 → 3b → 4 → 5 → 5b → 6 → 6b → 7`

One phase per session. Do not start phase N+1 with phase N unmerged. Never cut 2b; if
time runs out the cut order is `6b → 5b → advisor tier`. If work comes up outside the
current phase's scope, write it to `docs/BACKLOG.md` rather than building it.

**Read `docs/BACKLOG.md` at the start of a phase** — it holds open items B-001…B-004.
The one that changes how Phase 0 is run: the design docs have only ever been checked
against each other where they *self-flag* a conflict (the three precedence rules above).
No pass has been made for contradictions the docs don't announce. Phase 0's `AUDIT.md`
should carry a doc-vs-doc coherence section; resolve conflicts by precedence there
rather than editing the docs mid-flight.

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

## Target layout and budgets (not yet built)

```
cmd/forge/        CLI
pkg/vault/        frontmatter + markdown AST (goldmark), mtime-cached
pkg/similarity/   MinHash + LSH banding
pkg/graph/        note link graph: components, hubs, orphans, centrality
pkg/codeindex/    go-tree-sitter (start with Java + Kotlin only) — the only cgo package
pkg/gitsig/       go-git: churn, blame ownership, co-change coupling
pkg/drift/        the key package — ADDENDUM §B.6, AST comparison not line diffs
pkg/linkcheck/    HTTP HEAD on sources, cached, rate-limited
pkg/report/       renders analyses to markdown
pkg/store/        SQLite via modernc.org/sqlite, derived cache only
```

Latency budgets, to be **measured rather than assumed**: `forge drift` <100ms (the
binding constraint — it runs on the git-hook path), `forge index` <200ms, `forge check`
<10s warm, SessionStart hook <200ms.

## Commands

**There are none yet.** No `.go` files, no Makefile, no tests, no build or lint target —
do not assume `go build ./...` works. The intended surface, by the phase that creates it:

| Command | Phase |
|---|---|
| `forge slug`, `forge validate`, `forge index` | 1 |
| `forge recall` (deterministic scoring, JSON, `--explain`) | 2 |
| `forge drift`, `forge check`, `forge reindex`, cross-compile + goreleaser | 2b / 6 |
| `forge-init` wizard | 3 |
| `/forge-check`, `/forge-stats` | 5 |
| `/forge-export-dataset`, `/forge-dataset-stats` | 6b |

## Known discrepancies (record, don't fix)

- The Go module was renamed `TIL` → `knowledge-forge` on 2026-08-08 (bare path, no VCS
  host prefix — deliberately deferred, see BACKLOG B-004). Imports will read
  `knowledge-forge/pkg/vault`. The **directory is still `/Users/mimir45/TIL`** and the
  docs still call the artifact `knowledge-forge/`; that mismatch is cosmetic and stays
  (B-003). Don't rename the directory unasked.
- `docs/CLAUDE-CODE-PROMPT.md` says to put the docs in the repo root; they live in
  `docs/`. Don't shuffle files to match the prompt text.
- This repo has no `.git`, but Phase 0 is gated on `git init`.
