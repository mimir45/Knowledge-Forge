---
# ─────────────────────────────────────────────────────────────────────────────
# Knowledge Forge — packaged defaults.
#
# This file is the LOWEST layer of the configuration chain and is never edited by
# users. `forge init` copies what you answer into ~/.forge/forge.config.md; edit
# that, or a .forge.config.md at the root of a project, instead.
#
# Precedence, highest wins:
#   1. $FORGE_CONFIG                     one-run override, must exist if set
#   2. <project>/.forge.config.md        per-repo settings, committed
#   3. ~/.forge/forge.config.md          your settings, written by `forge init`
#   4. this file                         compiled into the binary
#
# Merge rule: maps merge key by key, scalars and lists replace wholesale. So
# `static: {languages: [go]}` in a project gives you exactly [go], not [go] plus
# everything below — you can always narrow a list.
# ─────────────────────────────────────────────────────────────────────────────

vault_path: ""             # `forge init` fills this; empty falls back to --vault
repo_path: auto            # auto = the git repo the command was run in

paths:
  notes: notes
  moc: moc
  inbox: _inbox
  archive: _archive
  index: _index.md

trigger:
  mode: ask                # ask | auto | manual
  # ask    = offer to write, wait for accept
  # auto   = write without asking when confidence is high
  # manual = only on an explicit /forge

# ── recall ───────────────────────────────────────────────────────────────────
# DESIGN §5.3's decision tree.
#   score >= answer_threshold, fresh  -> ANSWER_FROM_VAULT
#   score >= answer_threshold, stale  -> UPDATE(refresh)
#   score >= update_threshold         -> UPDATE(extend)
#   otherwise                         -> CREATE, linking every neighbour
#                                        at or above neighbour_min_score
#
# neighbour_min_score is the one of the three that has moved, re-derived twice against a
# labelled query sweep as the scoring blend itself changed (see
# references/recall-spec.md). Raise it for fewer,
# surer links; lower it for a denser graph. The other two are DESIGN §5.3's and should not
# be touched to paper over a recall scoring gap — re-derive the calibration table instead.
recall:
  strategy: lexical        # lexical | hybrid (hybrid is a v2.2 upgrade, not built)
  answer_threshold: 0.85
  update_threshold: 0.55
  neighbour_min_score: 0.150

# Days before a note of each type is considered stale. 0 = never stale.
freshness_days:
  concept: 365
  howto: 180
  api: 90
  pattern: 365
  pitfall: 365
  incident: 0
  decision: 0              # decisions are superseded, not refreshed

# ── engines ──────────────────────────────────────────────────────────────────
# The four tiers. `none` is the static Go core and makes zero model calls.
engines:
  default: host            # none | host | api | advisor

  api:
    provider: anthropic
    model: claude-sonnet-5
    key_env: ANTHROPIC_API_KEY   # the variable name, never the key itself
    base_url: ""

  advisor:
    model: claude-opus-5
    mode: critique         # critique-only: returns disputed claims and a patch,
                           # never a rewrite

  local:
    enabled: false
    model: ""

  budget:
    advisor_usd_per_day: 2.00
    api_usd_per_day: 1.00
    on_exhausted: queue    # queue | degrade | stop
    # queue defers paid work to the next window. It cannot stall a git hook:
    # this governs T2/T3 spend only, and drift/index/recall are locked to T0.

  routing:
    advisor_when:
      type: [decision, pattern]
      confidence_below: medium
      stack_in: [security, auth, payments]
      user_flag: "--deep"

# ── pipeline ─────────────────────────────────────────────────────────────────
# Which engine runs each stage. Three of them are LOCKED to `none`: recall,
# write and index are the static core, and the binary refuses to start if any
# layer assigns them anything else.
pipeline:
  intake:     {engine: host}
  recall:     {engine: none}                              # locked
  plan:       {engine: host}
  research:   {engine: api, fallback: host}
  synthesize: {engine: host}
  verify:     {engine: advisor, fallback: local, then: host}
  write:      {engine: none}                              # locked
  link:       {engine: none}
  index:      {engine: none}                              # locked

research:
  max_sources: 6
  prefer: [official-docs, source-code, rfc]
  allow_domains: []
  deny_domains: []
  use_docs_mcp: true       # a context7-style docs MCP, if one is available
  scan_codebase: true      # ground the note in the repo you are in

