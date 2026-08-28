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
item recorded rather than fixed this phase: **B-027** (half-closed 2026-08-21, closed
2026-08-23;
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
new `.golangci.yml`, `errcheck` disabled repo-wide, recorded as **B-029** and re-enabled
2026-08-23) round out the
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
topic slugs and D5's profile values are kept.** They are the semantic and conditioning
features those tiers carry and hashing them makes the corpora untrainable, so anything
spelled as a plain kebab-case name — a topic named after a product, a framework named after
an in-house SDK — survives redaction. Making that observable is also why rendered lines
carry an `id`: without it, path hashing and SHA blanking reached no output format at all. Two items opened rather than built: **B-034** (D6,
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
consequence, plus two `pkg/codeindex` doc comments; ADDENDUM §B.6 / DESIGN §15 kept the
singular name until 2026-08-23, when the doc half closed too). Re-triaged: **B-025** is blocked on observing a live
`PostToolUse`/WebFetch payload, not open work — **do not re-attempt the WebFetch**. New:
**B-030** (`dataset.capture` accepts five tiers but only `d2`/`d4` gate anything; `d3` is
implemented and never reads the list, so removing `d3` silently does not stop capture —
**closed in Phase 6b, 2026-08-22**, by making the control real rather than documenting it).
Two items were re-sized rather than fixed, so the next session does not start on a wrong
estimate: **B-029** is roughly double its recorded scope — **and that re-size was itself
wrong; see the B-029 closure note below for the four measured numbers.** Its triage item 1
landed 2026-08-22 — `cmd/forge/recall_load.go`'s `refresh()` returned `nil` on every path —
but **not the way the entry prescribed**: propagating reaches `runRecall`, which exits 1
without emitting candidates it already scored correctly, so a transient SQLite lock held by
a concurrent `forge intent` would cost the answer. The signature was the defect, not the
missing propagation; `refresh` no longer returns `error` and `writeRows` checks all three
errors the old body dropped. The entry's item 3 (`catfile.go`'s `cmd.Wait()`) was re-traced
in the same pass and its claim is **weaker** than recorded — that re-trace held up, and
`cmd.Wait()` was the sweep's one real fix when B-029 closed 2026-08-23. **B-008's** hidden prerequisite was
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
adjacent-topic query now verdicts CREATE with *zero* neighbours — **closed 2026-08-23**,
see below). **The thresholds did not
move and still must not.** `references/recall-spec.md` §2.3/§2.4 were stale by two changes,
since 2b's IDF weighting was never documented at all; there is now a §2.3.1, a §2.5 note,
a generated §3.1 and a §4.1 example matching real output.

**B-029 and B-027 closed 2026-08-23, same worktree, out-of-phase — two items, one at each
end of `docs/TODO.md`'s queue, sharing no anchor with anything between them.**
B-029 re-enabled `errcheck`: it is off `.golangci.yml`'s `disable:` list and
`golangci-lint run ./...` is clean under the **v1.64.8** `ci.yml` pins. **Every count the
entry carried was wrong, in both directions** — the worklist is **35** (26 test / 9
production), measured with the pinned linter against the repo's own config; the bare
`errcheck` binary reports 105 raw and golangci-lint **v2** reports 50, and neither is this
repo's gate. The finding worth carrying forward is not an errcheck finding at all:
golangci-lint v1 defaults to `max-issues-per-linter: 50` and `max-same-issues: 3`, so the
same tree reports **22** at stock limits — thirteen repeats silently dropped, for every
linter in the default set, for all of Phase 6. Both limits are now `0`. Two of the entry's
three prescriptions did not survive a trace: `stamp()`/`persist()` do **not** need signature
changes (one call each, self-healing failure, and neither caller has an error channel —
unlike `refresh()`, which promised propagation it never performed), and `cmd.Wait()` — the
one real fix — is adopted **only when `drainBlobs` succeeded**, because on a truncated
stream `Wait` reports a broken pipe that would bury the read error the caller needs. Not
re-measured and not claimed: `forge drift`'s <100ms hook budget — `catfile.go` adds no work
(Wait was already called there; only its return value is now read) and the three repos the
vault's cached indexes name are not on this machine.
B-027 closed by **editing the eight design-doc sites**, which is a change of practice worth
knowing about: the standing "record, don't fix" rule and AUDIT §8.4's mechanism govern
**decisions** — a doc line superseded by a later ruling. B-027 is not one; the docs named a
file that never existed on disk under that name, so following the mechanism sent a reader
to a doc that is factually wrong about a path. Correcting a filename is not overriding a
decision, no §8.4 entry was added, and the two normative sites (ADDENDUM §B.6, DESIGN §15)
carry a dated marker so the edit is traceable. **The rule still stands for decisions.**

