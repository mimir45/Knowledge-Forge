# Knowledge Forge — Design & Execution Plan

**Working name:** `knowledge-forge` (skill invocation: `forge`)
*Alternatives if the name is taken: `lorekeep`, `vaultsmith`, `codex-notes`.*

**What it is:** a Claude Code skill that turns "explain X" moments into permanent,
linked, verified markdown notes in an Obsidian vault — so the second time you or
your agent needs that knowledge, it's a file read instead of a research run.

**Status of this document:** design + execution plan, **revision 2**. It is written
to be handed to Claude Code as a spec. Companion documents:

| File | Contents |
|---|---|
| `KNOWLEDGE-FORGE-ADDENDUM.md` | engine tiers (T0–T3), the no-AI static core, drift detection, weekly checker, fine-tuning datasets |
| `KNOWLEDGE-FORGE-B2B.md` | "no context loss" — MCP ingestion (Slack/Teams/Atlassian/mail), hybrid retrieval, ACL model |
| `KNOWLEDGE-FORGE-STACK.md` | ADR-001: the deterministic core is **Go** (single static binary in `bin/`), not Python; ADR-002: B2B backend leans Go, decided at B1 |
| `CLAUDE-CODE-PROMPT.md` | the consolidated phase-by-phase prompts to paste into Claude Code |

> **Rev 2 note:** where this document names `scripts/*.py`, read them as subcommands
> of a single Go binary (`forge recall`, `forge index`, `forge validate`, …) per
> ADR-001 in the STACK doc. The pipeline, schemas, and decision logic are unchanged.
> The one-time vault migration script is the only piece that may stay Python — it
> runs once and never ships.

---

## Table of contents

