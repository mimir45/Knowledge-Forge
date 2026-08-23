# Claude Code execution prompts — consolidated (rev 2)

Paste these into Claude Code **in your skill's repo**, in order. One phase per
session (or per `/clear`) — they're sized to fit comfortably in context.

Put all four docs in the repo root first:
`KNOWLEDGE-FORGE-DESIGN.md` · `KNOWLEDGE-FORGE-ADDENDUM.md` ·
`KNOWLEDGE-FORGE-B2B.md` · `KNOWLEDGE-FORGE-STACK.md`

Phase order: **0 → 1 → 2 → 2b → 3 → 3b → 4 → 5 → 5b → 6 → 6b → 7**

> **Before Phase 0:** `git init` / commit your vault. Several phases mutate it.
> **Stack (rev 2):** the deterministic core is Go (STACK ADR-001) — a single
> static binary `forge` in the plugin's `bin/`. Drift detection is git-tree
> anchored with symmetric rollback (addendum §B.6).

---

## Phase 0 — Audit

```
Read KNOWLEDGE-FORGE-DESIGN.md in full, plus KNOWLEDGE-FORGE-ADDENDUM.md sections
A and B, and KNOWLEDGE-FORGE-STACK.md. Together they are the target design.

I want a factual audit of what exists today, before we change anything.

1. Map the current implementation:
   - Every file that belongs to the skill: SKILL.md, prompts, scripts, agent
     definitions, hooks, config, plugin manifest. Give me a tree with a one-line
     purpose for each.
   - Quote the SKILL.md frontmatter verbatim (name + description).
   - Trace the exact control flow from "user asks a question" to "file written to
     the vault". Note where the model decides vs. where code decides.

2. Locate the vault and profile it (read-only, do not modify):
   - total notes, total words
   - how many have YAML frontmatter, and which keys appear (frequency table)
   - how many have >=1 outbound [[link]]; how many have >=1 INBOUND link
   - orphan count (zero inbound links)
   - likely duplicate clusters: group notes whose titles/slugs are >0.7 similar
   - directory structure and whether it's consistent
   - is there an index/MOC file, and what updates it

3. Grade the current implementation against sections 6 through 14 of the design
   doc. For each section: PRESENT / PARTIAL / ABSENT, plus one line of evidence
   from the actual files. No guessing — if you can't find it, say ABSENT.

4. Confirm or correct the failure modes F1-F10 in section 2 of the design doc.
   Which ones are real here? Which did I get wrong? Are there failure modes I
   missed that you can see in the code?

5. For each of the 9 pipeline stages (design doc 5.2), state what currently
   executes it: a script, an inline model instruction, a subagent, or nothing.
   Flag every place model judgment is used for something a script could do
   deterministically — those are the T0 conversion candidates.

6. Assess my existing dedup/recall honestly against design doc section 8 and
   addendum A.3. I believe most of it exists. Tell me specifically: is scoring
   deterministic and reproducible? Is there a cache? Does it run BEFORE research,
   always, with no bypass path? Does anything model-driven leak into it?
   List only the deltas needed — do not propose rebuilding what works.

7. Is a code repository available alongside the vault? If so: languages, build
   system, size, and whether any note already references file paths or symbols
   (grep for file:line patterns and class names). Report how many such references
   exist and how many are currently BROKEN against the current git HEAD — that's
   the Phase 2b drift baseline.

Write all of this to docs/AUDIT.md. Include a "Baseline metrics" table at the top
with the numbers from steps 2 and 7 — we re-measure against it after every phase.

Do not write any other file. Do not modify the vault or the skill yet.
```

---

## Phase 1 — Note contract & vault migration

