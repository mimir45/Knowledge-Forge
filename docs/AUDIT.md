# Phase 0 — Audit

**Date:** 2026-08-09 · **Scope:** the v1 `til-writer` skill, the real vault at
`/Users/mimir45/Documents/Base`, and the design docs in `docs/`.
**Deliverable of Phase 0.** Nothing was written outside this file; the vault and the
skill were not modified.

Every scalar below was measured, not estimated. Where a number came from a subagent it
was re-derived in the main session before landing in the baseline table — the table is
what later phases re-measure against, so a wrong number here propagates.

---

## Baseline metrics

| Metric | Value | Source |
|---|---|---|
| **Vault** | | |
| Total `.md` files (excl. `.obsidian`, `.trash`, `.git`) | 109 | step 2 |
| Total words | 58,108 | step 2 |
| — external symlinked content (`raw/daily/`) | 13,056 | step 2 |
| — vault-authored | ~45,050 | step 2 |
| Mean words/note | 533 | step 2 |
| Notes per top-level dir | `TIL/` 27 · `concepts/` 19 · `issues/` 15 · `decisions/` 14 · `entities/` 12 · `sources/` 9 · root 5 · `raw/` 5 · `syntheses/` 3 | step 2 |
| Directory depth | 1–3 | step 2 |
| Notes with YAML frontmatter | 94 / 109 (86.2%) | step 2 |
| Notes with **no** frontmatter | 15 | step 2 |
| Distinct frontmatter keys in use | **6** (schema §6.1 defines 19) | step 2 |
| Distinct `type` values | **0 — key never appears** | step 2 |
| Distinct `status` values | 2 (`active` ×72, `design-complete-no-code` ×1) | step 2 |
| Distinct tags | 131 across 431 assignments; **83 (63%) used exactly once** | step 2 |
| Wikilink instances (code-fence stripped) | 452 | step 2 |
| Notes with ≥1 **outbound** wikilink | 78 / 109 (71.6%) | step 2 |
| Notes with ≥1 **inbound** wikilink | 80 / 109 (73.4%) | step 2 |
| Zero-inbound notes | 29 | step 2 |
| — **content orphans** | **21 — every one of them a `TIL/` note** | step 2 |
| — orphan-by-design (`raw/daily/` symlinks) | 5 | step 2 |
| — non-note files (`CLAUDE.md`, `CCFA_…_TR.md`) | 2 | step 2 |
| — graph roots (`index.md`) | 1 | step 2 |
| TIL notes containing **any** `[[wikilink]]` | **6 / 27** — and all 6 are the `food-ordering-system/` subtree, written by a different process | step 2 |
| TIL notes **outside** `food-ordering-system/` containing a wikilink | **0 / 21** | step 2 |
| Near-duplicate slug pairs (ratio > 0.7) | 11 (3 high-confidence, 1 exact-basename collision) | step 2 |
| Dangling wikilinks | 4 (all one cause: missing `raw/daily/` symlinks for 04-18/20/21/22) | step 2 |
| Dangling `source:` paths | 4 (same four notes, same cause) | step 2 |
| Vault git history | 1 commit (`404d85e`), clean, no remote | step 2 |
| **v1 skill** | | |
| Files in `~/.claude/skills/til-writer/` | **1** (`SKILL.md`, 106 lines, 4.0K) | step 1 |
| Occurrences of `index` in SKILL.md | **0** | step 1 |
| Occurrences of `[[` in SKILL.md | **0** | step 1 |
| Scripts / agent definitions / hooks / config / tests | **0 of each** | step 1 |
| **Code corpus (Phase 2b drift baseline)** | | |
| Genuine `file:line` refs vault-wide | **0** (1 raw regex hit, disqualified — a sample JSON blob) | step 7 |
| Distinct code filenames referenced | 80 (54 `.java`, 0 `.kt`) | step 7 |
| Code repos reachable | 3 — 2 git, 1 working-tree-only | step 7 |
| Broken `.java` refs vs HEAD (basename matching) | 6 / 54 (11%) — `MeterReadingsService` only | step 7 |
| Broken FE refs vs HEAD | 5 / 8 real path refs — 4 **never committed**, 1 deleted | step 7 |
| Path-shaped Java refs that fail suffix resolution | 14 / 19 | step 7 |

---

## 1. What the v1 system actually is

The whole skill is **one markdown file**. `find ~/.claude/skills/til-writer -type f`
returns exactly one path:

```
/Users/mimir45/.claude/skills/til-writer/
└── SKILL.md    106 lines, 4.0K
```

No `scripts/`, no `references/`, no `agents/`, no hooks, no config, no plugin manifest,
no tests. Frontmatter, verbatim:

```yaml
---
name: til-writer
description: Use when the user asks an explanatory question about a technology, integration, or concept — e.g. "how does X work", "how can I integrate X with Y", "what is X", "explain X". Triggers background research and writes a TIL note to Obsidian.
---
```

### Control flow — model-decides vs code-decides

| # | Step | Decided by | Where |
|---|---|---|---|
| 1 | Skill description injected into context at session start | **CODE** | harness skill discovery |
| 2 | "Explanatory question or task request?" | **MODEL** | `SKILL.md:10-17`, prose criteria only |
| 3 | Ask the user the 1/2/3 confirmation | **MODEL** | `~/.claude/CLAUDE.md:5` — an instruction, not a hook |
| 4 | Derive `<kebab-case-topic>` | **MODEL** | no slugifier exists anywhere |
| 5 | Pick the topic subfolder | **MODEL** | `~/.claude/CLAUDE.md:7` mapping table, applied by judgment |
| 6 | Dedupe — "same topic already exists → update instead" | **MODEL** | `SKILL.md:98`. **No index, no search, no tool call specified** |
| 7 | Spawn the background agent | CODE executes, **MODEL** composes | `SKILL.md:24-29` is illustrative pseudocode |
| 8 | Research | **MODEL** | `SKILL.md:45` |
| 9 | Humanize | **MODEL** | `SKILL.md:47-57` — a 10-bullet checklist the model self-applies; no linter |
| 10 | Emit the note schema | **MODEL** | `SKILL.md:60-89` |
| 11 | Write the file | CODE executes, **MODEL** picks the path | `SKILL.md:58`, `SKILL.md:91` |
| 12 | Permission check on the write | **CODE** | `~/.claude/settings.json:8` — `Write(/Users/mimir45/Documents/Base/TIL/**)` |

