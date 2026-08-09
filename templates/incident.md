---
title: "{{title}}"                         # 3-120 chars, human-readable; slug is derived from this, not the reverse
slug: "{{slug}}"                           # kebab-case, immutable once assigned; collisions get -2, -3 from `forge slug`
type: incident                             # concept|howto|pattern|pitfall|decision|api|incident
stack: ["{{stack-1}}"]                     # 1-6 values, CLOSED vocabulary — see references/taxonomy.md §2
tags: ["{{tag-1}}"]                        # 1-8 values, OPEN, kebab-case, alias-normalized — see taxonomy.md §3
depth: "{{depth}}"                         # int 1-5, default 3 (1 = skim, 5 = deep-dive)
confidence: "{{confidence}}"               # high|medium|low, default medium; a note failing a §12 gate goes to _inbox/ with low, never silently higher
created: "{{created}}"                     # YYYY-MM-DD
updated: "{{updated}}"                     # YYYY-MM-DD, >= created
verified: "{{verified}}"                   # YYYY-MM-DD, >= created; staleness keys off THIS field, not updated
freshness_days: 730                        # incident default; the record of what happened doesn't age even if the system did
sources: []                                # may be legitimately empty — incident is a first-party record, not a citation of someone else's claim
related: ["[[{{related-slug-1}}]]"]        # wikilinks; quality gate wants >= 2 outbound, schema floor is 0
supersedes: []                             # slugs of notes this one absorbed
forge_version: "{{forge_version}}"         # e.g. 2.0.0
origin: "{{origin}}"                       # ask|session-capture|garden|import
# engine_trail: []                         # rev-2, optional. Per-stage engine tier, e.g. [recall=none, write=none]
# drift_checked_at: "{{git_sha}}"          # rev-2, optional. Git sha the note's code refs were last checked against
---

# {{title}}

> **TL;DR** — {{one sentence a tired engineer can act on}}
<!-- What broke and what the fix was, in one line. Do NOT hedge. -->

## Timeline
| Time | Event |
|---|---|
<!-- Chronological, factual events only — what was observed and when, no
interpretation. Do NOT put the root cause analysis in this table; that's the next
section. Do NOT compress the timeline down to "it broke, we fixed it" — the
sequence is the diagnostic value. -->

## Root cause
{{the actual mechanism that produced the incident, not just the trigger}}
<!-- The real cause, traced past the immediate trigger to the mechanism. Do NOT
stop at "deploy X caused it" if X only exposed a pre-existing bug — name the bug. -->

## What changed
{{the concrete remediation and the prevention put in place afterward}}
<!-- What was actually changed — code, config, process — to fix it and stop it
recurring. Do NOT list a prevention idea that was discussed but not implemented;
if it's still open, that belongs in Open questions, not here. -->

## Impact
{{who/what was affected, for how long, how it was noticed}}
<!-- Scope and duration, and the detection path (alert, user report, etc). Do NOT
speculate on impact you didn't measure — state what's known and flag the rest as
an open question. -->

## Open questions
{{explicit "I could not verify X" — never silently omit}}
<!-- Anything about scope, cause, or fix effectiveness that remains unconfirmed.
Do NOT leave empty by default. -->

## Sources
{{auto-rendered from frontmatter}}
<!-- Do not hand-edit. Rendered from the `sources` list above by `forge index`;
legitimately empty for most incidents. -->