```
Read KNOWLEDGE-FORGE-DESIGN.md sections 6, 7 and 12, docs/AUDIT.md, and
KNOWLEDGE-FORGE-STACK.md section 2 (library choices).

Implement the note contract. Validation and indexing are GO subcommands of a
single binary (module layout per STACK: cmd/forge/ + pkg/). The one-time vault
migration may be a throwaway Python script — it runs once and never ships.

Build:
1. references/schema.yaml — the frontmatter schema from section 6.1, as a real
   validatable schema. Every field: type, required/optional, allowed values,
   default. Adapt the controlled vocabularies in `stack` and `tags` to the actual
   tags found in my vault during the audit — don't invent a taxonomy I'm not using.
   Also write references/taxonomy.md documenting it.
   Include the rev-2 fields: engine_trail (list), drift_checked_at (sha, optional).

2. templates/{concept,howto,pattern,pitfall,decision,api}.md — section 6.2. Each
   is a real markdown file with {{placeholders}} and inline HTML comments telling
   the model what belongs in each section and, importantly, what does NOT.

3. `forge slug` — deterministic slug generation (pkg/vault). Same input must
   always give the same slug. Collisions get a numeric suffix.

4. `forge validate <path|--all>` — validates notes against schema.yaml.
   Exit 0 = pass. Exit 1 = fail with a precise, actionable error list.
   Support --fix for mechanically fixable problems (missing dates, key order,
   normalizing tag case). Frontmatter parsing via goldmark + yaml.v3.

5. `forge index` — generates _index.md per design doc section 7.1. Idempotent,
   under 200ms warm on my vault, output under 4KB (truncate long lists, always
   keep counts and the stale list). Cache in SQLite via modernc.org/sqlite
   (pure Go — the binary must stay static). The DB is derived only: `forge
   reindex` must fully rebuild it from the markdown.

6. migrate_vault.py (throwaway) — backfills the contract onto existing notes.
   Non-negotiable requirements:
   - --dry-run is the DEFAULT; writing requires an explicit --apply
   - refuses to run if the vault git tree is dirty
   - --backup copies the vault before touching anything
   - infers type/stack/tags from content, but marks every inferred field with
     confidence: low so I can review
   - never deletes or reorders body content; only adds frontmatter
   - prints a summary diff of exactly what would change

7. The D3 human-edit capture hook (addendum D.1): a post-commit hook on the VAULT
   repo that records (generated note, human-edited note) pairs into
   .forge/datasets/d3.jsonl when a note is edited within 7 days of generation.
   Install it now — this data only accumulates forward.

Then:
- run migrate_vault.py --dry-run and show me the summary. STOP and wait for my
  approval before --apply.
- after I approve, run it, then `forge validate --all` and `forge index`, and
  report: % of notes now valid, and any that still fail.

Go constraints: table-driven tests with fixtures; `forge --help` on every
subcommand; CGO_ENABLED=0 (nothing in this phase needs cgo).
Do not touch SKILL.md in this phase.
```

---

## Phase 2 — Recall (the important one)

```
Read KNOWLEDGE-FORGE-DESIGN.md sections 5.2, 5.3, and 8, and the audit's recall
assessment (step 6 of Phase 0). Most of recall already exists — this is a
hardening pass plus the SKILL.md rewrite. Do not rebuild what works; port what
exists into the Go binary and close the listed deltas.

Build:
1. `forge recall`
   - input: --question "..." [--stack java,spring-boot] [--vault PATH]
   - output: JSON array of {slug, path, title, score, updated, verified, stale,
     matched_on} sorted by score desc, top 10
   - scoring exactly as section 8: 0.4 slug/title similarity + 0.3 tag overlap
     + 0.2 stack overlap + 0.1 body term-hit density (ripgrep or pkg/vault scan)
   - frontmatter cache in SQLite keyed by path+mtime; cold run may be slow, warm
     run must be under 200ms on my vault
   - --explain flag prints the score breakdown per candidate

2. references/recall-spec.md — documents the scoring, the thresholds, and the
   decision tree from 5.3, so SKILL.md can stay short and point here.

3. Rewrite skills/forge/SKILL.md around the pipeline:
   - Keep it UNDER 200 lines. It orchestrates; it does not contain content.
     Templates, rules and specs are read on demand from templates/ and references/.
   - The frontmatter description must be written in trigger language: the phrasings
     I would actually type ("how does X work", "explain X", "what's the difference
     between X and Y", "best practices for X") and an explicit negative clause for
     what it must NOT fire on (debugging, refactoring, "what did I do yesterday").
   - Stage 1 always runs `forge recall` FIRST. Non-negotiable. No research happens
     before recall has reported.
   - Implement the three decision branches, including both UPDATE sub-modes:
       * extend  = insert a new section, add sources, bump `updated`, append a
                   Changelog line. NEVER rewrite existing body text.
       * refresh = re-verify existing claims against current sources, correct only
                   what changed, bump `verified`. Show me a diff before writing.
   - On ANSWER_FROM_VAULT: read the note, answer from it, show the [[link]], and
     offer to deepen it. Do not create a file.

Test it in front of me:
- pick 3 topics my vault already covers -> each must resolve ANSWER_FROM_VAULT
  with score >0.85 and create zero files
- pick 2 topics adjacent to existing notes -> each must resolve UPDATE(extend)
- pick 1 genuinely new topic -> must resolve CREATE and then link to its nearest
  neighbours
Report the actual scores. If the thresholds are wrong for my vault, tell me the
numbers you'd pick instead and why — don't silently change them.
```