Step 12 is the only code-side gate in the entire flow, and it is a path allowlist — it
constrains *where* a write may land and decides nothing about content, naming, freshness,
duplication, or whether to write at all.

**Everything else is a model call.** That is the finding that sizes the project: the T0
static core does not exist in any form, so Phase 1 is not a port, it is a build.

---

## 2. Vault profile

Topology as it stands:

```
Base/
├── index.md, log.md                          roots (index.md itself has 0 inbound)
├── CLAUDE.md, lint-report.md, CCFA_…_TR.md   non-notes at root
├── Untitled.base, Untitled 1.base            Obsidian Bases (excluded from counts)
├── raw/daily/            5   symlinks to /Users/mimir45/claude-memory-compiler/daily/
├── sources/daily/        9
├── entities/            12
├── concepts/            19
├── decisions/           14
├── issues/              15
├── syntheses/            3
├── archive/              0   declared in Base/CLAUDE.md, exists but empty
└── TIL/                 27   databases docker git go java-spring{,/food-ordering-system,/keycloak} react-nextjs tools
```

**Consistency verdict: two vaults sharing one directory.**

- The `llm-wiki` region (`raw` → `sources` → `entities|concepts|decisions|issues|syntheses`,
  77 notes) follows `Base/CLAUDE.md` faithfully — flat, kebab-case, 5-key frontmatter,
  registered in `index.md`, densely linked.
- The **`TIL/` region (27 notes) is not described in `Base/CLAUDE.md`'s folder structure
  at all**, nests 2–3 levels deep, and is 78% orphaned.
- `TIL/java-spring/food-ordering-system/` is a third pattern again: `NN-slug.md` numeric
  prefixes, no frontmatter, its own local `00-index.md`.

Neither region matches DESIGN §7's prescribed
`notes/{concept,howto,…}/ moc/ _inbox/ _archive/ profiles/`. Phase 1 is a topology change,
as `CLAUDE.md` already warns.

**Root cause of the orphan cluster, verified directly:** `SKILL.md` contains the string
`index` **zero** times and `[[` **zero** times. The `llm-wiki` skill and `Base/CLAUDE.md`'s
ingest workflow steps (g)/(h) do maintain `log.md` and `index.md`. Two writers, one
bookkeeping path — and `til-writer` is the one that skips it. Of the 21 notes it authored,
**not one contains a wikilink**.

Near-duplicates worth acting on in Phase 1:

| Ratio | A | B |
|---|---|---|
| **1.000** | `TIL/docker/testcontainers-docker-socket.md` | `issues/testcontainers-docker-socket.md` |
| 0.980 | `TIL/tools/intellij-test-sources-root.md` | `issues/intellij-test-source-root.md` |
| 0.857 | `decisions/testcontainers-daemon-socket.md` | `issues/testcontainers-docker-socket.md` |
| 0.829 | `TIL/databases/keyset-cursor-pagination.md` | `concepts/keyset-pagination.md` |

The 1.000 pair is also a link-resolution hazard: identical basenames make a bare
`[[testcontainers-docker-socket]]` ambiguous under Obsidian resolution. Testcontainers
docker-socket is documented in three places across `TIL/`, `issues/`, `decisions/`.

Excluding the single `raw/daily/` ingest gap, the vault has **zero** dangling references —
link hygiene in the `llm-wiki` region is genuinely clean. The problem is confined to `TIL/`.

---

## 3. Grading against DESIGN §6–14

Rule applied: PRESENT means it exists and works; PARTIAL means something in that direction
exists but does not meet the spec; ABSENT means not found. No guessing.

| § | Requirement | Grade | Evidence |
|---|---|---|---|
| 6.1 | 19-key frontmatter schema | **PARTIAL** | 6 keys exist vault-wide (`tags date title source status created`); til-writer emits 3 (`title date tags`). `type`, `slug`, `depth`, `confidence`, `verified`, `freshness_days`, `sources[]`, `related`, `supersedes`, `forge_version`, `origin` — all ABSENT |
| 6.1 | `slug` as immutable identity | **ABSENT** | no slug key; identity is the filename, which the model invents per note |
| 6.1 | `verified` separate from `updated` | **ABSENT** | neither key exists; `date` is write-time only |
| 6.2 | Body template | **PARTIAL** | `SKILL.md:60-89` prescribes a real structure, but it is inline prose in a prompt, not `templates/concept.md`, and its sections differ from §6.2's |
| 7 | Vault topology (`notes/ moc/ _inbox/ _archive/ profiles/`) | **ABSENT** | actual topology is `concepts/ decisions/ entities/ issues/ raw/ sources/ syntheses/ archive/ TIL/` |
| 7.1 | `_index.md` regenerated by `forge index` | **PARTIAL** | `index.md` exists (123 lines, "Last updated 2026-04-26") but is agent-maintained by hand; no `forge`, no automation, and `til-writer` never touches it |
| 7.2 | ≥2 outbound + ≥1 inbound per new note | **ABSENT** | 0 of 21 til-writer notes have any wikilink |
| 8 | `forge recall` before research | **ABSENT** | see §6 below |
| 9 | `profiles/me.md`, depth personalization | **ABSENT** | no profile file. The nearest thing is `USER LEVEL (infer from conversation)` inside the agent prompt — an unrecorded per-invocation guess |
| 10 | `config/forge.config.md` | **ABSENT** | vault path is hardcoded twice in `SKILL.md`; no config file exists |
| 11 | Four `forge-*` subagents | **ABSENT** | `forge-researcher`, `forge-codebase-scout`, `forge-verifier`, `forge-librarian` do not exist. The five agents in this repo's `.claude/agents/` are **workflow** agents for building the project and are not these |
| 12 | Quality gates (schema/citation/code/freshness/anti-slop/link/duplicate) | **ABSENT** | no gate of any kind. Nothing routes to `_inbox/`; `_inbox/` does not exist; `confidence` is not a key |
| 12 | Code verification in a throwaway dir | **ABSENT** | nothing compiles anything |
| 13 | `hooks/hooks.json` (SessionStart / UserPromptSubmit / SessionEnd / PostToolUse) | **ABSENT** | `~/.claude/settings.json` hooks are all GSD + claude-memory-compiler; none reference the skill or the vault |
| 13 | Gardener / `/forge-check` | **ABSENT** | no merge, no prune, no re-index |
| 14 | `.forge/log.jsonl` telemetry | **ABSENT** | no `.forge/` directory; nothing is logged. `log.md` is a human ingest ledger, not events |

