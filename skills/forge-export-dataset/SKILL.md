---
name: forge-export-dataset
description: Use when the user wants to get their captured training data out — "export the dataset", "give me the d3 pairs", "export for fine-tuning", "dump the routing data as CSV", "anonymize and export". Not for asking how much data there is (that's `forge-dataset-stats`, which writes nothing) and not for scrubbing a whole vault into a shareable copy (that's `forge scrub`).
---

# Forge Export Dataset

A thin wrapper around one deterministic command. **This skill writes only to `--out`.**
`forge export-dataset` reads one dataset, filters it, redacts it, and writes a corpus
plus a datasheet. Zero model calls. Nothing transmits anywhere — the export lands on
disk and moving it is a decision the user makes by hand.

---

## Step 0 — Settle three things before running anything

**Which set.** `d1` routing, `d2` advisor critiques, `d3` human corrections, `d4` gate
repairs, `d5` style, `d6` code↔knowledge (derived, not captured). If the user says "the
dataset", ask — the six have different shapes and there is no combined export.

`d6` is different in kind, not just shape: it is **derived**, not captured. `d1`-`d5`
each accumulate over time via a write path; `d6` has none and is recomputed fresh on
every export, from the vault's code citations resolved against whatever
`.forge/code-index-<repo>.json` caches this machine currently holds. If those caches are
stale or absent, tell the user to run `forge logback` first rather than treating an
empty `d6` export as "nothing captured yet".

**Which format.** Not every combination exists, and the command refuses the ones that do
not rather than inventing an output shape:

| set | valid formats |
|---|---|
| `d1` | `jsonl-sft`, `csv` |
| `d2` | `jsonl-sft` |
| `d3` | `jsonl-sft`, `jsonl-dpo` |
| `d4` | `jsonl-sft`, `jsonl-dpo` |
| `d5` | `jsonl-sft` |
| `d6` | `jsonl-sft` |

`jsonl-dpo` needs a chosen/rejected pair actually present in the data; only `d3` (the
human's edit over the generated note) and `d4` (the fixed draft over the failing one)
have one. `csv` is `d1`'s alone.

**Whether to anonymize.** It is on by default, from `dataset.anonymize_on_export`, for
`d1`-`d5`. Do not pass `--no-anonymize` unless the user asked for raw output in those
terms, and when they do, relay the warning the command prints instead of summarizing it
away.

`d6` is the one exception: `--anonymize` is **always refused** for it — repo, path and
symbol names are the entire feature, and no redaction of them has been found that keeps
the export useful. Exporting `d6` at all means passing `--no-anonymize`; do not treat
that as "the user asked for raw output", it is the only way this set exports.

## Step 1 — Run it

```bash
forge export-dataset --vault "$VAULT" --set d3 --format jsonl-dpo \
  --since 2026-05-01 --out ./export
```

`--since` takes `YYYY-MM-DD` and filters on the record's own timestamp — capture time
for four tiers, the *edit* time for `d3`, since that is when the correction happened.
`d6` has no per-record timestamp at all and refuses `--since` outright rather than
silently filtering nothing.

## Step 2 — Read the exit code before reading the output

- **0** — the export and its datasheet are in `--out`.
- **2** — a usage error: an unknown `--set`, an undefined `(set, format)` pair, a
  malformed `--since`, both anonymize flags at once, or (`d6` only) `--since` or
  `--anonymize` at all. Nothing was attempted.
- **3** — the export failed and `--out` was not created. For `d1`-`d5` the usual cause is
  a line in the capture file that no longer parses; the error names the file and the
  line number. For `d6` the usual cause is a `.forge/code-index-<repo>.json` cache that
  cannot be read (built by an older extractor, or corrupt) — the error names that file
  instead, and the fix is `forge logback` against that repo, not a hand edit. Either way,
  relay it verbatim rather than retrying.

Exit 3 is fail-closed working as designed: a line nobody can parse is a line nobody
could redact, so no partial export is produced.

## Step 3 — Relay the report and point at the datasheet

The JSON report on stdout carries `records`, `available`, the date range, the stack
distribution, and the redaction count. Two lines are worth calling out directly:

- **`anonymized: false`** — say so plainly. The file contains captured text exactly as
  recorded. For `d6` this is *every* successful export, not a flag the user chose.
- **`records: 0`** — for `d1`-`d5`, the tier has captured nothing; diagnose with
  `/forge-dataset-stats`. For `d6`, it means no citation resolved against a cached code
  index on this machine — point at `forge logback`, not at capture volume.

Then point the user at the datasheet written beside the export. Its **Limitations**
section is the part that matters before the data is used or shared, and it names things
the report's numbers cannot: that topic slugs and profile values survive redaction, that
`d1` has no outcome label, that `d5` only sees notes that went through `forge gate`, that
`d6`'s record count is citations, not notes.

## Step 4 — Before the user shares an export

Two checks, in this order. Confirm the run was anonymized (for `d6`, it never is — say
that instead). Then tell them to read the topic and profile fields: topic slugs and
`d5`'s profile values are deliberately **not** hashed, because they are the semantic and
conditioning features those tiers carry. Anything spelled as a plain kebab-case name gets
through — a topic named after a product, a framework named after an in-house SDK. That is
the gap the scrubber does not close, and no flag closes it. For `d6`, the repo, path and
symbol strings are never redacted at all — read the repo names specifically before
sharing.

---

## Invariants

- Never pass `--no-anonymize` on the user's behalf, and never to work around an error —
  except for `d6`, whose only export path *is* `--no-anonymize`; still relay the raw-
  export warning rather than summarizing it away.
- Never write outside `--out`. This skill does not touch the vault's notes.
- Never re-run a failed export with different flags to make it succeed — exit 3 means
  data needs fixing, not that the invocation was wrong. For `d6`, "fixing" means running
  `forge logback`, not editing anything by hand.
- The datasheet is always mentioned, never skipped as boilerplate.
- Every export is recorded in `.forge/exports.jsonl`; do not edit or remove that log.