**B-033 closed 2026-08-23, same worktree, out-of-phase — and it is the case that rule
governs.** The neighbour floor moved **0.30 → 0.125** at both default sites
(`pkg/recall/doc.go`, `config/forge.config.example.md`; `config/presets/` restates neither
threshold, checked). `DESIGN:257`'s "0.3–0.55" was **deliberately not edited** — a decision
superseded by a later ruling is exactly what AUDIT §8.4 governs, unlike B-027's wrong
filename — and the ruling is **D-9**, the first §8.4 entry with no C-number.
The derivation is the point: 0.30 left six of fifteen adjacent-topic queries emitting
**zero** neighbours, and re-deriving against §3.1's nine would have been circular, so
`cmd/forge/testdata/neighbour-labels.txt` is a separate set — fifteen queries, 58 expected
neighbours, written from the corpus file list before any score was measured and **committed
one commit ahead of the sweep that reads them**, so the ordering is checkable in git rather
than asserted. 0.125 is F1's maximum (0.611) and the only swept value that recovers the
Storybook family; measured end to end, that query now emits seven neighbours where it
emitted zero. **Precision 0.548 is a lower bound and was left uncorrected on purpose** —
several counted false positives are defensible links the labels did not name, and
re-labelling after seeing scores is how a derivation becomes a fit.
**The intent gate got the opposite answer, deliberately.** B-033's entry called the two
numbers "one question, not two"; that is true of the root cause and false of the answer,
because `printIntent` interrupts a live session under a never-disturb contract. Measured
against 25 labelled prompts, the old `0.7` admitted **3 of 10** FIRE prompts — but the two
classes separate at 0.402/0.407, a **0.005** margin, so the labels rule 0.7 out and cannot
choose its replacement. The gate is **0.50**, the lowest value still a clear step above the
QUIET ceiling that admits every FIRE prompt whose phrasing tracks a note title (8 of 10),
and `minFireAdmitted` is pinned to exactly 8 so any decay fails the build rather than
passing quietly the way 0.7's did. **`DefaultThresholds.Update` was tried first and
rejected** — the argument for it (below Update the verdict is CREATE, so the message would
contradict the scorer) does not survive contact with the code: `printIntent` computes no
verdict and `emitIntentHit` hedges with "may". It would have cost three near-verbatim title
matches for an alignment nothing in the function asserts. Same failure as
`TestNeighbourBandEdges`, inverted — that test pinned a *number* while claiming to test a
*rule*; this would have pinned the gate to a *rule* that does not govern it. It stays a
plain constant, because `printIntent` runs under the 50ms hook budget and loads no config.
Two things worth carrying forward: `pkg/recall`'s `TestNeighbourBandEdges` had `0.30`
spelled into its fixture and failed as if the band had broken (it now expresses both edges
via `DefaultThresholds` — a test that pins a *number* while claiming to test a *rule* reads
as a real regression), and **B-036** was opened rather than built: three of §3.1's nine
queries now emit ten neighbours because two general Spring notes score on every Spring
question, which no single scalar cut separates. **Do not respond to B-036 by raising the
floor.**