Eleven ABSENT, four PARTIAL, zero PRESENT. That is the correct output for a system that
is one prompt file.

---

## 4. Failure modes F1–F10 (DESIGN §2)

All ten confirmed. Two need their severity restated upward, and five new ones are added
below under a distinct `NF-` prefix so they cannot be confused with the F1–F12 fixture
defects in `testdata/README.md`.

| # | Verdict | Evidence |
|---|---|---|
| F1 | **CONFIRMED — worse than stated** | Not merely "no pre-flight search": `SKILL.md:98` states the *intent* ("update existing file instead of creating new one") with no mechanism at all. The measured result is 11 near-duplicate pairs including an exact-basename collision |
| F2 | **CONFIRMED — partially mitigated** | A template does exist (`SKILL.md:60-89`), so notes are not fully freeform. But 3 frontmatter keys and 15 notes with no frontmatter at all means the vault is still not machine-queryable — `type` has **zero** values vault-wide |
| F3 | **CONFIRMED — the single worst one** | 21 content orphans, 100% of them `TIL/`, 0 of 21 with a wikilink. `SKILL.md` mentions neither `index` nor `[[` |
| F4 | **CONFIRMED** | No `verified`, no `freshness_days`. `date` records when it was written, never when it was last checked |
| F5 | **CONFIRMED** | `USER LEVEL (infer from conversation)` is inferred fresh each time and never persisted |
| F6 | **CONFIRMED** | `/Users/mimir45/Documents/Base/TIL/` hardcoded in `SKILL.md`; the harness allowlist at `settings.json:8` hardcodes the same path a second time |
| F7 | **CONFIRMED** | No verification step exists. The 6 broken `.java` refs in step 7 are the downstream cost |
| F8 | **CONFIRMED** | Trigger criteria are prose in `SKILL.md:10-17`. No evals, no test cases, nothing measures fire rate |
| F9 | **CONFIRMED** | `archive/` is declared in `Base/CLAUDE.md` and is empty. Nothing has ever been pruned or merged |
| F10 | **CONFIRMED** | No `.forge/`, no event log, nothing captured |

### New failure modes found in this audit

| # | Failure mode | Evidence | Fixed by |
|---|---|---|---|
| **NF-1** | **No gate whatsoever between the model and the vault.** The background agent writes a finished note straight in. There is no `_inbox/`, no `confidence: low`, no diff, no human review — the note is published the instant it is generated | `SKILL.md:58`, `SKILL.md:91`; the only check is the path allowlist at `settings.json:8` | DESIGN §12 |
| **NF-2** | **The routing logic is duplicated in two files that have already drifted.** `~/.claude/CLAUDE.md:5` is a near-verbatim copy of the `SKILL.md` description, and the two now contradict each other in three places (below) | see the contradiction table | DESIGN §10 — one config file |
| **NF-3** | **The write path is model-decided and unauthorized nesting has appeared.** `SKILL.md:58` writes flat to `TIL/<topic>.md`; `CLAUDE.md:7` mandates a subfolder. CLAUDE.md won in practice — and went further than either doc allows: `TIL/java-spring/food-ordering-system/` and `TIL/java-spring/keycloak/` are two levels deep, plus an ad-hoc `TIL/go/` | vault topology, depth 3 | DESIGN §7 + `forge slug` |
| **NF-4** | **Code references are not machine-resolvable as written.** Notes cite code by logical module shorthand (`common-domain/valueobject/Money.java`) rather than repo-relative path; 14 of 19 path-shaped Java refs fail suffix resolution. Genuine `file:line` refs: **0** | step 7 | Phase 2b — needs a citation format before drift can work |
| **NF-5** | **Notes record work that was never committed.** 4 frontend refs point at files with zero commits on any branch — one note claims "Wrote `index.stories.tsx`" for a file that has never existed in git | `sources/daily/2026-04-16-frontend-code-review.md:26-28`, `sources/daily/2026-04-17-storybook-llm-wiki.md:18,28` | DESIGN §12 code gate |

### The three SKILL.md ↔ CLAUDE.md contradictions (NF-2 detail)

| # | `SKILL.md` | `~/.claude/CLAUDE.md` | Resolved by |
|---|---|---|---|
| 1 | `:21` "Spawn a background agent **immediately** — do NOT wait before answering" | `:5` "**BEFORE** spawning the background agent, ask the user … **Wait** for their reply" | nothing. The model picks per invocation |
| 2 | `:58` flat `TIL/<kebab-case-topic>.md` | `:7` topic subfolder, with a mapping table | nothing. CLAUDE.md won in practice, then exceeded itself (NF-3) |
| 3 | `:60-89` template has no "Why it matters" | `:6` requires a "Why it matters" section | nothing |

---

## 5. The 9 pipeline stages — what executes each one today

ADDENDUM §A.3 stages. "Inline instruction" means prose inside `SKILL.md` that a model
interprets; there is no code path for any of these.

