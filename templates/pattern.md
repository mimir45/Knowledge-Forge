---
title: "{{title}}"                         # 3-120 chars, human-readable; slug is derived from this, not the reverse
slug: "{{slug}}"                           # kebab-case, immutable once assigned; collisions get -2, -3 from `forge slug`
type: pattern                              # concept|howto|pattern|pitfall|decision|api|incident
stack: ["{{stack-1}}"]                     # 1-6 values, CLOSED vocabulary — see references/taxonomy.md §2
tags: ["{{tag-1}}"]                        # 1-8 values, OPEN, kebab-case, alias-normalized — see taxonomy.md §3
depth: "{{depth}}"                         # int 1-5, default 3 (1 = skim, 5 = deep-dive)
confidence: "{{confidence}}"               # high|medium|low, default medium; a note failing a §12 gate goes to _inbox/ with low, never silently higher
created: "{{created}}"                     # YYYY-MM-DD
updated: "{{updated}}"                     # YYYY-MM-DD, >= created
verified: "{{verified}}"                   # YYYY-MM-DD, >= created; staleness keys off THIS field, not updated
freshness_days: 365                        # pattern default; the forces a pattern responds to change slowly
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
<!-- The shape, in one line — "use N instances of X to solve Y". Do NOT hedge. -->

## Problem
{{the recurring forces that make an ad-hoc solution keep failing the same way}}
<!-- The forces in tension, not a single instance of the problem. Do NOT describe
a specific bug you hit — that's a pitfall.md, not this section. -->

## Solution shape
{{the structure — participants, responsibilities, how they interact}}
<!-- The reusable structure, abstracted from any one codebase. Do NOT jump to
code yet — that's the next section. Do NOT restate the problem as the solution. -->

## In {{primary_stack}}
{{concrete code, verified, minimal, idiomatic}}
<!-- The pattern instantiated once, minimally, in this stack. Do NOT include a
second variant "for completeness" — one clean instance is the point. -->

## Trade-offs
| You gain | You give up |
|---|---|
<!-- What applying this pattern costs, in the same table so gain and cost read
side by side. Do NOT list a cost that's actually a Pitfall (a mistake in applying
it) — that belongs in "When NOT to use this" below, this table is about the
inherent cost of the pattern applied correctly. -->

## When NOT to use this
{{the section everyone skips and everyone needs}}
<!-- Concretely, what problem shape makes this pattern overkill or wrong, and what
to reach for instead. Do NOT repeat a Trade-offs row here verbatim. -->

## Open questions
{{explicit "I could not verify X" — never silently omit}}
<!-- Named claims about the pattern's behavior you could not confirm. Do NOT leave
empty by default — if nothing is open, say so explicitly. -->

## Sources
{{auto-rendered from frontmatter}}
<!-- Do not hand-edit. Rendered from the `sources` list above by `forge index`. -->
