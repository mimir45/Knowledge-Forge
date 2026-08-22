# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Status

**Phases 0, 1, 2, 2b, 3, 3b, 4, 5, 5b, 6 and 6b are done** (2026-08-09 through 2026-08-22;
Phase 1 merged as `1c9df95`, Phase 2 as `3619b72`, Phase 2b committed straight to `main`,
`cb12a08`…`15a795f`; Phase 3 committed straight to `main` in one commit; Phase 3b
likewise, on top of `847098a`; Phase 4 likewise, on top of `884e42e`; Phase 5 likewise,
on top of Phase 4's commit; Phase 5b likewise, on top of Phase 5's commit; Phase 6
likewise, on top of Phase 5b's commit). The repo is
a git repo with a Go source tree: `cmd/forge`
(`slug validate index reindex capture recall drift check config init engine verify-code
gate session-context intent session-capture cache-source stats logback`) over `pkg/vault`,
`pkg/graph`, `pkg/report` (now including `weekly.go`'s rollup renderer,
`weekly_store.go`'s week-over-week `.forge/weekly-stats.json` persistence, and, new in
Phase 5b, `knowledgemap.go`'s `RenderKnowledgeMap`), `pkg/store`,
`pkg/dataset`, `pkg/recall`, `pkg/similarity`, `pkg/codeindex`, `pkg/coderef`,
`pkg/gitsig`, `pkg/drift`, `pkg/linkcheck`, `pkg/config` (the four-layer config chain),
`pkg/engine` (the four-backend engine abstraction — `none/host/api/advisor`, per-stage
routing with fallback chains, SQLite-backed budget accounting), `pkg/qualitygate` (the
seven DESIGN §12 gates + `Run`/`Report` orchestration + `_inbox/` quarantine),
`pkg/telemetry` (DESIGN §14's `ask` event, sha256 topic hashing, never
raw question text; gated fully behind `cfg.Telemetry.Enabled` and wired into `forge
recall`), `pkg/sentinel` (new in Phase 5b — the id-based begin/end managed-block
primitive `forge logback` uses for CLAUDE.md fragments and inline markers; `Upsert`/
`UpsertBefore`/`Remove`, idempotent, position-independent), plus seven note templates in
`templates/`, `skills/forge/SKILL.md`,
`skills/forge-init/SKILL.md`, `skills/forge-check/SKILL.md`, `skills/forge-stats/
SKILL.md` and (new in 6b) `skills/forge-export-dataset/SKILL.md` +
`skills/forge-dataset-stats/SKILL.md`, `references/recall-spec.md`,
`references/writing-rules.md`,
eight packaged presets in `config/presets/`, a Makefile with a six-target cross-compile
matrix, a hash-verifying `bin/forge` shim, a `hooks/` + `scripts/` pair that now installs
both the vault's D3 capture hook (`vault-post-commit`) and four Claude
Code lifecycle shims (`session-context`, `user-prompt-intent`, `session-end-capture`,
`post-tool-cache-source`, declared in `hooks/hooks.json`) plus three git-anchored
drift shims for code repos (`code-post-commit`, `code-post-merge`, `code-post-checkout`,
installed via `scripts/install_drift_hook.sh`), and four `agents/*.md` spec
files (`forge-researcher`, `forge-codebase-scout`, `forge-verifier`, `forge-librarian` —
spec only, see the packaging-gap note below). Build and test with `CGO_ENABLED=0 go
build ./...` / `go test ./...` — 18 packages report `ok` (`config`, `profiles`,
`references` are data-only, no test files), all green under both `CGO_ENABLED=0` and
`CGO_ENABLED=1`.
Phase 5's own `forge session-context` / `forge intent` warm-latency check ran 20
iterations against synthetic stdin; both landed well under budget (<200ms / <50ms).
`forge check` against the real vault confirmed the new `moc/weekly/YYYY-WW.md` rollup
renders, is byte-identical on a second immediate run, and `.forge/weekly-stats.json`
persists across runs without zeroing or duplicating deltas within the same ISO week. Two
new backlog items recorded rather than fixed this phase: **B-025** (`forge cache-source`'s
`PostToolUse`/WebFetch `tool_response` JSON shape was never confirmed from official docs,
so `cacheBody` deliberately caches the raw bytes rather than guessing a field name) and
**B-026** (a citation to a fully deleted file can never verdict BROKEN, because
`registryOf` always builds `pkg/coderef`'s registry from the current `HEAD` tree — found
smoke-testing Phase 5's drift hooks, pre-existing in Phase 2b's `pkg/drift`, not this
phase's to fix). The packaging gap already on file for root-level `agents/` now also
covers `hooks/hooks.json`: nothing in this repo auto-installs it into
`~/.claude/settings.json` or a project's `.claude/settings.json` — closed by Phase 6's
plugin manifest (below).
**B-026 and its sequel B-028 were later closed by standalone fixes on `main`, not tied to
a phase.** B-026 (2026-08-16): `pkg/coderef`'s registry-based resolution now falls back to
a verified-era `ResolveAt` scan on a full sweep (`opts.Deep`, no `--since-commit`), so a
fully deleted cited file verdicts `Broken` there instead of `Skipped` forever — but
`forge check`'s full sweep never calls `drift.Apply`, so B-026 alone made `drift.md`
accurate and demoted nothing automatically, and its weekly cadence plus the hook path's
(`forge drift`'s default, what `code-post-commit`/`code-post-merge`/`code-post-checkout`
actually run) total lack of immediacy on a same-commit deletion is what got filed as
B-028. B-028 (2026-08-17): the hook's already-computed cheap gate
(`coderef.ChangedFilesStatus`, `--name-status` instead of `--name-only`) now carries
deletion evidence (`drift.Changed{Touched, Deleted}`), so an `Unresolved` citation
matching a same-commit deletion verdicts `Broken` immediately under `--apply` — the only
automated demotion path a deleted-file citation has — with no `--deep` and no historical
registry scan required, plus a gate-ordering correction over the backlog's own sketch (an
unmatched hook-path miss produces no finding at all, never `Skipped`, so an unrelated
later commit cannot flip a still-broken note back to `high`; see
`TestRollbackSymmetryOnDeletion`). See BACKLOG.md's B-026 and B-028 entries for the full
closure notes, including each fix's residual basename-collision limitation.
Phase 5b (ADDENDUM §B.7 / DESIGN §15, "Log-back into the codebase") built `forge logback`:
T0, deterministic, idempotent, generates `docs/knowledge-map.md` and per-module `CLAUDE.md`
fragments in the target code repo (both gated independently by
`static.logback.{knowledge_map,claude_md_fragment}`), plus opt-in inline
`// forge:logback:<symbol>` markers (`static.logback.inline_markers`, default off,
revertible via `--remove-markers`) — new package `pkg/sentinel` is the id-based
begin/end managed-block primitive both the CLAUDE.md fragments and the inline markers
share, and it never touches anything outside its own marker pair. `.forge/code-index-
<repo>.json` freshness was already solved by Phase 2b's `pkg/drift`/`pkg/codeindex` and
is reused as-is, not rebuilt. Verified via `logback_test.go` (dispatch, full pipeline,
idempotent rerun, `--remove-markers` round-trip, per-flag config gating) passing under
both `CGO_ENABLED=0` and `CGO_ENABLED=1`, plus a hand-built smoke test against a real
temp git repo confirming byte-identical reruns (`diff`, no output) and a byte-for-byte
`--remove-markers` revert. One correctness fix worth naming: inline-marker resolution
must key off `coderef.Ref.Symbol != ""`, not `Ref.Kind == KindSymbol` — the canonical
`code_refs:` citation form (`repo:path#Symbol`) parses to `KindPath` with `Symbol` set,
so filtering on `Kind` alone would silently skip nearly every real citation. New backlog
item recorded rather than fixed this phase: **B-027** (half-closed 2026-08-21;
`pkg/drift/gitindex.go` caches
per-repo as `.forge/code-index-<repo>.json`, while ADDENDUM/DESIGN both describe the
singular `.forge/code-index.json` — correct behavior, since one shared name would
collide across repos, but undocumented; found during Phase 5b's explore pass,
pre-existing since Phase 2b, not this phase's to fix).
Phase 6 (2026-08-18, "Package & release") turned the working binary into an installable
Claude Code plugin. `.claude-plugin/plugin.json` + `.claude-plugin/marketplace.json`
(repo `mimir45/Knowledge-Forge`) close the packaging gap noted above for `agents/` and
`hooks/hooks.json`; `hooks/hooks.json` itself now uses `${CLAUDE_PLUGIN_ROOT}`-relative
paths instead of the hardcoded absolute ones. New package `pkg/scrub` + `forge scrub
<src> <dst>` is AUDIT §8.4 D-6's binding, previously-unlisted deliverable: it redacts
email/absolute-home-path/key-prefixed-token/long-token-shaped content from a copied
vault and **fails closed** — a note that can't be re-validated against
`references/schema.yaml` after scrubbing aborts the whole run with nothing written to
`--dst`. Two real false-positive classes in its long-token heuristic
(`reLongToken`, "any 32+ char run of `[A-Za-z0-9]`") were found and fixed via real-vault
dry runs, not caught by the fixture test alone: kebab-case slugs and dated filenames
(e.g. `2026-04-13-local-ai-continue-rag-spring`) were corrupting `sources:`/wikilink
citations, fixed by dropping `-`/`_` from the character class; camelCase Java
identifiers in embedded code samples (e.g.
`getPaymentOutboxMessageBySagaIdAndSagaStatus`) were also false-positiving, fixed by
requiring at least one digit in the match (RE2 has no lookahead, so the filter runs
post-match in `redactLongTokens`) — a random 32+ char draw from `[A-Za-z0-9]` contains a
digit with near-certainty, so real hex/base64/JWT-shaped secrets are still caught.
Redactions on the real vault's 122 notes went 637 → 86 → 43 across the two fixes; both
have regression tests (`TestScrubDoesNotRedactSlugsOrFilenames`,
`TestScrubDoesNotRedactCamelCaseCodeIdentifiers`). One residual false positive is
deliberately left unchased: an identifier that happens to embed a digit (e.g.
`TestE2ESessionContextRespectsTheBudget`, via "E2E") still trips the heuristic — one
occurrence in the shipped `examples/vault/`.
`examples/vault/` (93 files: 91 notes under `notes/{pitfall,concept,decision,howto}/`
plus `moc/codebase.md` and `moc/weekly/2026-W33.md`) was generated by running `forge
scrub` against the real vault, scoped to `notes/`+`moc/` only by explicit user decision
overriding DESIGN §16.4's "15-20 real notes" line — `raw/`, `sources/`, `reports/`, and
root-level loose files excluded. The user reviewed the scrubbed output and signed off
before it was committed, per this phase's own binding review-gate requirement. Two ADR
files were added per AUDIT §8.4 D-3: `docs/adr/0001-lexical-recall-vs-embeddings.md`
(DESIGN §8) and `docs/adr/0002-go-for-static-core.md` (STACK §1); no third ADR for STACK
§6, since B2B is a separate project (B-021). README.md/LICENSE/CHANGELOG.md/
CONTRIBUTING.md were added — README and LICENSE are release-blocking, not just
documentation, since `.goreleaser.yml`'s `archives.files` lists `README.md` as a
non-glob entry. `evals/` scaffolding and a `ci.yml` lint/evals step (`golangci-lint` via
new `.golangci.yml`, `errcheck` disabled repo-wide, recorded as **B-029**) round out the
phase. **B-013 closed this phase**: `pkg/codeindex.Extractor`'s doc comment now
explicitly covers cache-format/serialized-shape versioning, not just extraction-logic
versioning, per its own "must land before the first released binary" text. Verified:
`CGO_ENABLED=0 go build ./...` and `go test ./...` green across all packages including
the new `pkg/scrub`. **Not yet verified, and not claimed**: the shim's real
download-and-checksum path and `claude plugin marketplace add mimir45/Knowledge-Forge`
from a genuinely clean machine — both need the tagged release this phase's remote-push
step produces, which is the phase's actual "done when" condition.
Phase 6b (2026-08-22, "Dataset capture & export") closed the roadmap. Its first step was
implementation, not the verification the phase prompt asked for: `pkg/dataset` held
`d2.go`/`d3.go`/`d4.go` and **no `d1.go` or `d5.go`**, while
`config/forge.config.example.md` shipped `capture: [d1, d2, d3, d4, d5]`. New
`pkg/dataset/tier.go` is the five-tier registry that removed the trap behind that gap —
`Enabled()` read as general but hardcoded `D2Tag`, and `D4Enabled()` existed only because
of it, so a third tier added by reaching for the general-sounding name would silently have
taken D2's gate. `Tier.Enabled` takes `config.Dataset` rather than the bare capture list,
which also closed a quieter hole neither call site had: `dataset.enabled` was checked
nowhere, so `{enabled: false, capture: [d2]}` captured anyway. **B-030 closed here** —
`forge capture` now honours the list, keeping `hooks/vault-post-commit`'s two rules (exit 0
always, stderr only), and an unreadable config **skips** capture rather than proceeding,
because fail-open is the wrong default for a consent check.
D1 captures in `runRecall` beside `logAsk` and is deliberately scoped to `forge recall`:
`forge intent` also ranks the vault on every `UserPromptSubmit` but carries a 50ms budget
and a never-disturb-the-session contract, and `intent.go` builds its own `recall.Query`
rather than calling `runRecall`, so the limit is structural. **ADDENDUM's "every run" is
wrong on this point and every D1 datasheet says so.** D5 captures in `forge gate`'s
non-quarantine branch — the only acceptance signal in the tree — carrying seven fixed-shape
`profiles/me.md` fields and deliberately *not* the four free-text ones (`assume_known`,
`never_assume`, `code_style`, `avoid`), which the template invites employer-specific prose
into. Nothing in Go enforces that gate runs; it is an invariant of `skills/forge/SKILL.md`,
so **D5 is a subset of accepted notes, not a census**, and that too is in the datasheet.
Alongside: `logAsk` now fills `telemetry.Event.Stack`, which existed since Phase 5 and
production never wrote, starving two real readers (`pkg/report/index.go:67`,
`coverage.go:43`).
The export path needed a reader that inverts every reader already in the tree: `jsonl.go`,
`check_asklog.go` and `session_capture.go` all skip an unparseable line on purpose ("a
truncated tail must not wedge every future commit"), but on the export path a line nobody
can parse is a line nobody could redact, so `pkg/dataset/read.go` refuses it and drops the
run — naming file and line so the failure is a hand-edit rather than a bug report, and
separating `bufio.ErrTooLong` from EOF, which a default Scanner silently conflates.
Fail-closed is two layers proving two different things: buffer-then-commit (pkg/scrub's
shape) means nothing reaches `--out` until the whole run succeeded; the per-record re-decode
proves redaction did not corrupt a record's *shape*. **Neither proves no secret escaped** —
only `TestAnonymizeRemovesEverySeededSecret` does, which is why that test is D-6's
regression guard. `pkg/scrub` gained exactly one exported wrapper (`Redact`, zero behaviour
change); export-only strictness (internal-URL patterns) lives in `pkg/dataset/anonymize.go`,
because a note body legitimately cites `http://localhost:8080` and a corpus meant to leave
the machine does not.
**Two corrections to the phase plan, both load-bearing.** D2 is *not* a DPO tier: §D.1's
table describes its pair as "draft → critique → accepted patch" but Phase 3b captures
`D2Pair{Draft, Critique}`, so the chosen side does not exist and emitting DPO would
fabricate a preference. And an undefined `(set, format)` combination exits **2, not 3** —
exit 3 promises "a real attempt was made, `--out` untouched", which misdescribes a request
rejected before a record was read. **One limit is stated rather than solved, everywhere:
topic slugs are kept.** They are the only semantic feature D1 and D5 carry and hashing them
makes those corpora untrainable, so a topic named after a product survives redaction. Making
that observable is also why rendered lines carry an `id` — without it, path hashing and SHA
blanking reached no output format at all. Two items opened rather than built: **B-034** (D6,
which is a derivation over `forge logback`'s map, not a capture tier — four of five sources
say five datasets and AUDIT never flagged the disagreement) and **B-035** (D1's missing
outcome label; the blocker is that no `run_id` correlates a recall call to the note write
that follows). Verified: both build lanes green, `go vet` clean, smoke-tested against a temp
fixture vault only.

Everything else below is still design spec; **the roadmap is complete.**

**A defect-cleanup pass ran 2026-08-21 on `simplify/codebase-cleanup` — out-of-phase work,
not a phase, and it does not reorder the roadmap.** It took the doc-sync and one-line tier
of the open backlog and deliberately left the two items that need their own session.
Closed: **B-024** (`D2Tag` renamed `"d2_advisor"` → `"d2"`; the reason it shipped green was
that no test asserted config and code agree, so `pkg/dataset/capture_gate_test.go` now pins
the packaged layer against both live tags — verified to fail when the old spelling is
restored; **not** verified end to end, since both `captureD2` call sites sit behind a live
metered advisor call) and **B-009** (it was already satisfied by `README.md:25-31` in Phase
6; only the status line was stale). Half-closed: **B-023** (the four doc sites now say
`stop`, matching the code; the behavior question — `stop` halts nothing and `degrade` is
not a distinct path from the default fallthrough — is untouched and still open, kept out of
a doc commit on purpose) and **B-027** (`agents/forge-codebase-scout.md` was telling that
agent to seed from a path that never exists on disk, which was the one operational
consequence, plus two `pkg/codeindex` doc comments; ADDENDUM §B.6 / DESIGN §15 still say
the singular name, deliberately). Re-triaged: **B-025** is blocked on observing a live
`PostToolUse`/WebFetch payload, not open work — **do not re-attempt the WebFetch**. New:
**B-030** (`dataset.capture` accepts five tiers but only `d2`/`d4` gate anything; `d3` is
implemented and never reads the list, so removing `d3` silently does not stop capture —
**closed in Phase 6b, 2026-08-22**, by making the control real rather than documenting it).
Two items were re-sized rather than fixed, so the next session does not start on a wrong
estimate: **B-029** is roughly double its recorded scope (measured **95** raw errcheck
findings, not "~20"; ~37 after default exclusions, 10 of them production). Its triage item 1
landed 2026-08-22 — `cmd/forge/recall_load.go`'s `refresh()` returned `nil` on every path —
but **not the way the entry prescribed**: propagating reaches `runRecall`, which exits 1
without emitting candidates it already scored correctly, so a transient SQLite lock held by
a concurrent `forge intent` would cost the answer. The signature was the defect, not the
missing propagation; `refresh` no longer returns `error` and `writeRows` checks all three
errors the old body dropped. The entry's item 3 (`catfile.go`'s `cmd.Wait()`) was re-traced
in the same pass and its claim is **weaker** than recorded — read B-029's closing section
before scheduling the sweep. **B-008's** hidden prerequisite was
that **no harness producing the §3.1 table existed** — closed since, see below. See BACKLOG
for the rest.

**B-008 closed 2026-08-22, on `worktree-b-008-recall-recalibration` — out-of-phase work,
not a phase.** The harness came first: `cmd/forge/calibration_test.go` runs §3.1's nine
queries against `examples/vault` (92 docs, staged into a temp dir per run, because
`loadDocs` writes a SQLite cache under `<root>/.forge` and scoring in place would mutate a
tracked directory) and diffs `cmd/forge/testdata/calibration.golden`; `-update` re-records
it. The "before" column was measured against the unmodified scorer and committed before the
fix was written. The fix is the two changes the entry predicted: the vocabulary filter
changed sides (it now filters `--stack` hints, not question terms — a hint is a user filter,
a question term is evidence), and an absent term weighs **the mean of the present ones**,
assigned in the weight-map builder so `idf(0, n) == 0` stays true and its test was
*preserved rather than inverted*, contrary to what the entry predicted. Measured: the
0.740 false positive falls to **0.415** and out of first place. **Two of the entry's own
sizing findings were wrong** and are corrected there: `docker-compose-init-container-pattern`
*is* in `examples/vault` (the file name is longer than the backlog's shorthand), and the
corpus had not drifted — eight of nine "before" scores reproduce §3.1's originals exactly.
The cost is real and was shipped knowingly on the user's decision: every "before" UPDATE
verdict came from one artifact (a channel reading 1.000 off a single surviving term), and
removing it also costs the Storybook row, where that artifact happened to land on the right
note. Four narrower admission rules were measured and none recovers it. Three items opened
rather than folded in: **B-031** (Kafka/Testcontainers is a coverage defect, split out —
admission is strictly decreasing, so one knob cannot move a false positive down and a miss
up), **B-032** (an untagged note escapes the absent-term penalty entirely, §2.5's asymmetry
running the other way), **B-033** (the 0.30 neighbour floor predates the scale change, so an
adjacent-topic query now verdicts CREATE with *zero* neighbours). **The thresholds did not
move and still must not.** `references/recall-spec.md` §2.3/§2.4 were stale by two changes,
since 2b's IDF weighting was never documented at all; there is now a §2.3.1, a §2.5 note,
a generated §3.1 and a §4.1 example matching real output.

**B-022 closed in Phase 4**
(the schema pattern now covers all nine `cfg.Pipeline` stages minus `critique`); **B-007
closed in Phase 4** (`agents/forge-librarian.md`'s prompt stamps `Forge-Write: true` on
every commit it authors, and `pkg/dataset/d3_forge_write_test.go` pins the guard both
ways).
**Packaging gap, recorded rather than implied fixed:** nothing in this repo loads agents
from a root-level `agents/` directory — Claude Code loads `.claude/agents/`, and no
plugin manifest exists yet (Phase 0's finding, still true). The four `agents/*.md` files
are correct spec for when packaging exists but are not live, dispatchable agents today;
`skills/forge/SKILL.md`'s dispatch to them is verified today via the generic Agent tool
with an explicit tool allowlist, not live agent auto-discovery.
`testdata/vault/` is a markdown fixture, described below.

Phase 2b's measured actuals, so no later phase re-derives them: `forge index` 0.02s,
`forge drift --since-commit` 0.06–0.07s (budget 100ms, the binding one), `forge check`
0.93s cold / 0.39s warm. Nine reports render deterministically — six consecutive runs,
md5-identical. Against the real vault: drift finds **9 notes referencing changed code**
(2 broken, 7 suspect) over 140 citations; 21 of 94 orphans; 23 graph components; 3
duplicate pairs ≥0.40; 39 of 41 stacks covered. Two knowing deviations: `pkg/gitsig`
shells out to the `git` CLI rather than go-git (**B-009**), and **B-008's IDF weighting
shipped here and fixed neither named case** — the terms carrying a question's meaning were
filtered out of the denominator before any weight was computed. Closed 2026-08-22 by
admitting them; see the Status note above. Do not respond to any of it by moving the
thresholds.

Two Phase 2 decisions that later phases must not undo without reading
`references/recall-spec.md` first: the score is a weighted **mean over active
channels**, not DESIGN §8's literal weighted sum (§2.5), and the title measure is **F₂,
not Dice** (§2.2). Both are argued from measured vault behaviour. The verdict ships
inside `forge recall`'s JSON envelope so nothing downstream restates the threshold tree
— AUDIT §8.4 D-7 moves those thresholds into Phase 3's config chain. Thresholds stay at
DESIGN §5.3's 0.85 / 0.55; the calibration sweep is spec §3.1, which since B-008's closure
is **generated, not transcribed** — `go test ./cmd/forge -run TestCalibration -update`
rewrites `cmd/forge/testdata/calibration.golden` and the diff is the review surface. Any
change touching `pkg/recall`'s scoring must show that diff.

The project ("Knowledge Forge") is a Claude Code plugin that turns "explain X" moments
into permanent, linked, verified markdown notes in an Obsidian vault. Its defensible
core is a Go static-analysis engine that runs with **zero model calls**; an optional
four-tier LLM layer (none / host / API / advisor-critique) sits on top, configurable
per pipeline stage.

## Read the docs in this order

0. **`docs/AUDIT.md` §8.4** — the binding decision record (D-1 … D-8). Read it *first*,
   because the design docs below were deliberately **not** edited: where §8.4 marks a line
   stale, the doc still says the old thing and §8.4 is what you follow. Details under
   "Phase workflow".
1. **`docs/ROADMAP.md`** — condensed index over everything else. Always start here.
2. **`docs/KNOWLEDGE-FORGE-STACK.md`** (ADR-001) — **wins on every stack question.** It
   supersedes ADDENDUM §B (which specified Python — the doc itself says "that was
   wrong") and B2B §8 (which assumed Spring Boot — now an open decision, ADR-002).
3. **`docs/KNOWLEDGE-FORGE-DESIGN.md`** — the master spec (schema, pipeline, gates,
   vault topology, subagents). Its rev-2 note means every `scripts/*.py` reference reads
   as a `forge` subcommand.
4. **`docs/KNOWLEDGE-FORGE-ADDENDUM.md`** — engine tiers (§A), no-AI capability boundary
   and the 10 reports (§B), drift detection (§B.6), weekly checker (§C), datasets (§D),
   full config YAML + presets (§E).
5. **`docs/CLAUDE-CODE-PROMPT.md`** — the actual execution mechanism: a ready-to-paste
   prompt per phase.
6. `docs/KNOWLEDGE-FORGE-B2B.md` — describes a **separate project**, not a phase of this
   one (BACKLOG B-021). Kept in this repo only for reference/history.

Only surviving Python: the one-time `migrate_vault.py` and the offline dataset /
fine-tuning tooling. Neither ships in the binary.

## Things that live outside this repo

- **Vault:** `/Users/mimir45/Documents/Base`, a git repo. **Migrated by Phase 1** on
  2026-08-09: 91 notes moved to DESIGN §7's `notes/<type>/ moc/ _inbox/ _archive/
  profiles/` topology, 345 wikilinks rewritten, 0 broken, 60/91 schema-valid (the 31
  failures are 47 issues needing human judgment — see `lint-report.md` in the vault).
  All seven `notes/<type>/` subdirs exist per B-005. Rollback: backup at
  `/Users/mimir45/Documents/Base-backup-2026-08-09`, or vault commit `b3168f0`.
  `raw/` (5) and `sources/` (9) stay live and outside the note contract; the other old
  topology dirs survive as empty `.gitkeep` shells.
- **v1 skill:** `/Users/mimir45/.claude/skills/til-writer/` — this is the system Phase 0
  audits and this project replaces. The user's global `~/.claude/CLAUDE.md` already
  routes "explain X" prompts into the same vault through it. It contains **only
  `SKILL.md`** — no scripts, no agent definitions, no hooks, no plugin manifest. Phase 0's
  file-map step should expect ABSENT for most rows.
- **The D3 hook is installed and live** in the vault: `.git/hooks/post-commit` runs
  `forge capture` from **`~/.forge/bin/forge`** (the absolute path is pinned in
  `<vault>/.forge/forge-bin`; `$FORGE_BIN` overrides it). That binary is a **copy**, not
  the repo's build output — **rebuild it after any change to `pkg/dataset` or
  `cmd/forge/capture.go`**: `CGO_ENABLED=0 go build -o ~/.forge/bin/forge ./cmd/forge`.
  By design the hook can never fail a commit and never prints, so a stale or broken
  binary is silent: if pairs stop appearing, read `<vault>/.forge/capture.log`. It
  captures nothing today — every migrated note is `origin: import` — and starts paying
  off in Phase 4. Uninstall is `rm .git/hooks/post-commit`.

## Fixture vault (`testdata/vault/`)

A 13-note fixture reproducing the real vault's **pre-migration** topology plus twelve
deliberate defects (F1–F12) — mixed frontmatter shapes, a dangling wikilink, a dangling
`source:` path, an orphan, a near-duplicate pair, notes with no frontmatter at all,
status carried as body prose. Catalogue: `testdata/README.md`.

- Rehearse anything that mutates a vault here first. Phase 1's migration is irreversible
  and the real vault has no backups.
- It has **no `.git`, deliberately** — a nested repo would become a stray gitlink once
  this repo is `git init`-ed. The harness copies the fixture into a temp dir and
  `git init`s the copy; that is how the migration's "refuses a dirty tree" precondition
  and drift's `--since-commit` get exercised. Never `git init` it in place.
- The defects are the test surface. **Do not fix them.**
- It is **not** `examples/vault/`, a separate Phase 6 deliverable (**built**) — 93 files
  generated from the real vault via `forge scrub`, scoped to `notes/`+`moc/` only.

## Phase workflow

`0 → 1 → 2 → 2b → 3 → 3b → 4 → 5 → 5b → 6 → 6b`

This repo's roadmap ends at 6b — B2B (`docs/KNOWLEDGE-FORGE-B2B.md`) is a fully separate
project, not a phase gated inside this one; see BACKLOG B-021. One phase per session. Do not start phase N+1 with phase N unmerged. Never cut 2b; if
time runs out the cut order is `6b → 5b → advisor tier`. If work comes up outside the
current phase's scope, write it to `docs/BACKLOG.md` rather than building it.

**Read `docs/BACKLOG.md` at the start of a phase** — B-002…B-004, **B-029**, **B-031**,
**B-032**, **B-033**, **B-034**, **B-035** and most of the twelve findings 2b recorded are
open; **B-025 is blocked**, not open. B-001 (doc coherence), B-005 (seven note types) and
B-006 (link rewrite) closed on 2026-08-09; B-007 and B-022 in Phase 4; B-009 and B-024 on
2026-08-21, when B-023 and B-027 were also half-closed (docs synced, the behavior/design-doc
halves still open); **B-008 on 2026-08-22**, which opened B-031/B-032/B-033 in its place;
**B-030 in Phase 6b the same day**, which opened B-034/B-035.
**The one still needing its own session is B-029** — re-sized on 2026-08-21 rather than
attempted, and its closing section says what the estimate actually is.

Before touching `pkg/recall`'s scoring, read B-008's closure note and spec §2.3.1. Three
things there are easy to undo by accident: the vocabulary filter applies to `--stack` hints
and **not** to question terms (the reverse looks like the obvious reading and is the bug
B-008 fixed); `idf(0, n) == 0` is correct and its test must not be inverted, because the
absent-term policy lives one layer up in `weightsOver`; and §3.1's table is generated, so a
scoring change that does not update `cmd/forge/testdata/calibration.golden` is unreviewed
rather than harmless.

**Then read `docs/AUDIT.md` §8.** It is the output of that pass: thirteen contradictions
the docs do *not* self-flag, eight resolved by the precedence rule above. **§8.4 is a
binding decision record** (D-1 … D-8) covering the six precedence could not settle. The
design docs were deliberately **not** edited, so where §8.4 marks a line stale the doc
still says the old thing — §8.4 is what you follow. It changes Phase 3, 3b, 6 and 6b:

- **3** — config is a four-layer chain (`FORGE_CONFIG` > `.forge.config.md` > `~/.forge/forge.config.md` > packaged `config/forge.config.example.md`); the schema is the *union* of ADDENDUM §E and the DESIGN §10 keys §E never restates; `forge init` is the **only** writer of `~/.forge/forge.config.md` and `<vault>/profiles/me.md` (rendered from `profiles/me.template.md`) — never `config/forge.config.md`, which stays a packaged template; `skills/forge-init/` asks the questions and shells out.
- **3b** — `on_exhausted` defaults to `queue`; `cost.md` is built here, not in 2b; budget counters live in SQLite and must survive `forge reindex`.
- **2b** — ships **nine** reports, not ten.
- **6** — build `pkg/scrub` / `forge scrub` and use it to generate `examples/vault/`; it needs a fixture test before the phase passes. Ship exactly two ADR files: `0001-lexical-recall-vs-embeddings` (from DESIGN §8) and `0002-go-for-static-core` (from STACK §1).
- **6b** — `--anonymize` calls `pkg/scrub` and **fails closed**; it never exports raw on scrubber error.

## Agent crew (`.claude/agents/`)

The top-level session manages; three project-scoped Sonnet subagents do the work. Route
by verb: **find → `finder`** (read-only search, reports `file:line` hits, also searches
the vault at `/Users/mimir45/Documents/Base`), **do → `executor`** (Read/Write/Edit/Bash;
stays in scope, verifies with real command output), **explain → `explainer`**
(read-only; writes nothing, so TIL notes stay with the `til-writer` skill).

Two more for audit work: **`vault-analyst`** (read-only vault metrics — counts,
frontmatter key frequency, inbound links, orphans, near-dupes) and **`doc-auditor`**
(finds contradictions between the design docs that they don't self-flag — Backlog B-001).

And one competing run: **`cross-checker`** — independently re-derives another agent's
numbers or findings and returns strict JSON, one verdict per claim, each `links`-ed to the
primary's finding ID. Spawn it **in parallel with** the primary, not after: a checker that
has already seen the answer anchors to it. `vault-analyst` and `doc-auditor` therefore end
their reports with a JSON block whose IDs match their prose, so the two runs join
mechanically. Use it when a number is going into a document later phases re-measure
against — the Phase 0 baseline table is the case that motivated it.

Prefer delegating over doing it inline. Independent tasks go out in parallel.

These are **workflow** agents for building the project. They are not the four **product**
agents (`forge-researcher`, `forge-codebase-scout`, `forge-verifier`, `forge-librarian`)
that DESIGN §11 specifies and Phase 4 builds into `agents/`. Deferred to Phase 1, when
there is code to point them at: a `go-reviewer` and a `test-writer`.

## Invariants

Each is stated in a different doc and each is easy to violate by accident:

- The T0 static core makes **zero model calls**. If a design seems to require one, stop
  and ask.
- Stages `recall`, `write`, and `index` accept engine `none` only. On a config that says
  otherwise, **refuse to start with a clear error** — never silently override.
- Drift is git-anchored: post-commit / post-merge / post-checkout, `--since-commit <sha>`.
  Never on file save, never against the uncommitted working tree. Verdicts are a pure
  function of (note refs, tree state) so a revert restores demoted notes symmetrically.
  Demotion history lives in `.forge/`, never in note bodies.
- `CGO_ENABLED=0` for every package except `pkg/codeindex` (go-tree-sitter needs cgo).
- Markdown is the only source of truth. SQLite (`modernc.org/sqlite`, pure Go) is a
  derived cache; `forge reindex` must rebuild it entirely from markdown.
- `pkg/similarity` is hand-rolled MinHash + LSH. **No embeddings.**
- Never auto-mutate the vault on a schedule. Quality-gate failures go to `_inbox/` with
  `confidence: low`, never a silent publish.
- Code verification compiles in a throwaway directory, never in the user's project.
- The advisor tier (T3) is critique-only: it returns disputed claims and a patch, never
  a rewrite.
- Telemetry logs the topic and a hash. Never raw question text, code, or file contents.
- CLI only for v1. Do not build the daemon on speculation — measure first.

## Layout and budgets — all built as of 3b

```
cmd/forge/        CLI
pkg/vault/        frontmatter + markdown AST (goldmark), mtime-cached
pkg/recall/       deterministic question -> note scoring; zero model calls
pkg/similarity/   MinHash + LSH banding
pkg/graph/        note link graph: components, hubs, orphans, centrality
pkg/codeindex/    go-tree-sitter (Java + TypeScript) — the only cgo package, tag-gated
pkg/coderef/      extracts code citations from note bodies and frontmatter
pkg/gitsig/       churn, ownership, co-change coupling — via the git CLI, not go-git (B-009)
pkg/drift/        the key package — ADDENDUM §B.6, AST comparison not line diffs
pkg/linkcheck/    HTTP HEAD on sources, cached, rate-limited
pkg/report/       renders analyses to markdown; must not import pkg/codeindex
pkg/store/        SQLite via modernc.org/sqlite, derived cache only except the budget table
pkg/engine/       none/host/api/advisor backends, per-stage select+fallback, engine_trail
pkg/config/       the four-layer config chain
pkg/sentinel/     id-based begin/end managed comment blocks; Upsert/UpsertBefore/Remove
pkg/scrub/        redacts secret/PII-shaped content from a vault copy; fails closed
```

Latency budgets and the **measured** actuals on an Apple M4: `forge drift` <100ms → **60–70ms**
(the binding constraint — it runs on the git-hook path), `forge index` <200ms → **20ms**,
`forge check` <10s warm → **390ms** (930ms cold). Phase 4 adds two more, measured, since
DESIGN sets no combined gate-pipeline budget: `pkg/qualitygate.Run`'s six in-process
gates minus `code` (schema, citation, freshness, antislop, link, duplicate) →
**~0.13ms** per run, far under the informal sub-100ms target set against `forge check`'s
existing warm figure above; `forge verify-code` per invocation, dominated by toolchain
startup, not gate logic → bash **~10ms warm** (~470ms cold, one-time OS page-cache
effect), java **~170ms warm** (~370ms cold). `tsc` is not installed in this environment,
so the TypeScript lane is untested here — `TestCompileTSSkippedWhenToolchainAbsent`
covers the absent-toolchain path instead. Phase 5 measured its two Claude Code hook
commands the same warm/cold way: `forge session-context` <200ms budget → measured **well
under budget warm** over 20 iterations against synthetic stdin; `forge intent` <50ms
budget → likewise **well under budget warm**, the reuse of `forge recall`'s already-warm
SQLite cache being what makes that budget plausible at all. `hooks/hooks.json` declares
the bindings but nothing in this repo installs it into a live `settings.json` yet (see
the packaging-gap note in Status), so these are direct-invocation measurements, not a
measurement of a live session.

## Commands

`CGO_ENABLED=0 go build ./...` and `go test ./...` both work, and 2b added a `Makefile`:
`make build test bench dist install-hook`. There is still no lint target. Two build lanes,
because `pkg/codeindex` is the one cgo package and is build-tag gated — the default lane is
pure Go and cross-compiles; the codeindex lane needs cgo and a host toolchain. Phases 1, 2,
2b, 3, 3b, 4, 5, 5b and 6's commands ship; the rest is the intended surface, by the phase that creates it:

| Command | Phase |
|---|---|
| `forge slug`, `forge validate`, `forge index`, `forge reindex`, `forge capture` | 1 — **built** |
| `forge recall` (deterministic scoring, JSON, `--explain`) | 2 — **built** |
| `forge drift`, `forge check`, cross-compile + goreleaser | 2b — **built** |
| `forge config` (`--layers`, `--json`), `forge init`, `skills/forge-init/` wizard | 3 — **built** |
| `forge engine select/run/record` — the zero-model-call binary's one named exception | 3b — **built** |
| `forge verify-code` (sandboxed compile check, bash/ts/java), `forge gate` (seven-gate `_inbox/` quarantine) | 4 — **built** |
| `forge session-context`, `forge intent`, `forge session-capture`, `forge cache-source`, `forge stats`, `/forge-check`, `/forge-stats`, git-anchored drift hooks (`scripts/install_drift_hook.sh`) | 5 — **built** |
| `forge logback` (`docs/knowledge-map.md`, per-module `CLAUDE.md` fragments, opt-in inline markers, `--remove-markers`, `--dry-run`) | 5b — **built** |
| `forge scrub <src> <dst>` (redacts secret/PII-shaped content, fails closed) | 6 — **built** |
| `forge export-dataset` (one tier, format matrix, `--since`, anonymized by default, datasheet), `forge dataset-stats`, `/forge-export-dataset`, `/forge-dataset-stats` | 6b — **built** |

## Known discrepancies (record, don't fix)

- The Go module was renamed `TIL` → `knowledge-forge` on 2026-08-08 (bare path, no VCS
  host prefix — deliberately deferred, see BACKLOG B-004). Imports will read
  `knowledge-forge/pkg/vault`. The **directory is still `/Users/mimir45/TIL`** and the
  docs still call the artifact `knowledge-forge/`; that mismatch is cosmetic and stays
  (B-003). Don't rename the directory unasked.
- `docs/CLAUDE-CODE-PROMPT.md` says to put the docs in the repo root; they live in
  `docs/`. Don't shuffle files to match the prompt text.
- BACKLOG **B-005** decided seven note types against DESIGN §7's five-directory tree; all
  seven `notes/<type>/` subdirs now exist in the vault, three of them empty `.gitkeep`
  shells. Don't prune them to match §7.
