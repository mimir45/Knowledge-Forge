---
name: forge-stats
description: Use when the user asks whether Knowledge Forge is paying off — "is the vault helping", "what am I asking about most", "show forge stats", "what gaps are there", "how's the hit rate", "am I saving time with this". Not for a single explanatory question (that's `forge`), not for the weekly health rollup of drift/orphans/duplicates (that's `forge-check`), and not for how much training data has accumulated in the capture tiers (that's `forge-dataset-stats` — this skill reads `.forge/log.jsonl`, that one reads `.forge/datasets/`).
---

# Forge Stats

A thin wrapper around one deterministic command. **This skill writes nothing.**
`forge stats` reads `.forge/log.jsonl` (the telemetry log `forge recall` appends to,
gated on `telemetry.enabled`) and `.forge/weekly-stats.json` (the snapshot history
`forge check` maintains); this file runs it and narrates the table. Unlike the four
hook subcommands, `forge stats` is a direct user command with a normal error-exit
contract — if it fails, say so, don't paper over it.

---

## Step 0 — Check telemetry is actually on

If the hit rate and most-asked sections come back empty on a vault that's clearly been
used for a while, check `forge config --json`'s `telemetry.enabled` before concluding
"nothing's been asked" — telemetry off means `forge recall` never wrote to the log in
the first place, which reads identically to "an unused vault" in the output. Say which
one it actually is rather than letting the ambiguity stand.

## Step 1 — Run it

```bash
forge stats --vault "$VAULT"
```

Zero model calls, no flags beyond `--vault`. If it exits non-zero, show the real error —
this command is not fail-silent like the four hook subcommands, so a nonzero exit here
is a real problem (most likely a bad `--vault` path), not expected behavior.

## Step 2 — Relay the five sections, plainly labeled

- **Hit rate** — the share of asks that resolved `ANSWER_FROM_VAULT` rather than a
  research run. This is the single number that answers "is this paying off".
- **Most-asked topics** (top 15) — each tagged `(written)` or `(gap)`. Lead with any
  `(gap)` entries near the top; a topic asked often with no note is the most actionable
  line in the whole report.
- **Gaps** — topics asked twice or more and never written, same data
  `forge-check`'s weekly rollup surfaces under `## 🎯 Gaps`. If the user is running both
  skills in one session, it's fine to say "same list as the weekly rollup" rather than
  re-narrating it.
- **Research-time-saved estimate** — `written_hits × 15 min`. Always relay this as an
  approximation, using the command's own wording — there is no real measurement behind
  the 15-minute constant anywhere in this project, and presenting it as a hard number
  would misrepresent it.
- **Weekly trend** — one row per week from `.forge/weekly-stats.json`, notes/hit-rate/
  orphans/drift over time. The trend's last column is literally `Drift`, used as a
  stand-in for a dedicated staleness metric that doesn't exist yet — say that plainly if
  asked what it means, don't imply it's a different measurement than it is.

## Step 3 — If asked "should I write about X"

That's exactly what the Gaps section is for — point at it directly rather than
re-running `forge recall` for a topic that's already sitting in the gap list with a
count attached.

---

## Invariants

- This skill never writes a note, a config value, or the telemetry log itself — it only
  reads and narrates what `forge stats` already computed.
- The time-saved figure is always presented as an approximation, never a hard metric.
- An empty report gets diagnosed (telemetry off vs. genuinely unused) before being
  reported as "nothing to show."
- A nonzero exit is shown to the user verbatim, not silently retried or hidden.
