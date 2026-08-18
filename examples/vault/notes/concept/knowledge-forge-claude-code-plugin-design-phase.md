---
title: "Knowledge Forge — Claude Code Plugin (design phase)"
slug: knowledge-forge-claude-code-plugin-design-phase
type: concept
stack: [go, claude-code, obsidian]
tags: [knowledge-management, static-analysis]
depth: 3
confidence: low
created: 2026-08-08
updated: 2026-08-14
verified: 2026-08-14
freshness_days: 365
sources: []
related: []
supersedes: []
forge_version: 2.0.0
origin: import
---

# Knowledge Forge — Claude Code Plugin

A Claude Code plugin that turns "explain X" moments into permanent, linked, verified
markdown notes in an Obsidian vault, so the second time a question comes up it's a vault
read instead of a research run. Intended replacement for the v1 `til-writer` skill.

## Project Path
`[REDACTED-PATH]` — module `knowledge-forge` (bare path, no VCS host prefix), Go 1.26.
The directory is still named `TIL`; only the module was renamed.

## Status (2026-08-14) — Phases 0–5b done, Phase 6 next
Phase 0 produced `docs/AUDIT.md`: the factual baseline plus a doc-vs-doc coherence pass
that found thirteen contradictions the docs never self-flag. Seven resolved under the
precedence rule; the six that could not are **§8.4, D-1 … D-8**, a binding decision record
that changes Phases 3, 3b, 6 and 6b. No design doc was edited — where §8.4 marks a line
stale, the doc still says the old thing and §8.4 is what wins.

Phase 1 shipped the note contract and the migration. Numbers, all measured:

| | |
|---|---|
| Notes migrated to the DESIGN §7 topology | **91** |
| Wikilinks rewritten | **345** |
| Links broken by the move | **0** |
| Schema-valid after migration | **60 / 91 (66%)** |

The 31 remaining failures all need human judgment, not a script: 28 `sources: uncited`,
9 `stack: missing`, 7 `tags: missing`, 2 `tags: too-few`, 1 `stack: too-many` (a note can
fail on more than one rule). Rollback if ever needed: backup at
`[REDACTED-PATH]`, or vault commit `b3168f0`.

Code now exists: `cmd/forge` (`slug`, `validate`, `index`, `reindex`, `capture`) over
`pkg/vault`, `pkg/graph`, `pkg/report`, `pkg/store`, `pkg/dataset`, plus seven note
templates. Repo commit `1c9df95` merges the Phase 1 branch to `main`.

### Phase 2 — `forge recall` (repo commit `3619b72`)

The dedup engine. It scores a question against every note in this vault with **zero model
calls**, in ~5 ms warm, and returns one verdict: answer from an existing note, extend it,
refresh it, or write a new one. `skills/forge/SKILL.md` runs it before any research —
that ordering is the whole point, because research done first anchors the judgement.

The phase brief called this "a hardening pass over existing code". It wasn't: `AUDIT.md`
§6 had already measured none of it existing, and the tree agreed. Built greenfield.

Measured against this vault, 91 notes, zero files created:

| Question | Verdict | Score |
|---|---|---|
| how does keyset pagination work | ANSWER_FROM_VAULT | 0.917 |
| explain the transactional outbox pattern | ANSWER_FROM_VAULT | 1.000 |
| what is hexagonal architecture | ANSWER_FROM_VAULT | 0.867 |
| storybook decorator for react router context | UPDATE(extend) | 0.708 |
| JpaSpecificationExecutor dynamic sorting | UPDATE(extend) | 0.604 |
| JPA entity graph to avoid N+1 | CREATE + 5 neighbours | 0.333 |