---

## Phase 2b — Static analysis core ⭐ (Go)

```
Read KNOWLEDGE-FORGE-ADDENDUM.md section B and KNOWLEDGE-FORGE-STACK.md in full.

Build the T0 static core in GO. Single static binary, shipped in the plugin's
bin/. Everything here is deterministic — no model calls anywhere in this code.
If you find yourself wanting one, stop and tell me instead.

Layout:
  cmd/forge/            CLI (cobra or stdlib flag)
  pkg/vault/            frontmatter + markdown AST (goldmark), mtime-cached
  pkg/similarity/       MinHash + LSH banding, hand-rolled. NO embeddings.
  pkg/graph/            note link graph: components, hubs, orphans, centrality
  pkg/codeindex/        go-tree-sitter for java/kotlin (start there; add py/ts on
                        demand); pom.xml / build.gradle / package.json -> dep +
                        version map. Output .forge/code-index-<repo>.json,
                        one per --repo (B-027).
  pkg/gitsig/           go-git: churn, blame ownership, co-change coupling
  pkg/drift/            THE KEY PACKAGE — addendum B.6. AST comparison, not line
                        diffs. BROKEN / SUSPECT / auto-repair-line-numbers.
  pkg/linkcheck/        HTTP HEAD on sources, cached, rate-limited
  pkg/report/           renders analyses to markdown
  pkg/store/            SQLite via modernc.org/sqlite (PURE GO, no cgo).
                        Derived cache only — `forge reindex` must fully rebuild it
                        from the markdown. Never a source of truth.

Hard requirements:
- CGO_ENABLED=0 for every package except codeindex (tree-sitter needs cgo).
  Keep the cgo surface isolated to that one package.
- File parsing uses a worker pool sized to GOMAXPROCS, with a path+mtime+size cache.
- Drift is GIT-ANCHORED per addendum B.6:
    * `forge drift --since-commit <sha>` — installed as post-commit / post-merge /
      post-checkout git hooks on the CODE repo. Never triggered on file save.
      Never evaluates uncommitted working-tree changes — half-finished edits are
      not drift.
    * `git diff --name-only <sha>..HEAD` is the cheap gate; AST comparison runs
      only on files in that set. Never re-scan the vault.
    * Each note records drift_checked_at: <sha>. Demotion history lives in
      .forge/ (slug+sha keyed), NOT in note body churn.
    * ROLLBACK SYMMETRY: verdicts are a pure function of (note refs, tree state).
      On revert/reset/checkout, a symbol that reappears in the new HEAD restores
      the note automatically — confidence back to its pre-demotion value,
      drift.md entry cleared, log line citing both SHAs.
    * Branch-aware: verdicts are per-HEAD; the weekly checker reports against the
      default branch only.
- Table-driven tests with fixtures for every package. Benchmarks for parse,
  similarity, and drift. Include a rollback test: break a symbol, assert demotion;
  git revert, assert restoration.
- `forge --help` for every subcommand.

Performance targets — measure and report actuals, do not assume:
  forge drift --since-commit <sha> < 100ms   (this is the binding constraint)
  forge index                      < 200ms
  forge check (full vault + repo)  < 10s warm

Then generate all 10 reports from addendum B.4 into vault/reports/. Ranking is
the point: staleness.md sorts by (ask_frequency x days_overdue), not
alphabetically. Also generate vault/moc/codebase.md per addendum B.5 including
the "high churn + high complexity + zero notes" section.

Build/release:
- Makefile targets for local build and a cross-compile matrix
  (darwin/linux/windows x amd64/arm64). Use `zig cc` for local cgo
  cross-compilation; a GitHub Actions native-runner matrix for releases.
- goreleaser config, SHA-256 checksums published per artifact.
- bin/forge as a shell shim: resolve ~/.forge/bin/<version>/forge, download from
  the GitHub release for the detected platform if missing, VERIFY the pinned
  SHA-256 before executing. Honor FORGE_BIN override for airgapped installs.

Do NOT build the daemon from STACK section 5. Ship the CLI, measure the hook
path, and report the number to me. We decide on a daemon from data.

Finish by running everything against my vault and repo and showing me all 10
reports plus the benchmark numbers. I especially want drift.md — how many of my
existing notes reference code that has already changed?
```

