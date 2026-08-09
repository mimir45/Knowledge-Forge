# Knowledge Forge — Roadmap (brief)

> Condensed index over the full design. For details on any phase, follow the
> link in that row — this doc doesn't restate what's already written there.
>
> Source docs: [`KNOWLEDGE-FORGE-DESIGN.md`](./KNOWLEDGE-FORGE-DESIGN.md) ·
> [`KNOWLEDGE-FORGE-ADDENDUM.md`](./KNOWLEDGE-FORGE-ADDENDUM.md) ·
> [`KNOWLEDGE-FORGE-STACK.md`](./KNOWLEDGE-FORGE-STACK.md) ·
> [`CLAUDE-CODE-PROMPT.md`](./CLAUDE-CODE-PROMPT.md)
>
> [`KNOWLEDGE-FORGE-B2B.md`](./KNOWLEDGE-FORGE-B2B.md) describes a **separate
> project**, kept here only for reference/history — see "Sequencing notes"
> below and BACKLOG B-021.

## What it is

A Claude Code plugin that turns "explain X" moments into permanent, linked,
verified markdown notes in an Obsidian vault, so the second time the question
comes up it's a vault read instead of a research run. The defensible core is a
Go static-analysis engine (dedup recall, schema validation, drift detection,
report generation) that runs with **zero model calls**; an optional four-tier
LLM layer (none / host / API / advisor-critique) sits on top, configurable per
pipeline stage. The B2B "no context loss" product (Slack/GitHub ingestion →
org wiki → MCP server) that earlier revisions of this roadmap groundworked
for is now **out of this repo entirely** — a separate project, started only
after this one's own readiness gate (see "Sequencing notes").

**Status (2026-08-09):** Phases 0, 1, 2 and 2b are **done** — 13 Go packages,
`go test ./...` green, real vault migrated and measured (drift, dedup,
coverage numbers in `CLAUDE.md`'s Status section). Phase 3 (config chain +
`forge init`) is **in progress**, uncommitted in the working tree. Everything
from 3b on is still design only.

## Phase order

`0 → 1 → 2 → 2b → 3 → 3b → 4 → 5 → 5b → 6 → 6b`

One phase per Claude Code session (sized to fit context). Do not start phase
N+1 with phase N unmerged.

| # | Phase | Goal | Key deliverables | Done when | Est. |
|---|---|---|---|---|---|
| 0 ✅ | **Audit** | Establish the factual baseline before changing anything | `docs/AUDIT.md`: current file map, vault metrics (notes, frontmatter coverage, links, orphans, dup clusters), grading vs. design §6-14, F1-F10 confirmation, drift baseline if a repo sits alongside the vault | Baseline numbers exist and every later phase is measured against them | 0.5 day |
| 1 ✅ | **Contract & migration** | Give every note a machine-readable, validated shape | `references/schema.yaml`, `templates/*.md` per type, `forge slug`, `forge validate`, `forge index`, one-time `migrate_vault.py` (dry-run default), D3 human-edit capture hook | 100% of notes validate; `_index.md` builds in one command | 1-2 days ⭐ highest value/hour |
| 2 ✅ | **Recall** | The compounding feature: answer from the vault instead of re-researching | `forge recall` (deterministic scoring, JSON output, `--explain`), `references/recall-spec.md`, `SKILL.md` rewritten around the pipeline (<200 lines) | Known question → `ANSWER_FROM_VAULT` in <5s, zero new files created | 1 day (mostly hardening — recall largely exists) |
| 2b ✅ | **Static core** ⭐ | The no-AI engineering layer — the most defensible part of the project | Go binary (`bin/forge`): `pkg/{vault,similarity,graph,codeindex,gitsig,drift,linkcheck,report,store}`, all 10 reports, `moc/codebase.md`, git-anchored drift with rollback symmetry, cross-compile + goreleaser | `forge drift` <100ms, `forge index` <200ms, full check <10s warm — measured, not assumed | 4-6 days (incl. Go ramp) — **never cut** |
| 3 🔧 | **Config & personalization** | Installable by a stranger with zero code edits | `forge.config.md`, presets (`java-backend`, `frontend`, `devops`, `minimal`), `profiles/me.md`, `forge-init` wizard, every hardcoded path removed | A stranger installs and runs it without editing plugin files; same question at different `seniority`/`depth` produces visibly different notes | 1-2 days |
| 3b | **Engine abstraction** ⭐ | One interface, four backends, cost-aware routing | `pkg/engine` interface, per-stage config with fallback chains, hard locks on recall/write/index, advisor in critique-only mode, budget accounting, `engine_trail` in frontmatter | Same question runs cleanly under all 4 presets (`offline`/`claude-only`/`byo-api`/`max`); `offline` degrades usefully instead of failing | 1-2 days |
| 4 | **Subagents & verification** | Catch bad notes before they publish; ground notes in the actual repo | 4 agent defs (`forge-researcher`, `forge-codebase-scout`, `forge-verifier`, `forge-librarian`), parallel researcher+scout, quality gates (schema/citation/code/freshness/anti-slop/link/duplicate), `_inbox/` quarantine | A deliberately wrong snippet gets caught and demoted, not published | 2-3 days |
| 5 | **Hooks & weekly checker** | Make it a system, not a command you remember to run | `hooks/hooks.json` (SessionStart/UserPromptSubmit/SessionEnd/PostToolUse shims), `/forge-check` (T0-only weekly report), `/forge-stats` | A fresh session's first response cites an existing note unprompted | 2 days |
| 5b | **Log-back into codebase** | Knowledge discoverable from code, not just the vault | `docs/knowledge-map.md`, per-module `CLAUDE.md` fragments (sentinel-managed), `.forge/code-index.json` kept fresh; inline markers opt-in only | Generated knowledge-map and one CLAUDE.md fragment shown and correct | 1 day |
| 6 | **Package & release** | Ship as a public, installable plugin | `.claude-plugin/plugin.json` + `marketplace.json`, `bin/forge` shim w/ checksum verify, evals (`triggers.yaml`, golden notes, CI), README in the prescribed order, ADRs, `examples/vault/` | `claude plugin marketplace add <you>/forge` works from a clean box | 2-3 days |
| 6b | **Dataset capture & export** | Turn normal use into an owned, exportable training corpus | D1-D5 capture wired (`.forge/datasets/*.jsonl`), `/forge-export-dataset`, datasheets per export, `/forge-dataset-stats` with honest trainability claims | Exports run with anonymization + datasheet; stats report doesn't overclaim on volume | 1-2 days |