Two departures from the literal spec, both argued from this vault's behaviour in
`references/recall-spec.md`: the blend is a weighted **mean over active channels** rather
than a raw weighted sum (§2.5 — under a raw sum a perfect title match with no tag input
tops out at 0.49 and resolves CREATE, i.e. the dedup engine's headline case fails), and
the title measure is **F₂, not Dice** (§2.2 — Dice punishes a note whose title is *more*
specific than the question, which is the normal shape of a good note).

A channel only activates when the query supplied input **and** the note carries the
field. That is a direct consequence of the 31 notes above that still have no `tags:` —
scoring them zero ranked them below well-tagged irrelevant notes.

### Phase 2b — drift detection and the nine reports

The static core, finished. New packages: `pkg/similarity` (hand-rolled MinHash + LSH, no
embeddings), `pkg/codeindex` (go-tree-sitter, the only cgo package, build-tag gated),
`pkg/coderef`, `pkg/gitsig`, `pkg/drift`, `pkg/linkcheck`. New commands: `forge drift`,
`forge check`.

**This vault's `reports/` and `moc/codebase.md` are now generated output** — nine reports
written by `forge check` and `forge drift`, ranked rather than alphabetical (staleness
sorts by ask_frequency × days_overdue). Measured against this vault and my code repos:

| Report | Result |
|---|---|
| **drift** | **9 notes reference code that has already changed** — 2 broken, 7 suspect. Of 140 code citations: 42 ok, 11 suspect, 3 broken, 84 skipped (repo not on disk) |
| orphans | 21 of 94 (22%) |
| graph-health | 23 components, largest holds 71% |
| duplicates | 3 pairs ≥ 0.40, out of 1547 compared |
| coverage | 39 of 41 stacks have a note; `kotlin` and `postman` do not |
| staleness | 0 of 91 overdue |
| deadlinks | 0 broken of 0 external URLs — 63 citations are all first-party paths |

Latency, Apple M4, all inside budget: `forge index` **0.02 s** (budget 200 ms),
`forge drift --since-commit` **0.06–0.07 s** (budget 100 ms — the binding one, it runs on
the git-hook path), `forge check` **0.93 s cold / 0.39 s warm** (budget 10 s). Benchmarks:
parse 10038 ns/op at 520 MB/s, drift check 79769 ns/op, MinHash pairs over 500 notes ~90 ms,
Java parse 19495 ns/op, gitsig analyze 294857 ns/op.

Determinism was verified, not assumed: six consecutive `forge check` runs produced
md5-identical output for all ten files. Getting there took fixing three defective
`sort.Slice` comparators out of nineteen audited — a comparator whose last term is not
unique in the collection leaves ties unordered, and `drift.md` oscillated between 9 and 10
notes until that was fixed.

Two deviations, both recorded rather than hidden: `pkg/gitsig` shells out to the `git` CLI
instead of go-git (BACKLOG **B-009** — go-git's log walk was slower and its rename
detection weaker), and B-008's IDF weighting shipped without fixing B-008 (see Open Items).

Release plumbing: Makefile with a six-target cross-compile matrix, goreleaser config,
SHA-256 checksums, and a `bin/forge` shim that **verifies the pinned hash before exec**.
No hash is committed — a pin in git is stale the moment anyone rebuilds.

### Phase 3 — config chain and `forge init` (repo commit `847098a`)

`pkg/config`: a four-layer chain (`$FORGE_CONFIG` > `.forge.config.md` project >
`~/.forge/forge.config.md` home > packaged `config/forge.config.example.md`), tolerant
markdown-frontmatter parsing, list-**replace**/map-**merge** semantics, and a post-merge
check that `recall`/`write`/`index` resolve to `engine: none` — a project layer cannot
sneak a real engine past the T0 lock even if each layer looks valid alone. Eight packaged
presets in `config/presets/`.

New commands: `forge init` (the sole writer of `~/.forge/forge.config.md` and
`<vault>/profiles/me.md`, per `AUDIT.md` §8.4 **D-4**) and `forge config --layers/--json`.
`skills/forge-init/SKILL.md` is the conversational wrapper — asks at most five questions,
shells out, never writes either file itself. It installs the vault's D3 hook by the
**pinned** binary path (`$HOME/.forge/bin/forge`, `$FORGE_BIN` override), never a `PATH`
lookup — a lesson from review, since the hook is silent on failure by design.

**B-008's §3.1 recalibration was deliberately not attempted this phase.** The backlog
entry's own words: verifying the candidate fix honestly means re-deriving the whole
calibration table, not re-running two queries. It stays open, owned by its own session.

### Phase 3b — `pkg/engine` abstraction

One interface, four backends (`none`/`host`/`api`/`advisor`), routed per pipeline stage
with a fallback chain (`Stage.Fallback` → `Then`). `local` is **not** a fifth backend —
it's a routing alias resolving to `api` pointed at `engines.local.base_url`. Hard locks on
`recall`/`write`/`index` are enforced twice: once at config load (Phase 3, unchanged) and
again inside `engine.Select` itself, independent of config validation, as defense in depth
against a tampered config reaching the call site directly.

`forge engine` is now this binary's **one named exception** to zero model calls —
`select` (dry resolve, no spend, always exits 0 even offline), `run` (spends budget, calls
api/advisor via an injected `http.RoundTripper` — never real network in tests), `record`
(stamps `engine_trail` onto a note, used by the skill after a `host`-tier step it ran
itself in its own context, since the binary can't call that tier).

Budget is per-day USD caps, persisted in a new SQLite `budget` table — the **one**
documented exception to "SQLite is purely derived"; `forge reindex` rebuilds every other
table but leaves `budget` alone by simply never listing it in `Store.Reset()`.

`engine_trail` only stamps six of `cfg.Pipeline`'s nine stage names
(`recall|research|write|verify|critique|index`) — `intake/plan/synthesize/link` are host
orchestration bookkeeping, not audited model-call decisions, and widening the schema
pattern to cover them is left to **BACKLOG B-022** rather than done here.

`reports/cost.md` ships — the tenth report, deliberately absent from 2b's nine so its
counters would exist before the report tried to read them (AUDIT §8.4 D-1). Verified with
a mock `httptest.Server` standing in for `engines.api.base_url`: `offline`/`claude-only`
report $0.00, `byo-api`/`max` report nonzero spend against the mock's fixed price, `max`
additionally shows the advisor tier firing.

One wrinkle found, not fixed: the design docs say `on_exhausted: degrade | queue | fail`;
the code (already shipped in Phase 3) uses `stop`, not `fail`, and grepping every read site
this phase found that `stop` and `degrade` are behaviorally identical to the default
silent fallthrough — nothing branches on either value except `queue`. Recorded as
**BACKLOG B-023**, docs left stale per this project's own "record, don't fix" rule.

## This vault now has a git hook
`.git/hooks/post-commit` runs `forge capture` after every commit (installed 2026-08-09;
binary at `~/.forge/bin/forge`). It harvests ADDENDUM §D.1 **D3 pairs**: a note forge
generated that a human corrected within seven days becomes one (generated, preferred)
training example in `.forge/datasets/d3.jsonl` — gitignored, local, never transmitted.

Four things it refuses to capture: notes with `origin: import`, edits outside the seven-day
window, edits in the same commit that created the note, and commits carrying a
`Forge-Write:` trailer. It also refuses a pair whose edit predates its generation, which
clock skew can produce.

**It captures nothing today, by design.** All 91 migrated notes carry `origin: import`, so
nothing in this vault is model output. The dataset only accumulates forward, from Phase 4
on — which is exactly why the hook had to be installed now rather than in Phase 6b where
the export lives. Removing it is `rm .git/hooks/post-commit`; it can never fail a commit
and never prints to the terminal (see `.forge/capture.log`).

### Phase 4 — subagents & verification (repo commit `884e42e` + Phase 4)
`pkg/qualitygate` ships the seven DESIGN §12 gates (schema, citation, code, freshness,
antislop, link, duplicate) behind `Run(cfg, vaultRoot, draft, mode) (Report, error)` —
one file per gate, each under 20 lines via helper delegation. Each gate returns a
`Verdict` (Pass/Fail/Skipped) plus a `Remedy`; only `RetryOnce`, `MarkUnverified`,
`DropConfidence`, `RewritePass` block the write (`Report.Quarantine = true`) —
`DelegateToLibrarian` (link) and `SwitchToUpdate` (duplicate) are routing decisions the
skill acts on post-write, not defects. `forge verify-code --lang <java|ts|bash>` shells
out to the system toolchain (`javac`/`tsc`/`bash -n`) in a throwaway temp dir — never a
dependency resolver, no network, no `npm install`/Maven. Its three-valued outcome is the
load-bearing decision: a snippet whose only diagnostics are unresolved imports is
`skipped` ("syntax-checked only, not compiled against its classpath"), not `fail` —  but
a real syntax error co-occurring with an unresolved import still `fail`s, because any
parse/syntax diagnostic dominates regardless of what else is in the same compiler run.

`forge gate` is the deterministic CLI entry point (not an agent) that runs all seven
gates and, on `Quarantine=true`, writes the draft to `_inbox/` with `confidence: low` and
a `## Open questions` bullet per failed gate, in fixed gate order (schema → citation →
code → freshness → antislop → link → duplicate) so two runs on unchanged state produce
byte-identical output (pinned as the B-020 determinism convention). CREATE and UPDATE
diverge on failure: CREATE writes straight to `_inbox/`; UPDATE never touches the
already-published, already-linked note — it writes the proposed edit to `_inbox/` instead
with a `supersedes` back-pointer to the target slug, so a failing edit is demoted, never
silently dropped and never allowed to demote a trusted note.

Four product subagents are specced as `agents/*.md` (`forge-researcher`,
`forge-codebase-scout`, `forge-verifier`, `forge-librarian`) with `.claude/agents/`-style
frontmatter, tool allowlists, and call-budget limits — but **not live today**: nothing in
this repo loads agents from a root-level `agents/` directory (Claude Code loads
`.claude/agents/`, and no plugin manifest exists yet). `skills/forge/SKILL.md` was
rewired to dispatch researcher+scout in parallel, run `forge gate` before write, and
dispatch `forge-librarian` after a passing gate — verified today through the generic
Agent tool with an explicit tool allowlist, not live agent auto-discovery.

B-007 closed this phase both ways: `pkg/dataset/d3_forge_write_test.go` pins that a
librarian-authored commit (carrying the `Forge-Write: true` trailer) yields zero D3
pairs and an otherwise-identical commit without the trailer yields exactly one; and
`agents/forge-librarian.md`'s prompt now literally instructs
`git commit --trailer "Forge-Write: true"`. B-022 also closed:
`references/schema.yaml`'s `engine_trail` pattern now covers all nine `cfg.Pipeline`
stages minus `critique` (which was never a pipeline key). A new item was found and
recorded, not fixed: **B-024** — `pkg/dataset/d2.go`'s `D2Tag = "d2_advisor"` never
matches the packaged config's `dataset.capture: [..., d2, ...]`, so D2 capture is
silently inert under the shipped config; D4's own tag was spelled to avoid repeating it.

Measured, Apple M4: the six in-process gates minus `code` run in **~0.13ms** combined —
`forge verify-code` is dominated by toolchain startup instead, bash **~10ms warm**, java
**~170ms warm** (`tsc` not installed in this environment, untested here).

### Phase 5 — hooks & weekly checker (done)

Goal: turn the static core from "a command you remember to run" into a system — a
SessionStart hook, a cheap intent check before a question gets re-explained from scratch,
a weekly T0-only rollup, and `/forge-stats`. Plan at
`~/.claude/plans/so-phae-4-done-soft-coral.md`, all 15 tracked sub-tasks done:

- `pkg/telemetry` (event, sha256 question hash — never raw text, JSONL writer), gated
  fully behind `cfg.Telemetry.Enabled`, wired into `forge recall`.
- The ask-capture log now feeds `staleness.md` and `gaps.md`'s ranking (`Asks` was a
  hardcoded nil since 2b, waiting on exactly this).
- `pkg/report/weekly.go` (ADDENDUM §C's rollup, literal section headers) and a new
  `.forge/weekly-stats.json` store keyed by ISO week (own `ISOWeek()` year, not
  `t.Year()`, so a week spanning New Year's keys right) — diffs against the most recent
  **prior** week, never "last run", so a same-week rerun neither zeroes the delta nor
  duplicates the snapshot. `Prune(12)` caps stored history.
- `forge check`'s `jobs()` now writes `moc/weekly/<ISO-week>.md` unconditionally (no
  `--repo` dependency, unlike drift/codebase).

One real bug found and fixed on the way, worth remembering: wiring the weekly file in
without excluding it from the link graph made `moc/weekly/…` a *new graph node* on the
very next scan, which cascaded into non-idempotent `duplicates.md`/`orphans.md`/
`graph-health.md` — caught by the existing idempotency test. Fixed the same way
`reports/` itself is excluded (`pkg/vault/note.go`'s `excludedPrefixes`), plus a
dedicated regression test asserting the weekly file's bytes and the stats-store entry
count are both stable across two same-week runs. `moc/weekly/` and `moc/codebase.md`
now read as the two opposite answers to "is this generated file a stable map (stays in
the graph) or dated accumulating output (excluded, like `reports/`)".

- `check.ai_pass` (new `cmd/forge/check_ai_pass.go`): when `cfg.Check.AIPass` is true,
  three sub-tasks run after `jobs()` — draft refresh for the top BROKEN drift finding
  (`synthesize`), duplicate-merge proposal for the top pair clearing the **0.85 spec
  threshold** (`synthesize`; B-019's bar, not `duplicates.md`'s lower 0.40 operating one —
  new `report.TopDuplicatePair`), ADR stub for the top undocumented churny module (`plan`;
  reuses `report.TopUncovered`, `moc/codebase.md`'s own ranking, exported rather than
  redefined). Each calls `engine.Host{}.Run(req)` **directly**, bypassing
  `engine.Resolve`/budget entirely — the plan's literal wording ("dispatched through
  `pkg/engine`'s `Host.Run()`") and `Host.Run`'s own no-I/O contract both point at the
  simple path, not `forge engine run`'s full tier-resolution/spend machinery. Prints the
  returned `Instruction` only; no auto-apply. Verified against the real vault
  (`--offline`, no `--repo`): correctly no-ops all three with explicit messages, since
  there's no drift/code data without a repo. Unit-tested: `topBroken`'s (Note, Ref)
  tiebreak, `TopUncovered`, `TopDuplicatePair`'s threshold gate.

- `check.drain_advisor_queue` (new `cmd/forge/check_drain.go`): runs after `aiPass`, gated
  on **both** `cfg.Check.DrainAdvisorQueue` and `cfg.Engines.Budget.OnExhausted == "queue"`
  — ADDENDUM §A.4's "budget queue drain" batches deferred T3 spend into this one scheduled
  run. Walks the already-collected `d.notes` (no second vault scan), reusing a new shared
  `isQueued` predicate (`check_collect.go`) so `countQueued`'s cost.md count and the drain
  loop can never read the `pending_advisor` flag two different ways. Unlike `check.ai_pass`,
  this dispatches to the **real** advisor tier (`buildEngine(cfg, "advisor")`, an actual
  HTTP call, `engine.Advisor` critique contract) — `--offline` skips it outright, the same
  posture `deadlinks.md` takes toward its own network probes. Per queued note: re-loads it
  fresh (`loadNoteAndSchema`, `queueNote`'s own pattern) rather than reusing the in-memory
  copy, sends the body as the prompt on the `synthesize` stage, books the returned cost via
  `st.Spend("advisor", ...)`, captures the critique via D2 and prints it (path + confidence
  + whether a patch came back) for human approval, and only clears `pending_advisor`
  (`vault.SetScalars(..., "false")`) once both the call and the spend succeeded — a note
  never comes off the queue with nothing recorded, and its T3 output is never silently
  dropped. Stops mid-drain, rather than looping, the moment the advisor tier's own ledger
  reads exhausted — **fixed post-review**: the first cut gated on
  `engine.Exhausted(cfg, st, clock, "synthesize")`, but `drainOne` always dispatches to the
  advisor tier regardless of what `pipeline.synthesize` names, and the packaged default sets
  that stage's engine to `host` — so the chain walk found nothing metered and the guard
  never fired, meaning drain had no real ceiling under the shipped config. Replaced with a
  direct `st.Remaining("advisor", cfg.Engines.Budget.AdvisorUSDPerDay, clock)` check
  (`advisorExhausted`) so the guard reads the tier that actually runs. Regression test
  `TestDrainAdvisorQueueStopsWhenAlreadyExhausted` pins this: `pipeline.synthesize:
  {engine: host}` (the packaged shape) plus a pre-spent advisor budget must still stop the
  drain. Verified against the real vault (`--offline`): the active config chain already
  has both gates true (packaged default), so this run hit the intended `--offline` skip path
  and printed exactly that, confirming the gate is reachable there. Unit-tested against a
  fixture-copy vault with an `httptest.Server` standing in for the advisor endpoint: flag
  clears and spend books on a successful call, `--offline` leaves the flag untouched, the
  DrainAdvisorQueue gate alone (without `on_exhausted: queue`) dispatches nothing, and an
  already-exhausted advisor budget stops the drain before any dispatch. Known gap left as-is
  (not worth restructuring per review): if the advisor call succeeds but `st.Spend` then
  fails, the flag is never cleared and the note re-dispatches next run — a double charge
  with the first spend unbooked. B-024 still applies: D2 capture is inert under the shipped
  config's `D2Tag` mismatch, so the capture call above is correct but unverifiable end-to-end
  here.

- Two small prerequisites for the five hook subcommands, both done: `pkg/report/index.go`'s
  unexported `truncate` is now exported as `Trim` (doc comment updated to say why —
  `cmd/forge`'s upcoming `session-context` hook reuses the same 4KB-budget cut-at-a-line-
  boundary logic for the profile it appends after the index, rather than reimplementing
  it). One call site changed, no external callers existed under the old name, build clean.
  And `pkg/config.Static` gained `CacheTTLDays int` (`yaml:"cache_ttl_days"`, packaged
  default `30` in `config/forge.config.example.md`) for the still-unwritten `forge
  cache-source` hook's `.forge/cache/<hash>.md` TTL — follows `Check.ChurnDays`'s existing
  convention of "zero means unset; the command applies its own default at the call site"
  rather than baking 30 into the config chain's zero value.

- All five CLI subcommands are done and tested — four are Claude Code hooks
  (fail-silent, always exit 0), the fifth (`forge stats`) is a direct user/skill-invoked
  command with a normal error-exit contract, not a hook. `cmd/forge/session_context.go`
  (`forge session-context`, the `SessionStart` hook): prints `_index.md` then
  `profiles/me.md`, each independently `Trim`-capped at 4KB (not a shared 4KB total —
  "the same byte cap" read as the same cap *value*, applied per section), fail-silent to
  `.forge/session-context.log`, exit 0 always even against a nonexistent vault. Three
  tests (happy path, missing-profile-still-prints-index, always-exits-zero).
  `cmd/forge/intent.go` (`forge intent`, the `UserPromptSubmit` hook): reads the hook's
  stdin JSON (`user_prompt` field, confirmed against Claude Code's own hooks doc rather
  than guessed), reuses `loadDocs`+`recall.Rank` exactly like `forge recall` — no
  `Thresholds.Result`, since intent only needs the top score, not a verdict — and prints
  the `UserPromptSubmit` output schema's `additionalContext`+`continue:true` only above a
  hardcoded 0.7 (deliberately not the config-driven recall thresholds), with a code
  comment citing B-008's known false-positive risk near 0.740. Same fail-silent, always-
  exit-0 posture as `session-context`. `cmd/forge/session_capture.go` (`forge
  session-capture`, the `SessionEnd` hook): parses the hook's stdin JSON
  (`session_id`/`transcript_path`/`reason`), reads the transcript as JSONL, keeps only
  assistant-role text blocks (deliberately never the raw file — tool results and file
  contents live in the same transcript and the telemetry no-raw-content invariant extends
  to what this hook may write into a note), regex-scans for "we established/decided/
  concluded/agreed that…" sentences, and writes up to 3 stub notes via the exact
  `vault.WriteToInbox` convention `forge gate` already uses (`confidence: low` +
  `## Open questions`). Stub frontmatter only sets what's honestly derivable (title/slug
  from the sentence, `type: concept` and `depth`/`freshness_days` at schema defaults,
  `origin: session-capture`, dates, `forge_version`, empty `sources`/`related`/
  `supersedes`) — `stack` and `tags` are left out entirely rather than invented, named in
  the Open questions bullets instead. Deduped by `session_id`+content-hash in
  `.forge/session-capture-seen.json` (derived cache, not source of truth — losing it costs
  duplicate stubs, not data); the key includes session id deliberately, so a
  `--resume`-refired `SessionEnd` against the same transcript writes nothing twice, but the
  same conclusion restated in a genuinely new session is treated as new evidence, not a
  duplicate. `cmd/forge/cache_source.go` (`forge cache-source`, the `PostToolUse`/WebFetch
  hook): three separate `WebFetch` research calls against Claude Code's own hooks and
  tools-reference docs never turned up a literal `tool_response` JSON schema for
  `PostToolUse` — only `tool_input.url`/`tool_input.prompt` are confirmed, from WebFetch's
  own published parameters. Rather than guess a field name (`content`/`result`/`text`),
  `cacheBody` unmarshals `tool_response` as a plain JSON string and uses it verbatim on
  success; any other shape (object, array) is cached as the raw JSON bytes Claude Code
  sent, unmodified — two branches, both true regardless of the real schema, zero
  speculative field probing. Cache key is `sha256(tool_input.url)` truncated to 16 hex
  chars, written to `.forge/cache/<hash>.md` with a small `url`/`fetched`/`ttl_days`
  header; `ttl_days` comes from `cfg.Static.CacheTTLDays`, falling back to 30 when unset —
  nothing enforces the TTL on read yet, this command only writes. Confirmed safe against
  the note contract: `.forge/` is in `pkg/vault`'s `skipDirs`, so `Walk` never descends
  into it — these cache files can never be mistaken for notes by `validate`/`index`/the
  report suite. Recorded as BACKLOG B-025 rather than chased further, since resolving the
  real payload shape needs a live hook firing to inspect, not more doc fetches. Four tests
  (happy-path string response, non-WebFetch tool_name no-op, malformed-stdin fail-silent,
  object-shaped response cached verbatim). All four subcommands registered in `main.go`'s
  `commands` map and `usage` string; full suite (`go build`/`go test`, both cgo and
  `CGO_ENABLED=0` lanes) green after each, including a dedup-on-refire regression test for
  `session-capture`. `session_capture.go`'s `stubTitle` also picked up a small hardening
  fix this pass: truncation now cuts on a UTF-8 rune boundary (`truncateValidUTF8`) rather
  than a raw byte index, since a non-English captured conclusion could otherwise split a
  multi-byte rune and write invalid UTF-8 into a note's `title:` field — `validate.go`'s
  own `MaxLength` check counts bytes (`len(v)`), same as before, so the schema-length
  contract is unchanged; only the split-safety improved.

- `cmd/forge/stats.go` (`forge stats`, the fifth and final subcommand — not a hook, a
  direct command): reads `.forge/log.jsonl` and `.forge/weekly-stats.json` via the exact
  functions `forge check` already uses (`loadNotes`, `slugMap`, `loadAskLog`,
  `report.OpenWeeklyStore`) rather than reimplementing or calling the heavier
  `collectVault` pass. Renders five `text/tabwriter` sections: hit rate
  (`report.HitRate`), most-asked topics (top 15, tagged `(written)`/`(gap)`), gaps
  (asked ≥2, never written — Step 3's exact-slug rule restated as a 3-line filter rather
  than exporting `pkg/report/gaps.go`'s unexported `unwritten`, both using B-020's
  count-desc/topic-asc tiebreak), an approximate research-time-saved estimate
  (`written_hits × 15 min`, a named constant and both a doc comment and the printed line
  calling it a rough estimate — `VaultStats` tracks no such metric anywhere else), and a
  weekly trend table using `Drift` as a stand-in for staleness since `VaultStats` has no
  dedicated staleness field (`{Notes, HitRate, Orphans, Drift}` only) — substitution is
  named plainly in the printed header, not silently misleading. Four tests: happy path
  (a real fixture slug, `hibernate`, resolves written; an unknown topic asked twice
  surfaces as a gap), graceful empty output on a missing log/store, a weekly-trend row
  seeded via `report.OpenWeeklyStore`/`Record`/`Save`, and — since this command is not
  fail-silent like the other four — a nonzero-exit assertion against a nonexistent vault
  path (`vaultOrExit`'s ordinary error contract). All green, full suite (`go build`/
  `go test`, both cgo and `CGO_ENABLED=0` lanes) clean afterward.

- The four Claude Code hook shims + `hooks/hooks.json` are done. `hooks/session-context`
  and `hooks/user-prompt-intent` let stdout pass straight through unredirected — their
  entire job is printing into context (index+profile, or the `{additionalContext,
  continue}` JSON above the 0.7 threshold) — while `hooks/session-end-capture` and
  `hooks/post-tool-cache-source` redirect both streams to `/dev/null`, since
  `SessionEnd`/`PostToolUse` output isn't consumed as context and their underlying Go
  commands already log failures internally. All four resolve the forge binary the same
  way: `$FORGE_BIN`, then `~/.forge/bin/forge` (the global install location CLAUDE.md
  already documents for the D3 vault hook's binary, reused here since these hooks fire
  inside arbitrary project directories, not necessarily the vault repo), then `$PATH` —
  and all four `|| true; exit 0` unconditionally, mirroring `vault-post-commit`'s
  discipline. `hooks/hooks.json` declares the four event bindings in real Claude Code
  settings.json shape (confirmed against this machine's own `~/.claude/settings.json`,
  not guessed), with a `_notes` array documenting the resume gotcha (`SessionStart`
  re-runs on `--continue`/`--resume`; every other hook's output is replayed from the
  saved transcript, not re-executed — nothing time-sensitive belongs in
  `UserPromptSubmit`) and the packaging gap: nothing auto-merges this file into a real
  `settings.json` yet, same class of gap as the root-level `agents/` one, closed only
  when Phase 6's plugin manifest lands. Smoke-tested end-to-end against a scratch vault
  with `FORGE_BIN` pointed at a fresh build: `session-context` printed the index+profile,
  `user-prompt-intent` stayed silent below threshold, both silent hooks exited 0 with no
  output. `hooks.json` validated via `python3 -m json.tool`; all four shims validated via
  `sh -n`.

- Git-anchored drift hooks for code repos are done: `scripts/install_drift_hook.sh
  <code-repo> <vault-dir> [forge-binary]` mirrors `install_vault_hook.sh`'s idempotent
  marker-grep pattern (refuses to overwrite a non-Forge hook, safe to re-run) but installs
  three hooks at once (`post-commit`, `post-merge`, `post-checkout`) and writes
  `<code-repo>/.forge/vault-path` — a code repo has no other way to know which vault it's
  paired with, and the hooks must always pass `--vault` explicitly to hit the drift budget.
  Three new shims: `hooks/code-post-commit` (since-SHA `HEAD^`, skips cleanly on a repo's
  first commit via `git rev-parse --verify -q HEAD^`), `hooks/code-post-merge` (since-SHA
  `ORIG_HEAD`, the ref git itself sets to the pre-merge commit — deliberately not `HEAD^`,
  since a merge can bring in many commits through one parent), `hooks/code-post-checkout`
  (git's own `$1`/`$2`/`$3` args — skips a file-level checkout, `$3 = 0`, and a no-op
  checkout, `$1 == $2`, both silently). All three derive the citation `NAME` (the `repo:`
  prefix a note's code citations use) as the basename of the repo's own `git rev-parse
  --show-toplevel` — deliberately not a second pin file, since it's self-evident at
  hook-fire-time and the real vault has no existing citations under any name for this repo
  yet to conflict with. Same fail-silent discipline as the four session hooks: `set -u`,
  `$FORGE_BIN` → `~/.forge/bin/forge` → `$PATH`, `|| true; exit 0` unconditionally.
  Verified end-to-end against a scratch code repo + a scratch copy of the fixture vault
  (never the real vault, to keep the test side-effect-free): `sh -n` on all three shims
  and the installer; a real re-run of the installer confirmed idempotent. The first pass's
  1163ms cold / 55ms warm / 411ms post-merge numbers were caught during Task #15's
  verification as **floor numbers from a null test** — the scratch repo had zero real
  citations, so every run proved only that the plumbing (binary resolution, pin file,
  skip logic, exit codes) worked, not that `forge drift` found or demoted anything.
  Re-tested with an actual citation-bearing note: `code-post-commit` originally passed
  the literal string `"HEAD^"` as `--since-commit` (the other two shims already resolved
  to a concrete SHA first) — fixed to `since=$(git rev-parse --verify -q HEAD^) || exit 0`
  so the value, and anything drift persists from it, stays reproducible. The installer's
  `install_one` also had a real atomicity gap — it checked-and-copied one hook at a time,
  so a conflict on the second or third hook left the first one installed and the run
  half-done; fixed by splitting into a `check_one`/`install_one` pair that verifies all
  three markers before copying any. Installer now also warns (stderr, non-fatal) when the
  target repo's `.gitignore` doesn't exclude `.forge/`, since `vault-path` is local
  pairing state, not something to commit; loosened the vault precondition from "must be a
  git repo" to "must be a directory", since the script never touches the vault's own git
  state.

  Re-verified with a real citation this time: a scratch note citing `repo:App.java#extra`
  (symbol-level, `code_refs:` canonical form) went through the full pipeline twice.
  First, a body-changing commit that didn't touch the cited symbol correctly left
  `confidence: high` and only advanced `drift_checked_at` to the new SHA. Then a commit
  that removed the `extra()` method fired `code-post-commit` for real — 90ms warm — and
  correctly flipped `confidence: high → low`, matching `.forge/demotions.json` and
  `.forge/drift.log` written into the scratch vault. This is the discriminating positive
  result the null test couldn't produce. One real defect found and recorded rather than
  fixed, since it's pre-existing in Phase 2b's `pkg/drift`, not Phase 5's hooks: **B-026**
  — a citation to a file that's been *fully deleted* (not just a symbol inside a
  surviving file) can never verdict BROKEN, because `cmd/forge/drift.go`'s registry is
  built from `ScanRepo(..., "HEAD")` — the current tree — so a deleted file's basename is
  simply absent from it and `Registry.Resolve` reports `Unresolved`/`Skipped` before
  `checkPath`'s file-existence branch is ever reached. Confirmed by reproducing it live:
  deleting `App.java` outright and citing it by path (no `#symbol`) left the note at
  `confidence: high` forever, `forge drift --deep --json` reporting
  `"verdict":"skipped","reason":"no registered repository contains this path"` instead of
  `"broken"`.

- `skills/forge-check/SKILL.md` and `skills/forge-stats/SKILL.md` are done, following
  `skills/forge/SKILL.md`'s frontmatter+staged-steps template. `forge-check`: Step 0
  frames the command as T0-only/cron-safe; Step 1 covers the exact flag set read straight
  from `cmd/forge/check.go` (`--vault`, `--repo NAME=PATH` repeatable, `--months`,
  `--days`, `--offline`); Step 2 leads with the always-written `moc/weekly/<ISO-week>.md`
  rollup and instructs relaying B-017/B-019's caveat lines verbatim rather than presenting
  an empty "Act now" section as "nothing needs attention"; Step 3 covers `check.ai_pass`
  (only mentioned when the config has it on; proposals are shown and approved
  individually, never auto-applied, routed through the normal gate pipeline); Step 4
  covers `check.drain_advisor_queue`. `forge-stats`: Step 0 disambiguates "telemetry off"
  from "genuinely unused" before reporting an empty table; Step 1 covers the single
  `--vault` flag and the command's non-fail-silent contract; Step 2 relays all five
  sections (hit rate, most-asked topics tagged written/gap, gaps, the approximate
  time-saved estimate, the weekly trend — noting its `Drift` column stands in for a
  staleness metric that doesn't exist yet); Step 3 redirects "should I write about X" to
  the Gaps section instead of re-running `forge recall`.

Task #15's verification pass closed out Phase 5: `go test ./...` now reports **17**
packages `ok` (up from 16 — `pkg/telemetry` is the new one), `[REDACTED-KEY]`
landed alongside the existing `TestE2EIndexRespectsTheBudget`, and `forge
session-context`/`forge intent` were timed 20 iterations warm+cold against synthetic
stdin — both landed well inside their <200ms/<50ms budgets. A real `forge check --vault
[REDACTED-PATH]` run (followed live via `skills/forge-check/SKILL.md`'s own
steps) wrote `moc/weekly/2026-W33.md` (94 notes · hit-rate 0% · orphans 19 · drift 0,
correctly flagged "First recorded week"), was byte-identical on an immediate second run,
and left `.forge/weekly-stats.json` holding exactly one entry — confirming the
same-week-rerun-doesn't-duplicate contract holds against the real vault, not just the
fixture. `check.ai_pass` is `false` in the resolved config (the shipped default), so the
skill correctly said nothing about it; `check.drain_advisor_queue` is `true` but the
run's own stdout already showed `0 dispatched, 0 failed`, so there was nothing to drain.
`git status --short` on the vault after all of this showed only the expected report
regeneration plus the two new Phase-5 artifacts (`moc/weekly/`, `reports/cost.md`) —
nothing unexpected, nothing left over from the compaction-boundary backup/restore cycle
earlier in the session. The `CLAUDE.md` status section is now updated to name Phase 5
done, list the five new subcommands and `pkg/telemetry`, and record the measured hook
latencies. B-008's §3.1 recalibration, B-023 (`on_exhausted: stop` vs. every doc's
`fail`), B-025 (`cache-source`'s unconfirmed `tool_response` shape), and B-026
(deleted-file citations never verdict BROKEN) all remain open — recorded, not fixed, in
Phase 5, per this project's own convention.

### Phase 5b — `forge logback` (repo commit on top of Phase 5)

ADDENDUM §B.7 / DESIGN §15, "Log-back into the codebase": makes the vault's knowledge
discoverable *from the code repo itself*, not just from the vault. T0, deterministic,
idempotent, never touches code semantics — comments and separate files only.

New package `pkg/sentinel`: an id-based begin/end managed-block primitive
(`Upsert`/`UpsertBefore`/`Remove`, `Style{Open,Close}` for Markdown/Slash/Hash comment
syntax) — finds a block by its id, not its position, so hand-written prose above and
below stays untouched and a second write with the same body is byte-identical. Nothing
like it existed before; `pkg/vault/fix.go`/`quarantine.go` only ever did whole-file
regeneration.

`forge logback --repo NAME=PATH [--vault DIR] [--dry-run] [--remove-markers]` does three
things, each independently gated by config (`static.logback.knowledge_map` /
`claude_md_fragment` / `inline_markers`, packaged defaults `true`/`true`/`false`):
renders `docs/knowledge-map.md` (module → note links, reusing `check_codebase.go`'s
existing `groupsOf`/`citedPaths`/`locate` join — no new grouping logic needed), upserts a
managed `CLAUDE.md` fragment per documented module, and — opt-in — writes a
`// forge:logback:<symbol>` comment immediately above each cited symbol
(`codeindex.Symbol.Start` gives the anchor line). `--remove-markers` strips inline
markers only, byte-for-byte, independent of the config gate (so it still works after the
gate's been turned back off). `.forge/code-index-<repo>.json` freshness was already
solved by Phase 2b's `pkg/drift`/`pkg/codeindex` and is reused as-is.

One real bug caught before it shipped: inline-marker resolution must key off
`coderef.Ref.Symbol != ""`, not `Ref.Kind == KindSymbol` — the canonical `code_refs:`
citation form (`repo:path#Symbol`) parses to `Kind: KindPath` with `Symbol` populated,
so filtering on `Kind` alone would silently skip nearly every real citation in the vault
and only match the rare bare-symbol inline-body form. Found by reading
`pkg/coderef/ref.go`/`extract.go` closely before writing tests, not by a failing test.

Verified via `logback_test.go` (dispatch, full pipeline, idempotent rerun,
`--remove-markers` round-trip, per-flag config gating — all green under both
`CGO_ENABLED=0` and `CGO_ENABLED=1`) plus a hand-built smoke test against a real temp git
repo: `docs/knowledge-map.md` and a `CLAUDE.md` fragment both rendered correctly, an
inline marker landed immediately above the cited method, a second run produced a
byte-identical diff (true idempotence), `--remove-markers` reverted the source file
byte-for-byte, and `--dry-run` wrote nothing to disk at all while still printing what it
would do.

New item found and recorded, not fixed: **BACKLOG B-027** — `pkg/drift/gitindex.go`
caches each repo's symbol table as `.forge/code-index-<repo>.json` (correct, since a
shared singular name would collide across `--repo`s), while ADDENDUM §B.6/DESIGN §15
both describe the singular `.forge/code-index.json` — a documentation fix, not a code
one; pre-existing since Phase 2b, found during this phase's explore pass.

## Key Architecture Choices
- Go static binary for the T0 core (ADR-001) — chosen for ~1–5ms startup, not compute;
  the binding constraint is `forge drift` <100ms on the git-hook path
- Zero model calls in the static core; optional 4-tier engine layer on top
  (none / host / API / advisor-critique), configurable per pipeline stage
- Stages `recall`, `write`, `index` are hard-locked to engine `none`
- Advisor tier is critique-only — returns a patch, never a rewrite ("generate cheap,
  critique expensive")
- Plain markdown is the source of truth; SQLite is a derived cache `forge reindex`
  rebuilds from scratch
- Lexical recall, deliberately **no embeddings** — a four-channel weighted blend in
  `pkg/recall` for question→note scoring; MinHash + LSH stays for near-duplicate
  detection in `pkg/similarity`
- Drift detection is git-anchored: post-commit/merge/checkout only, never on file save
  or the uncommitted tree, with symmetric rollback

## Phase Order
`0 → 1 → 2 → 2b → 3 → 3b → 4 → 5 → 5b → 6 → 6b` — this repo's roadmap now **ends at 6b**
(BACKLOG **B-021**, decided 2026-08-09). The B2B "no context loss" product
(`KNOWLEDGE-FORGE-B2B.md`) is no longer a phase inside this repo at all — it's a fully
separate project, informally gated on the same readiness signal (30 days real usage,
≥3 outside users) but not enforced here. Phases 0, 2 and 2b matter most; 2b is never cut.
Per-phase prompts live in `docs/CLAUDE-CODE-PROMPT.md`.

## Open Items
- **31 notes still fail the schema** — the `sources:` backfill in particular is a reading
  job, not a scripting one. `lint-report.md` in this vault has the per-note list
- **BACKLOG B-008 (Phase 3's, still open after Phase 3)**: the IDF weighting shipped in 2b and **neither case was
  fixed** — "Redis caching in Spring Boot" is unchanged at 0.740 (still extend into the
  wrong note), "Kafka consumers with Testcontainers" moved 0.469 → 0.501 (still CREATE).
  The stated cause was wrong. In this vault `redis` and `caching` have df 0 as both tag and
  stack, so they are filtered out of the denominator before any weight is computed; the one
  surviving tag term, `spring`, has df 1, and a weighted ratio over a single term is 1.0 on
  any hit. The fix is to let question terms the vault carries nowhere count against a note —
  which re-creates the active-zero channel §2.5 rejected from measurement, so it owes a full
  §3.1 recalibration and was not taken here. The thresholds stay at 0.85 / 0.55
- **BACKLOG B-009 (Phase 6)**: `pkg/gitsig` runs the `git` CLI, so a packaged binary
  assumes `git` is on `PATH`
- **BACKLOG B-007 — closed in Phase 4**: `forge-librarian`'s prompt stamps
  `Forge-Write: true` on every commit it authors; `d3_forge_write_test.go` pins the
  guard both ways
- **BACKLOG B-022 — closed in Phase 4**: `engine_trail`'s schema pattern now covers all
  nine `cfg.Pipeline` stage names minus `critique`
- **BACKLOG B-023 (found in 3b, still open)**: code's `on_exhausted: stop` vs. every
  doc's `fail`, and `stop`/`degrade` are byte-identical in behavior to the default today —
  only `queue` is ever read. Docs left stale on purpose; someone still has to decide
  whether `stop` and `degrade` should diverge before either gets renamed
- **BACKLOG B-024 (found in Phase 4, not fixed)**: `pkg/dataset/d2.go`'s
  `D2Tag = "d2_advisor"` never matches the packaged config's `"d2"` capture-list entry,
  so D2 capture is silently inert under the shipped config
- **Packaging gap (Phase 4, widened in Phase 5)**: the four `agents/*.md` files are spec
  only — nothing in this repo loads a root-level `agents/` directory yet, so they aren't
  live, dispatchable agents until packaging exists. Phase 5's `hooks/hooks.json` is the
  same class of gap: nothing auto-merges it into a real `settings.json` either. Both close
  only when Phase 6's plugin manifest lands
- **BACKLOG B-025 (found in Phase 5, not fixed)**: `forge cache-source`'s
  `PostToolUse`/WebFetch `tool_response` JSON shape was never confirmed from official
  docs, so `cacheBody` deliberately caches the raw bytes rather than guessing a field name
- **BACKLOG B-026 — closed 2026-08-16, out-of-phase session** (found in Phase 5,
  pre-existing in Phase 2b's `pkg/drift`): a citation to a fully deleted file could never
  verdict BROKEN, because `registryOf` always built the citation registry from the
  current `HEAD` tree, so a deleted file's basename was simply absent from it rather than
  resolving to a broken reference. Fixed via a new `Source.ResolveAt(ref, asOf)` —
  `GitSource.registryAt` rebuilds the registry against the historical tree at the note's
  `verified` date via `coderef.ScanRepo` (memoised per `asOf`, one `git ls-tree` per
  (repo, date) pair), gated to fire only on `forge check`'s deep sweep, never on the
  100ms-budget `forge drift` hook path. Built via `subagent-driven-development`
  (implementer + task review + a whole-branch review that surfaced and got three
  doc-only fixes), merged to `main` at `305fb53`, both build lanes green. One caveat the
  final review surfaced and BACKLOG now documents as item 5 of the closure note: no
  shipped invocation actually applies the `Broken` verdict today (`forge check` never
  calls `drift.Apply`; the `code-post-commit` hook applies but isn't `--deep`) — this fix
  makes `drift.md` accurate on a full sweep, it doesn't yet demote a note through any
  automated path. That immediacy gap is filed separately as **BACKLOG B-028**.
- **BACKLOG B-027 (found in Phase 5b, not fixed, pre-existing in Phase 2b's
  `pkg/drift`)**: `pkg/drift/gitindex.go` caches each repo's symbol table as
  `.forge/code-index-<repo>.json`, while ADDENDUM §B.6/DESIGN §15 both describe the
  singular `.forge/code-index.json`. The per-repo suffix is correct behavior (a shared
  name would collide across `--repo`s) but undocumented — a docs fix, not a code one
- Module rename **done** (2026-08-08). Directory rename and the eventual
  `github.com/<user>/…` module path stay deferred — BACKLOG B-003 / B-004
- The old topology directories (`concepts/ decisions/ entities/ issues/ syntheses/
  archive/ TIL/`) survive as empty shells; `raw/` (5) and `sources/` (9) are still live
  and deliberately outside the note contract
- The fixture vault at `[REDACTED-PATH]/` (13 notes, pre-migration
  topology, 12 catalogued defects) did its job — the migration was rehearsed there before
  it ran here. It carries no `.git` on purpose: the harness copies it to a temp dir and
  `git init`s the copy, avoiding a nested-repo gitlink

## Sources
- `[REDACTED-PATH]` (index over the other four)
- `[REDACTED-PATH]` (ADR-001, wins on stack questions)
- `[REDACTED-PATH]` (Phase 2's scoring, thresholds, output
  contract — the calibration table is §3.1)
- `[REDACTED-PATH]` (Phase 4 — antislop gate's banned
  phrases + structural requirements, parsed not hardcoded)
- `[REDACTED-PATH]` §6 (Phase 4 — write-time gate vs.
  passive report threshold asymmetry)
- `[REDACTED-PATH]`