---

## Phase 3 — Config & personalization

```
Read KNOWLEDGE-FORGE-DESIGN.md section 9 and KNOWLEDGE-FORGE-ADDENDUM.md
section E (the revised config — it supersedes design doc section 10).

Goal: a stranger can install this and use it without editing a single plugin file.

1. Grep the entire repo for hardcoded values — vault paths, my name, language
   assumptions, "java", tag names, thresholds. List every one you find, then move
   all of them into config/forge.config.example.md using the exact schema in
   addendum E (including the engines, static, check, and dataset blocks). Config
   loading with a clear precedence chain:
   env var > project .forge.config.md > user config > packaged defaults.
   Implement loading in Go (pkg/config) so the binary and the skill read the
   same file.

2. config/presets/ — the four presets from addendum E: offline, claude-only,
   byo-api, max. Plus stack presets java-backend, frontend, devops, minimal.
   Each with a comment at the top saying who it's for.

3. profiles/me.template.md — the developer profile from design doc section 9.

4. skills/forge-init/SKILL.md — an onboarding wizard that:
   - finds or asks for the vault path, validates it's writable
   - asks at most 5 questions (primary language, frameworks, infra, seniority,
     trigger mode) and picks presets from the answers
   - writes forge.config.md and profiles/me.md
   - installs the git hooks (drift on the code repo, D3 capture on the vault repo)
   - runs `forge index` and prints a "you're set up" summary
   Target: under 2 minutes from install to first use.

5. Wire the profile into synthesis so it demonstrably changes output. In SKILL.md,
   spell out the concrete effect of each profile field — the table in design doc
   section 9. A profile the model reads but doesn't act on is worse than no profile.

Prove it: run the same question twice, once with seniority:junior/depth:2 and once
with seniority:senior/depth:5, and show me both notes side by side. If they look
similar, the wiring is wrong — fix it.
```

---

## Phase 3b — Engine abstraction ⭐

```
Read KNOWLEDGE-FORGE-ADDENDUM.md sections A and E.

1. Define ONE engine interface (Go, pkg/engine) that all four tiers implement:
      Run(stage, prompt, context, constraints) -> {output, tokens, costUSD, tier}
   Backends: none (errors for stages that need generation), host (session model —
   surfaced to the skill as instructions, since the binary can't call the host
   model directly; document this seam precisely), api (anthropic/openai/
   openrouter/ollama via config), advisor (critique mode).

2. Per-stage engine selection from config, with fallback chains
   (e.g. verify: advisor -> local -> host).

3. HARD LOCKS: recall, write, and index accept ONLY engine `none`. If config sets
   anything else, refuse to start with a clear error explaining that these stages
   must be reproducible and why. Do not silently override — refuse.

4. Advisor in CRITIQUE mode, not rewrite. It receives draft + sources + schema +
   writing rules, and returns ONLY: disputed claims with reasoning, what's missing,
   a confidence verdict, and a minimal patch. It must not regenerate the note.
   Log the critique verbatim — it's dataset D2.

5. Budget accounting: per-day USD caps per tier, persisted in SQLite.
   on_exhausted: degrade | queue | stop. `queue` writes pending_advisor:true to
   the note and the weekly checker drains it.

6. Add engine_trail: [host, advisor] to note frontmatter (already in the schema
   from Phase 1 — now populate it). Generate reports/cost.md from it.

Prove it: run the same question under all four presets (offline, claude-only,
byo-api, max). Show me the four results, the cost of each, and confirm `offline`
produced a usable outcome — it should do recall, report the match, and tell me
clearly what it cannot do rather than failing.
```