**B-032 closed 2026-08-23, same worktree, out-of-phase, landing right after B-033 per
`docs/TODO.md`'s execution order.** `tagsChannel` and `stackChannel` now activate on
`len(hits) > 0` rather than `len(tags) > 0` / `len(stack) > 0` (`pkg/recall/score.go`), so
a note whose tags or stack don't overlap the query is inactive exactly like a note with
none at all — parity, not the exemption §2.5 argues against deleting. **Two candidate
fixes were computed by hand against the entry's own cited row before either was written,
and the more literal reading of the entry's English (drop `len(tags) > 0` entirely) is the
one that turned out disqualified** — it fixes the row but returns B-008's false positive
(`spring-cli`, 0.415) to first place, which the entry's own step 5 forbids, and it does so
by deleting the untagged-exemption rule the entry says not to delete. The row itself does
not move under the shipped fix — `meterreadingsservice` stays 0.500, `spring-cli` stays
0.415 — matching the entry's own note that the row "is not a regression"; what moves is
every note in between that carried an *irrelevant* tag (128 active tags/stack channels
across the nine calibration queries drop to 84, all 43 that scored a hard 0.000 are gone).
One calibration winner changes: the Docker query's top-1 moves from a note with one real
`docker` tag hit (0.163) to one with none (0.170) — the same shape of trade recurring one
level down, and `weighted`'s own documented behavior (a channel capped low by corpus-wide
term absence still drags an active note below where exclusion would leave it), not a new
mechanism this fix introduces.
**Both of B-033's derivations were re-run against the new scale, not skipped.** The
neighbour floor moved again, **0.125 → 0.150** — F1's peak on the same unedited
`neighbour-labels.txt` sweep shifted to 0.150 (0.578, up from 0.125's now-0.550) — at both
default sites, with a second AUDIT §8.4 entry (**D-10**) recording the supersession over
D-9 the same way D-9 recorded it over `DESIGN:257`. The intent gate's derivation held
*mechanically* — gate stays 0.50, still 8/10 FIRE admitted, still 0 QUIET admitted — but
its measured FIRE/QUIET separation margin went from +0.005 to **-0.036**, opening
**B-037** rather than being nudged: the lowest gate-admitted FIRE prompt and the highest
QUIET prompt now overlap in score between 0.407 and 0.443, and the gate is safe only
because it sits above that whole band, not because the two classes still separate
everywhere the way B-033's derivation described. **Do not respond to B-037 by moving the
gate** — nothing is failing today, and a margin going negative in a slice the gate doesn't
sit inside is a reason to gather more labelled prompts, not to re-derive a number.

**B-023 closed 2026-08-24, same worktree, out-of-phase — the behaviour half, after the doc
half half-closed it 2026-08-21.** Chosen from the three options TODO.md recorded: `stop`
now gets a real non-zero exit; `degrade` stays today's silent fallthrough, unchanged,
because that already is the honest reading of the word. `cmd/forge/engine_run.go`'s
exhaustion check now runs whenever `engine.Resolve` degrades a stage to `"none"` for lack
of budget (previously gated on `rel != "" && OnExhausted == "queue"` alone) and dispatches
in a new `onExhausted` helper: `"queue"` keeps its existing `--rel`-gated
`pending_advisor: true` stamp and falls through to `none`; `"stop"` prints to stderr and
returns exit 1 without calling the tier; any other accepted value is the unmodified
fallthrough. `pkg/engine/select.go:30`'s unconditional degrade is untouched —
`on_exhausted` is still read only one layer up, in `cmd/forge`. New tests
(`TestOnExhaustedBehaviorDiverges`, `TestOnExhaustedStopDoesNotFireWhenBudgetAvailable`)
cover all three values against both the exhausted and non-exhausted branch. The four doc
sites were not touched — they only ever named the three values, never claimed a behaviour
for `stop`, so there was nothing wrong to fix. See BACKLOG.md's B-023 closure note.

**B-031 closed 2026-08-24, same worktree, out-of-phase — the head of the queue closed with
no code change.** Its own TODO.md plan asked for the choice between its two shapes to be
written down before coding; doing that first showed neither survives contact with the
corpus. Shape 1 ("tag the note `kafka`") is not merely non-generalising as BACKLOG framed
it — `kafka` and `consumer` appear nowhere in `testcontainers-docker-based-integration-
testing.md`, title through body, so the tag would misdescribe the note, not under-curate
it. Shape 2 ("the body channel is the only one that sees `kafka` here") is true of the
*system* — `cqrs-and-event-driven-messaging.md` (44 hits) and `transactional-outbox-
pattern.md` (28 hits) carry it heavily — but not of *this row*: both notes sit in
`notes/howto/`, and `bodyPass`'s top-20 window is filled by `notes/concept/*` first purely
because ~84 of 92 docs tie at 0.000 frontmatter and the tie-break sorts by path.
`BodyPassSize` bumped 20→200 locally (not committed) confirmed both kafka-bearing notes
still score below the 0.150 neighbour floor and below the current winner's 0.311 — they
carry zero frontmatter signal, being architecture notes that mention Kafka as an example
rather than testing-infrastructure notes. TODO.md step 2's "measure which one is binding"
resolves to neither: raising `wBody` or `BodyPassSize` moves nothing for this query without
a DESIGN §8 ratio change this one row can't justify. Verdict for this row: today's CREATE,
linking the four testcontainers-family neighbours, is the correct answer to a question the
vault has no dedicated note for — not a coverage miss. **This closes the row, not the
general question** — whether a term the body carries strongly and the frontmatter carries
nowhere should ever lift a candidate is untouched, and the reason it doesn't today (which
of the seven type directories a note lives in decides its place in `bodyPass`'s tie-break,
not its content) reproduces on any query, not just this one. Filed as **B-038**. No golden
diff on B-031 itself; `TestCalibration` unchanged. See BACKLOG.md's B-031 closing section
for the measurement table and B-038 for the general defect.