| # | Stage | Executed by today | T0 conversion candidate? |
|---|---|---|---|
| 0 | intake | **Inline instruction** — `SKILL.md:10-17` prose criteria, plus `CLAUDE.md:5`'s confirmation prompt | Partly. Trigger classification stays a model call; the confirm/route step becomes a hook |
| 1 | recall | **Nothing** | **Yes — highest value.** `forge recall`, hard-locked T0. This is the F1/F3 fix |
| 2 | plan | **Nothing** — there is no CREATE/UPDATE/ANSWER decision at all | **Yes.** DESIGN §5.3's scoring blend is arithmetic over recall output |
| 3 | research | Inline instruction → the background agent | No. Genuinely needs a model + web |
| 4 | synthesize | Inline instruction (`SKILL.md:45`) | No |
| 4b | *humanize* | Inline instruction — a 10-bullet anti-AI-vocabulary checklist the model self-applies | **Yes.** A banned-phrase linter is a deterministic string check. Today nothing verifies the checklist was followed |
| 5 | verify | **Nothing** | **Yes, for the mechanical gates** — schema, citation-presence, link-count, duplicate-score, freshness. Compilation stays out-of-process but is still code, not a model call |
| 6 | write | Model picks the path, `Write` executes; allowlist at `settings.json:8` | **Yes — hard-locked T0.** Slug derivation, path resolution, and frontmatter emission are pure functions |
| 7 | link | **Nothing** — 0 of 21 notes have a wikilink | **Yes.** Backlink insertion + MOC append are graph operations |
| 8 | index | **Nothing** — `SKILL.md` never mentions `index` | **Yes — hard-locked T0.** `index.md` regeneration is a fold over frontmatter |

Five of ten rows are "Nothing" (recall, plan, verify, link, index). Seven are T0 candidates. This is the concrete case for the
static core: the stages the v1 system skips entirely are almost exactly the stages that
need no model at all.

---

## 6. Dedup and recall, assessed honestly

The Phase 0 prompt says *"I believe most of it exists."* **None of it exists.**

Point by point against DESIGN §8 and ADDENDUM §A.3:

- **Deterministic and reproducible?** No — there is nothing to be deterministic about.
- **Cached?** No cache, no `.forge/state.json`, no `.forge/` at all.
- **Always runs before research, with no bypass?** It never runs. Research is stage 3 and
  it is the first thing that happens after intake.
- **Anything model-driven leaking into a stage that should be T0?** Every stage is
  model-driven, including all three that DESIGN hard-locks to T0 (recall, write, index).

The only artifact pointing this direction is `SKILL.md:98` — *"Same topic TIL already
exists in vault → update existing file instead of creating new one"* — a statement of
intent under a "Red Flags" heading, with no vault read, no search, no scoring, and no
tool call specified. It is a hope, not a mechanism, and the 11 near-duplicate pairs are
what that costs.

Delta list vs §8's five steps: **all five absent** (slug candidates, frontmatter scan,
ripgrep hit-density, score & rank, JSON emit).

---

## 7. Code corpus and the Phase 2b drift baseline

Three repos are reachable from the machine and referenced by the vault:

| Repo | Git | HEAD | Tracked files | Stack | Size |
|---|---|---|---|---|---|
| `~/Code/BE/MeterReadingsService` | yes (1 dirty file) | `7c1c8bf` (2026-04-30) | 74 | Java 42 + YAML 15; Maven, Docker Compose, Keycloak | 7.1 MB |
| `~/Code/FE/leprecoin` | yes | `72990ab` (2026-05-15) | 392 | TypeScript — 165 `.tsx` + 143 `.ts`; Next.js, Jest, Storybook | 11.6 MB packed |
| `~/Code/BE/food/food-ordering-system` | **no** | — | — | Java multi-module Maven, 6 modules | working tree only |

None of them sits alongside the vault; each is a separate project directory.

**Broken-reference baseline:**

- `MeterReadingsService` — **6 of 54 distinct `.java` refs broken (11%)**. All six are in
  notes from 2026-04-14/15 naming classes (`AdminUserController`, `UserService`,
  `UserResponse`, `SecurityConfig`, `MonthlySummaryService`, `IntegrationTestBase`) that
  are not in the repo. HEAD's message is *"Update Keycloak config, remove unused files,
  and rename field"* — textbook drift.
- `leprecoin` — **5 of 8 real path refs broken.** Four were **never committed to any
  branch** (`git log --all` returns nothing), one existed and was deleted.
- `food-ordering-system` — **not measurable.** No git, so no HEAD to diff against. Roughly
  half the Java reference corpus (the 6 `TIL/java-spring/food-ordering-system/` notes) has
  no history at all.

**The number that matters more than the broken count:** genuine `file:line` references
vault-wide are **0**. The one regex hit is a fabricated example inside a sample JSON blob.
`.kt` refs: **0** — Kotlin appears nowhere, which bears on `pkg/codeindex` starting with
"Java + Kotlin only."

So the honest Phase 2b position is: a broken-ref *rate* is measurable and is ~11% on the
one repo where the comparison is clean, but **line-level drift detection has no input
data**, and the reference format itself (NF-4) has to be fixed before AST comparison has
anything to compare. The vault also has a single commit and no remote, so there is no
vault-side history for a bidirectional diff either.

---

## 8. Doc-vs-doc coherence (Backlog B-001)

Per `CLAUDE.md`, the design docs have only ever been checked against each other where they
*self-flag* a conflict. The three self-flagged ones are not findings and are excluded:
STACK ADR-001 > ADDENDUM §B (Python); STACK ADR-002 > B2B §8 (Spring Boot); DESIGN rev-2
rereads every `scripts/*.py` as a `forge` subcommand.

**Status: pass complete, all findings resolved.** Findings are recorded here rather than
applied to the docs, per `CLAUDE.md`'s "resolve conflicts by precedence in `AUDIT.md`
rather than editing the docs mid-flight." The ones precedence could not settle were
escalated and decided by the user on 2026-08-09; **§8.4 is the decision record and is the
binding artifact** — where it marks a doc line stale, the doc still says the old thing and
§8.4 is what an implementer follows.

### 8.1 Findings

Thirteen reportable contradictions, worst impact first. Every `file:line` below was
re-checked in the main session against the working tree. **Seven** resolve without
escalation (C-1, C-2, C-3, C-4, C-6, C-11, C-12 — five by the precedence rule, C-2 and
C-12 by a scope call recorded as D-7 / D-8); the other **six** needed a human decision and
have one (**§8.4**, D-1 … D-6). 7 + 6 = 13. The `Resolution`
column below is preserved as it stood *before* those decisions, so the escalation is still
legible. Read it together with §8.4, which supersedes it for C-5, C-7, C-8, C-9, C-10 and
C-13.