---

## Phase 4 — Subagents & verification gates

```
Read KNOWLEDGE-FORGE-DESIGN.md sections 11 and 12.

1. Create the four agents in agents/ per the table in section 11. Give each the
   MINIMUM tool set listed — the researcher must not be able to write files. Each
   agent file states its job, its output contract (exact shape it returns), and
   its hard limits (max sources, max tool calls) so a run can't spiral.

2. forge-codebase-scout is the one that matters most. It answers "how is this
   ACTUALLY used in the repo I'm in right now" and returns file:line evidence plus
   local conventions — seeded from .forge/code-index-<repo>.json so it greps
   instead of broadly. Its findings go into the note's "In <stack>" section,
   attributed to the repo. Generic notes are a web search; this is the part that
   makes the note irreplaceable.

3. Run researcher and codebase-scout in PARALLEL. They have no dependency.

4. Implement the quality gates from section 12 as a pre-write stage:
   - schema, citation, code-compiles, freshness, anti-slop, link, duplicate-recheck
   - `forge verify-code` compiles/lints snippets in a THROWAWAY directory,
     never in my project, never with network access
   - a note that fails a gate goes to _inbox/ with confidence:low and an
     explicit "## Open questions" section listing what couldn't be verified.
     It does NOT get published silently and it does NOT get dropped.
   - failing draft + gate error + fixed draft -> append to dataset D4.

5. references/writing-rules.md — the anti-slop rules. Ban: "in this article",
   restating my question back at me, history lessons, marketing adjectives,
   bullet lists with one word per bullet. Require: >=1 concrete example, a
   "when NOT to use this" section, a word cap.

Adversarial test: feed the pipeline a topic where you deliberately include a
snippet that will not compile and a version-specific claim with no source. Show
me that it lands in _inbox/ with confidence:low and the failures named. If it
publishes clean, the gates are decorative — fix them.
```

---

## Phase 5 — Hooks, weekly checker, stats

```
Read KNOWLEDGE-FORGE-DESIGN.md section 13 and KNOWLEDGE-FORGE-ADDENDUM.md
sections C and 14.

1. hooks/hooks.json plus thin shims calling the forge binary:
   - SessionStart -> `forge session-context`: trimmed _index.md + active project
     profile to stdout, hard cap 4KB, fail silent (log to .forge/, never print
     errors into context).
   - UserPromptSubmit -> `forge intent`: cheap regex only, NO model call, must
     add under ~50ms. Emits the top recall hit if score >0.7.
   - SessionEnd -> `forge capture`: extracts "we established that..." moments
     into _inbox/ stubs.
   - PostToolUse (WebFetch) -> `forge cache-source`: caches into .forge/cache/
     with TTL.
   Be aware of the documented resume behaviour: SessionStart re-runs on resume,
   but mid-session hook output is REPLAYED from saved text on --continue/--resume
   rather than re-executed. So nothing time-sensitive goes in UserPromptSubmit.
   Note this in a comment in hooks.json.

   Drift is NOT wired to Claude Code hooks — it's on the code repo's git hooks
   (post-commit/merge/checkout), already installed in Phase 2b/3. Verify they
   fire and stay under 100ms.

2. skills/forge-check/SKILL.md — the weekly checker per addendum C, T0-ONLY by
   default:
   - runs all 10 reports; writes moc/weekly/YYYY-WW.md ranked by ACTIONABILITY
     using the exact structure in addendum C ("Act now" / "Review" / "Vault stats
     with week-over-week deltas" / "Gaps")
   - zero model calls, zero network except linkcheck, under 10s
   - safe to run from cron or CI
   - optional AI pass behind check.ai_pass (default false), doing ONLY what T0
     cannot: draft the refresh for the top BROKEN note, propose merge text for
     duplicate pairs, stub an ADR for the top undocumented module. Every mutation
     shows a diff and needs my approval. Never auto-mutate the vault on a schedule.
   - if engines.budget.on_exhausted is `queue`, drain pending_advisor notes here.

3. Event log: append to .forge/log.jsonl per design doc section 14. Log the TOPIC
   and a hash of the question — never the raw question text, never code, never
   file contents. telemetry.enabled config flag must actually disable it.

4. skills/forge-stats/SKILL.md — reads the log and reports: vault hit-rate over
   time, most-asked topics, gaps (asked >=2x, never written), estimated research
   time saved, staleness trend. Render as a compact terminal table.

Run /forge-check on my vault now and show me the week file.
```