verify:
  run_code: auto           # auto | never | ask — always in a throwaway directory
  require_citation_for: [version-specific, performance-claim, security-claim]
  duplicate_threshold: 0.40  # write-time gate trigger — separate from check.duplicate_threshold
                              # below (a report threshold); see references/duplicate-spec.md §6

write:
  language: en
  max_note_words: 1200
  diagrams: mermaid        # mermaid | ascii | none

# ── static core ──────────────────────────────────────────────────────────────
# Everything here runs with zero model calls.
static:
  code_index: true
  languages: [java, kotlin, python, typescript]
  git_signals: true
  cache_ttl_days: 30      # forge cache-source's .forge/cache/<hash>.md TTL (Phase 5)

  drift:
    enabled: true
    trigger: git           # git only. Never on file save, never on the
                           # uncommitted working tree.
    branch: default
    auto_repair_line_numbers: true
    on_broken: demote
    on_restored: undemote

  linkcheck:
    enabled: true
    timeout_s: 5

  logback:
    knowledge_map: true
    claude_md_fragment: true
    inline_markers: false  # opt-in: writes `// forge:` markers into your source

# ── weekly check ─────────────────────────────────────────────────────────────
check:
  enabled: true
  schedule: "0 9 * * MON"
  ai_pass: false
  reports: [coverage, staleness, duplicates, orphans, gaps, graph-health, churn, deadlinks, drift]
  drain_advisor_queue: true
  churn_days: 90           # git history window for the churn report
  duplicate_threshold: 0.40  # MinHash Jaccard above which a pair is reported

garden:
  enabled: true
  schedule: weekly

# ── dataset ──────────────────────────────────────────────────────────────────
dataset:
  enabled: true               # the master switch; false stops all five tiers
  capture: [d1, d2, d3, d4, d5]
                              # every tag here gates a real write path — remove one and
                              # that tier stops capturing. d1 routing (forge recall),
                              # d2 advisor critiques, d3 human edits (the vault's
                              # post-commit hook), d4 gate repairs, d5 accepted notes.
  anonymize_on_export: true   # --anonymize fails closed; it never exports raw
                              # when the scrubber errors

telemetry:
  enabled: true
  scope: local             # local | team
  # Logs the topic and a hash. Never raw question text, code, or file contents.
---

# What is configurable, and what deliberately is not

Phase 3 swept the repo for hardcoded values. Everything that is a *user preference*
moved into the frontmatter above. Three groups stayed in code, and this section is the
record of why, so a later phase does not "finish the job" by moving them.

## Moved here

| Was | Now |
|---|---|
| `pkg/recall.DefaultThresholds` — 0.85 / 0.55 / 0.30 | `recall:` |
| `cmd/forge/check.go` `-days 90` | `check.churn_days` |
| `pkg/similarity.DuplicateThreshold` 0.40 | `check.duplicate_threshold` |
| `--vault` default `"."` in eight subcommands | `vault_path` |
| `_index.md`, `notes/`, `moc/`, `_inbox/`, `_archive/` | `paths:` |
| `pkg/codeindex`'s java / typescript switch | `static.languages` |
| `pkg/vault.FreshnessDefault`'s per-type table | `freshness_days:` |

## Stayed in code, on purpose

- **`pkg/vault`'s `excludedPrefixes` and `hubNames`** (`raw/`, `sources/`, `_archive/`,
  `reports/`; `index`, `_index`, `log`, `readme`, `home`). These define what *is a note*,
  not what a user prefers. A vault where `reports/` counts as notes fails `forge check`
  in a way no setting should be able to cause.
- **`pkg/report/duplicates.go`'s `specThreshold = 0.85`.** It is printed as documentation
  of what DESIGN §8 asked for and compared against what actually ships; it is never
  applied. Wiring it to config would make the report agree with itself by construction
  and stop being evidence.
- **`$HOME/.forge/bin` in the Makefile.** That is the install location, not a preference;
  `$FORGE_BIN` already overrides it at run time.

## Not a hardcoded value anywhere in this repo

There is no `/Users/...` path, no author name, and no `os.Getenv` outside this package in
non-test Go source. The personal-path hardcoding this phase was written to remove did not
exist; what existed was defaults and taxonomy, listed above.
