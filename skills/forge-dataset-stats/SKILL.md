---
name: forge-dataset-stats
description: Use when the user asks what training data the vault has accumulated — "how much training data do I have", "can I fine-tune on this yet", "show dataset stats", "what's in d1/d3", "is there enough for a LoRA". Not for whether the vault is paying off day to day (that's `forge-stats`, which reads the telemetry log, not the capture tiers) and not for the weekly health rollup (that's `forge-check`).
---

# Forge Dataset Stats

A thin wrapper around one deterministic command. **This skill writes nothing.**
`forge dataset-stats` reads the five capture files under `.forge/datasets/`, plus `d6`
(a derived set, recomputed live rather than read from a file), and reports how much each
holds and what that volume is honestly enough for. It is a direct user command with a
normal error-exit contract — if it fails, show the error.

---

## Step 0 — Know which command was asked for

`/forge-stats` and `/forge-dataset-stats` read different files and answer different
questions. Telemetry (`.forge/log.jsonl`, hit rate, gaps) is the first; capture tiers
(`.forge/datasets/*.jsonl`, training volume) is this one. If the user's phrasing could
mean either — "show me the forge stats" — ask which, rather than running one and
presenting it as the answer to both.

## Step 1 — Run it

```bash
forge dataset-stats --vault "$VAULT"
```

Zero model calls, no flags beyond `--vault`. A tier with no file yet is a zero row, not
an error.

## Step 2 — Relay both sections

- **Pairs** — one row per tier: count, and the first and last capture date. A tier
  showing `unreadable` has a torn line in its JSONL (or, for `d6`, a
  `.forge/code-index-<repo>.json` cache built by an older extractor); the message names
  the file, so relay it verbatim, it is directly fixable. `d6`'s first/last columns are
  always `—` — it has no per-record timestamp, not an empty history — and its count is
  recomputed on this run, not a running total: it can go up *or down* between runs as
  citations and code indexes change, unlike `d1`-`d5`.
- **What this is enough for** — the volume ladder, read back per tier. Relay these
  strings as written. They are deliberately bounded above, and softening "and nothing
  else" into "a good start" undoes the only thing the section is for. `d6`'s line says
  plainly that it is not a training-adapter shape at all — its stated use is
  retrieval/RAG eval — so do not read a `d6` count against the same 100/1000 bands the
  other rows use.

## Step 3 — If a tier is at zero on a vault that has been used

For `d1`-`d5`: check whether the tier is actually gated on. `forge config
--json`'s `dataset.enabled` is the master switch and `dataset.capture` is the per-tier
list; a tag missing from that list means the write path was never reached, which looks
identical to "nothing happened yet". Say which one it is. This matters most for `d3`,
whose capture also depends on the vault's `post-commit` hook being installed at all.

For `d6`: there is no gate to check — it has no `dataset.capture` entry by design (it
never captures). A zero row means no citation in the vault resolves against a
`.forge/code-index-<repo>.json` cache on this machine; point at `forge logback`.

## Step 4 — If asked "so can I fine-tune"

Answer with the tier's own line and the sequencing, not with encouragement. The honest
order is eval sets → routing (d1) → style (d5) → drafting (d3+d4) → advisor
distillation (d2+d3), and the counts decide where the vault actually sits. `d6` sits
outside this ladder entirely — it is a retrieval/RAG eval set, not a fine-tuning input,
so its count never answers "can I fine-tune". A user who starts a fine-tune on 200 pairs
and concludes fine-tuning does not work has been failed by this skill.

---

## Invariants

- This skill never writes anything, never exports, and never touches
  `.forge/datasets/`. Exporting is `/forge-export-dataset`.
- The adequacy strings are relayed as-is, never upgraded.
- A zero row is diagnosed (tier gated off vs. genuinely empty) before being reported.
- A nonzero exit is shown verbatim.
