---
title: "{{title}}"                         # 3-120 chars, human-readable; slug is derived from this, not the reverse
slug: "{{slug}}"                           # kebab-case, immutable once assigned; collisions get -2, -3 from `forge slug`
type: decision                             # concept|howto|pattern|pitfall|decision|api|incident
stack: ["{{stack-1}}"]                     # 1-6 values, CLOSED vocabulary — see references/taxonomy.md §2
tags: ["{{tag-1}}"]                        # 1-8 values, OPEN, kebab-case, alias-normalized — see taxonomy.md §3
depth: "{{depth}}"                         # int 1-5, default 3 (1 = skim, 5 = deep-dive)
confidence: "{{confidence}}"               # high|medium|low, default medium; a note failing a §12 gate goes to _inbox/ with low, never silently higher
created: "{{created}}"                     # YYYY-MM-DD
updated: "{{updated}}"                     # YYYY-MM-DD, >= created
verified: "{{verified}}"                   # YYYY-MM-DD, >= created; staleness keys off THIS field, not updated
freshness_days: 730                        # decision default; a decision's rationale ages slowly, even if superseded later
sources: []                                # may be legitimately empty — decision is a first-party record, not a citation of someone else's claim
related: ["[[{{related-slug-1}}]]"]        # wikilinks; quality gate wants >= 2 outbound, schema floor is 0
supersedes: []                             # slugs of notes this one absorbed
forge_version: "{{forge_version}}"         # e.g. 2.0.0
origin: "{{origin}}"                       # ask|session-capture|garden|import
# engine_trail: []                         # rev-2, optional. Per-stage engine tier, e.g. [recall=none, write=none]
# drift_checked_at: "{{git_sha}}"          # rev-2, optional. Git sha the note's code refs were last checked against
---

# {{title}}

> **TL;DR** — {{one sentence a tired engineer can act on}}
<!-- The decision itself, in one line — not the problem, not the reasoning. Do
NOT hedge; a decision note that isn't sure it decided anything shouldn't exist yet. -->

## Context
{{the forces and constraints in play at the time of the decision}}
<!-- What made this a decision rather than an obvious default — constraints,
deadline, prior state. Do NOT include the reasoning for the choice itself here;
that's the Decision section. Write it as it was true then, not with hindsight. -->

## Decision
{{what was chosen, stated as a decision, not as a description of the option}}
<!-- One clear statement of what was decided. Do NOT list the option alongside
its alternatives here — this section is the verdict, not the deliberation. -->

## Consequences
{{what this decision commits you to, including the costs, not only the benefits}}
<!-- Concrete downstream effects, positive and negative, including "what becomes
harder now". Do NOT write only the upside — an ADR that reads as pure marketing
for its own decision is a red flag for the gate. -->

## Alternatives considered
| Option | Why not |
|---|---|
<!-- Every option seriously considered, each with the actual reason it lost — not
a strawman. Do NOT include an alternative nobody actually weighed just to pad the
table. Do NOT re-argue for the chosen option here — that's Decision/Consequences. -->

## Open questions
{{explicit "I could not verify X" — never silently omit}}
<!-- What this decision leaves unresolved, or what would overturn it. Do NOT
leave empty by default. -->

## Sources
{{auto-rendered from frontmatter}}
<!-- Do not hand-edit. Rendered from the `sources` list above by `forge index`;
legitimately empty for most decisions. -->
