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
6. `docs/KNOWLEDGE-FORGE-B2B.md` — out of scope until Phase 7's gate ("OSS v2.0 shipped,
   30 days of real usage, ≥3 outside users reporting value").

Only surviving Python: the one-time `migrate_vault.py` and the offline dataset /
fine-tuning tooling. Neither ships in the binary.

## Things that live outside this repo

- **Vault:** `/Users/mimir45/Documents/Base` — 109 markdown notes. **Is a git repo** as of
  2026-08-09: 1 commit, clean tree.
- Its current topology is `concepts/ decisions/ entities/ issues/ raw/ sources/
  syntheses/ archive/ TIL/`, which does **not** match DESIGN §7's prescribed
  `notes/{concept,howto,…}/ moc/ _inbox/ _archive/ profiles/`. Phase 1's migration is
  therefore a *topology change*, not just a frontmatter backfill — plan it that way.
- **v1 skill:** `/Users/mimir45/.claude/skills/til-writer/` — this is the system Phase 0
  audits and this project replaces. The user's global `~/.claude/CLAUDE.md` already
  routes "explain X" prompts into the same vault through it. It contains **only
  `SKILL.md`** — no scripts, no agent definitions, no hooks, no plugin manifest. Phase 0's
  file-map step should expect ABSENT for most rows.
- Phase 0/1 preconditions are **met**: the vault repo exists and its tree is clean, so
  the migration's dirty-tree refusal and D3's post-commit hook both have something to
  attach to. The D3 hook itself is not installed yet (Phase 1 installs it).

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

**Read `docs/BACKLOG.md` at the start of a phase** — B-002…B-004 are open; B-001 (the
doc-vs-doc coherence pass) closed on 2026-08-09.

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
| `forge slug`, `forge validate`, `forge index`, `forge capture` | 1 |
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
