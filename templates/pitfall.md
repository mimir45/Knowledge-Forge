---
title: "{{title}}"                         # 3-120 chars, human-readable; slug is derived from this, not the reverse
slug: "{{slug}}"                           # kebab-case, immutable once assigned; collisions get -2, -3 from `forge slug`
type: pitfall                              # concept|howto|pattern|pitfall|decision|api|incident
stack: ["{{stack-1}}"]                     # 1-6 values, CLOSED vocabulary — see references/taxonomy.md §2
tags: ["{{tag-1}}"]                        # 1-8 values, OPEN, kebab-case, alias-normalized — see taxonomy.md §3
depth: "{{depth}}"                         # int 1-5, default 3 (1 = skim, 5 = deep-dive)
confidence: "{{confidence}}"               # high|medium|low, default medium; a note failing a §12 gate goes to _inbox/ with low, never silently higher
created: "{{created}}"                     # YYYY-MM-DD
updated: "{{updated}}"                     # YYYY-MM-DD, >= created
verified: "{{verified}}"                   # YYYY-MM-DD, >= created; staleness keys off THIS field, not updated
freshness_days: 365                        # pitfall default; the underlying cause changes on the ecosystem's own timeline
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
<!-- Lead with the symptom in plain language, not the cause — that's what gets
matched against someone's search. Do NOT hedge. -->

## Symptom
{{what you actually observed — error text, behavior, log line}}
<!-- Exact, quotable symptom: an error message, a repro, an observed behavior. Do
NOT summarize it into a category ("performance issue") — the raw symptom is what
makes this note findable. -->

## Root cause
{{the actual mechanism that produces the symptom}}
<!-- The real cause, traced to a mechanism, not "misconfiguration" as a vague
label. Do NOT restate the symptom in different words instead of explaining it. -->

## How to tell it's this one
{{what distinguishes this from superficially similar problems}}
<!-- The differential — what else produces a similar symptom, and how this one is
told apart. Do NOT skip this just because you're confident; that confidence is
exactly what this section exists to justify to the next reader. -->

## Fix
{{the concrete change that resolves it}}
<!-- The actual fix you applied or verified, minimal and idiomatic. Do NOT offer
three alternative fixes "just in case" — pick the one you verified and say why if
a second exists. -->

## Prevention
- [ ] {{actionable, imperative, testable}}
<!-- Checklist to stop this from recurring — lint rule, test, config default. Do
NOT restate the Fix as a checklist item; prevention is upstream of needing the fix. -->

## Open questions
{{explicit "I could not verify X" — never silently omit}}
<!-- Claims about the root cause you could not fully confirm (e.g. reproduced once,
not across versions). Do NOT leave empty by default. -->

## Sources
{{auto-rendered from frontmatter}}
<!-- Do not hand-edit. Rendered from the `sources` list above by `forge index`. -->
