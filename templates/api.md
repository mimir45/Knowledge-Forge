---
title: "{{title}}"                         # 3-120 chars, human-readable; slug is derived from this, not the reverse
slug: "{{slug}}"                           # kebab-case, immutable once assigned; collisions get -2, -3 from `forge slug`
type: api                                  # concept|howto|pattern|pitfall|decision|api|incident
stack: ["{{stack-1}}"]                     # 1-6 values, CLOSED vocabulary — see references/taxonomy.md §2
tags: ["{{tag-1}}"]                        # 1-8 values, OPEN, kebab-case, alias-normalized — see taxonomy.md §3
depth: "{{depth}}"                         # int 1-5, default 3 (1 = skim, 5 = deep-dive)
confidence: "{{confidence}}"               # high|medium|low, default medium; a note failing a §12 gate goes to _inbox/ with low, never silently higher
created: "{{created}}"                     # YYYY-MM-DD
updated: "{{updated}}"                     # YYYY-MM-DD, >= created
verified: "{{verified}}"                   # YYYY-MM-DD, >= created; staleness keys off THIS field, not updated
freshness_days: 180                        # api default; signatures and constraints drift with the library's own release cadence
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
<!-- What this surface is for, in one line. Do NOT hedge. -->

## Surface
| Method / Endpoint | Signature | Description |
|---|---|---|
<!-- One row per method/endpoint actually documented below, verified against the
real signature. Do NOT list the entire library's API "for completeness" — only
what this note is actually about. Do NOT turn a row into a tutorial; that's the
"In {{primary_stack}}" section's job. -->

## Auth & constraints
{{auth model, rate limits, required scopes, versioning behavior}}
<!-- Operational constraints a caller must know before using the surface —
authentication, limits, breaking-change policy. Do NOT restate the Surface table
in prose here. -->

## In {{primary_stack}}
```{{lang}}
{{concrete code, verified, minimal, idiomatic}}
```
<!-- One minimal, verified call against the surface above. Do NOT build this into
a multi-step walkthrough — a full workflow belongs in howto.md, this is "how you
call it", not "how you accomplish a task with it". Fenced block is required —
antislop's structural rule fails any api note with zero code. -->

## Pitfalls
| Pitfall | Why it happens | Fix |
|---|---|---|
<!-- One row per pitfall actually hit or reported in a cited source, specific to
this surface (not the underlying protocol/library in general). Do NOT pad, do NOT
restate Auth & constraints as a pitfall. -->

## Open questions
{{explicit "I could not verify X" — never silently omit}}
<!-- Behavior you could not confirm against the real service/library (e.g.
undocumented rate-limit values). Do NOT leave empty by default. -->

## Sources
{{auto-rendered from frontmatter}}
<!-- Do not hand-edit. Rendered from the `sources` list above by `forge index`. -->