| ID | Conflict | Side A | Side B | Surfaces at | Resolution |
|---|---|---|---|---|---|
| C-1 | Drift trigger: file-save vs git-anchored | `DESIGN:709` "drift on Edit/Write"; `ADDENDUM:562` "drift hook on Edit/Write" | `ADDENDUM:287` "Git-tree anchored, not file-watch anchored"; `ADDENDUM:510` "never on save"; `STACK:302` "never on file save"; `PROMPT:597` "reject it" | 2b, 5 | **Side B.** STACK wins on the hook path; `STACK:41` labels the git-hook row "rev 2", making the Edit/Write lines rev-1 residue. `STACK:165`'s `PostToolUse` mention reads as index-refresh, not drift |
| C-2 | §E declared the replacement config schema but omits keys other docs require | `ADDENDUM:461` "Replaces §10 of the main doc"; `PROMPT:279` "the exact schema in addendum E" | `DESIGN:502-507` `paths:`; `:516-518` recall thresholds/strategy; `:520-525` `freshness_days`; `:528` `research.max_sources`; `:536` `verify.run_code`; `:541` `write.max_note_words`; `:546` `garden.schedule`; `:549` `telemetry.enabled` | 3 | **Neither alone — Phase 3 emits the union.** §E supersedes DESIGN §10 only for the surface it actually covers; it never restates or removes the rest, and four later passages still read them. "Exact schema" must not be read literally |
| C-3 | Evals harness language: Go vs Python | `STACK:25` "**Go** for harness, YAML for cases" | `DESIGN:770`, `:959` `evals/run_evals.py`; `PROMPT:496` | 6 | **STACK.** Stack question. Not covered by the rev-2 note — the path is `evals/`, not `scripts/`, and `CLAUDE.md:35` scopes surviving Python to `migrate_vault.py` + offline dataset tooling |
| C-4 | `codeindex` languages: four grammars vs two | `ADDENDUM:506` `languages: [java, kotlin, python, typescript]`; `STACK:286` | `STACK:345` "start with Java + Kotlin only"; `PROMPT:206` | 2b, 3 | **Side B.** Later and more specific inside the top-precedence doc, and `CLAUDE.md:136` records it as the invariant. §E's default shrinks to `[java, kotlin]`. STACK contradicts itself here; `:286` is the eventual target |
| C-5 | `cost.md` is a 2b deliverable but its only data source lands in 3b | `PROMPT:247` "all 10 reports"; `ROADMAP:39`; `ADDENDUM:222` `cost.md` … "from `engine_trail`" | `ADDENDUM:675` 3b "Generate reports/cost.md from it"; `PROMPT:342` engine_trail "now populate it" | 2b | **UNRESOLVED — human decision.** Data dependency implies 2b ships 9 reports + a `cost.md` stub, but that changes a *gate criterion* (`ROADMAP:39`), which is not an implementer's call |
| C-6 | Weekly report path: flat file vs subdirectory | `DESIGN:627` `moc/weekly-review-YYYY-WW.md` | `ADDENDUM:336`, `:693`; `PROMPT:421` — `moc/weekly/YYYY-WW.md` | 5 | **ADDENDUM**, by delegation not raw precedence: `DESIGN:629` hands the weekly checker's spec to §C, so DESIGN's inline path is a leftover in a section DESIGN itself delegated. Not cosmetic — index globbing and orphan classification differ |
| C-7 | Config location: one file vs a four-layer chain, under three names | `DESIGN:496` "One file: `config/forge.config.md` … Nothing else … should be edited by users" | `PROMPT:284` "env var > project `.forge.config.md` > user config > packaged defaults"; `DESIGN:765` + `PROMPT:281` `config/forge.config.example.md` | 3 | **UNRESOLVED — human decision** on (a) the user-editable filename and (b) whether env-var and project overrides exist. DESIGN's own file tree already contradicts its §10 sentence |
| C-8 | ADR numbering vs ADR filenames | `STACK:33` "ADR-001 — Go for the T0 static core"; `STACK:183` "ADR-002 — B2B backend language" | `STACK:246` keep §1 as `docs/adr/0002-go-for-static-core.md`; `DESIGN:958` + `PROMPT:512` require `0001-lexical-recall` and `0002-go-for-static-core` | 6 | **UNRESOLVED — human decision.** Both sides sit inside STACK, so precedence cannot break the tie. As written, ADR-001 ships as file `0002`, ADR-002 gets no file, and `0001-lexical-recall` matches no ADR heading anywhere |
| C-9 | `profiles/me.md` has two producers | `DESIGN:455-456` "generated by an interactive `forge init`"; `:555` | `PROMPT:294-298` `skills/forge-init/SKILL.md` "writes forge.config.md and profiles/me.md" | 3 | **UNRESOLVED — human decision.** Output ownership, not naming: building both yields two writers to the same two files with no precedence. DESIGN nominally wins but lists the skill in its own tree (`DESIGN:753`) |
| C-10 | The anonymization scrubber is reused but no phase produces it | `ADDENDUM:436`, `:734` and `PROMPT:546` — "the same scrubber as `examples/vault/`" | `examples/vault/` is hand-curated notes (`PROMPT:515`, `DESIGN:823`); no phase deliverable names a scrubber | 6b | **Side B is the factual state.** `--anonymize` is specified by reference to a non-deliverable, so the export path has no defined redaction behaviour — the one place a mistake leaks employer content. Needs a human call on which phase owns it |
| C-11 | Positioning copy still says the internals are Python | `DESIGN:876` "Don't hide that it's Python-and-markdown" | `STACK:238-240` §8 positioning is a Go static core; ADR-001 at `STACK:33` | 6 | **STACK.** Stack question; `DESIGN:876` is rev-1 residue. Not covered by the rev-2 note, which only remaps `scripts/*.py` paths — this is prose about the implementation language |
| C-12 | Budget accounting store: `.forge/state.json` vs SQLite | `ADDENDUM:670` "persisted in .forge/state.json" | `PROMPT:338` "persisted in SQLite"; `STACK:104` "Replaces the `.forge/state.json` design" | 3b | **STACK**, with a caveat: the swap is self-announced, but `STACK:104` frames SQLite as a *derived cache*, and budget spend is durable state, not rebuildable. Recorded explicitly: **budget counters live in SQLite and must survive `forge reindex`** |
| C-13 | `on_exhausted` has two different defaults | `ADDENDUM:117` `on_exhausted: degrade` | `ADDENDUM:485` `on_exhausted: queue` | 3b | **UNRESOLVED — human decision.** Both are the global `engines.budget` block, same scope, so it is a true contradiction; §E announces it replaces DESIGN §10 (`:461`), not §A.4, so no precedence rule applies. Evidence leans `queue` (§A.4 calls it "the good default for teams"; §C and §E's `max` preset assume a queue exists) — inference, not a stated default |

*(C-13 was recorded as C-1 in the draft of this file, before the full pass returned.)*

### 8.2 Considered and dropped

Not reportable — either satisfiable by both docs, differently scoped, or already
self-flagged: ripgrep vs "zero runtime deps" (`PROMPT:157` hedges); `*.sh` hooks vs `forge`
subcommands (`STACK:19` permits either); engine-preset phase ownership (omission, not
contradiction); `<2s warm full-vault` vs `<10s full vault + repo` (different scopes);
`freshness_days: 180` vs `concept: 365` (a per-note override, by design); stage-7 link owner
(`none` is an overridable default, not a hard lock — underspecification); duplicate
threshold ≥0.85 gate vs >0.85 report (different artifacts, same number); B2B's vector-store
reversal (self-flagged at `B2B:120-122`).

### 8.3 The ones that needed a human — all decided

C-5 (2b report count), C-7 (config filename + override chain), C-8 (ADR numbering), C-9
(who writes `profiles/me.md`), C-13 (`on_exhausted` default), and C-10 (who owns the
scrubber). None could be settled by the precedence rule. §8.1 listed C-10's resolution as
"needs a human call" but §8.3 originally counted only five — that undercount is corrected
here: it was six.

The user made the call on 2026-08-09 ("fix all of them"). Decisions are recorded in §8.4.

### 8.4 Decision record

These are binding for later phases. Each states the decision, why, the phase that applies
it, and the doc lines that become stale once it lands. **No design doc was edited** —
`CLAUDE.md` requires conflicts be resolved here rather than mid-flight in the docs, and
"known discrepancies: record, don't fix". This section is the resolution artifact; where a
doc line below is marked stale, the doc still says the old thing on purpose, and this file
is what an implementer reads.

---

**D-1 — `cost.md` moves from Phase 2b to Phase 3b. (C-5)**

2b ships **nine** reports; `cost.md` is a 3b deliverable. No stub.

*Why:* 2b's gate criterion at `ROADMAP:39` is the latency triple — "`forge drift` <100ms,
`forge index` <200ms, full check <10s warm — measured, not assumed". "All 10 reports" sits
in the *deliverables* column, not the gate. So this moves a deliverable and leaves the gate
untouched, which is what made C-5 look like a gate change and is not. `cost.md`'s only
input is `engine_trail`, which does not exist until 3b populates it (`ADDENDUM:675`,
`PROMPT:342`). A stub would put an unbacked file in the user's real vault.

*Applies at:* 2b (nine reports), 3b (add `cost.md`).
*Stale:* `ROADMAP:39` and `PROMPT:247` "all 10 reports" → nine at 2b, ten from 3b.

---

**D-2 — Config: the four-layer chain stands; DESIGN §10's "one file" is superseded text. (C-7)**

This one turns out to be resolvable by a supersession the pass missed. `ADDENDUM:461` says
§E "Replaces §10 of the main doc" and `PROMPT:275` repeats it. `DESIGN:496`'s "One file …
Nothing else … should be edited by users" **is inside §10**. It is superseded prose, not a
live constraint.

Names and precedence, highest first:

| Layer | Path | Written by |
|---|---|---|
| env var | `FORGE_CONFIG` | the user, ad hoc |
| project | `.forge.config.md` at repo root | the user |
| user | `~/.forge/forge.config.md` | `forge init` (see D-4) |
| packaged | `config/forge.config.example.md` | the repo; **never** edited by users |

`~/.forge/` is already an established location (`ADDENDUM:487`, `~/.forge/models/`). What
survives from §10 is its *intent*, restated as a rule: **nothing inside the plugin tree is
user-edited** — `forge.config.example.md` is a template that gets copied out, never edited
in place.

*The supersession rule this and D-7 both apply* — "§E replaces §10" is read clause by
clause, not wholesale: a §10 line that §E **contradicts** is superseded (`DESIGN:496`,
here); a §10 line §E is merely **silent** on survives (`DESIGN:502-549`, D-7). Without that
distinction D-2 and D-7 look like opposite readings of the same sentence.

*Applies at:* 3 (`pkg/config` implements the chain).
*Stale:* `DESIGN:496`.

---

**D-3 — ADR numbers are two schemes, not a collision. (C-8)**

STACK's "ADR-001" / "ADR-002" are **section labels inside STACK**. `docs/adr/NNNN-*.md` are
**record numbers in the ADR log**. They were never meant to match. The mapping:

- `docs/adr/0001-lexical-recall-vs-embeddings.md` ← **DESIGN §8**, "Why not embeddings?"
  (`DESIGN:434-441`), which `DESIGN:872` already names as the ADR to write. It is not
  orphaned; it just does not live in STACK.
- `docs/adr/0002-go-for-static-core.md` ← **STACK §1** (STACK's "ADR-001"), exactly as
  `STACK:246` instructs.
- **STACK §6** (STACK's "ADR-002", B2B backend language) gets **no file in Phase 6**. Its
  own heading says "deferred to B1", and B2B is out of scope until Phase 7's gate. It
  becomes `0003-` if and when B1 runs.

Phase 6 therefore ships exactly the two files `DESIGN:958` and `PROMPT:512` list. Nothing
is renumbered.

*Applies at:* 6.
*Stale:* `PROMPT:512`'s parenthetical "(from STACK sections 1 and 6)" — `0001` comes from
DESIGN §8, not STACK §6. The Rust paragraph it tells you to keep is in STACK §1, so it
belongs to `0002`.

---

**D-4 — One writer for the config and profile files: the Go binary. (C-9)**

`forge init` is the only code that writes either file, and it writes them at these exact
paths — **`~/.forge/forge.config.md`** (the user layer of D-2's chain, *not*
`config/forge.config.md`, which D-2 forbids as user-edited) and **`<vault>/profiles/me.md`**
(rendered from `profiles/me.template.md`, `PROMPT:293`). `skills/forge-init/SKILL.md` is a
conversational **wrapper**: it finds the vault, asks the ≤5 questions (`PROMPT:294-298`),
then shells out to `forge init` with the answers. The skill writes nothing itself.

*Why:* building both as specified yields two writers to the same two files with no
precedence — the actual failure, which is about output ownership, not naming. Putting the
bytes behind the binary keeps the wizard deterministic and testable without a model, and
matches the T0-first posture everywhere else. It also preserves both docs' visible
behaviour: `DESIGN:455` still gets "generated by an interactive `forge init`", and the
skill in `DESIGN:753` / `PROMPT:294` still exists and still drives the interaction.

*Applies at:* 3. Phase 3 must not build a second writer in the skill.
*Stale:* nothing, once `PROMPT:294-298`'s "writes forge.config.md and profiles/me.md" is
read as "causes them to be written".

---

**D-5 — `on_exhausted` default is `queue`. (C-13)**

§E's value (`ADDENDUM:485`) wins over §A.4's (`ADDENDUM:117`).

*Why:* three converging reasons. §E is the later, replacement config surface and Phase 3 is
directed to emit it (`PROMPT:275`). §A.4 itself calls queue "the good default for teams"
(`ADDENDUM:128`) — it argues for the value its own YAML does not set. And the drain path
(`ADDENDUM:365`, `:704`; `PROMPT:430`) plus §E's `max` preset (`ADDENDUM:545`) both assume
a queue exists; with `degrade` as the default, that drain is dead code in every default
install.

*Safety check, made explicitly because it would have flipped this:* `queue` **cannot stall
a git hook**. `on_exhausted` governs paid-tier (T2/T3) spend only, and `drift`, `index` and
`recall` are hard-locked to T0 with zero model calls — no hook path can reach the budget
check. The <100ms drift budget is unaffected.

*Applies at:* 3b.
*Stale:* `ADDENDUM:117`.
*Carry-over:* when Phase 3 emits the union (C-2 / D-7), take §A.4's `advisor_when`
**including** `stack_in: [security, auth, payments]`, which `ADDENDUM:491` drops. §E's
budget block is narrower than §A.4's; the union is the wider one.

---

**D-6 — The scrubber is a Phase 6 deliverable: `pkg/scrub`, exposed as `forge scrub`. (C-10)**

Phase 6 builds it and uses it to produce `examples/vault/` (`PROMPT:515`) rather than
hand-scrubbing. Phase 6b's `--anonymize` calls the same package (`ADDENDUM:435`,
`PROMPT:545`).

*Why:* both docs already say "the same scrubber as `examples/vault/`", and 6 precedes 6b,
so the only reading under which that sentence is true is one where `examples/vault/` is the
scrubber's first consumer. The docs assumed the tool; no phase named it.

*Two conditions, because this is the one path where a miss publishes employer content:*

1. Phase 6 does not pass until `pkg/scrub` has a fixture test with known-secret inputs —
   implemented is not the bar, tested is.
2. `--anonymize` **fails closed**. If the scrubber is unavailable or errors, the export
   refuses. It never falls back to exporting raw.

*Applies at:* 6 (build + test), 6b (consume).
*Stale:* nothing. This adds a deliverable the docs referenced but never listed.

---

**D-7 — Phase 3's config is the union of §E and DESIGN §10's uncovered keys. (C-2)**

Recorded here as a decision, not just a precedence outcome, because Phase 3 is measured
against it. Under D-2's clause-by-clause supersession rule, §E displaces the §10 lines it
contradicts and is silent on the rest — so the eight key groups at `DESIGN:502-549` that §E
never restates stay live. "The exact schema in addendum E" (`PROMPT:279`) must not be read
literally.

**D-8 — Budget counters live in SQLite and must survive `forge reindex`. (C-12)**

Also a decision rather than a pure precedence call. `STACK:104` swaps `.forge/state.json`
for SQLite, but frames SQLite as a *derived cache* rebuildable from markdown — and budget
spend is durable state that is not. The reindex path must preserve the budget tables. This
is the one documented exception to "SQLite is purely derived"; note it in `pkg/store`.

---

### 8.5 What is still open after §8.4

Nothing from §8.1. All thirteen findings now have a resolution. The split: **seven** needed
no escalation — five settled by the precedence rule (C-1, C-3, C-4, C-6, C-11) and two
settled by a scope call recorded anyway because later phases are measured against them (C-2
→ D-7, C-12 → D-8). The other **six** were decided by the user: C-5 → D-1, C-7 → D-2,
C-8 → D-3, C-9 → D-4, C-13 → D-5, and C-10 → D-6, which assigns the owner §8.1 could not
find. 7 + 6 = 13.

Phase 3 is unblocked (C-7, C-9 decided; C-13 decided for 3b). Phase 6's deliverable list
grows by one package (D-6).

---

## 9. Known discrepancies (recorded, not fixed)

Per `CLAUDE.md`'s standing rule:

- `testdata/README.md` says the real vault has **108** notes. It has **109**. Stale by one.
- `Base/CLAUDE.md` documents `status: active | draft | archived`. `draft` and `archived`
  are **never used**; `design-complete-no-code` is used and **undocumented**. With 72 of 73
  values identical, the field carries essentially no information.
- `Base/CLAUDE.md`'s folder structure does not mention `TIL/` at all, though `TIL/` is the
  largest directory in the vault (27 notes).
- `index.md` has zero inbound links. Treated here as a graph root, not an orphan, but it
  means the vault has no single entry point that anything links to.
- Tag vocabulary is uncontrolled: 63% of tags are used once, with visible unnormalized
  collisions (`config`/`configuration`, `go`/`golang`, `junit`/`junit5`, `issue`/`issues`,
  `migration`/`migrations`).
- The vault contains 2 Obsidian `.base` files, excluded from every count above. They can
  act as dynamic MOC-like views, so index coverage may be slightly understated.
- The fixture is described everywhere as a **13-note** vault but `testdata/vault/` holds
  **15** `.md` files. Both are right: 13 counts content notes, and `index.md` / `log.md` are
  the two roots on top. Recorded so Phase 1 does not re-litigate it as an off-by-two.

---

## 10. Method and caveats

**Measured in the main session** (not delegated): note/word counts, per-directory counts,
frontmatter presence and key frequency, the full link graph (fenced-code and inline-code
spans stripped before wikilink extraction), orphan list, dangling wikilinks, `SKILL.md`
file listing and its `index`/`[[` occurrence counts, TIL wikilink coverage, vault-wide
`file:line` count. Frontmatter keys were extracted from inside the `---` block only, so
body prose cannot be miscounted as a key.

**Delegated and then spot-checked:** the near-duplicate clustering (`difflib` character
ratio on date-stripped slugs, threshold 0.7, validated against the fixture's known
`soft-delete`/`soft-deletion` pair at 0.833), the `source:` path check, and the step 7
repo/broken-ref analysis.

**§8's coherence pass was delegated and then citation-verified, not independently
re-derived.** The main session opened roughly twenty of the cited `file:line` pairs across
DESIGN, ADDENDUM, STACK and PROMPT and confirmed each quote matches the line it points at.
That establishes the citations are real; it does **not** establish that no conflict was
missed, which would require a second, independent pass over the same docs. No competing run
was made — the `cross-checker` agent was written after these findings had already been read,
so any run of it now would be anchored rather than independent.

Caveats:

1. **Semantic duplication is unmeasured.** Similarity here is filename-only. Two notes
   with unrelated slugs and near-identical bodies score 0. That is exactly what
   `pkg/similarity`'s MinHash is for, and it does not exist yet.
2. **Markdown links `[text](path.md)` were not checked** — only `[[wikilinks]]`. Inbound
   counts are therefore conservative.
3. **Ambiguous basename resolution.** The identical `testcontainers-docker-socket.md` pair
   was resolved by first match. Every actual inbound link to it is path-qualified
   (`[[issues/testcontainers-docker-socket]]`), verified by grep, so the TIL copy's
   zero-inbound status is real and not an artifact.
4. **No update cadence available.** One vault commit means per-note age, churn, and
   staleness cannot be derived from history at all.
5. **Frontmatter parsed by regex, not a YAML library.** Multi-line scalars and quoted
   colons could be mis-keyed; spot checks found none.
6. **`raw/daily/*.md` are symlinks** to `/Users/mimir45/claude-memory-compiler/daily/`.
   Their 13,056 words are external content inside the 58,108 total.

---

## 11. What Phase 0 did not do

- **The vault was not modified.** The global rule to update Obsidian notes after each task
  was deliberately not applied here: the Phase 0 prompt says *"do not modify the vault or
  the skill yet,"* and `CLAUDE.md`'s invariant says never auto-mutate the vault. Both are
  more specific than the global rule, and the vault has no backups.
- **During the Phase 0 execution, no file other than `docs/AUDIT.md` was written.**
  `CLAUDE.md` and `.claude/agents/` were changed elsewhere in this session under separate
  user instructions (plan step 3, and a later request for a competing-verification agent) —
  outside the Phase 0 prompt's scope, not as part of it.
- **Agent registry verification: done, 2026-08-09 (after restart).** All six agents in
  `.claude/agents/` — `finder`, `executor`, `explainer`, `vault-analyst`, `doc-auditor`,
  `cross-checker` — registered and are callable. `vault-analyst` was smoke-tested against
  `testdata/vault/` and passed all three acceptance assertions: exactly one content orphan
  (`archive/old-rag-notes.md`, F6), the `soft-delete`/`soft-deletion` near-duplicate pair
  (F7, difflib basename ratio 0.833), and `index.md`/`log.md` classified as roots.

  A `cross-checker` run was spawned **in parallel**, before the primary returned, on a
  strictly neutral prompt that named no note and pre-classified nothing. It re-derived all
  three independently and returned 5/5 CONFIRMED. This matters because the `vault-analyst`
  prompt *did* state that `index.md`/`log.md` are entry points — that assertion was
  contaminated on the primary run and is carried here on the checker's evidence alone.

  Two things the smoke test surfaced that Phase 1 needs:

  1. **`index.md` and `log.md` are not zero-inbound in the fixture.** Each has exactly one,
     and it comes from the other (`log.md` → `[[index]]`, `index.md` → `[[log.md]]`); neither
     receives a link from any content note. So "orphan == zero inbound" classifies them
     correctly here only by accident, and would misclassify the **real** vault's `index.md`,
     which has genuinely zero inbound (§9). `pkg/graph` needs root detection that does not
     reduce to an inbound count.
  2. **An uncatalogued link-form variation (F-Extra1) — a resolver requirement, not a
     defect.** `index.md` contains `[[log.md]]`: the only one of the fixture's 38 wikilinks
     carrying a literal `.md` extension (verified by grep — 38 total, 1 with the extension).
     Obsidian resolves that form, so the fixture link is **valid** and the defect catalogue
     stays at twelve (F1–F12); this is not a thirteenth. What it constrains is *our* code:
     a resolver that naively appends `.md` reads the link as dangling and drops `log.md`'s
     inbound count to zero. `vault-analyst` reported both resolutions and they differ —
     1 vs 2 dangling links, 14 vs 13 notes with inbound. **`pkg/vault`'s link resolver must
     be extension-agnostic**, and this link is the case that tests it. Not fixed, per the
     fixture rule; `testdata/README.md` correctly still catalogues F1–F12.