---

## Phase 5b — Log-back into the codebase

```
Read KNOWLEDGE-FORGE-ADDENDUM.md section B.7.

Build `forge logback` (T0, deterministic, idempotent):

1. docs/knowledge-map.md in the code repo — module -> relevant notes, generated
   from the code↔note map (addendum B.5). Regeneration must produce zero diff
   when nothing changed.
2. A CLAUDE.md fragment per module ("Relevant notes: [[...]]") so ANY agent
   session in that repo gets the vault's knowledge without the plugin. Managed
   begin/end sentinel blocks — never touch content outside them.
3. Keep .forge/code-index-<repo>.json fresh (already built in 2b) and documented for
   other tools.
4. Inline `// forge: [[note-slug]]` markers: implement behind
   static.logback.inline_markers, DEFAULT OFF, same sentinel discipline,
   fully revertible with `forge logback --remove-markers`.

Never modify code semantics. Comments and separate files only.

Run it on my repo and show me the generated knowledge-map.md and one CLAUDE.md
fragment.
```

---

## Phase 6 — Package, document, release

```
Read KNOWLEDGE-FORGE-DESIGN.md section 16 and KNOWLEDGE-FORGE-STACK.md
sections 3, 4, and 8.

1. Restructure the repo to the layout in design doc 16.1 (rev 2: Go source in
   cmd/ + pkg/, binary shim in bin/). .claude-plugin/plugin.json holds ONLY the
   manifest — every component directory (skills/, agents/, hooks/, bin/) sits at
   the plugin ROOT, not inside .claude-plugin/. Set an explicit
   "version": "2.0.0" so users only get updates when I bump it.
   Then run `claude plugin validate .` and fix everything it reports.

2. .claude-plugin/marketplace.json so it's installable via
   `claude plugin marketplace add <me>/knowledge-forge`.
   Verify from a clean directory that install actually works — including the
   bin/forge shim downloading and checksum-verifying the right platform binary.

3. evals/:
   - triggers.yaml with >=15 should_fire and >=15 should_not_fire cases. Pull the
     should_not_fire cases from real adjacent intents: debugging, refactoring,
     git operations, project status questions.
   - golden/ — 3 frozen question+sources fixtures with expected outputs. Assert
     schema validity, required sections present, >=1 official source cited.
   - run_evals.py with a pass/fail summary and a non-zero exit on failure.
   - .github/workflows/ci.yml: go vet + golangci-lint, go test ./..., schema
     validation, evals, AND the goreleaser cross-compile matrix from STACK
     section 3 (native runners for cgo; zig cc documented for local builds).

