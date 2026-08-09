---
title: "{{title}}"                         # 3-120 chars, human-readable; slug is derived from this, not the reverse
slug: "{{slug}}"                           # kebab-case, immutable once assigned; collisions get -2, -3 from `forge slug`
type: howto                                # concept|howto|pattern|pitfall|decision|api|incident
stack: ["{{stack-1}}"]                     # 1-6 values, CLOSED vocabulary — see references/taxonomy.md §2
tags: ["{{tag-1}}"]                        # 1-8 values, OPEN, kebab-case, alias-normalized — see taxonomy.md §3
depth: "{{depth}}"                         # int 1-5, default 3 (1 = skim, 5 = deep-dive)
confidence: "{{confidence}}"               # high|medium|low, default medium; a note failing a §12 gate goes to _inbox/ with low, never silently higher
created: "{{created}}"                     # YYYY-MM-DD
updated: "{{updated}}"                     # YYYY-MM-DD, >= created
verified: "{{verified}}"                   # YYYY-MM-DD, >= created; staleness keys off THIS field, not updated
freshness_days: 180                        # howto default; instructions rot faster than concepts, flagged stale sooner
sources:
  - url: "{{source-1-url}}"                # absolute http(s) URL, or a vault-relative path for a first-party source
    accessed: "{{source-1-accessed}}"      # YYYY-MM-DD
    kind: "{{source-1-kind}}"              # official|spec|blog|paper|video|forum|code|session
related: ["[[{{related-slug-1}}]]"]        # wikilinks; quality gate wants >= 2 outbound, schema floor is 0
supersedes: []                             # slugs of notes this one absorbed
forge_version: "{{forge_version}}"         # e.g. 2.0.0
origin: "{{origin}}"                       # ask|session-capture|garden|import
# engine_trail: []                         # rev-2, optional. Per-stage engine tier, e.g. [recall=none, write=none]
# drift_checked_at: "{{git_sha}}"          # rev-2, optional. Git sha the note's code refs were last checked against
---

# {{title}}

> **TL;DR** — {{one sentence a tired engineer can act on}}
<!-- The outcome, not the topic — "do X to get Y", not "about X". Do NOT hedge. -->

## Goal
{{what you end up with when this is done, and how you'll know}}
<!-- The concrete end state — a working thing, a passing check — not a restated
title. Do NOT describe motivation at length here; one clause on why is enough,
this isn't the Mental model section from concept.md. -->

## Prerequisites
{{versions, access, prior setup this assumes without re-explaining}}
<!-- Only things that would silently break the steps below if missing. Do NOT
re-teach a prerequisite concept inline — link to a concept.md note instead of
inlining it. -->

## Steps
1. {{imperative, one action per step, in the order you actually ran them}}
<!-- Numbered, imperative, each step independently checkable. Do NOT bundle two
actions into one numbered step. Do NOT include steps you assume work but never
ran — Verified means run. -->

## Verify it worked
{{the exact command or observation that confirms success}}
<!-- A concrete check, not "you should see it work now". Do NOT restate the last
step of "Steps" — this is the independent proof the steps above achieved the Goal. -->

## Pitfalls
| Pitfall | Why it happens | Fix |
|---|---|---|
<!-- One row per pitfall you have actually hit while doing this procedure. Do NOT
list theoretical pitfalls, do NOT pad to fill the table, do NOT re-describe a step
that's already correct in "Steps" above. -->

## Open questions
{{explicit "I could not verify X" — never silently omit}}
<!-- Steps or claims you could not fully confirm (e.g. only tested on one OS/version).
Do NOT leave empty just because it ran cleanly once — state that explicitly if true. -->

## Sources
{{auto-rendered from frontmatter}}
<!-- Do not hand-edit. Rendered from the `sources` list above by `forge index`. -->
