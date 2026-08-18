# Changelog

Seeded here for Phases 0 through 5b, condensed to one entry per phase rather than one
per commit — the phases were built and merged before this file existed. From Phase 6
on, entries follow `.goreleaser.yml`'s own changelog convention: commits prefixed
`docs:`, `test:`, or `chore:` are excluded from release notes.

## Phase 6 — Package & release (in progress)

- `pkg/scrub` / `forge scrub`: redacts secret/PII-shaped content from a vault copy;
  fails closed (AUDIT §8.4 D-6).
- `.claude-plugin/plugin.json` + `.claude-plugin/marketplace.json`: this repo is now an
  installable Claude Code plugin (`claude plugin marketplace add mimir45/Knowledge-Forge`).
- `docs/adr/0001-lexical-recall-vs-embeddings.md`, `docs/adr/0002-go-for-static-core.md`
  (AUDIT §8.4 D-3).
- README, LICENSE (MIT), CONTRIBUTING, this changelog.

## Phase 5b — Log-back into the codebase

- `forge logback`: generates `docs/knowledge-map.md` and per-module `CLAUDE.md`
  fragments in a target code repo, plus opt-in inline `// forge:logback:<symbol>`
  markers (`--remove-markers`, `--dry-run`).
- New package `pkg/sentinel`: id-based begin/end managed comment blocks
  (`Upsert`/`UpsertBefore`/`Remove`), shared by the CLAUDE.md fragments and inline
  markers.

## Phase 5 — Subagents, hooks & weekly checker

- Claude Code lifecycle hooks: `forge session-context`, `forge intent`,
  `forge session-capture`, `forge cache-source`; `hooks/hooks.json` declares the
  bindings.
- `forge stats`, `/forge-check`, `/forge-stats`.
- Git-anchored drift hooks for code repos: `scripts/install_drift_hook.sh`
  (`code-post-commit`, `code-post-merge`, `code-post-checkout`).
- Four product-agent spec files in `agents/` (`forge-researcher`,
  `forge-codebase-scout`, `forge-verifier`, `forge-librarian`).

## Phase 4 — Subagents & verification

- `forge verify-code`: sandboxed compile check (bash, TypeScript, Java) — never in the
  user's project, always a throwaway directory.
- `forge gate` + `pkg/qualitygate`: the seven DESIGN §12 gates, `_inbox/` quarantine on
  a blocking failure.
- B-007 closed: `agents/forge-librarian.md` stamps `Forge-Write: true` on every commit
  it authors.
- B-022 closed: the schema pattern now covers all nine `cfg.Pipeline` stages.

## Phase 3b — Engine abstraction

- `pkg/engine`: the four-backend abstraction (`none`/`host`/`api`/`advisor`), per-stage
  routing with fallback chains, SQLite-backed budget accounting, `engine_trail`
  stamping, `reports/cost.md`.
- `forge engine select/run/record` — the zero-model-call binary's one named exception.

## Phase 3 — Config chain

- `pkg/config`: the four-layer config chain (`FORGE_CONFIG` >
  `.forge.config.md` > `~/.forge/forge.config.md` > packaged
  `config/forge.config.example.md`).
- `forge config` (`--layers`, `--json`), `forge init`, `skills/forge-init/` wizard.

## Phase 2b — Drift detection & weekly reports

- `pkg/drift` (git-anchored AST comparison, not line diffs), `pkg/similarity`
  (hand-rolled MinHash + LSH — no embeddings), `pkg/gitsig`, `pkg/linkcheck`.
- `forge drift`, `forge check` — nine ADDENDUM §B.4 reports plus the codebase MOC.
- `bin/forge` shim, `.goreleaser.yml`, CI, and the six-target cross-compile Makefile
  lane.

## Phase 2 — Deterministic recall

- `pkg/recall`: deterministic question → note scoring, zero model calls.
- `forge recall` (JSON output, `--explain`).
- Two measured corrections against DESIGN §8's literal spec: a weighted **mean** over
  active channels (not a literal weighted sum), and **F₂** (not Dice) for the title
  measure — see `references/recall-spec.md`.

## Phase 1 — Note contract & vault migration

- `pkg/vault` (frontmatter + markdown AST, mtime-cached), `pkg/graph`.
- `forge slug`, `forge validate`, `forge index`, `forge reindex`, `forge capture`.
- The one-time `migrate_vault.py`: 91 notes moved to the seven-`notes/<type>/`
  topology, 345 wikilinks rewritten, 0 broken.
- The D3 human-correction capture hook (`git` `post-commit` in the vault).

## Phase 0 — Audit baseline

- `docs/AUDIT.md` (including the §8.4 binding decision record, D-1…D-8), the design
  docs, `go.mod`, the `testdata/vault/` fixture (F1–F12), the workflow-agent crew
  (`finder`, `executor`, `explainer`, `vault-analyst`, `doc-auditor`, `cross-checker`).