**Total: ~4-5 weeks of focused evenings** to a public v2.0 (Go ramp + release
tooling adds ~1-1.5 weeks vs. the original Python estimate). This repo's
engineered roadmap **ends at 6b** — see "After 6b" below for what follows
without being a numbered phase, and BACKLOG B-021 for why Phase 7 (as
`KNOWLEDGE-FORGE-DESIGN.md:726` and `CLAUDE-CODE-PROMPT.md:563` still name
it) isn't in this table.

## After 6b — run it and measure (not a phase)

Not gated, not sized, not owned by a session the way 0-6b are — this is
ongoing use, not a deliverable. It's what the old Phase 7 row described,
kept here because it's real OSS-lifecycle work, just not a phase:

- Run it daily for ~1 month.
- `docs/RESULTS.md`: hit-rate trend, dedup rate, drift events, time saved —
  `/forge-stats` (Phase 5) is what generates the numbers that go in it.
- Launch post drafted once those numbers exist.
- First LoRA fine-tuning experiment if D1 capture (Phase 6b) reaches ≥300
  pairs.

The same three numbers — 30 days of real usage, a real hit-rate, ≥3 outside
users reporting value — are also the informal signal for when the separate
B2B project (below) might reasonably start. They aren't a gate *inside* this
repo anymore; nothing here blocks on them.

## Sequencing notes

- **Phases 0, 2, and 2b matter most.** 0 gives you the before-numbers, 2 is
  the feature that makes the tool worth using, 2b is the engineering that
  makes it defensible rather than "a prompt with extra steps."
- **Never cut 2b.** If time runs out, cut in this order instead:
  `6b (dataset export can wait, capture stays)` → `5b (log-back)` →
  `advisor tier (3b/4's T3 path)`.
- **Phase 1's vault migration is the one irreversible step.** Dry-run is the
  default everywhere; it refuses to run on a dirty git tree.
- **B2B (`KNOWLEDGE-FORGE-B2B.md`) is a fully separate project, not a phase
  inside this repo.** It is not phase-gated here — see BACKLOG B-021 for the
  decision record. The old readiness condition ("OSS v2.0 shipped, 30 days
  of real usage, ≥3 outside users reporting value") still applies, but
  informally, to that separate project's own start — not as a gate this
  repo's roadmap enforces.
- Each phase has a ready-to-paste Claude Code prompt in
  [`CLAUDE-CODE-PROMPT.md`](./CLAUDE-CODE-PROMPT.md) — that's the actual
  execution mechanism once you're ready to start Phase 0.
