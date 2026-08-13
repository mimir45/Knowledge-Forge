---
name: forge-check
description: Use when the user asks to run the weekly vault health check — "run forge check", "what needs attention in the vault", "give me the weekly rollup", "any drift/duplicates/orphans to look at", or on a scheduled/cron-style invocation of the weekly job. Not for a single explanatory question (that's `forge`) and not for onboarding a new vault (that's `forge-init`).
---

# Forge Check

A thin wrapper around one deterministic command. **This skill writes nothing itself** —
`forge check` (the Go binary) renders every report; this file runs it, reads what it
wrote, and narrates. If `check.ai_pass` is on, this file also carries out the printed
proposals as in-session steps — see Step 3 — because `forge check` itself only ever
prints an `Instruction`, never edits a file.

---

## Step 0 — What this actually is

`forge check` is T0-only: zero model calls, safe to run from cron/CI/a scheduled agent
at the packaged default config. It collects the vault **once** and renders nine reports
(`reports/*.md`) plus `cost.md`, a codebase map (`moc/codebase.md`, only with `--repo`),
and the weekly rollup (`moc/weekly/<ISO-week>.md`, always). A second run on the same day
rewrites nothing new — headers carry a date, not a timestamp — so re-running this skill
twice in one sitting is safe and cheap to show the user again.

## Step 1 — Run it

```bash
forge check --vault "$VAULT" [--repo NAME=PATH ...] [--days N] [--offline]
```

- Pass `--repo NAME=PATH` (repeatable) for every code repo the vault cites, if known —
  without at least one, `drift.md` and `moc/codebase.md` are skipped, not written empty,
  and say so plainly rather than treating the skip as an error.
- Pass `--offline` on a bad network or in CI with no egress; `deadlinks.md` then reports
  cached verdicts only, and an unreachable URL is not counted as a dead one.
- `--days` overrides the code-churn window (`check.churn_days` in config, default 90) —
  only pass it if the user asked for a different window than usual.

If it exits non-zero, show the real error. A single renderer failing costs only its own
file — the command's own stdout says which files it wrote and which it skipped; relay
that summary rather than re-deriving it from the file list yourself.

## Step 2 — Show the weekly rollup

`moc/weekly/<ISO-week>.md` is the one file this command always writes (no `--repo`
dependency, unlike drift/codebase). Read it and lead with it:

- `## 🔴 Act now` and `## 🟡 Review` — the ranked items. If either section prints a
  caveat line instead of a list (BACKLOG B-017's churn window, or B-019's duplicate
  threshold), relay that caveat verbatim — it's the honest reason the section reads
  empty, not evidence nothing needs attention.
- `## 📊 Vault` — counts and week-over-week deltas. Omitted on the very first run ever;
  say so rather than treating a missing section as a bug.
- `## 🎯 Gaps` — topics asked twice or more and never written, straight from the
  telemetry log. This is the same data `forge stats` shows in more detail.

Mention the other reports (`orphans.md`, `duplicates.md`, `staleness.md`, etc.) exist
under `reports/` for anyone who wants the un-ranked detail, but don't dump all nine into
the chat — the weekly file is the one built to be read end to end.

## Step 3 — `check.ai_pass`, only if the config has it on

Check `forge config --json` for `check.ai_pass`. If **false** (the default), skip this
step entirely and say nothing about it — most vaults never touch it.

If **true**, `forge check` already ran three sub-tasks and printed one `Instruction`
each to its own stdout (draft refresh for the top BROKEN drift finding, a duplicate-merge
proposal for the top pair clearing the 0.85 spec threshold, an ADR stub for the top
undocumented churny module) — or an explicit no-op line when there was nothing to act on
(no `--repo`, nothing above threshold, etc.). Show each printed proposal to the user and
ask before acting on any of them:

- **Never auto-apply.** These are proposals, not writes — `forge check` itself performs
  no file mutation for `ai_pass`, by design (no diff UI, no auto-apply, matching
  `Host.Run()`'s no-I/O contract).
- If the user approves a proposal, carry it out the same way `forge`'s own CREATE/UPDATE
  branches would (draft → `forge gate` → publish only on `Quarantine: false`) — do not
  invent a shortcut that skips the gate just because the proposal came from `ai_pass`.

## Step 4 — `check.drain_advisor_queue`, only if it fired

If the run's stdout mentions draining the advisor queue (gated on both
`check.drain_advisor_queue` and `engines.budget.on_exhausted: queue`), relay which notes
had their `pending_advisor` flag cleared and which critiques came back with a patch —
each one still needs human approval before the patch is applied; this step only reports
what the advisor tier said, it never writes the patch in.

---

## Invariants

- This skill never renders a report or writes a note itself — `forge check` and (for
  `ai_pass`'s approved proposals) the normal `forge` write pipeline do that.
- The weekly file is always the lead; the nine `reports/*.md` files are detail, not the
  headline.
- An empty "Act now" section gets its caveat relayed verbatim (B-017/B-019), never
  reported as "nothing needs attention" without that context.
- `ai_pass` proposals are shown and approved individually, never auto-applied.
- `--offline` runs are reported as offline, not silently presented as a full network run.
