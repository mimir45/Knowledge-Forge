# Release readiness — Knowledge Forge

Written 2026-08-27, ahead of the first tagged release. This is a snapshot, not a living
spec: for day-to-day work, `CLAUDE.md`'s "Read the docs in this order" list and
`docs/BACKLOG.md` remain authoritative. This file exists to answer one question without
reading either in full: **what's left before release, and what does the system look
like today.**

## Status

The roadmap (`0 → 1 → 2 → 2b → 3 → 3b → 4 → 5 → 5b → 6 → 6b`) is complete. Every phase
merged to `main`; `go build ./...` and `go test ./...` are green under both
`CGO_ENABLED=0` and `CGO_ENABLED=1`. What remains is out-of-phase backlog work (below)
and the two unverified release mechanics also listed below — nothing structural.

## Architecture, in one screen

A Go static-analysis engine (`pkg/vault`, `pkg/recall`, `pkg/similarity`, `pkg/graph`,
`pkg/codeindex`, `pkg/coderef`, `pkg/gitsig`, `pkg/drift`, `pkg/linkcheck`, `pkg/report`,
`pkg/store`) makes **zero model calls** and does the deterministic work: scoring a
question against an Obsidian vault, detecting drift between notes and the code they
cite, rendering reports. An optional four-tier engine layer (`pkg/engine`: none / host /
API / advisor-critique) sits on top, configured per pipeline stage via a four-layer
config chain (`pkg/config`). `pkg/qualitygate` runs seven gates before a note is
accepted; failures go to `_inbox/` with `confidence: low`, never a silent publish.
`pkg/dataset` exports six datasets (five captured, one — D6 — derived from
`forge logback`'s code↔knowledge map) for fine-tuning, anonymized by default via
`pkg/scrub`, which fails closed. Everything ships as a Claude Code plugin
(`.claude-plugin/`) wrapping the `forge` CLI (`cmd/forge`) plus four Claude Code
lifecycle hooks and three git-anchored drift hooks.

Full detail: `docs/ROADMAP.md` (index), `docs/KNOWLEDGE-FORGE-DESIGN.md` (master spec),
`docs/KNOWLEDGE-FORGE-ADDENDUM.md` (engine tiers, datasets, config).

## Not yet verified — real release blockers

These are the "done when" conditions later phases named but couldn't check on this
machine, per `CLAUDE.md`'s Phase 6/6b status notes:

- The shim's real download-and-checksum path, and `claude plugin marketplace add
  mimir45/Knowledge-Forge` from a genuinely clean machine — both need the tagged
  release that hasn't been cut yet.
- The TypeScript verify-code lane (`tsc`) — untested here because `tsc` isn't installed
  in this environment; `TestCompileTSSkippedWhenToolchainAbsent` covers the
  absent-toolchain path instead, not the real one.

## Open backlog (from `docs/BACKLOG.md`'s index — 15 items)

Not release-blocking on their own; recorded so nothing gets lost. `docs/TODO.md` has a
six-field execution plan for each workable one.

| Item | One line |
|---|---|
| B-002 | Fixture vault exists at `testdata/vault/`; it is not `examples/vault/` |
| B-003 | Repo directory is still named `TIL` — **user confirmed keep-as-is, 2026-08-24** |
| B-004 | Module path has no VCS host prefix |
| B-010 | `AUDIT.md` §7 says `food-ordering-system` has no git history; it does |
| B-011 | `reports/` and `moc/` are graph nodes but not contract notes |
| B-012 | `code_refs` is in the schema and nothing writes it yet |
| B-014 | The code index parses TypeScript, not Kotlin |
| B-016 | The vault carries both `sources:` and `source:` |
| B-017 | §B.5's 90-day churn window shows nothing on these repos |
| B-018 | A bare symbol citation is credited to one arbitrary declaration |
| B-019 | Duplicate detection ships three deviations from DESIGN §8 |
| B-020 | `sort.Slice` comparators need a tiebreak unique in their collection |
| B-025 | `forge cache-source`'s WebFetch `tool_response` JSON shape is unconfirmed — **blocked on observing a live payload, do not re-attempt the WebFetch** |
| B-037 | `forge intent`'s FIRE/QUIET margin went negative (-0.036) under B-032's rescaled floor — **do not respond by moving the gate**; nothing is failing today |
| B-038 | `bodyPass`'s top-20 candidate window is allocated by path, not relevance |

Not on this list: **B-021** (B2B is a separate project, not a defect — see below) and
**B-039** (retracted outright during B-036's measurement pass, not a real finding).

## Pending PR

**#16 `chore/drop-windows-release-target`** — open, 1 commit ahead of `main`, drops
Windows from the build/release surface. Not merged by this pass: merging into `main` is
outside what this session does — it's yours to review and merge.

## Housekeeping done this pass

- `.gitignore` now excludes `/.claude/worktrees/` (per-session isolation, never release
  surface). `.claude/agents/*.md` (six files) stays tracked — that's the documented
  "Agent crew," not clutter; see `CLAUDE.md`'s "Agent crew" section.
- Confirmed nothing is uncommitted or unmerged anywhere this session could reach: every
  worktree branch's work is already in `main` via a merged PR (worktree-b-008 → PR #1,
  phase-6b family → PRs #2–4, fix/b032 → #5, docs/todo-simplify-index → #6,
  fix/b-023 → #7, feat/b-015 → #8, docs/catchup → #9, feat/b-035 → #10,
  feat/b-034 → #11, dev → #12–13, worktree-b-036 → #14, this worktree's B-037 work →
  #15). PR #16 is the only exception (see above).

## Housekeeping *not* done — needs your call

- **`docs/KNOWLEDGE-FORGE-B2B.md` and `docs/CLAUDE-CODE-PROMPT.md`** are the two stale
  docs worth removing before release: B2B describes a fully separate project (already
  recorded as B-021, kept only for reference), and the execution prompt is
  phase-by-phase instructions for a roadmap that's now complete. Deleting them touches
  ~9 other files that cite them by name (`CLAUDE.md`, `docs/ROADMAP.md`,
  `docs/KNOWLEDGE-FORGE-DESIGN.md`, `docs/KNOWLEDGE-FORGE-STACK.md`,
  `docs/KNOWLEDGE-FORGE-ADDENDUM.md`, `references/schema.yaml`,
  `references/taxonomy.md`, `.claude/agents/doc-auditor.md`, `docs/tr/*`) plus adding an
  `AUDIT.md` §8.4 entry recording the removal, the same mechanism B-027 used for a stale
  filename. The `git rm` for this was blocked by this session's permission classifier —
  flagging it back to you rather than working around the denial.
- **Two worktrees look safe to remove** — their branches are fully merged (PR #1 and
  #14) and this session's sandbox can't reach outside its own worktree to remove them.
  From the main checkout (`/Users/mimir45/knowledge-forge` — the directory was
  `/Users/mimir45/TIL` when this snapshot was written; renamed 2026-09-01, BACKLOG B-003):
  ```
  git worktree remove .claude/worktrees/b-008-recall-recalibration
  git branch -d worktree-b-008-recall-recalibration
  git worktree remove .claude/worktrees/b-036-neighbour-window
  git branch -d worktree-b-036-neighbour-window
  ```
  (This session's own worktree, `b-037-intent-gate-plan`, is left for you or the harness
  to clean up on session end — same reason.)
- **`docs/BACKLOG.md` and `examples/vault/` were deliberately left untouched.** BACKLOG
  is a historical record under this repo's own "record, don't fix" rule — its entries
  describe what was true when written, including mentions of the two removed docs.
  `examples/vault/` is a frozen, already-shipped, `forge scrub`-generated deliverable;
  editing it isn't a doc-sync fix, it's re-shipping the artifact.
