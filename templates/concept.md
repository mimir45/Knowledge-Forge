---
title: "{{title}}"                         # 3-120 chars, human-readable; slug is derived from this, not the reverse
slug: "{{slug}}"                           # kebab-case, immutable once assigned; collisions get -2, -3 from `forge slug`
type: concept                              # concept|howto|pattern|pitfall|decision|api|incident
stack: ["{{stack-1}}"]                     # 1-6 values, CLOSED vocabulary — see references/taxonomy.md §2
tags: ["{{tag-1}}"]                        # 1-8 values, OPEN, kebab-case, alias-normalized — see taxonomy.md §3
depth: "{{depth}}"                         # int 1-5, default 3 (1 = skim, 5 = deep-dive)
confidence: "{{confidence}}"               # high|medium|low, default medium; a note failing a §12 gate goes to _inbox/ with low, never silently higher
created: "{{created}}"                     # YYYY-MM-DD
updated: "{{updated}}"                     # YYYY-MM-DD, >= created
verified: "{{verified}}"                   # YYYY-MM-DD, >= created; staleness keys off THIS field, not updated
freshness_days: 365                        # concept default; flagged stale past this many days since verified
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
<!-- One sentence, imperative or declarative, that stands alone if every other section
were deleted. Do NOT restate the title. Do NOT hedge ("it depends") — if it depends,
say on what, in one clause. -->

## Mental model
{{the analogy or diagram that makes it click — max 1 diagram}}
<!-- The single mental picture that makes the mechanism predictable without reading
the next section. One analogy or one diagram, not both. Do NOT put step-by-step
mechanism here — that belongs in "How it actually works". Do NOT pad with a second
analogy "just in case"; pick the one that is actually load-bearing. -->

## How it actually works
{{mechanism, in order of execution}}
<!-- The real mechanism, in the order it executes, with enough precision that a
reader could predict behavior on an input you didn't show. Do NOT describe the API
surface in the abstract — that's what "In {{primary_stack}}" is for. Do NOT restate
the mental model in prose. -->

## In {{primary_stack}}
{{concrete code, verified, minimal, idiomatic}}
<!-- Code that actually runs, in the stack this note is filed under, trimmed to the
minimum that demonstrates the mechanism above. Verified means you ran it, not that it
looks right. Do NOT include unrelated boilerplate (imports/config) unless the pitfall
is in the boilerplate itself. Do NOT turn this into a full tutorial — that's howto.md's job. -->

## Best practices
- [ ] {{actionable, imperative, testable}}
<!-- Checklist of things to DO, each one independently actionable and checkable.
Do NOT include a practice that is just "avoid the pitfall below" restated as a
positive — that's redundant with the Pitfalls table, not a second data point. Do NOT
include generic advice ("write tests") that isn't specific to this concept. -->

## Pitfalls
| Pitfall | Why it happens | Fix |
|---|---|---|
<!-- One row per pitfall you have actually hit or seen reported in a cited source.
Do NOT list theoretical pitfalls, do NOT pad to fill the table, and do NOT restate
the Best practices section inverted. -->

## When NOT to use this
{{the section everyone skips and everyone needs}}
<!-- The honest cases where this concept is the wrong tool, with what to reach for
instead. Do NOT write "it's always fine" — if that were true this section would be
empty, which is itself suspicious for a note worth writing. Do NOT reuse a Pitfalls
row here; a pitfall is a way this fails when used correctly, this section is about
whether to use it at all. -->

## Open questions
{{explicit "I could not verify X" — never silently omit}}
<!-- Named claims you could not confirm, with what would confirm them. Do NOT leave
this empty just because everything felt solid — if genuinely nothing is open, say so
explicitly rather than deleting the section. Do NOT use this as a TODO list for the
note itself (missing sections); it's about unresolved claims in the subject matter. -->

## Sources
{{auto-rendered from frontmatter}}
<!-- Do not hand-edit. Rendered from the `sources` list above by `forge index`. -->
