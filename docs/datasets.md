# Datasets

Knowledge Forge produces labelled training data as a byproduct of normal use. This
document is the whole story of what is captured, where it lives, what leaves the
machine, and what the data is honestly good for.

**The short version: nothing is transmitted. Ever.** There is no upload path in this
codebase. Capture writes to files inside your vault, export writes to a directory you
name, and both are local filesystem operations. The data is yours in the ordinary sense
that it is on your disk and no one else has a copy.

---

## The five tiers

| set | what a pair is | produced by |
|---|---|---|
| `d1` | question features → the routing decision | `forge recall` |
| `d2` | draft → advisor critique | `forge engine run` on the advisor tier |
| `d3` | generated note → your edited note | the vault's `post-commit` hook |
| `d4` | failing draft + gate error → fixed draft | `forge gate --previous-draft` |
| `d5` | (topic, profile, sources) → the accepted note | `forge gate`, on a pass |

Each lands in `<vault>/.forge/datasets/<set>.jsonl`, one JSON object per line, appended.
`.forge/` is derived state and gitignored in the vault, so capture files are not
committed by the vault's own history.

D6 in `ADDENDUM §D.1`'s table — code↔knowledge retrieval pairs — is **not** built. It is
derivable from what `forge logback` already generates rather than from a capture path,
and it is filed as BACKLOG **B-034**.

## Consent, and how to turn it off

Two gates, both in the config chain, and every write path checks both:

```yaml
dataset:
  enabled: true                # the master switch
  capture: [d1, d2, d3, d4, d5]  # per-tier; remove a tag to stop that tier
  anonymize_on_export: true
```

`dataset.enabled: false` stops all five. Removing a tag stops that tier alone. Both take
effect immediately — there is no cached copy of the decision.

Telemetry is a **separate** consent. `telemetry.enabled` governs `.forge/log.jsonl`,
DESIGN §14's ask log, and turning it off does not stop `d1` capture, nor the reverse.
They are different questions: one is a local log, the other is a corpus meant to be
exportable.

## What is never captured

- **Raw question text.** `d1` stores a sha256 hash and the extracted topic slug; `d5`
  stores the slug alone. This is `ADDENDUM §D`'s rule and it is enforced at the point of
  capture, not at export.
- **The four free-text profile fields.** `d5` carries `primary_language`, `frameworks`,
  `infra`, `seniority`, `default_depth`, `note_language` and `explain_style` from
  `profiles/me.md`. It does **not** carry `assume_known`, `never_assume`, `code_style`
  or `avoid` — the template invites employer-specific prose in those, and not capturing
  is cheaper than scrubbing.

## Export

```bash
forge export-dataset --set d3 --format jsonl-dpo --since 2026-05-01 --out ./export
```

Manual, one tier at a time, never scheduled and never automatic. Two files land in
`--out`: the corpus and a datasheet describing it.

Anonymization is on by default. It runs the same redaction patterns `forge scrub` uses
on `examples/vault/`, plus one export-only rule for internal URLs (`.internal`, `.local`,
`.corp`, `.lan`, `.intranet`, `.test`, `localhost`, RFC 1918 addresses). Structural
fields are handled separately: note paths become `notes/<type>/<hash>.md`, commit SHAs
are blanked.

It **fails closed**. A capture line that no longer parses is a line nothing could redact,
so the run aborts with `--out` never created. Exit 3 always means "nothing was written".

Every export appends its report — counts and metadata, never record content — to
`<vault>/.forge/exports.jsonl`, before the files are written. An export that happened
without being recorded is not a reachable state.

### The gap the scrubber does not close

**Topic slugs and profile values are kept.** Topics are the only semantic feature `d1`
and `d5` carry; profile values are `d5`'s entire conditioning half. Hashing either makes
those corpora untrainable, so both survive redaction — which means anything spelled as a
plain kebab-case name gets through. A topic named after a product
(`acme-billing-outbox`), a framework named after an in-house SDK
(`frameworks: [acme-internal-sdk]`): neither is token-shaped, address-shaped or
path-shaped, so no pattern catches them.

Every datasheet says this under Limitations. Read the topic and profile fields before
sharing an export.

## What the data is worth

`forge dataset-stats` answers this per tier, against `ADDENDUM §D.2`'s ladder. The
summary, which is deliberately unexciting:

| volume | what is realistically achievable |
|---|---|
| under 100 | an evaluation set, nothing trainable |
| 100–1,000 | a 1–3B routing adapter (`d1`), or a style adapter (`d5`) — and nothing else |
| 1,000–10,000 | a 7–8B drafting LoRA in your stack and voice, for that task alone |
| 10,000+ | DPO on `d2`+`d3`: distilling advisor judgement becomes realistic |

The honest sequencing is eval sets → routing → style → drafting → advisor distillation.

Every corpus here is also **single-author**, which the datasheets state first: what looks
like a learned preference is at least partly one person's habit.
