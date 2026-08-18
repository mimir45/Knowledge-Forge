# 0002 — Go for the T0 static core

- **Status:** Accepted
- **Source:** `docs/KNOWLEDGE-FORGE-STACK.md` §1 ("ADR-001 — Go for the T0 static
  core"), per `docs/AUDIT.md` §8.4 D-3. STACK.md's own doc numbers this ADR-001; this
  repo's `docs/adr/` sequence numbers it 0002 so the file order matches AUDIT D-3's
  ordering, not STACK's internal numbering.

## Context

The T0 static core (`forge recall`, `forge drift`, `forge check`, the quality gates) is
an invariant of this project: it must run with **zero model calls**, and it sits on a
git hook path, so it has real latency budgets, not aspirational ones:

| Command | Budget | Where it runs |
|---|---|---|
| `forge drift` | <100ms | post-commit / post-merge / post-checkout hook |
| `forge index` | <200ms | after most writes |
| `forge check` | <10s warm | weekly full pass |
| `forge session-context` / `forge intent` | <200ms / <50ms | Claude Code lifecycle hooks |

The original spec (ADDENDUM §B) assumed a Python implementation. STACK.md's own text
says plainly: "that was wrong."

**The problem with Python:**
- Startup latency alone threatens the git-hook budget before any real work happens.
- A hook-installed dependency (PyYAML, a markdown parser, a tree-sitter binding) means
  asking every user to manage a Python environment just to get a commit through.
- No credible path to a single, dependency-free binary a user can just download and run.

## Decision

Go, for the entire T0 static core. `CGO_ENABLED=0` for every package except
`pkg/codeindex` (go-tree-sitter needs cgo), so the default build lane stays pure Go and
cross-compiles cleanly; the codeindex lane needs cgo and a host toolchain, kept as a
second, separate build lane rather than forcing cgo on everything that doesn't need it.

**What Go buys:**
1. A single static binary, `go build`, no runtime dependency to install alongside it —
   the whole reason `bin/forge` can be a thin hash-verifying shim instead of a language
   runtime bootstrapper.
2. Startup and steady-state latency low enough that the git-hook budgets above are
   comfortably hittable, not a stretch goal.
3. A standard library good enough for most of this system's actual work (`encoding/
   json`, `regexp`, file walking, a pure-Go SQLite driver in `modernc.org/sqlite`)
   without reaching for a dependency for every piece.
4. Cross-compilation as a built-in feature, not a bolted-on toolchain — this is what
   makes the Makefile's six-target cross-compile matrix and `.goreleaser.yml`'s release
   pipeline straightforward rather than a project of their own.

**Alternatives considered:**

| Option | Verdict |
|---|---|
| Python | Rejected — see "The problem with Python" above. |
| Rust | Considered seriously — matches or beats Go on raw latency and produces an equally dependency-free binary. Rejected for this project on iteration speed and ecosystem maturity for this domain (markdown/YAML parsing, a pure-language SQLite driver): Go's stdlib and `modernc.org/sqlite` get to a working T0 core faster, and the latency budgets above don't actually require Rust's ceiling — Go's measured actuals (see Phase 2b, below) clear every budget with room to spare. |
| Node/TypeScript | Rejected — startup latency and dependency-tree weight work against the same git-hook budget problem Python has, without Python's story-telling advantage of "everyone already has it." |

## Consequences

- Two build lanes exist because of the one cgo package (`pkg/codeindex`), not because Go
  itself needed splitting: `CGO_ENABLED=0 go build ./...` for the default, pure-Go lane;
  a second, cgo-enabled lane for the tree-sitter-backed code index. A portable release
  binary built from the pure-Go lane has no code index — `.goreleaser.yml` documents
  this trade-off directly at the point it matters.
- STACK §1's original text mentions a `pkg/forge/...` library-package prefix that never
  shipped — this repo's actual layout is the flat `pkg/vault`, `pkg/recall`,
  `pkg/graph`, etc. tree named in `CLAUDE.md`'s package map, not a `pkg/forge/`
  subtree. This ADR records the *decision to use Go*, not that one naming detail; the
  mismatch is a known, stale artifact of the source doc, not a correction this ADR
  makes.
- The measured actuals bear the decision out, not just the intent: Phase 2b clocked
  `forge drift --since-commit` at 60–70ms against a 100ms budget (the binding
  constraint, since it runs on the hook path), `forge index` at ~20ms against 200ms, and
  `forge check` at 390ms warm / 930ms cold against a 10s budget. None of this required
  tuning beyond ordinary Go — no budget was a near-miss.
- `pkg/gitsig` shells out to the `git` CLI instead of using a go-git library (**B-009**,
  open, kept as a known, accepted deviation) — one place this ADR's "no external runtime
  dependency" ideal is not quite absolute, since it assumes `git` is already on PATH,
  which every environment this tool runs in already satisfies.
