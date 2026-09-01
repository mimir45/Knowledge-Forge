# Knowledge Forge — Addendum: Engine Tiers, the No-AI Core, and the Data Flywheel

Companion to `KNOWLEDGE-FORGE-DESIGN.md`. Read that first.

This addendum revises the design based on three corrections:

1. **Dedup/recall is largely already built.** Phase 2 is a hardening pass, not new
   construction. The roadmap shifts accordingly.
2. **Every pipeline step gets a configurable execution tier** — four of them, and the
   real engineering goes into the no-AI tier.
3. **The collected data is a fine-tuning corpus**, for the user's own models and for
   an org's. This is promoted from a side effect to a first-class output.

Plus the weekly checker, which is now the primary surface of the no-AI tier.

---

## Table of contents

- [A. The engine tier system](#a-the-engine-tier-system)
- [B. The no-AI engine (the actual product)](#b-the-no-ai-engine-the-actual-product)
- [C. The weekly checker](#c-the-weekly-checker)
- [D. The fine-tuning flywheel](#d-the-fine-tuning-flywheel)
- [E. Revised config](#e-revised-config)
- [F. Revised roadmap](#f-revised-roadmap)
- [G. Revised CV framing](#g-revised-cv-framing)
- [H. Claude Code prompts for the new phases](#h-claude-code-prompts-for-the-new-phases)

**Assumptions I'm making explicit** (correct me if wrong):

- *"log back … if they use Karpathy setup they have codebase access"* → I've read this
  as: when the vault sits alongside a repo, the tool indexes the code, writes
  code↔note links **both directions**, and detects when code drift invalidates a note.
  That's §B.5–B.7.
- *"fine tuning … for also their models too"* → datasets are exportable and owned by
  the user/org, usable to fine-tune their own models — not only yours. §D.

---

## A. The engine tier system

### A.1 The four tiers

| Tier | Key | What executes | Marginal cost | Deterministic? | Offline? |
|---|---|---|---|---|---|
| **T0** | `none` | Local scripts only — parsers, ripgrep, git, tree-sitter | $0 | yes | yes |
| **T1** | `host` | The Claude Code session model the user is already running | included in their plan | no | no |
| **T2** | `api` | BYO key: Anthropic / OpenAI / OpenRouter / **local Ollama** | their tokens | no | yes if Ollama |
| **T3** | `advisor` | Opus 5 / Fable 5 as a **second-pass critic** over T1/T2 output | premium, budgeted | no | no |

### A.2 The one non-obvious design decision

**T3 is not a bigger T1. It's a different job.**

The naive design is a quality dial: cheap model → good model → best model, same task.
That's wasteful — you pay Opus prices to do the boring 80% (parsing the question,
filling a template) that a small model handles fine.

The right design is **generate cheap, critique expensive**:

```
  T1/T2 drafts the note  ──▶  T3 advisor receives:
                                  • the draft
                                  • the sources
                                  • the schema + writing rules
                              and returns ONLY:
                                  • claims it disputes, with reasoning
                                  • what's missing (esp. "when NOT to use")
                                  • a confidence verdict
                                  • a patch, not a rewrite
```

Three things follow from this that are worth stating:

1. **Cost drops ~5–8×** versus running the advisor over the whole pipeline, because
   critique tokens are a fraction of generation tokens.
2. **The advisor's output is a supervision signal.** Draft + critique + accepted patch
   is exactly a preference pair. This is what makes §D possible at all — you get the
   training data for free as a byproduct of using the product.
3. **It degrades cleanly.** No advisor budget left? The draft still ships with
   `confidence: medium` instead of `high`. Nothing breaks.

### A.3 Per-stage configuration

Each of the nine pipeline stages declares an engine. Defaults:

| # | Stage | Default | Allowed | Lock |
|---|---|---|---|---|
| 0 | intake | `host` | T0–T3 | — |
| 1 | **recall** | `none` | **T0 only** | 🔒 hard |
| 2 | plan | `host` | T0–T3 | — |
| 3 | research | `api` → fallback `host` | T1–T3 | ⚠️ needs a model |
| 4 | synthesize | `host` | T1–T3 | ⚠️ needs a model |
| 5 | verify | `advisor` → fallback `host` | T0–T3 | — (T0 = mechanical gates only) |
| 6 | **write** | `none` | **T0 only** | 🔒 hard |
| 7 | link | `none` | T0–T2 | — |
| 8 | **index** | `none` | **T0 only** | 🔒 hard |

**Why hard-lock recall, write, and index.** These must be reproducible and cheap.
A model-driven recall returns different candidates for the same question on different
days, which destroys caching, makes the dedup guarantee unprovable, and makes bugs
unreproducible. Same logic as not putting an LLM in your database index. If a user
config sets `recall: {engine: host}`, **refuse to start and explain why** — an
opinionated refusal here is a feature, and it's the kind of thing a reviewer notices.

Stage 7 (`link`) is a deliberate middle case: deterministic linking by tag/stack
overlap covers most of it, but T1+ can catch conceptual relationships that share no
vocabulary. Default T0, upgrade optional.

### A.4 Fallback, budget, and routing

```yaml
engines:
  budget:
    advisor_usd_per_day: 2.00
    api_usd_per_day: 1.00
    on_exhausted: degrade        # degrade | queue | stop
  routing:
    advisor_when:                # when T3 is worth it
      - type: [decision, api, incident]      # high-stakes note types
      - confidence_below: medium
      - stack_in: [security, auth, payments]
      - user_flag: --deep
```

- `degrade` — fall to the next tier down, mark the note's `engine_trail` in frontmatter.
- `queue` — write to `_inbox/` with `pending_advisor: true`; the weekly checker
  drains the queue when budget resets. **This is the good default for teams.**
- `fail` — for CI/eval runs where silent degradation would corrupt results.

Add `engine_trail: [host, advisor]` to the note frontmatter. It makes cost auditable,
and it's a required column in the §D dataset.

---

## B. The no-AI engine (the actual product)

> This is where I'd spend the engineering time, and it's the part that's genuinely
> defensible. Everyone can write a prompt that calls a model. Almost nobody ships the
> static-analysis layer underneath.

### B.1 Why the no-AI tier is strategically the most important

Four reasons, in order of how much they'd matter to an outside evaluator:

1. **It's the only part that can make hard guarantees.** "No duplicate notes,"
   "every note validates," "this note references code that no longer exists" — these
   are provable claims. Nothing model-driven is provable.
2. **It runs where models can't.** Airgapped enterprises, CI pipelines, pre-commit
   hooks, offline. That's a real segment for the B2B version and it's a segment
   AI-wrapper competitors structurally cannot serve.
3. **It's $0 and it's fast**, so it can run constantly — on every commit, every
   session start — rather than occasionally.
4. **It generates the ground truth the AI tiers are graded against.** The evals in
   §16.3 of the main doc are T0 code. So is every metric in `docs/RESULTS.md`.

Framing to use in the README: **the AI tiers are optional accelerators on top of a
working static-analysis tool.** Not "AI tool with an offline mode."

### B.2 Honest capability boundary

| Capability | T0 | Notes |
|---|:--:|---|
| Search & rank vault by relevance | ✅ | lexical scoring, §8 main doc |
| Detect duplicates / near-duplicates | ✅ | shingling + MinHash on note bodies |
| Validate schema, dates, links | ✅ | |
| Build index, MOCs, tag pages, graphs | ✅ | |
| Detect stale notes | ✅ | date arithmetic |
| Detect **dead source URLs** | ✅ | HTTP HEAD, no model |
| Detect **code drift invalidating notes** | ✅ | git + AST — §B.6, the killer feature |
| Index a codebase: symbols, deps, ownership, coupling | ✅ | tree-sitter + git |
| Generate wiki scaffolding & stub notes | ✅ | from javadoc/docstrings + structure |
| Produce all reports & metrics | ✅ | §B.4 |
| Route/classify a question by topic | ⚠️ | keyword + taxonomy; ~good enough, not great |
| Summarize prose | ❌ | extractive only (TextRank), and it reads like it |
| Write explanatory prose | ❌ | |
| Research the open web | ❌ | can *fetch* known doc URLs, can't decide what to read |
| Verify a semantic claim | ❌ | can verify *mechanical* claims: does it compile, does the version exist |

Put this exact table in the docs. Being upfront about the ❌ column is what makes the
✅ column believable.

### B.3 The static core

> **Rev 2:** implemented in **Go** as one static binary (`bin/forge`) — see
> `KNOWLEDGE-FORGE-STACK.md` (ADR-001) for the full rationale, library choices,
> and the cgo/cross-compilation plan.

```
cmd/forge/            CLI
pkg/vault/            frontmatter + markdown AST (goldmark), mtime-cached
pkg/similarity/       MinHash + LSH banding, hand-rolled, no embeddings
pkg/graph/            note link graph: components, hubs, orphans, centrality
pkg/codeindex/        go-tree-sitter → symbols; build files → dep graph  (cgo isolated here)
pkg/gitsig/           go-git: churn, blame-ownership, co-change coupling
pkg/drift/            code↔note reference integrity  ← §B.6
pkg/linkcheck/        source URL liveness (HEAD, cached, rate-limited)
pkg/report/           renders any analysis to markdown
pkg/store/            SQLite (modernc.org/sqlite, pure Go) — derived cache only;
                      `forge reindex` fully rebuilds it from the markdown
```

No network required except `linkcheck`. Warm full-vault analysis <2s; `forge drift`
on the hook path <100ms (the binding constraint — STACK §1).

### B.4 The report suite

Every report is a generated markdown file in `vault/reports/`, so it's readable in
Obsidian, diffable in git, and greppable by the agent later.

| Report | Answers | Signal it produces |
|---|---|---|
| `coverage.md` | which stacks/topics have notes, which don't | where the wiki is thin |
| `staleness.md` | notes past `freshness_days`, ranked by ask-frequency | what to refresh **first** |
| `duplicates.md` | note pairs >0.85 similar | merge candidates |
| `orphans.md` | zero inbound links | invisible knowledge |
| `gaps.md` | asked ≥2×, never written | your personal curriculum |
| `graph-health.md` | components, hubs, isolated clusters | is the wiki one thing or fifteen |
| `churn.md` | most-updated notes | volatile knowledge = unstable areas of the system |
| `deadlinks.md` | source URLs returning 4xx/5xx | citations that have rotted |
| `drift.md` | **notes whose code references moved or changed** | notes that are now lying |
| `cost.md` | tokens/$ per stage per tier, from `engine_trail` | where the money goes |

Ranking is the part that makes these useful rather than noise: `staleness.md` sorted
by *ask frequency × days overdue* tells you the one note to fix today. A flat list of
80 stale notes tells you nothing.

### B.5 Codebase indexing without a model

When a repo is available (the Karpathy-style setup where the vault sits next to code):

```
tree-sitter parse
   ├── symbol table: classes, methods, annotations, exported functions
   ├── import graph  → which modules depend on which
   └── framework fingerprints:
         pom.xml / build.gradle   → Spring Boot version, starters, Kafka, JPA…
         package.json             → framework + versions
         Dockerfile / compose     → runtime topology

git
   ├── churn:      files changed most in last N months
   ├── ownership:  blame aggregation per file/package
   └── coupling:   files that change together (co-change matrix)
```

Everything above is deterministic and takes seconds. It produces
`.forge/code-index-<repo>.json` — **one file per configured `--repo name=path`, not one
shared file.** `forge drift`/`check`/`logback` all take `--repo` repeatably, so a single
name would let the second repo's index overwrite the first's on the very next run.

**The map:** join code → notes with pure matching — a note with `stack: [kafka]` maps
to files importing `org.apache.kafka`; a note mentioning `OrderConsumer` maps to the
file defining that symbol. No embeddings, no model. Emit `vault/moc/codebase.md`:

```markdown
# Codebase map — order-service

## payments  (14 files · churn: high · owners: @you, @dmitri)
Depends on: messaging, common
Notes: [[idempotent-payment-processing]] · [[stripe-webhook-retries]]
⚠️ No note covers: `RefundOrchestrator` (312 LOC, changed 9× in 90d, 0 notes)
```

That last line is the whole B2B pitch, generated by a `for` loop: **high-churn,
high-complexity, zero-documentation code, ranked.** No AI needed to find it, and it's
the report an engineering manager will actually pay for.

### B.6 Drift detection — the killer feature

The main doc's `forge-codebase-scout` writes notes containing `OrderConsumer.java:47`.
Those references rot within weeks. Every wiki in existence has this problem and nobody
solves it, because solving it requires exactly the static index above.

```
for each note:
  for each code reference (file:line, symbol, config key, version):
     ├── file gone?                      → BROKEN
     ├── symbol renamed/removed?         → BROKEN  (AST diff, not line diff)
     ├── line moved but symbol intact?   → auto-repair the line number, silent
     ├── enclosing function body changed since note.verified?  → SUSPECT
     └── declared dep version bumped?    → SUSPECT (note may describe old behaviour)
```

- **BROKEN** → note demoted to `confidence: low`, listed in `drift.md`, queued for the
  weekly checker.
- **SUSPECT** → flagged, `verified` date pinned, surfaced with the specific diff.
- **auto-repair** → line numbers fixed silently; this alone eliminates most rot.

**Git-tree anchored, not file-watch anchored.** Drift verdicts are computed against
commits, and this is a hard rule:

- Each note's drift state records the commit it was evaluated at:
  `drift_checked_at: <sha>`. The next run diffs `<sha>..HEAD` (`git diff --name-only`
  first as a cheap gate, AST comparison only on files in that set) — never a full
  rescan, never triggered by uncommitted editor noise.
- The check runs **only when the codebase actually changed**: post-commit /
  post-merge / post-checkout on the repo, not on every file save. Uncommitted working
  tree changes never demote a note — half-finished edits aren't drift.
- **Rollbacks reverse verdicts symmetrically.** Verdicts are a pure function of
  (note refs, tree state), so on `revert`/`reset`/branch switch, a symbol that
  reappears in the new HEAD automatically restores the note: `confidence` back to its
  pre-demotion value, `drift.md` entry cleared, with a log line citing both SHAs.
  Demotion history lives in `.forge/` (keyed by note slug + sha), NOT in note body
  churn — otherwise every rebase writes garbage into the vault's git history.
- Branch-aware: verdicts are per-HEAD. A note BROKEN on a feature branch is not
  BROKEN on `main`; the weekly checker reports against the default branch only.

A wiki that tells you which of its own pages have gone stale — with evidence, and
that heals itself when the code change is rolled back — is a materially different
artifact from a folder of markdown.

### B.7 Log-back into the codebase

Bidirectional, so knowledge is discoverable from the code, not just from the vault.

**Safe (default on):**
- `docs/knowledge-map.md` in the repo — module → relevant notes, generated.
- A `CLAUDE.md` fragment per module: *"Relevant notes: [[…]]"* — so **any** agent
  session in that repo gets the vault's knowledge in context without the plugin.
- `.forge/code-index-<repo>.json` for other tools (one per `--repo`; see §B.6).

**Opt-in (default off):**
- Inline `// forge: [[note-slug]]` markers above the symbols a note documents, in a
  managed block with begin/end sentinels so regeneration is idempotent and revertible.
  Off by default — many teams won't accept generated comments in source, and shipping
  it on by default is how you get a bad first issue.

**Never:** modifying code semantics. Comments and separate files only.

---

## C. The weekly checker

`/forge-check` — runs at **T0 by default**, which means it's free, offline, fast, and
can be a cron job or a CI job without anyone worrying about it.

**T0 pass (always):** run all ten reports of §B.4 → `vault/reports/` → write
`moc/weekly/YYYY-WW.md` combining them, ranked by actionability:

```markdown
# Week 32, 2026

## 🔴 Act now (3)
- [[spring-security-6-config]] — BROKEN: `SecurityConfig.filterChain` removed in
  commit a3f9c21 (6d ago). Note describes an API that no longer exists.
- `RefundOrchestrator` — 312 LOC, 9 changes/90d, 0 notes, asked about 4×.
- 2 dead source URLs in [[kafka-exactly-once]].

## 🟡 Review (7)
- 4 notes stale >180d · 2 merge candidates · 1 orphan

## 📊 Vault
312 notes (+7) · hit-rate 44% (+3pt) · orphans 4 (−2) · drift 3 (+3)

## 🎯 Gaps (asked, never written)
1. "ECS task draining" (5×)   2. "Testcontainers + Kafka" (3×)
```

**Optional AI pass** (`check.ai_pass: true`) does only what T0 can't:
- draft the refresh for the top-ranked BROKEN note
- propose merge text for duplicate pairs
- write the ADR stub for an undocumented high-churn module

Every mutation shows a diff and requires approval. **Never auto-mutate the vault on a
schedule** — that's how you get an issue titled "forge ate my notes."

**Budget queue drain.** If `on_exhausted: queue`, this is where `pending_advisor`
notes get their T3 pass. Nice property: expensive work batches into one scheduled run
instead of blocking interactive sessions.

---

## D. The fine-tuning flywheel

The system produces labelled training data as a byproduct of normal use. Six datasets,
all exportable, all owned by the user.

### D.1 The datasets

| # | Dataset | Pair | Source | Format | Realistic volume (1 dev, 3mo) |
|---|---|---|---|---|---|
| D1 | **Routing** | question → `ANSWER`/`UPDATE`/`CREATE` + topic + stack | every run, auto-labelled by recall + outcome | classification | 300–800 |
| D2 | **Advisor distillation** | draft → advisor critique → accepted patch | T3 runs | DPO / SFT | 50–200 |
| D3 | **Human correction** | model note → your edited note | git diff of notes after write | DPO ★ | 100–400 |
| D4 | **Gate repair** | failing draft + gate error → fixed draft | verification stage | SFT | 100–300 |
| D5 | **Style** | (question, profile, sources) → accepted note | successful runs | SFT | 200–500 |
| D6 | **Code↔knowledge** | repo symbol/module → the note explaining it | the §B.5 map | retrieval / RAG eval | = note count |

★ **D3 is the highest-value dataset in the list and it's nearly free.** Every time you
hand-edit a generated note, git captures a (model output, human-preferred output) pair.
That is the exact shape of a preference dataset, and it's genuinely scarce — it can't
be scraped, bought, or synthesized. Capturing it is a `post-commit` hook on the vault
repo. Build this in Phase 1, before you need it, because the data only accumulates
forward.

### D.2 Being honest about what's trainable at what volume

This is where these pitches usually overclaim. The credible version:

| Volume | What's realistically achievable |
|---|---|
| **100–500 pairs** | LoRA on a 1–3B model for **D1 routing**. Narrow classification, small label space — this genuinely works and replaces a T1 call with a local 20ms one. Also a style adapter (D5) that mimics your note voice. |
| **1k–5k pairs** (1 dev, ~1 year, or a 6-dev team in a quarter) | LoRA on a 7–8B model for note *drafting* in your stack and voice. Quality comparable to a mid-tier API model **for this narrow task only**. |
| **10k+ pairs** (org-scale) | DPO on D2+D3 to distill advisor judgment. This is where "our model reviews docs like our staff engineer does" becomes real. |
| **Any volume** | **Evaluation sets.** D1 and D6 make excellent eval benchmarks long before they're enough to train on. |

So the honest sequencing is: **eval sets first → routing classifier → style adapter →
drafting model → advisor distillation.** Anyone technical who reads the plan will
trust it more for saying so, and it stops you from burning a month fine-tuning on 200
examples and concluding fine-tuning doesn't work.

### D.3 The distillation loop

```
   month 1-3          run T3 advisor on high-stakes notes
        │             collect D2 (critiques) + D3 (your edits)
        ▼
   month 3-4          export → LoRA a small local model on D2+D3
        │
        ▼
   month 4+           local model becomes the `verify` engine at T0.5:
                      runs free, offline, on EVERY note
                      T3 advisor reserved for what the local model flags uncertain
        │
        └──────────▶  advisor still corrects the local model → more D2 → retrain
```

Result: verification coverage goes from "the 15% of notes that justified Opus" to
100%, while advisor spend drops. That's a real systems result with a number attached,
and it's the strongest single item this project can put on a CV.

### D.4 Export

`/forge-export-dataset --set d3 --format dpo --since 2026-05-01 --anonymize`

- Formats: `jsonl-sft` (messages), `jsonl-dpo` (chosen/rejected), `csv` (classification).
- `--anonymize`: strips repo names, paths, identifiers, internal URLs via the same
  scrubber as `examples/vault/`.
- Emits a **datasheet** alongside: counts, date range, engine trail distribution,
  known biases (e.g. "83% Java/Spring — do not expect generalization"). Shipping a
  datasheet is a small thing that signals you've read the ML literature.
- Every export is logged. Nothing leaves the machine without an explicit command.

### D.5 Ownership and privacy

Non-negotiable, and state it in the README in exactly these terms:

- Datasets live in the user's vault, in plain JSONL. They are the user's property.
- Nothing is transmitted anywhere. There is no phone-home. Export is manual.
- Raw question text is never stored — topic + hash only (§14 main doc). Note bodies
  are stored, because they're the user's own notes.
- Org mode (Appendix A) requires explicit per-repo opt-in and is self-hostable.

For B2B this inverts into the actual sales argument: *"the dataset is yours, generated
by your engineers, on your code, and you can fine-tune your own model with it. We
don't touch it."* That's a much easier procurement conversation than any pitch that
involves sending code to a vendor.

---

## E. Revised config

Replaces §10 of the main doc.

```yaml
---
vault_path: ~/Obsidian/second-brain
repo_path: auto              # auto-detect git root of cwd

# ── engines ────────────────────────────────────────────────
engines:
  default: host
  api:
    provider: anthropic      # anthropic | openai | openrouter | ollama
    model: claude-haiku-4-5
    key_env: ANTHROPIC_API_KEY
    base_url: null           # set for ollama/self-hosted
  advisor:
    model: claude-opus-5     # or claude-fable-5
    mode: critique           # critique | rewrite   (critique strongly preferred)
  local:                     # the distilled model from §D.3, once it exists
    enabled: false
    model: ~/.forge/models/forge-verify-lora
  budget:
    advisor_usd_per_day: 2.00
    api_usd_per_day: 1.00
    on_exhausted: queue      # degrade | queue | stop
  routing:
    advisor_when:
      - type: [decision, api, incident]
      - confidence_below: medium
      - user_flag: --deep

pipeline:
  intake:     { engine: host }
  recall:     { engine: none }        # 🔒 locked
  plan:       { engine: host }
  research:   { engine: api, fallback: host }
  synthesize: { engine: host }
  verify:     { engine: advisor, fallback: local, then: host }
  write:      { engine: none }        # 🔒 locked
  link:       { engine: none }
  index:      { engine: none }        # 🔒 locked

# ── no-AI core ─────────────────────────────────────────────
static:
  code_index: true
  languages: [java, kotlin, python, typescript]
  git_signals: true
  drift:
    enabled: true
    trigger: git                 # git = post-commit/merge/checkout only, never on save
    branch: default              # verdicts reported against default branch
    auto_repair_line_numbers: true
    on_broken: demote            # demote | flag_only
    on_restored: undemote        # rollback/revert restores prior confidence
  linkcheck:
    enabled: true
    timeout_s: 5
  logback:
    knowledge_map: true          # docs/knowledge-map.md
    claude_md_fragment: true
    inline_markers: false        # opt-in only

check:                            # weekly checker
  enabled: true
  schedule: "0 9 * * MON"
  ai_pass: false                  # T0-only by default
  reports: [coverage, staleness, duplicates, orphans, gaps,
            graph-health, churn, deadlinks, drift, cost]
  drain_advisor_queue: true

dataset:
  enabled: true
  capture: [d1_routing, d2_advisor, d3_human_edits, d4_gate_repair, d5_style]
  anonymize_on_export: true
---
```

**Preset shortcuts** so nobody has to read the above on day one:

| Preset | Meaning |
|---|---|
| `offline` | everything T0. Reports, drift, index, search. No model, ever. |
| `claude-only` | T0 + T1. No API key needed. **Default for Claude Code users.** |
| `byo-api` | T0 + T2 for research, T1 for synthesis. Cheapest per note. |
| `max` | T0 + T1 + T3 advisor on high-stakes notes, queue on budget exhaustion. |

---

## F. Revised roadmap

Changes from the main doc's §15:

| Phase | Change |
|---|---|
| 0 — Audit | + inventory which engine each existing step implicitly uses; + capability audit of existing recall |
| 1 — Contract | **unchanged, still first.** Add the D3 capture hook here — data only accumulates forward. |
| 2 — Recall | **shrinks to a hardening pass** (you have most of it): add `--explain`, the mtime cache, the ≥0.85 lock, and eval coverage |
| **2b — Static core** ⭐ NEW | `scripts/core/*`, the ten reports, code index, drift detection. **The biggest new chunk of work and the most defensible.** |
| 3 — Config | + the engine block, presets, the hard-lock refusal |
| **3b — Engine abstraction** ⭐ NEW | one interface, four backends, budget accounting, `engine_trail` |
| 4 — Subagents | + advisor as critique-mode second pass |
| 5 — Hooks | → **weekly checker** becomes the centrepiece; drift hook on Edit/Write |
| **5b — Log-back** ⭐ NEW | knowledge-map, CLAUDE.md fragments, code-index-&lt;repo&gt;.json |
| 6 — Package | + document the tier matrix and the honest capability table |
| **6b — Datasets** ⭐ NEW | capture + export + datasheets |
| 7 — Real use | + first LoRA on D1 routing once ~300 pairs exist |

Revised estimate: **~4 weeks of focused evenings** (2b is real work, and the Go
ramp + release toolchain adds 1–1.5 weeks — STACK §9).

**If you cut anything, cut in this order:** 6b datasets (capture is cheap, export can
wait) → 5b log-back → advisor tier. **Never cut 2b.** The static core is the part that
makes this an engineering project rather than a prompt collection.

---

## G. Revised CV framing

The story is materially stronger now. Lead with this ordering:

**Bullet:**

> **Knowledge Forge** — open-source Claude Code plugin giving AI agents persistent,
> self-verifying technical memory. Built a deterministic static-analysis core
> (tree-sitter + git) that indexes a codebase, detects when code changes invalidate
> documentation, and generates ranked doc-debt reports — with an optional four-tier
> LLM layer (offline / host / API / advisor-critique) configurable per pipeline stage.
> Advisor corrections and human edits are captured as preference datasets, used to
> distill a local verification model.

**The four talking points, revised:**

1. **"The core doesn't use AI at all."** Drift detection, dedup, the report suite,
   the code index — pure static analysis. The LLM tiers are optional accelerators.
   This is the answer to "isn't this just a prompt?" and it ends that conversation.
2. **"Generate cheap, critique expensive."** The advisor is a second-pass critic, not
   a bigger generator — ~5–8× cheaper for most of the quality, and its corrections are
   the training signal. Cost-aware architecture is a senior signal.
3. **"I hard-locked the stages that must be deterministic."** Recall, write, and index
   refuse to run on a model. Knowing where *not* to put the LLM is the judgment call
   most people in this space get wrong.
4. **"The product generates its own training data."** Advisor critiques and human
   edits are preference pairs; the loop distills a premium model into a local one that
   runs on every note instead of 15% of them. And the honest volume table — knowing
   300 pairs trains a router, not a coding model.

Point 1 plus point 3 is a Java/Spring backend engineer's story, not an AI-hobbyist
story: static analysis, determinism guarantees, tiered execution with budget and
graceful degradation, idempotent regeneration. Same instincts as running a production
service. That's the framing.

---

## H. Claude Code prompts for the new phases

Insert into `CLAUDE-CODE-PROMPT.md` at the indicated positions.

### Phase 0 addendum — append to the Phase 0 prompt

```
Additionally, for the engine/tier work:

5. For each of the 9 pipeline stages, state what currently executes it: a script, an
   inline model instruction, a subagent, or nothing. Be specific about where model
   judgment is used for something a script could do deterministically — those are
   the T0 conversion candidates.

6. Assess my existing dedup/recall honestly against section 8 of the design doc and
   section A.3 of the addendum. I believe most of it exists. Tell me specifically:
   is scoring deterministic and reproducible? Is there a cache? Does it run BEFORE
   research, always, with no bypass path? Does anything model-driven leak into it?
   List only the deltas needed — do not propose rebuilding what works.

7. Is a code repository available alongside the vault? If so: languages, build
   system, size, and whether any note already references file paths or symbols
   (grep for file:line patterns and class names). Report how many such references
   exist and how many are currently BROKEN — that's the Phase 2b baseline.
```

### Phase 2b — Static analysis core ⭐ (new; runs after Phase 2)

> **Superseded.** The Phase 2b prompt now lives in `CLAUDE-CODE-PROMPT.md`
> (Go implementation per ADR-001; original spec in `KNOWLEDGE-FORGE-STACK.md` §10).
> Key deltas from the Python draft that used to sit here: single static Go binary in
> `bin/`, SQLite (pure-Go driver) instead of `state.json`, cgo isolated to the
> tree-sitter package, and git-anchored drift with symmetric rollback per §B.6.

### Phase 3b — Engine abstraction ⭐ (new; runs after Phase 3)

```
Read KNOWLEDGE-FORGE-ADDENDUM.md sections A and E.

1. Define ONE engine interface that all four tiers implement:
      run(stage, prompt, context, constraints) -> {output, tokens, cost_usd, tier}
   Backends: none (raises for stages that need generation), host (session model),
   api (anthropic/openai/openrouter/ollama via config), advisor (critique mode).

2. Per-stage engine selection from config, with fallback chains
   (e.g. verify: advisor -> local -> host).

3. HARD LOCKS: recall, write, and index accept ONLY engine `none`. If config sets
   anything else, refuse to start with a clear error explaining that these stages
   must be reproducible and why. Do not silently override — refuse.

4. Advisor in CRITIQUE mode, not rewrite. It receives draft + sources + schema +
   writing rules, and returns ONLY: disputed claims with reasoning, what's missing,
   a confidence verdict, and a minimal patch. It must not regenerate the note.
   Log the critique verbatim — it's dataset D2.

5. Budget accounting: per-day USD caps per tier, persisted in .forge/state.json.
   on_exhausted: degrade | queue | stop. `queue` writes pending_advisor:true to the
   note and the weekly checker drains it.

6. Add engine_trail: [host, advisor] to note frontmatter (update the schema and the
   validator). Generate reports/cost.md from it.

7. Implement the four presets from addendum section E: offline, claude-only,
   byo-api, max.

Prove it: run the same question under all four presets. Show me the four notes, the
cost of each, and confirm `offline` produced a usable (if unwritten) result — it
should do recall, report the match, and tell me clearly what it cannot do rather
than failing.
```

### Phase 5 revision — the weekly checker (replaces the gardener section)

```
Read KNOWLEDGE-FORGE-ADDENDUM.md section C.

Build /forge-check as a T0-ONLY run by default:
- runs all 10 reports from Phase 2b
- writes moc/weekly/YYYY-WW.md ranked by ACTIONABILITY, using the exact structure
  in addendum C: "Act now" (broken drift, dead links, undocumented high-churn code),
  "Review", "Vault stats with week-over-week deltas", "Gaps"
- must run with zero model calls, zero network except linkcheck, in under 10s
- safe to run from cron or CI

Optional AI pass behind check.ai_pass (default false), doing ONLY what T0 cannot:
draft the refresh for the top BROKEN note, propose merge text for duplicates, stub
an ADR for the top undocumented module. Every mutation shows a diff and needs my
approval. Never auto-mutate the vault on a schedule.

If engines.budget.on_exhausted is `queue`, drain pending_advisor notes here.

Drift triggering is GIT-ANCHORED, not save-anchored: install post-commit /
post-merge / post-checkout hooks on the repo that run `forge drift --since-commit
<last-checked-sha>` — diff-gated, incremental, under 100ms, fail silent. Never
demote on uncommitted working-tree changes. On revert/reset, verdicts must reverse
symmetrically (confidence restored, drift.md cleared) per addendum B.6.

Run /forge-check on my vault now and show me the week file.
```

### Phase 6b — Dataset capture & export ⭐ (new)

```
Read KNOWLEDGE-FORGE-ADDENDUM.md section D.

1. Capture, into .forge/datasets/*.jsonl, during normal operation:
   D1 routing        — question hash + features -> decision + outcome
   D2 advisor        — draft, critique, accepted patch  (from critique-mode runs)
   D3 human edits    — post-commit hook on the vault git repo: any note edited by a
                       human within 7 days of generation becomes a (generated,
                       human-preferred) pair. THIS IS THE MOST VALUABLE ONE.
   D4 gate repair    — failing draft + gate error -> fixed draft
   D5 style          — (question, profile, sources) -> accepted note

   Never store raw question text — hash + extracted topic only.

2. /forge-export-dataset --set <d1..d5> --format jsonl-sft|jsonl-dpo|csv
   --since DATE --anonymize
   --anonymize strips repo names, absolute paths, internal URLs, identifiers, using
   the same scrubber as examples/vault/.

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

## Open questions for you

1. **Is the no-AI tier a standalone product?** I've designed it as the substrate of
   one tool. But `offline` mode — codebase indexer + doc-drift detector + ranked
   doc-debt reports, zero AI — is arguably a sellable product on its own, with a much
   easier enterprise sale. Worth deciding before you write the README, because it
   changes the positioning.
2. **Fine-tuning: yours, theirs, or both?** "for also their models too" reads as both.
   The design supports both, but if you ever want an aggregated cross-customer
   dataset, the consent model has to be built in from day one — retrofitting it is
   effectively impossible.
3. **Does the advisor tier need to work with non-Anthropic models?** Supporting
   OpenAI/local as advisor is ~a day of work now and ~a week later.