1. [Problem & thesis](#1-problem--thesis)
2. [Current state (v1) and its failure modes](#2-current-state-v1-and-its-failure-modes)
3. [Goals / non-goals](#3-goals--non-goals)
4. [Design principles](#4-design-principles)
5. [Architecture](#5-architecture)
6. [The note contract](#6-the-note-contract)
7. [Vault topology & the memory graph](#7-vault-topology--the-memory-graph)
8. [Retrieval-before-research (the dedup engine)](#8-retrieval-before-research-the-dedup-engine)
9. [Personalization layer](#9-personalization-layer)
10. [Customization layer](#10-customization-layer)
11. [Subagent topology](#11-subagent-topology)
12. [Quality gates & verification](#12-quality-gates--verification)
13. [Hooks & automation](#13-hooks--automation)
14. [Telemetry (local-first, opt-in)](#14-telemetry-local-first-opt-in)
15. [Phased roadmap](#15-phased-roadmap)
16. [OSS release plan](#16-oss-release-plan)
17. [CV / portfolio framing](#17-cv--portfolio-framing)
18. [Appendix A — B2B extension sketch](#appendix-a--b2b-extension-sketch)
19. [Appendix B — risks & mitigations](#appendix-b--risks--mitigations)
20. [Appendix C — file listings](#appendix-c--file-listings)

---

## 1. Problem & thesis

The Karpathy framing that your setup already follows: *Obsidian is the IDE, the LLM
is the programmer, the wiki is the codebase.* The vault is persistent state; the
agent is the process that mutates it. That's the right foundation.

But a v1 "research and write a note" skill has an economics problem it doesn't solve
yet: **it writes, but it doesn't compound.** Every run produces a new artifact.
Nothing merges, nothing gets corrected, nothing gets promoted from "note I wrote once"
to "thing the agent reads before it starts working." The vault grows linearly and its
usefulness grows sub-linearly.

The thesis of v2:

> A knowledge skill is only valuable if writing a note makes the *next* run cheaper.
> Every design decision below is judged by whether it increases the hit-rate of
> "answer already exists in the vault."

Three levers to get there:

| Lever | Mechanism | Measured by |
|---|---|---|
| **Don't duplicate** | Search vault before researching; update existing notes instead of creating near-duplicates | duplicate rate, notes-updated ÷ notes-created |
| **Make notes findable** | Strict frontmatter, MOC hubs, backlinks, canonical slugs | vault-hit rate at session start |
| **Make notes trustworthy** | Citations, freshness dates, verified code, confidence field | stale-note rate, correction count |

---

## 2. Current state (v1) and its failure modes

*(Reconstructed from your description. The Claude Code prompt starts with an audit
phase that will replace these assumptions with facts from your repo.)*

**v1 flow as described:**

```
user asks "how does X work?" / "explain X"
        │
        ▼
 skill description matches → Claude Code offers the skill
        │
        ▼
 user accepts
        │
        ▼
 subagent: understand question + tech → research
        │
        ▼
 write markdown (best practices, use cases) → save to Obsidian vault
```

**Failure modes this shape has** (each maps to a v2 section):

| # | Failure mode | Symptom | Fixed in |
|---|---|---|---|
| F1 | No pre-flight vault search | Ask about `@Transactional` twice, get two notes | §8 |
| F2 | Freeform note shape | Notes aren't machine-queryable; no dataview, no reliable index | §6 |
| F3 | Orphan notes | No `[[links]]`, no MOC entry → note exists but is never retrieved | §7 |
| F4 | No freshness model | A 2024 Spring Boot note reads as authoritative in 2026 | §6, §12 |
| F5 | Same depth for everyone | Senior Java dev gets "what is a bean" boilerplate | §9 |
| F6 | Not configurable | Vault path, language, template all hardcoded → not shippable to others | §10 |
| F7 | Unverified code blocks | Snippets that don't compile land in the vault permanently | §12 |
| F8 | Trigger ambiguity | Skill fires on things it shouldn't, or misses things it should | §16 (evals) |
| F9 | Vault never gardened | No merge, no prune, no re-index over time | §13 |
| F10 | No signal captured | You have no idea what you actually ask about most | §14 |

**Audit checklist to run against the real repo** (Claude Code does this in Phase 0):

- Does `SKILL.md` frontmatter have a `description` written in trigger language?
- Is the vault path hardcoded or configurable?
- Is there a note template, or does the model freestyle the structure?
- Does the skill read anything from the vault before writing?
- Is there a subagent definition file, or is the "agent" just an inline instruction?
- Are there hooks? A `plugin.json`? Any tests/evals?
- What does the memory index look like today, and who updates it?

---

## 3. Goals / non-goals

### Goals

- **G1** — Second occurrence of a topic costs a vault read, not a research run.
- **G2** — Notes have a machine-readable contract (frontmatter) that supports
  Dataview queries, indexing, and programmatic gardening.
- **G3** — Depth, language, and style adapt to a declared developer profile.
- **G4** — Anyone can `claude plugin install` it and point it at their own vault in
  under 5 minutes, with zero code edits.
- **G5** — Code in notes is verified, and claims are cited with a captured date.
- **G6** — The vault self-maintains: an index that rebuilds, notes that merge, stale
  notes that get flagged.
- **G7** — Ships as a public Claude Code plugin with docs, evals, and a demo.

### Non-goals (v2)

- Not a general PKM system. It writes *technical* notes about code, tools, and
  architecture. Meeting notes, journals, and tasks are out of scope.
- Not a vector database. Ripgrep + frontmatter + an index file is the retrieval
  layer. If that stops scaling past ~2k notes, revisit — not before.
- Not a hosted service (that's Appendix A).
- Not an Obsidian plugin. It's a Claude Code plugin that happens to write files that
  Obsidian renders. This matters: it means it also works with plain markdown, VS Code,
  or any other editor.

---

## 4. Design principles

1. **Plain markdown or it didn't happen.** No sidecar DBs, no proprietary index. If
   the tool disappears, the vault is still a readable wiki. This is the single
   biggest reason the Karpathy pattern works.
2. **The index is a file, not a service.** `_index.md` and MOC notes are regular
   markdown that both humans and agents read.
3. **Update beats create.** Creating a new note is the fallback, not the default.
4. **Cite or mark uncertain.** Every non-obvious claim carries a source or gets a
   `confidence: low` flag. No confident hallucinations pinned to your wall.
5. **Config over code.** Everything a user might want to change lives in one config
   file, not in `SKILL.md`.
6. **Progressive disclosure.** `SKILL.md` stays short. Templates, rules, and
   references live in files the skill reads on demand — that's the documented
   `references/`, `templates/`, `scripts/` pattern.
7. **Deterministic where possible.** Slug generation, index rebuilds, and frontmatter
   validation are scripts, not model judgment.

---

## 5. Architecture

### 5.1 Component view

```
┌──────────────────────────────────────────────────────────────────────┐
│                          Claude Code session                          │
│                                                                       │
│  user: "how does Kafka consumer rebalancing work?"                    │
│                    │                                                  │
│                    ▼                                                  │
│      ┌──────────────────────────┐                                     │
│      │  SKILL.md  (forge)       │  ← trigger + orchestration only     │
│      │  ~150 lines, no content  │                                     │
│      └────────────┬─────────────┘                                     │
│                   │ reads on demand                                   │
│      ┌────────────┴──────────────────────────────┐                    │
│      │ config/forge.config.md   (user settings)  │                    │
│      │ templates/*.md           (note shapes)    │                    │
│      │ references/*.md          (rules, taxonomy)│                    │
│      │ bin/forge  (Go binary)   (deterministic)  │                    │
│      └────────────┬──────────────────────────────┘                    │
│                   │                                                   │
│   ┌───────────────┴────────────────────────────────────────┐          │
│   │                    PIPELINE                             │          │
│   │  0 intake → 1 recall → 2 plan → 3 research →            │          │
│   │  4 synthesize → 5 verify → 6 write → 7 link → 8 index   │          │
│   └───────────────┬────────────────────────────────────────┘          │
│                   │                                                   │
│      ┌────────────┴─────────────┐                                     │
│      │ subagents                │                                     │
│      │  • forge-researcher      │  (web + docs MCP, read-only)        │
│      │  • forge-codebase-scout  │  (repo grep, read-only)             │
│      │  • forge-verifier        │  (claim + code check)               │
│      │  • forge-librarian       │  (merge, link, index, garden)       │
│      └────────────┬─────────────┘                                     │
└───────────────────┼───────────────────────────────────────────────────┘
                    ▼
        ┌───────────────────────────────┐
        │      Obsidian vault           │
        │  notes/  moc/  _index.md      │
        │  _inbox/ _archive/ .forge/    │
        └───────────────────────────────┘
```

### 5.2 The pipeline, stage by stage

| # | Stage | Owner | Deterministic? | Output |
|---|---|---|---|---|
| 0 | **Intake** | main | no | normalized question, topic slug, `type`, `stack[]` |
| 1 | **Recall** | script + main | mostly | list of candidate existing notes with scores |
| 2 | **Plan** | main | no | decision: `ANSWER_FROM_VAULT` \| `UPDATE` \| `CREATE`; research questions |
| 3 | **Research** | `forge-researcher` (+`forge-codebase-scout`) | no | raw findings + sources |
| 4 | **Synthesize** | main | no | note body against template |
| 5 | **Verify** | `forge-verifier` | partly | citation check, code check, confidence assignment |
| 6 | **Write** | script | yes | file at canonical path, frontmatter validated |
| 7 | **Link** | `forge-librarian` | partly | `[[wikilinks]]` in + out, MOC entry |
| 8 | **Index** | script | yes | `_index.md` regenerated, `.forge/log.jsonl` appended |

**Stage 1 is the whole ballgame.** If Recall says `ANSWER_FROM_VAULT`, stages 3–8
are skipped and the user gets an answer in seconds with a link to the note. That's
the compounding.

### 5.3 Decision logic at stage 2

```
  best_match_score
        │
        ├─ ≥ 0.85  and note.updated within freshness_window
        │           → ANSWER_FROM_VAULT
        │             (read note, answer, offer "want me to deepen it?")
        │
        ├─ ≥ 0.85  and note is stale
        │           → UPDATE (refresh mode: re-verify claims, bump date)
        │
        ├─ 0.55–0.85
        │           → UPDATE (extend mode: add section, don't rewrite)
        │
        └─ < 0.55  → CREATE (new note, then link to the 0.3–0.55 neighbours)
```

Scoring is a cheap deterministic blend, not a model call:
`0.4·slug/title similarity + 0.3·tag overlap + 0.2·stack overlap + 0.1·body term
frequency (ripgrep hit density)`. Thresholds live in config so users can tune them.

---

## 6. The note contract

Every note the skill writes conforms to this. Enforced by
`forge validate` (Go), which runs before write and fails loudly.

### 6.1 Frontmatter schema

```yaml
---
title: "Kafka consumer group rebalancing"
slug: kafka-consumer-group-rebalancing     # canonical, kebab-case, immutable
type: concept                              # concept|howto|pattern|pitfall|decision|api|incident
stack: [kafka, spring-boot, java]          # controlled vocabulary, see references/taxonomy.md
tags: [messaging, distributed-systems, consumer-group]
depth: 3                                   # 1 skim … 5 deep-dive
confidence: high                           # high|medium|low
created: 2026-08-07
updated: 2026-08-07
verified: 2026-08-07                       # last time claims were re-checked
freshness_days: 180                        # after this, note is flagged stale
sources:
  - url: https://kafka.apache.org/documentation/#basic_ops_consumer_group
    accessed: 2026-08-07
    kind: official
  - url: https://...
    accessed: 2026-08-07
    kind: blog
related: ["[[kafka-partitions]]", "[[spring-kafka-listener-container]]"]
supersedes: []                             # slugs merged into this note
forge_version: 2.0.0
origin: ask                                # ask|session-capture|garden|import
---
```

Design notes on specific fields:

- **`slug` is the identity.** Filenames may change; slug must not. Merges record the
  absorbed slug in `supersedes` so old links can be redirected.
- **`verified` separate from `updated`.** A typo fix bumps `updated`. Only a
  re-check of the claims bumps `verified`. Staleness keys off `verified`.
- **`freshness_days` per note, not global.** A note on binary search never goes
  stale; a note on the Spring Boot 4 migration path goes stale in a quarter. Defaults
  by `type` live in config.
- **`origin`** is what makes §14 telemetry possible without a separate log format.

### 6.2 Body structure (`templates/concept.md`)

```markdown
# {{title}}

> **TL;DR** — {{one sentence a tired engineer can act on}}

## Mental model
{{the analogy or diagram that makes it click — max 1 diagram}}

## How it actually works
{{mechanism, in order of execution}}

## In {{primary_stack}}
{{concrete code, verified, minimal, idiomatic}}

## Best practices
- [ ] {{actionable, imperative, testable}}

## Pitfalls
| Pitfall | Why it happens | Fix |
|---|---|---|

## When NOT to use this
{{the section everyone skips and everyone needs}}

## Open questions
{{explicit "I could not verify X" — never silently omit}}

## Sources
{{auto-rendered from frontmatter}}
```

Separate templates per `type`: `concept.md`, `howto.md`, `pattern.md`, `pitfall.md`,
`decision.md` (ADR-shaped), `api.md`, `incident.md`.

**Rule:** the model fills a template. It does not invent structure. This is what
makes 400 notes feel like one document instead of 400 essays.

---

## 7. Vault topology & the memory graph

```
vault/
├── _index.md                  # generated. the agent's entry point.
├── _inbox/                    # low-confidence / unlinked notes awaiting triage
├── _archive/                  # superseded notes, kept for link integrity
├── moc/                       # Maps of Content — hand-curated hubs
│   ├── java.md
│   ├── spring-boot.md
│   ├── kafka.md
│   └── aws.md
├── notes/
│   ├── concept/
│   ├── howto/
│   ├── pattern/
│   ├── pitfall/
│   └── decision/
├── profiles/
│   ├── me.md                  # developer profile (§9)
│   └── projects/
│       └── <project>.md       # per-codebase context
└── .forge/
    ├── log.jsonl              # append-only event log (§14)
    ├── cache/                 # research cache, TTL'd
    └── state.json             # last index build, counters
```

### 7.1 `_index.md` — the single most important file

Regenerated by `forge index`. It is what a `SessionStart` hook feeds
into context, so it must be small (target < 4KB) and high-signal:

```markdown
# Vault index — 312 notes — rebuilt 2026-08-07

## By stack
- **spring-boot** (61): [[spring-transactional-propagation]] · [[spring-boot-4-migration]] · … → [[moc/spring-boot]]
- **kafka** (28): [[kafka-consumer-group-rebalancing]] · … → [[moc/kafka]]

## Recently updated
- 2026-08-06 [[jvm-g1-tuning]]
- …

## Stale (verify > 180d)
- [[spring-security-6-config]] (verified 2025-11-02)

## Gaps — asked but never written
- "ECS task draining" (asked 3×)
```

That last section is quietly the best feature: the log knows what you asked and the
index knows what exists, so the diff is a personal curriculum.

### 7.2 Link strategy

- Every new note gets **≥2 outbound links** and **≥1 inbound link** (the librarian
  edits a neighbour or the MOC to add it). A note with zero inbound links is
  invisible; enforce this or the graph is decoration.
- MOC files are append-only by the agent, hand-editable by you. The agent never
  deletes a human-written line in a MOC.
- `related` in frontmatter mirrors inline links for machine queries.

---

## 8. Retrieval-before-research (the dedup engine)

This is `forge recall` (Go). Deterministic, no model call, runs in < 200ms on a
few-thousand-note vault.

```
input:  normalized question, extracted terms, stack hints
        │
        ├─ 1. slug candidates:  slugify(question) + n-gram variants
        ├─ 2. frontmatter scan: parse all notes' title/slug/tags/stack (cached in .forge/state.json)
        ├─ 3. ripgrep pass:     term hit-density in bodies, top 20 files
        ├─ 4. score & rank
        └─ 5. emit JSON: [{slug, path, score, updated, verified, stale}]
```

Output goes back to the main agent, which applies §5.3's decision logic.

**Why not embeddings?** Three reasons, and they're worth writing down because a
reviewer will ask:

1. A model already read the question; lexical recall plus a *model* re-rank of the
   top 20 is as good as vectors at this scale, at zero infra cost.
2. Embeddings mean an index that can drift out of sync with the files — which breaks
   principle #1 (plain markdown or it didn't happen).
3. It's a `v2.2` upgrade behind a config flag (`recall.strategy: lexical|hybrid`)
   once someone hits the ceiling. Design the interface now, don't build it yet.

**UPDATE mode matters as much as dedup.** Two sub-modes:

- `extend` — append/insert a section, add sources, bump `updated`. Never rewrite
  what's there. Add a `## Changelog` line.
- `refresh` — re-verify existing claims against current sources, correct what
  changed, bump `verified`. Diff is shown to the user before write.

---

## 9. Personalization layer

Lives at `vault/profiles/me.md` — a human-editable markdown file, generated by an
interactive `forge init` on first run.

```yaml
---
primary_language: java
frameworks: [spring-boot, spring-cloud, hibernate]
infra: [aws-ecs, docker, kafka, postgres]
seniority: mid            # affects what gets explained vs assumed
default_depth: 3
note_language: en
explain_style: mechanism-first     # mechanism-first | example-first | analogy-first
assume_known: [oop, rest, sql, git, basic-concurrency]
never_assume: [k8s-internals, jvm-gc-internals]
code_style:
  java: "constructor injection, no field @Autowired, records for DTOs, Java 21"
avoid: ["marketing language", "history lessons", "'in this article we will'"]
---
```

How each field is actually used (this must be explicit in `SKILL.md` or the model
ignores the profile):

| Field | Concrete effect on output |
|---|---|
| `primary_language` | all code examples in Java unless the topic is language-specific |
| `assume_known` | those concepts are referenced, never re-explained |
| `never_assume` | those get a one-line primer inline |
| `seniority` | `mid` → skip definitions, keep tradeoffs; `senior` → skip tradeoff basics, keep edge cases |
| `explain_style` | orders the template sections |
| `code_style` | passed verbatim into the synthesis prompt as constraints |
| `avoid` | negative constraints in the final pass |

**Profile learning (v2.2):** the librarian periodically reads `.forge/log.jsonl` and
proposes profile edits — "you've asked about ECS 7× and always requested more depth;
raise `default_depth` for `aws` to 4?" — as a diff you approve. Never silent.

---

## 10. Customization layer

One file: `config/forge.config.md` (frontmatter-only markdown, so it's readable in
Obsidian too). Nothing else in the plugin should be edited by users.

```yaml
---
vault_path: ~/Obsidian/second-brain
paths:
  notes: notes
  moc: moc
  inbox: _inbox
  archive: _archive
  index: _index.md

trigger:
  mode: ask                  # ask | auto | manual
  # ask    = current behaviour: offer, wait for accept
  # auto   = write without asking when confidence is high
  # manual = only on explicit /forge

recall:
  strategy: lexical          # lexical | hybrid
  answer_threshold: 0.85
  update_threshold: 0.55

freshness_days:
  concept: 365
  howto: 180
  api: 90
  pattern: 365
  decision: 0                # decisions never go stale, they get superseded

research:
  max_sources: 6
  prefer: [official-docs, source-code, rfc]
  allow_domains: []
  deny_domains: []
  use_docs_mcp: true         # context7-style docs MCP if available
  scan_codebase: true        # ground the note in the repo you're in

verify:
  run_code: auto             # auto | never | ask
  require_citation_for: [version-specific, performance-claim, security-claim]

write:
  language: en
  max_note_words: 1200
  diagrams: mermaid          # mermaid | ascii | none

garden:
  enabled: true
  schedule: weekly

telemetry:
  enabled: true
  scope: local               # local | team (team = Appendix A)
---
```

**Presets** ship in `config/presets/`: `java-backend.md`, `frontend.md`,
`devops.md`, `minimal.md`. `forge init` asks four questions and copies a preset.
This is the difference between "a skill I built for me" and "a plugin 500 people
install."

---

## 11. Subagent topology

Defined as markdown files in `agents/`. Each gets a *minimal* tool set — that's both
a safety property and a quality property (a researcher that can't write files can't
half-write a note).

| Agent | Tools | Job | Returns |
|---|---|---|---|
| `forge-researcher` | WebSearch, WebFetch, docs-MCP, Read | Answer N research questions from authoritative sources | findings + `sources[]` with dates |
| `forge-codebase-scout` | Glob, Grep, Read | How is this *actually* used in the current repo? | file:line examples, local conventions |
| `forge-verifier` | Read, Bash(sandboxed), WebFetch | Compile/run snippets; spot-check claims against sources | pass/fail per claim, confidence |
| `forge-librarian` | Read, Edit, Glob, Grep, Bash | Merge, link, MOC upkeep, index rebuild, staleness sweep | diff of vault changes |

Two things worth calling out:

- **`forge-codebase-scout` is the differentiator.** Generic notes about Kafka are a
  Google search. A note that says *"in this repo, `OrderConsumer` sets
  `max.poll.interval.ms=300000` because the enrichment call is slow — see
  `OrderConsumer.java:47`"* is knowledge that exists nowhere else. This is also
  exactly the seam where the B2B version becomes obvious.
- **Researcher and scout run in parallel.** They have no dependency on each other.

---

## 12. Quality gates & verification

Gates run at stage 5. A note that fails a gate goes to `_inbox/` with
`confidence: low` rather than being silently published.

| Gate | Check | On fail |
|---|---|---|
| **Schema** | frontmatter validates against `references/schema.yaml` | block write, fix, retry once |
| **Citation** | every claim tagged version-specific / perf / security has a source | mark that claim `⚠️ unverified`, drop confidence |
| **Code** | Java/Kotlin/TS snippets compile in a scratch dir; shell snippets `--dry-run` or shellcheck | replace with pseudocode + note, or drop confidence |
| **Freshness** | every source has an `accessed` date; version-specific claims name the version | add version, or generalize the claim |
| **Anti-slop** | no "in this article", no restating the question, no filler; ≥1 concrete example; word count under cap | rewrite pass |
| **Link** | ≥2 outbound, ≥1 inbound | librarian adds them before write completes |
| **Duplicate** | no existing note scores ≥0.85 (re-checked post-synthesis) | switch to UPDATE mode |

`verify.run_code: auto` should compile in a throwaway directory, never in the user's
project. Say so in the README — it's the first thing a security-minded reviewer asks.

---

## 13. Hooks & automation

Hooks are what turn this from a command you remember to run into a system.
Config at `hooks/hooks.json`.

| Event | Script | Purpose |
|---|---|---|
| `SessionStart` | `inject_index.sh` | Print a trimmed `_index.md` + active project profile to stdout → lands in context. Claude now *knows what it knows* before the first prompt. |
| `UserPromptSubmit` | `detect_learning_intent.sh` | Cheap regex for "how does/why does/what is/explain/difference between" → emit a nudge that the vault may already have it, with the top recall hit. |
| `SessionEnd` | `capture_session.sh` | Extract unresolved "we figured out that…" moments into `_inbox/` as stubs. Free knowledge you'd otherwise lose. |
| `PostToolUse` (WebFetch) | `cache_source.sh` | Cache fetched docs into `.forge/cache/` so a re-run doesn't re-fetch. |

Note the documented gotcha: `SessionStart` re-runs on resume (with `source: resume`),
while mid-session hook output is *replayed* from saved text on `--continue`/`--resume`
rather than re-executed. So don't put anything time-sensitive in a `UserPromptSubmit`
hook and expect it to be current. Put freshness in `SessionStart`.

**The gardener** (`garden.schedule: weekly`) is a `forge-librarian` run that:
1. Flags notes past `freshness_days`.
2. Proposes merges for pairs scoring ≥0.85 against each other.
3. Rescues `_inbox/` orphans — links them or archives them.
4. Rebuilds `_index.md` and the "Gaps" section.
5. Writes a `moc/weekly-review-YYYY-WW.md` summary.

Ship it as the `/forge-check` weekly checker (addendum §C — T0-only by default);
users who want it automated wire it to a cron
or the scheduling feature of their choice. Don't auto-run destructive gardening on a
timer in the OSS default — that's how you get an angry issue about a mangled vault.

---

## 14. Telemetry (local-first, opt-in)

`.forge/log.jsonl`, append-only, plain JSON, in the user's own vault:

```json
{"ts":"2026-08-07T10:14:22Z","event":"ask","q_hash":"9f2a…","topic":"kafka-consumer-rebalancing","stack":["kafka","spring-boot"],"decision":"CREATE","recall_top_score":0.31,"duration_ms":48200,"sources":6,"project":"order-service"}
{"ts":"2026-08-07T11:02:10Z","event":"ask","topic":"kafka-consumer-rebalancing","decision":"ANSWER_FROM_VAULT","recall_top_score":0.94,"duration_ms":1900}
```

Log the *topic*, never the raw question text (store a hash), never code, never file
contents. Say this in the README in one sentence and you defuse the entire privacy
objection before it's raised.

**What it powers immediately (single user):**

- `/forge-stats` — what you actually study, vault hit-rate over time, time saved.
- The "Gaps" section of the index — asked ≥2×, never written.
- Profile learning proposals (§9).

**Why it matters strategically:** this log is the whole B2B thesis in embryo. Get the
schema right now, single-user, and the team version is an aggregation layer — not a
rewrite. That's the correct sequencing, and it's the version of the idea you can
actually finish.

---

## 15. Phased roadmap

Each phase is independently shippable and independently demoable. Do not start
phase N+1 with phase N unmerged.

### Phase 0 — Audit & baseline *(0.5 day)*
- Inventory the current skill: files, triggers, prompts, vault layout.
- Snapshot metrics: note count, % with frontmatter, % with ≥1 inbound link,
  duplicate clusters, orphan count.
- Write `docs/AUDIT.md`. **Everything after this is judged against these numbers.**

### Phase 1 — Contract & migration *(1–2 days)* ← *highest value per hour*
- `references/schema.yaml`, `templates/*.md`, `forge validate` + `forge slug` (Go).
- Vault migration: backfill frontmatter on existing notes (dry-run first,
  git-commit the vault before running, always). Throwaway Python is fine here —
  it runs once and never ships.
- `forge index` + first `_index.md`. Add the D3 human-edit capture hook now
  (addendum §D.1) — data only accumulates forward.
- **Done when:** 100% of notes validate; `_index.md` builds in one command.

### Phase 2 — Recall *(1 day — hardening pass; most of it exists)*
- `forge recall` + scoring + JSON contract + `--explain` + mtime cache.
- Decision logic in `SKILL.md`; UPDATE modes (`extend` / `refresh`).
- **Done when:** asking a known question answers from the vault in <5s and creates
  no new file. This is the demo that sells the whole project.

### Phase 2b — Static analysis core ⭐ *(4–6 days incl. Go ramp)*
- The Go core: vault parse, similarity (MinHash+LSH), link graph, tree-sitter code
  index, git signals, **drift detection**, linkcheck, the 10 reports.
  Full spec: addendum §B; stack/ADR: STACK doc. The most defensible work in the project.

### Phase 3 — Config & personalization *(1–2 days)*
- `forge.config.md`, `profiles/me.md`, `forge init`, presets.
- Remove every hardcoded path from the codebase.
- **Done when:** a stranger can install and run it without editing plugin files.

### Phase 3b — Engine abstraction ⭐ *(1–2 days)*
- One interface, four backends (none/host/api/advisor), budget accounting,
  `engine_trail`, hard locks on recall/write/index. Spec: addendum §A + §E.

### Phase 4 — Subagents & verification *(2–3 days)*
- Four agent definitions; parallel researcher + scout.
- Verifier gates; `_inbox/` quarantine path; advisor as critique-mode second pass.
- **Done when:** a deliberately wrong snippet gets caught and demoted.

### Phase 5 — Hooks & weekly checker *(2 days)*
- `SessionStart` index injection, intent detection, session capture, incremental
  drift on Edit/Write.
- `/forge-check` (T0-only weekly report — addendum §C), `/forge-stats`.
- **Done when:** a fresh session's first response cites an existing note unprompted.

### Phase 5b — Log-back into the codebase *(1 day)*
- `docs/knowledge-map.md`, per-module CLAUDE.md fragments, `.forge/code-index-<repo>.json`
  — one per configured `--repo`, not one shared file. Said the singular name until
  2026-08-23; corrected under BACKLOG B-027.
  Inline markers stay opt-in. Spec: addendum §B.7.

### Phase 6 — Package & release *(2–3 days)*
- `plugin.json`, marketplace repo, README, docs site, evals, CI, demo GIF,
  goreleaser + cross-compile matrix + checksum-pinned `bin/forge` shim (STACK §4).
- **Done when:** `claude plugin marketplace add <you>/forge` works from a clean box.

### Phase 6b — Dataset capture & export *(1–2 days)*
- D1–D5 capture, `/forge-export-dataset`, datasheets. Spec: addendum §D. Python OK
  (offline tooling, never ships to user machines).

### Phase 7 — Polish on real usage *(ongoing, 1 month)*
- Run it daily. Fix what annoys you. Publish `/forge-stats` from your own month of
  use as the launch post's centrepiece. First LoRA on D1 routing at ~300 pairs.

**Realistic total: ~4 weeks of focused evenings to a public v2.0** (incl. Go ramp).
If you cut: 6b capture stays but export waits → 5b → advisor tier. **Never cut 2b.**

---

## 16. OSS release plan

### 16.1 Repository layout

Ship as a **plugin** (distribution format) containing **skills** (the content). Per
the current plugin reference: `.claude-plugin/plugin.json` holds the manifest, and
every other component directory lives at the plugin root — not inside
`.claude-plugin/`.

```
knowledge-forge/
├── .claude-plugin/
│   └── plugin.json
├── skills/
│   ├── forge/
│   │   ├── SKILL.md                 # trigger + orchestration, keep it short
│   │   ├── templates/{concept,howto,pattern,pitfall,decision,api}.md
│   │   └── references/{schema.yaml,taxonomy.md,writing-rules.md,recall-spec.md}
│   ├── forge-init/SKILL.md          # onboarding wizard
│   ├── forge-check/SKILL.md         # weekly checker (T0)
│   └── forge-stats/SKILL.md         # personal analytics
├── agents/
│   ├── forge-researcher.md
│   ├── forge-codebase-scout.md
│   ├── forge-verifier.md
│   └── forge-librarian.md
├── hooks/hooks.json
├── bin/forge                         # shim → Go binary (STACK §4); source in cmd/ + pkg/
├── cmd/forge/  pkg/                  # Go core: recall, validate, index, drift, reports
├── config/
│   ├── forge.config.example.md
│   └── presets/{java-backend,frontend,devops,minimal}.md
├── evals/
│   ├── triggers.yaml                 # should-fire / should-not-fire cases
│   ├── golden/                       # expected note outputs
│   └── run_evals.py
├── examples/vault/                   # 15–20 real notes — this IS the documentation
├── docs/{getting-started,configuration,architecture,note-schema,faq}.md
├── .github/workflows/ci.yml
├── README.md  LICENSE(MIT)  CHANGELOG.md  CONTRIBUTING.md
```

Set `version` explicitly in `plugin.json` so users only get updates when you bump it
(omitting it makes every commit an update — fine for internal, bad for published).

A **separate marketplace repo** (or a `.claude-plugin/marketplace.json` in the same
repo) is what users `claude plugin marketplace add`. Validate before every release:
`claude plugin validate ./knowledge-forge`.

### 16.2 The README (this is the product)

Order matters. Most plugin READMEs bury the point.

1. **One sentence + one GIF.** The GIF must show the *second* ask — the moment it
   says "you already know this: [[note]]" and answers in 2 seconds. Not the research
   run. The payoff.
2. **The problem, in 4 lines.** "You explain Kafka rebalancing to your agent. Next
   week, you explain it again. And again."
3. **Install** — 3 commands, one of them `/forge-init`.
4. **What a note looks like** — paste a real one, frontmatter and all.
5. **How it decides** — the §5.3 decision tree as a code block. Shows there's
   engineering here, not just a prompt.
6. **Configuration** — the config block, collapsed.
7. **Privacy** — one paragraph: local files, topics not questions, opt-out flag.
8. **Not for you if…** — honest scope. Buys enormous credibility.

### 16.3 Evals — the thing that separates this from 3,000 other plugins

`evals/triggers.yaml`:

```yaml
should_fire:
  - "how does @Transactional propagation work"
  - "explain kafka consumer rebalancing"
  - "what's the difference between ECS Fargate and EC2 launch type"
should_not_fire:
  - "fix this NullPointerException"        # debugging, not learning
  - "rename this variable"
  - "what did I do yesterday"
```

Plus golden-note tests: fixed question + frozen sources → assert the output validates
against the schema, contains required sections, and cites ≥1 official source. Run in
CI. **"Has evals" is a genuinely rare signal in this ecosystem** and it's the single
cheapest credibility purchase available to you.

### 16.4 Launch

- **Pre-launch:** use it for 3–4 weeks. Ship `examples/vault/` from your real notes
  (scrubbed). Record the GIF. Write the post.
- **Post:** *"I gave Claude Code a memory that gets cheaper the more I use it"* —
  lead with the metric from your own `/forge-stats`: *"after 30 days, 41% of my
  technical questions were answered from the vault instead of researched."* A real
  number from real use beats any feature list.
- **Where:** r/ClaudeAI, r/ObsidianMD, HN Show HN, the Claude Code plugin
  marketplaces that aggregate community plugins, X/LinkedIn.
- **First-week discipline:** answer every issue within 24h, tag 5 `good first issue`s,
  merge one outside PR. That's what turns a repo into a project.

---

## 17. CV / portfolio framing

You're a Java/Spring backend dev. Frame this as **systems engineering**, not
"I wrote some prompts." The substance is real — make the framing match it.

### Resume bullet

> **Knowledge Forge** — open-source Claude Code plugin (⭐ N) that gives AI coding
> agents persistent, self-maintaining technical memory in a plain-markdown vault.
> Designed a deterministic retrieval-before-research pipeline that cut redundant
> research runs by ~40%, a validated note schema with automated freshness and
> citation gates, and a multi-agent verification stage. Shipped with an eval suite
> and CI.

### The four talking points

1. **"I found the real bottleneck."** v1 wrote notes. The insight was that writing
   isn't the problem — *retrieval before writing* is. Everything followed from that.
   This is exactly the "identify the actual constraint" story interviewers probe for.
2. **"I chose the boring solution and wrote down why."** Lexical recall + frontmatter
   over a vector DB, with the upgrade path designed but not built. Engineering
   judgment about *not* building things is more senior-signalling than the opposite.
3. **"I designed for failure."** Quarantine to `_inbox/` instead of publishing junk,
   validation gates that block writes, dry-run migrations, git-commit-before-mutate.
   Anyone who's run a production Spring service recognizes the instincts.
4. **"I made it measurable."** Local event log → hit-rate, staleness, gaps. Then used
   those numbers to decide what to build next and to write the launch post.

### Artifacts to have ready

| Artifact | Why |
|---|---|
| The repo, with a real commit history across phases | shows iteration, not a dump |
| `docs/architecture.md` with the pipeline diagram | 60-second whiteboard-ready explanation |
| `docs/AUDIT.md` → before/after metrics | before/after numbers are rare and land hard |
| A 90-second demo video | most people won't clone it |
| One ADR (e.g. "lexical recall vs embeddings") | proves you can write a decision doc |

### For the Java/Spring angle specifically

Don't hide that it's Python-and-markdown. Say: *"the plugin internals are Python
because that's what the runtime wants; the interesting part is the pipeline design —
the same staged-validation, quarantine-on-failure, idempotent-rebuild patterns I use
in Spring services."* Then have the `forge-codebase-scout` demo pointed at a Spring
Boot repo. That's the moment it stops being a toy.

---

## Appendix A — B2B extension sketch

*(You deselected B2B for this doc; here's the seam so v2 doesn't foreclose it.)*

The single-user log (§14) is the product wedge. The team version is an aggregation
layer, not a rewrite:

```
 dev laptops (N)                    team plane                    outputs
┌──────────────┐            ┌────────────────────────┐      ┌─────────────────┐
│ forge plugin │──events───▶│ ingest (topics only,   │─────▶│ team gap report │
│ local vault  │            │ hashed identity)       │      │ onboarding paths│
│              │◀──sync─────│ shared note repo (git) │      │ doc-debt heatmap│
└──────────────┘            └────────────────────────┘      └─────────────────┘
```

- **What a team actually buys:** "your 12 devs asked about the payments service 40
  times last month and there is no doc for it" — doc debt, quantified, ranked by
  cost. That's a real budget line, unlike "AI notes."
- **Shared vault = a git repo of notes.** Personal notes stay local; notes tagged
  `scope: team` push to the shared repo via PR. Review workflow you already have.
- **Onboarding:** the gap report becomes a generated ramp-up path for new hires.
- **Hard constraints from day one:** topics not questions, no code contents, opt-in
  per repo, self-hostable. Anything less and no engineering org will run it.
- **Don't build this until** the OSS version has real users. The single-user tool is
  the distribution channel and the proof; building the server first is the classic
  way to end up with an unused pipeline.

---

## Appendix B — risks & mitigations

| Risk | Likelihood | Mitigation |
|---|---|---|
| Migration mangles your existing vault | med | dry-run default, `git commit` gate, `--backup` flag, refuse to run on a dirty tree |
| Skill fires too often, becomes noise | high | `trigger.mode: ask` default, tight description, eval suite for should-not-fire |
| Notes are confidently wrong | med | citation gate, `confidence` field, `## Open questions` section, `_inbox/` quarantine |
| Scope creep into a full PKM | high | non-goals in §3, pinned in CONTRIBUTING.md |
| Nobody installs it | med | the GIF shows the payoff, not the process; launch with real 30-day metrics |
| Recall doesn't scale past ~2k notes | low | frontmatter cache in `state.json`; hybrid strategy behind a config flag |
| You burn out before Phase 6 | **high** | Phases 1–2 alone are already worth it and demoable. Ship each phase. |

---

## Appendix C — file listings

Files to create, in build order (Go subcommands live in `cmd/forge` + `pkg/`):

**Phase 1** — `references/schema.yaml` · `templates/{concept,howto,pattern,pitfall,decision}.md` ·
`forge slug` · `forge validate` · `forge index` · one-time migration script (throwaway Python)

**Phase 2** — `forge recall` · `references/recall-spec.md` · rewritten `SKILL.md`

**Phase 2b** — `pkg/{vault,similarity,graph,codeindex,gitsig,drift,linkcheck,report,store}` ·
the 10 reports · `moc/codebase.md` (spec: addendum §B, STACK §10)

**Phase 3** — `config/forge.config.example.md` · `config/presets/*.md` ·
`skills/forge-init/SKILL.md` · `profiles/me.template.md`

**Phase 3b** — engine abstraction in `pkg/engine/` · presets (addendum §A, §E)

**Phase 4** — `agents/forge-researcher.md` · `agents/forge-codebase-scout.md` ·
`agents/forge-verifier.md` · `agents/forge-librarian.md` ·
`references/writing-rules.md` · `forge verify-code`

**Phase 5** — `hooks/hooks.json` · hook shims calling `forge` (`session-context`,
`intent`, `capture`, `drift --since-commit`) · `skills/forge-check/SKILL.md` ·
`skills/forge-stats/SKILL.md`

**Phase 5b** — `forge logback` → `docs/knowledge-map.md` · CLAUDE.md fragments ·
`.forge/code-index-<repo>.json`

**Phase 6** — `.claude-plugin/plugin.json` · `.claude-plugin/marketplace.json` ·
`README.md` · `LICENSE` · `CHANGELOG.md` · `CONTRIBUTING.md` · `docs/*.md` ·
`docs/adr/{0001-lexical-recall,0002-go-for-static-core}.md` ·
`evals/triggers.yaml` · `evals/run_evals.py` · `.github/workflows/ci.yml` ·
goreleaser + release matrix · `examples/vault/*`

**Phase 6b** — dataset capture (D1–D5) · `/forge-export-dataset` · datasheets
(Python, offline tooling)

---

## Sources

- [Plugins reference — Claude Code Docs](https://code.claude.com/docs/en/plugins-reference)
- [Hooks reference — Claude Code Docs](https://code.claude.com/docs/en/hooks)
- [Andrej Karpathy's LLM Wiki: self-updating AI second brain with Obsidian](https://www.mindstudio.ai/blog/andrej-karpathy-llm-wiki-obsidian-ai-second-brain)
- [Self-evolving Claude Code memory with Obsidian and hooks](https://www.mindstudio.ai/blog/self-evolving-claude-code-memory-obsidian-hooks)
- [Claude Code Plugin Marketplace & Skills Guide (2026)](https://www.alexcloudstar.com/blog/claude-code-plugin-marketplace-skills-2026/)
- [obsidian-second-brain — prior art](https://github.com/eugeniughelbur/obsidian-second-brain)
- [claude-code-memory-setup — prior art](https://github.com/lucasrosati/claude-code-memory-setup)