4. README.md in the exact order given in design doc 16.2 — GIF and payoff first,
   problem second, install third. Include the section 5.3 decision tree verbatim
   as a code block. Add the honest capability table from addendum B.2 and the
   tier matrix from addendum A.1. Publish the measured drift-hook latency ("adds
   Nms per commit") — a real number builds trust. Privacy gets its own short
   section (topics-not-questions, local-only datasets). Include the
   "Not for you if..." section — cut it and the README reads like marketing.

5. docs/: getting-started, configuration, architecture (with the pipeline
   diagram), note-schema, faq, datasets. Plus two ADRs:
   docs/adr/0001-lexical-recall-vs-embeddings.md and
   docs/adr/0002-go-for-static-core.md (from STACK sections 1 and 6, including
   the rejected alternatives — keep the Rust paragraph).

6. examples/vault/ — 15-20 real notes from my vault, scrubbed of anything
   private.

7. LICENSE (MIT), CHANGELOG.md, CONTRIBUTING.md. CONTRIBUTING must restate the
   non-goals from design doc section 3.

Finally: give me a release checklist I can run before tagging v2.0.0, and a list
of exactly what I still need to produce by hand (the demo GIF, the launch post,
the screenshots).
```

---

## Phase 6b — Dataset capture & export

```
Read KNOWLEDGE-FORGE-ADDENDUM.md section D.

Python is fine here — this is offline tooling and never ships to a user machine.

1. Confirm capture is wired for all five sets into .forge/datasets/*.jsonl:
   D1 routing        — question hash + features -> decision + outcome (every run)
   D2 advisor        — draft, critique, accepted patch (from critique-mode runs)
   D3 human edits    — the vault post-commit hook from Phase 1. THE MOST VALUABLE.
   D4 gate repair    — failing draft + gate error -> fixed draft (from Phase 4)
   D5 style          — (question, profile, sources) -> accepted note
   Never store raw question text — hash + extracted topic only.

2. /forge-export-dataset --set <d1..d5> --format jsonl-sft|jsonl-dpo|csv
   --since DATE --anonymize
   --anonymize strips repo names, absolute paths, internal URLs, identifiers,
   using the same scrubber as examples/vault/.

3. Every export emits a DATASHEET beside it: counts, date range, engine_trail
   distribution, stack distribution, and a stated limitations section (e.g. "83%
   Java/Spring; do not expect generalization").

4. /forge-dataset-stats shows current counts against the volume thresholds in
   addendum D.2, and tells me plainly what is and is not yet trainable. Do not
   overclaim — if I have 180 D1 pairs, say "enough for a routing LoRA on a 1-3B
   model, not enough for anything else."

5. Document in docs/datasets.md: local-only, no transmission, manual export, user
   owns the data. Put this in the README privacy section too.
```

---

## Phase 7 — After a month of real use

```
Read .forge/log.jsonl, .forge/datasets/, and docs/AUDIT.md.

I've been using this daily for a month. Produce docs/RESULTS.md:

- vault hit-rate over time (weekly buckets) — is it trending up?
- notes created vs. notes updated ratio
- duplicate rate now vs. the Phase 0 baseline
- drift events: how many BROKEN/SUSPECT verdicts fired, how many auto-repairs,
  how many rollback restorations — vs. the Phase 0 broken-reference baseline
- most-asked topics, and the gaps that are still unwritten
- estimated research time saved (ANSWER_FROM_VAULT runs x median research duration)
- which quality gates actually fired, and how often
- dataset counts per D1-D5 against the addendum D.2 thresholds; if D1 >= 300,
  propose the routing LoRA experiment
- honest list of what didn't work and should be cut

Then draft the launch post. Lead with the single strongest number from above, not
with a feature list. Structure: the problem in 4 lines, the one insight
(retrieval-before-research), the metric, how it works, install. Under 900 words.
```

---

## Notes on using these

- **Phases 0, 2 and 2b are the ones that matter.** Phase 0 gives you the
  before-numbers; Phase 2 is the feature that makes it worth using; Phase 2b is
  the engineering that makes it defensible.
- **Don't let Phase 1's migration run un-reviewed.** Vault mutation is the one
  irreversible step in this plan.
- Drift discipline: git-anchored, commit-triggered, rollback-symmetric. If any
  phase output suggests running drift on file saves or uncommitted changes,
  reject it — that's addendum B.6 being violated.
- If Claude Code proposes changes beyond a phase's scope, tell it to write them
  to `docs/BACKLOG.md` instead. Phase discipline is what gets this finished.