**B-015 closed 2026-08-24, same worktree, out-of-phase.** `codeindex.File` gained
`Imports []string` (Extractor bumped 2 → 3, same commit as the shape change), extracted in
`parse_cgo.go`'s existing `walk` for both languages; a new `cmd/forge/
check_codebase_deps.go` resolves each import to a repo directory and folds it into
`CodeGroup.DependsOn`, called from both places that build one (`check_codebase.go`,
`logback_map.go`). Two assumptions TODO.md's plan flagged as unverified were checked
against real tree-sitter output before the resolver was written, not assumed: Java's
`import_declaration` exposes the qualified name pre-stripped of `static` and a trailing
`.*` (a wildcard and a class import are the same shape of string, which is why
`resolveJavaImport` trims one dotted segment at a time rather than branching on which kind
it is), and `coderef.ScanRepo`'s file list is filtered to six source extensions, not a full
tree — so TypeScript resolution matches an exact file set, not directory existence, which
closes a false-edge case the plan's own review raised (a nonexistent sibling file
resolving anyway because its parent directory happens to exist for an unrelated reason).
Suffix matches (Java has no known source root to resolve from) pick the
lexicographically-first candidate on a tie, per B-020's determinism rule. Verified: both
build lanes green; `TestGitSourceRebuildsFromScratchOnStaleExtractor` (`pkg/drift`) proves
the real hook path takes the full-rebuild branch on an `Extractor`-mismatched cache rather
than patching a stale entry forward; `TestGroupsOfPopulatesDependsOnEndToEnd` drives a real
temp git repo through `Build → dependsOn → groupsOf` rather than a hand-built `Index`. One
real, one-time, unmeasured cost recorded rather than hidden: the Extractor bump forces a
full re-parse on the first hook run after upgrade on any repo with a persisted
`.forge/code-index-<repo>.json`, against a cache whose whole purpose is avoiding exactly
that under `forge drift`'s hook-path budget — not re-measured on this machine, per B-029's
precedent for repos it doesn't have. See BACKLOG.md's B-015 closing section.

**B-035 closed 2026-08-25, out-of-phase, on `feat/b-035-run-id`.** `telemetry.NewRunID`
mints the correlation key (16 random bytes, hex; no counter, no timestamp semantics) once
per `forge recall` call in `runRecall`. It rides `telemetry.Event.RunID` and `D1Pair.RunID`
(both new, `omitempty`) and the JSON envelope `forge recall` prints — `recall.Result`
itself stays untouched, so `cmd/forge/recall.go` wraps it in a local `recallEnvelope` only
at the emit layer, keeping the zero-model-call scoring package ignorant of dataset capture.
`forge gate` takes an optional `--run-id`; when set and D1's own tier is enabled,
`captureD1Outcome` appends a `D1Outcome{RunID, Published}` record to a **second file**,
`.forge/datasets/d1-outcomes.jsonl`, not a rewrite of the D1 line — that line is already on
disk and immutable by the time the gate call happens, sometimes minutes later in a
different process. The join happens at export: `loadD1` matches the two files by `RunID`
before `since`/`anonymizeAll`/`roundTripAll` run, landing on `D1Pair` as an export-time-only
`Outcome *bool` (nil = never joined, distinguishable from non-nil `false` = joined and
quarantined). `ExportReport.D1Joined` and the D1 datasheet now state the actual join
fraction rather than repeating "no outcome label" verbatim. `--run-id` is optional
everywhere and the empty-id no-op is a tested case, not an afterthought
(`TestCaptureD1OutcomeSkipsOnEmptyRunID`); `skills/forge/SKILL.md` and
`references/recall-spec.md` were updated at both hops. Verified: both build lanes green,
`go vet` clean, new tests in `pkg/telemetry`, `cmd/forge` and `pkg/dataset` covering
minting/uniqueness, the envelope, both outcome branches, both no-op paths, and the export
join in SFT and CSV. Nothing here touches `pkg/recall`'s scoring — `TestCalibration` and
the neighbour/intent-gate goldens are unchanged, no `-update` run. See BACKLOG.md's B-035
closing section. **This session also found and closed a doc-sync gap from the prior
session**: B-031's and B-015's doc closures (`CLAUDE.md`, `docs/TODO.md`) had never been
pushed/merged to `main` — PR #8 squash-merged B-015's code without its doc hunks, and
B-031's/B-003's doc-only commits were never opened as a PR at all. Recovered on
`docs/catchup-b031-b003` (PR #9), cherry-picked from the stranded local commits (code
hunks came back empty — already identical on `main`); this branch is rebased on top of it
so B-035 lands on a correct base once #9 merges.

**B-034 closed 2026-08-25, out-of-phase, on `feat/b-034-d6-dataset` — and the five-vs-six
question this entry opened resolved the opposite way from how it was framed.** ADDENDUM
§D.1's "six" datasets turned out to be right all along; ROADMAP and the phase prompts
outvoting it three sources to one was the wrong frame, since those three are a dated
planning snapshot and historical execution prompts, not living specs — so nothing is
superseded and no AUDIT §8.4 entry is owed, and `docs/ROADMAP.md`/`docs/CLAUDE-CODE-
PROMPT.md` were deliberately left saying five, same treatment B-027 gave a stale doc
line before its own ruling existed. D6 shipped as `pkg/dataset/d6.go`, a **derivation**,
not a sixth capture path: `Tier` gained `Derived bool` (`Path` stays empty for it) and
D6 is a full `Tiers()` member, so `--set d6`, `dataset-stats`, and the tag-uniqueness
test all see it without new plumbing — the cost being that `forge dataset-stats` now
re-derives D6 (glob every `.forge/code-index-<repo>.json`, walk every note, resolve
every citation) on every call rather than reading a file, not measured against a real
multi-repo vault on this machine. The derivation needs no live repo access and no new
`--repo` flag: it builds a `coderef.Registry` straight from the cached code indexes and
resolves citations through the same branch `locate()` (`cmd/forge/check_codebase.go`)
uses, so a citation `forge check` already treats as unresolved contributes no D6 pair
either; symbol lookup goes through `codeindex.File.Lookup` over each index's files in
sorted order, not `Index.FindSymbol`, whose map iteration would pick a different file on
different runs when two files share a trailing member. `loadIndexes` fails closed on a
cache `Load` can't read (stale `Extractor`, corrupt JSON) rather than silently deriving
a smaller corpus — the same reasoning `read.go` already applied to a torn capture line,
carried to a cache instead of a log; an *absent* cache is not an error, since it just
means `forge logback` hasn't run there yet. Anonymization is refused outright rather
than attempted, via a general `refuseDerivedOptions` guard keyed on `Tier.Derived` (so a
future derived tier inherits the refusal) that also refuses `--since`, since a derived
set has no per-record timestamp; both are `UsageError` (exit 2, `--out` untouched, no
`exports.jsonl` line) before a record is read. Both "do not"s from the original entry
held: no `AppendD6`, and no `d6` in `config/forge.config.example.md`'s
`dataset.capture` list. Verified: both build lanes green, `go vet` clean, new
`pkg/dataset` tests (a real code-index-cache-plus-citing-note fixture, dedup, both
refusals, fail-closed on a stale-Extractor cache) and a `cmd/forge` exit-code test.
Nothing here touches `pkg/recall`'s scoring — `TestCalibration` and the neighbour/
intent-gate goldens are unchanged, no `-update` run. See BACKLOG.md's B-034 closing
section. **This branch stacks three deep and is not yet mergeable on its own**:
`docs/catchup-b031-b003` (PR #9) → `feat/b-035-run-id` → `feat/b-034-d6-dataset`. A PR
for this branch must base on `feat/b-035-run-id`, not `main` — a `main`-based PR would
show B-035's diff as if this session had authored it. Merge order is #9, then B-035,
then this one.

**This session also caught a second squash-merge doc-loss, the same class B-035's session
found and fixed for B-015/PR #8.** PR #10 (B-035) and/or PR #11 (B-034) squash-merged
their code cleanly but dropped both PRs' `CLAUDE.md`/`docs/TODO.md` hunks — the closure
paragraphs above and B-034's TODO.md cleanup never reached `main`, though every line of
Go code did. Recovered by restoring both files from `feat/b-034-d6-dataset`'s last known-
good commit (`b2a390c`) onto the `dev` integration branch this session introduced, then
layering B-036's own closure on top. No code was ever at risk; only these two doc files.

**B-036 measured 2026-08-26, out-of-phase, on `dev` — two wrong readings were pushed and
corrected the same day, in two more commits on `dev`, not by rewriting history; the item
stays open.** Built to its own unblock condition: `cmd/forge/neighbour_frequency_test.go`
adds the per-note "appears in N of M query results" column TODO.md named, run over
`testdata/neighbour-labels.txt`'s fifteen queries at the shipped floor (0.150). First
error: the initial read — 5/15 and 4/15 for the two notes this entry named, "no note
comes close to universal," closed as rejected with no code change — used the wrong
denominator. This entry's hypothesis was never "on every query," it was "on every query
**in an ecosystem**" (TODO.md's own wording), and the fifteen labels span several
unrelated ecosystems. It also miscounted the wrong number as evidence for a new item
(**B-039**, since retracted outright): "14 of 15 queries hit `Rank`'s `TopN=10` cap" was
`nonZero`-truncation, not neighbour-floor saturation — the metric that actually counts
full-neighbour emission reads 2 of 15, both of them Spring-flavored, no separate
mechanism. Second error, caught by advisor review one commit later: the "corrected"
denominator (four queries containing the literal word "Spring") was itself picked *after*
seeing which queries the two notes hit — the same failure shape one layer up. Final state
reports both an honest narrow reading (literal "Spring," 4 of 15: both notes 4/4) and a
wide one (+ Maven + Hibernate/JPA, 6 of 15: 5/6 and 4/6) rather than one cherry-picked
number, and the updated unblock condition asks for a pre-committed ecosystem label before
any design, not a post-hoc grep. See BACKLOG.md's B-036 closing note for the full
correction and the corrected harness output (`cmd/forge/testdata/neighbour-frequency.
golden`, now carrying per-slug query lists, not just counts, specifically so this kind of
misread is
checkable from the artifact rather than trusted from prose). **Status: still open** — the
unblock condition is answered, but the design (which has to touch `Rank`'s internal
window, not just `Neighbours`' filter, since `Neighbours` never sees an 11th candidate
`Rank` didn't compute) is deliberately not attempted in the same pass that just corrected
a measurement error.

**B-036 closed 2026-08-26, same day, on `worktree-b-036-neighbour-window` — out-of-phase
work, not a phase.** Its own updated unblock condition (a pre-committed ecosystem label)
landed first, in its own commit: `cmd/forge/testdata/query-ecosystems.txt` labels all
fifteen `neighbour-labels.txt` queries by a mechanical, documented rule, `spring` = 6/15,
matching the entry's own already-disclosed "wide" boundary exactly. The fix widens
`Rank`'s internal window rather than filtering `Neighbours` more cleverly, per the entry's
own framing: `pkg/recall/rank.go` gained `RankPool` (the full nonzero-sorted computation,
un-truncated) and `NeighbourPool = truncate(pool, NeighbourWindow=20)` — a constant kept
separate from `BodyPassSize` even though equal today, pinned by a new test so the two
can't silently drift apart — while `Rank` itself is now `truncate(RankPool(...), TopN)`,
provably byte-identical to before. `Thresholds.Result` became `ResultFrom(q, pool)`:
`Candidates` still truncates to `TopN` (recall-spec.md §4's contract, unchanged),
`Neighbours` now band-filters the wider pool. A candidate alternative — a note-level IDF/
document-frequency exclusion signal, extending §2.3.1's term-level IDF to notes, which is
what the entry's own "shape when picked up" section asked about — was computed by hand
against the real corpus and rejected before being built: it does not separate the two
universal notes from labelled-wanted neighbours, since several wanted notes share their
exact min-idf value and `idfCap` saturates almost every tagged note in an ecosystem at the
same max regardless of specificity. Golden diffs matched a stated prediction, checked
before accepting `-update`'s output rather than after: `calibration.golden`'s 9 rows kept
byte-identical Top-1/score/verdict, only the 3 previously-10-capped rows' Neighbours grew
(bounded ≤20), the original 10 present as a leading subset plus genuinely Spring/Java
additions — two of them `neighbour-labels.txt`'s own labelled-wanted neighbours for an
adjacent query. `neighbour-sweep.golden`'s F1 peak stayed at floor 0.150 (0.578 → 0.587,
still the argmax) — **the floor did not move**, the one place the entry's standing
"do not raise the floor" rule could have been violated by accident.
`neighbour-frequency.golden` showed both predicted effects, not just one: the two
universal notes' overall frequency rose slightly, and a previously crowded-out specific
note (`preauthorize-spring-security-method-level-access-control`) gained admissions on
`spring`-cluster queries. `TestNeighbourBandEdges`, `TestDecideAtThresholdBoundaries` and
`TestIntentGateSeparation` all passed unmodified (no `-update`) — nothing outside the
neighbour band moved. Both build lanes green. Binding invariants held: 0.85/0.55
untouched, `score.go` untouched (`idf(0,n)==0` intact), `neighbour-labels.txt` unedited,
`candidates`' at-most-10 contract unchanged, the 0.150 floor not raised, `BodyPassSize`
not moved. See BACKLOG.md's B-036 closing section for the full measurement and diff
detail.

**B-025 closed 2026-08-28, same worktree, out-of-phase — the unblock condition was
observation, not another documentation lookup.** The three earlier attempts had all
WebFetched Claude Code's own doc pages looking for a written schema; the entry's own
unblock trigger asked for a live hook payload instead, which is a different action. A
throwaway diagnostic `PostToolUse`/`WebFetch` hook (`cat > /tmp/webfetch-payload.json`)
could not be wired from inside the session it was meant to observe — the auto-mode
classifier correctly refused that settings-file write as self-granting — so a human added
it to the main checkout's `.claude/settings.json` (not the worktree's), and it fired
anyway on the next live `WebFetch` call: project-level hook config does not care which
worktree a session is cd'd into. Confirmed shape: `tool_response` is an object
`{result, url, code, codeText, bytes, durationMs}`; `result` carries the fetched text —
neither of the entry's own guessed field names (`content`, `text`) was right. `cacheBody`
(`cmd/forge/cache_source.go`) now decodes into `map[string]json.RawMessage` and looks up
`"result"` by key, deliberately not a fixed struct, so a present-but-empty result stays
distinguishable from an absent one; two fallbacks (bare string, raw bytes) are unchanged in
spirit for any shape that isn't this one. Two new tests pin the real shape and the
empty-result edge case; all four existing `cache_source_test.go` tests pass unmodified.
Both build lanes green. See BACKLOG.md's B-025 closing section.

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

**Read `docs/BACKLOG.md` at the start of a phase** — B-002…B-004, **B-037**, **B-038** and
most of the twelve findings 2b recorded are open. **B-025 closed 2026-08-28** (see the
Status note above) — it was blocked, not open, until then. B-001 (doc coherence), B-005 (seven note types) and
B-006 (link rewrite) closed on 2026-08-09; B-007 and B-022 in Phase 4; B-009 and B-024 on
2026-08-21, when B-023 and B-027 were also half-closed (docs synced, the behavior/design-doc
halves still open); **B-008 on 2026-08-22**, which opened B-031/B-032/B-033 in its place;
**B-030 in Phase 6b the same day**, which opened B-034/B-035; **B-029, B-027, B-033 and
B-032 all on 2026-08-23** (see the notes below), B-033 opening **B-036** and B-032 opening
**B-037** in their place; **B-023's behaviour half on 2026-08-24**, closing it fully;
**B-031 also on 2026-08-24**, closed with no code change and opening **B-038** in its
place (see the Status note above). **B-015 also on 2026-08-24**, populating
`CodeGroup.DependsOn` (see the Status note above). **B-035 on 2026-08-25**, minting the
`run_id` correlation key, and **B-034 the same day**, building D6 as a derived export
view over `forge logback`'s map rather than a sixth capture tier (see the Status note
above). **B-036 measured, then closed, both 2026-08-26** — unlike B-031, its hypothesis
*did* survive measurement once read against the right denominator (a corrected reading,
after a wrong one briefly landed the same day), and the widened-`Rank`-window fix shipped
the same day once its ecosystem-label prerequisite landed (see the Status note above).
**`docs/TODO.md`'s PLANNED class stays empty** — B-037 and B-038 are NO STEPS by their own
argument, each naming a measurement to run before any design is chosen; so nothing in
BACKLOG's open list currently has a six-field plan to execute; the next phase or
out-of-phase item starts by writing one.

**`docs/TODO.md` is the execution half of that file** (written 2026-08-23). BACKLOG records
*why* an item exists; TODO records *how to close it* — a six-field plan (anchors,
prerequisites, steps, verification, done-when, and an explicit "do not") for each workable
item, an index row for all 37, and an unblock condition rather than steps for the items
open but not actionable. It also fixed the execution order in advance: **B-033 had to land
before B-032**, because B-032 moves `blend`'s denominator and would shift the scale B-033
re-derives against. It did, both on 2026-08-23 — B-032 owed a re-run of B-033's two
derivations, not just a re-recorded calibration golden, and both ran
(`TestNeighbourFloorSweep`, `TestIntentGateSeparation` — see the B-032 closure note above
for what each said).

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
