# Knowledge Forge — Stack Decisions

Fourth document. Revises the implementation language choices in
`KNOWLEDGE-FORGE-ADDENDUM.md` (§B specified Python — that was wrong) and
`KNOWLEDGE-FORGE-B2B.md` (§8 assumed Spring Boot — that's now an open decision).

**Summary of the change:** the T0 static core moves to **Go**. Not for raw compute
speed — for **startup time and distribution**. Those are the two constraints that
actually bind, and Python fails both.

---

## 0. The decision table

| Component | Language | Decisive reason |
|---|---|---|
| **T0 static core** (parse, similarity, code index, git, drift, reports) | **Go** | ships to laptops; must start in <10ms; single static binary, zero runtime deps |
| Skill / prompt layer | Markdown | — |
| Hook scripts | **Go** (thin wrappers) or `sh` | must not add >100ms to Edit/Write |
| Engine abstraction (T1–T3 routing) | **Go** | lives inside the core binary; HTTP + JSON is not a reason to add a runtime |
| B2B ingestion workers | **Go** | I/O-bound, high concurrency, and **reuses the core as a library** |
| B2B API / MCP server | **Go** (default) / Java+Spring (if a customer demands it) | see §6 — deferred to B1 |
| Vector store | Postgres + pgvector | unchanged; self-hostable, one system |
| Dataset tooling & fine-tuning | **Python** | no alternative, and it never ships to a user machine |
| Evals | **Go** for harness, YAML for cases | runs in CI next to the core |

**The organizing principle: language is chosen by deployment boundary, not by
preference.** Anything that lands on a developer's machine is Go. Anything that runs
offline in your own control can be whatever's fastest to write.

---

## 1. ADR-001 — Go for the T0 static core

### Context

The core runs in four places with different constraints:

| Context | Frequency | Latency budget |
|---|---|---|
| git post-commit/merge/checkout hook (incremental drift, rev 2 — addendum B.6) | every commit/merge/checkout | **<100ms total** |
| `SessionStart` hook (index injection) | every session | <200ms |
| `/forge-check` weekly, full vault + repo analysis | weekly | <10s |
| CI / pre-commit (validation, drift) | every push | <5s |

The first row is the binding constraint. A hook that adds 300ms to every Edit is a
hook users disable, and a disabled hook means the whole continuous-drift feature
doesn't exist.

### The problem with Python

- **Startup.** A Python process that imports PyYAML + tree-sitter + datasketch costs
  ~80–200ms *before executing a line of your code*. That alone blows the hook budget.
- **Distribution.** The plugin would require Python 3.11+, a venv or `--break-system-packages`,
  and native compilation of tree-sitter bindings. Every one of those is a GitHub issue
  waiting to happen, on a tool whose entire pitch is "install it in 5 minutes."
  This is how good plugins die — not on features, on install friction.
- **Concurrency.** The GIL is a real ceiling on parsing a few thousand files.

### The decision

Go, compiled to a single static binary per platform, shipped in the plugin's `bin/`
directory — which the plugin spec explicitly supports: *executables in `bin/` are
added to the Bash tool's PATH and invokable as bare commands while the plugin is
enabled.* That directory exists for exactly this.

What Go buys, in order of importance:

1. **~1–5ms startup.** The hook budget becomes achievable.
2. **Zero runtime dependencies.** No Python, no node, no venv. `forge drift
   --since-commit <sha>` just works.
3. **Cross-compilation** to darwin/linux/windows × amd64/arm64 from one machine.
4. **Real parallelism** — goroutines over a worker pool for file parsing, which is
   embarrassingly parallel.
5. **Static typing on a schema-heavy codebase.** Frontmatter parsing, AST walking and
   graph algorithms are exactly where Python's dynamism costs you.

### Alternatives considered

| Option | Verdict |
|---|---|
| **Python** | Rejected: startup + distribution, as above. |
| **Rust** | Genuinely tempting — tree-sitter is Rust-native, so no FFI, and it'd be faster still. Rejected on solo-project velocity: compile times and the borrow checker tax on graph/cache code aren't repaid here, since the workload is I/O-bound and Go is already inside budget. Revisit only if profiling says otherwise. |
| **TypeScript / Node** | Rejected: needs a runtime on the user's machine, and ~30–50ms startup is uncomfortably close to the budget. |
| **Java** | Rejected for this component: JVM startup (even with AOT) and a JRE dependency are disqualifying for a laptop-side binary. Not a criticism of Java — wrong deployment shape. |

### Consequences

- The core becomes a **library** (`pkg/forge/...`) with a thin CLI on top, so the B2B
  server imports it directly rather than shelling out or reimplementing drift
  detection. This is the single biggest downstream benefit and it's why §6 leans Go.
- You'll write real Go. Budget ~1 week of ramp before you're fast. Worth it.
- **cgo is a landmine** — see §3.

---

## 2. Library choices

| Need | Choice | Note |
|---|---|---|
| Code parsing | `go-tree-sitter` + per-language grammars | **cgo** — see §3 |
| Git | `go-git` (pure Go) | avoids shelling out; keeps cross-compilation clean |
| Frontmatter/YAML | `goldmark` + `gopkg.in/yaml.v3` | goldmark also gives you a real markdown AST, which you need for safe section-level edits |
| Cache / index | **SQLite via `modernc.org/sqlite`** | **pure Go, no cgo.** Replaces the `.forge/state.json` design |
| Full-text / BM25 | `bleve` (pure Go) or hand-rolled | bleve is probably overkill at vault scale; use it for the B2B events corpus |
| MinHash / LSH | hand-rolled (~150 LOC) | no dominant Go lib and the algorithm is small; keeps deps down |
| CLI | `cobra` or stdlib `flag` | stdlib is enough |

### On replacing `state.json` with SQLite

The addendum specified a JSON cache file. Reconsider: SQLite (pure-Go driver, so still
one static binary) gives you incremental updates without rewriting the whole file,
concurrent readers (hooks firing while `/forge-check` runs), indexed lookups, and
sane growth to hundreds of thousands of rows for the B2B case.

It does **not** violate the "plain markdown is the source of truth" principle, because
the DB is strictly derived — `forge reindex` rebuilds it from the markdown in seconds.
Put that sentence in the docs; it's the same argument that defuses the vector-store
objection in the B2B security review.

---

## 3. The cgo problem — flag this now, it will bite on day one

`go-tree-sitter` uses cgo. The moment cgo is enabled you lose Go's clean
cross-compilation: `GOOS=darwin GOARCH=arm64 go build` stops working from a Linux box
without a cross toolchain, and you get a dynamic libc dependency.

Three ways out, best first:

1. **`zig cc` as the cross-compiler.** `CC="zig cc -target aarch64-macos"` makes cgo
   cross-compilation work from one machine. Well-trodden, ~an afternoon to set up.
2. **CI matrix build.** GitHub Actions on `ubuntu`/`macos`/`windows` runners, each
   building natively, artifacts attached to the release. Simplest, and you need the
   release workflow anyway.
3. **Avoid cgo entirely** — use a pure-Go parser for a reduced language set. Only
   worth it if the toolchain becomes a recurring tax.

Take (2) for correctness now and (1) for local iteration speed. Either way, statically
link and set `CGO_ENABLED=0` for every package that doesn't need tree-sitter, so only
the parsing path carries the constraint.

---

## 4. Distribution: getting the binary onto the machine

Two options, and the tradeoff is real:

| Approach | Pro | Con |
|---|---|---|
| **Commit binaries to `bin/`** | install just works, offline, no postinstall | ~6 binaries × 15–25MB = repo bloat, ugly git history |
| **Download on first run** | tiny repo | needs network at install, needs checksum verification, one more failure mode |

**Recommendation:** a small `bin/forge` shell shim that checks for the real binary in
`~/.forge/bin/<version>/`, and if absent downloads it from the GitHub release for the
detected platform, **verifying a SHA-256 pinned in the repo**. Repo stays small,
install stays one step, and the checksum pin means a compromised release can't
silently ship. Document the offline path (`FORGE_BIN=/path/to/binary`) for airgapped
users — who are exactly the audience for the `offline` preset.

---

## 5. The hot path: CLI vs daemon

Even at 3ms startup, the `PostToolUse` hook has to re-open SQLite and re-resolve
config on every file edit. Fine at 20 files, wasteful at scale.

**Phase it:**

- **v1 — CLI only.** `forge drift --since-commit <sha>` per invocation. Composable,
  trivially testable, easy to debug. Ship this. Measure it. It is very likely already
  under 100ms and you should not build a daemon on speculation.
- **v2 — optional daemon**, only if measurement says you need it: `forge daemon`
  listening on a unix socket in the vault dir, holding the parsed index warm; the hook
  becomes a ~1ms socket write. Must degrade to direct CLI execution if the socket is
  absent, so there's never a state where a stale daemon breaks the tool.

Writing "measure before building the daemon" into the plan is worth more than the
daemon. It's also a better interview answer than having built one.

---

## 6. ADR-002 — B2B backend language (deferred to B1, leaning Go)

You said no need to stick with Spring. Agreed that it shouldn't be automatic. Here's
the honest weighing rather than a preference.

**The decisive argument is code reuse.** The B2B plan's biggest leverage is
*"B1: reuse the T0 core as-is"* — the same drift detection, code index, similarity
and reporting that runs on a laptop also runs server-side over the org's repos. If the
core is Go and the server is Java, you cannot import it. You'd shell out to a binary
(awkward, but workable) or maintain two implementations of drift detection (fatal —
that's the component with the most subtle logic in the whole system).

| | Go | Java + Spring Boot |
|---|---|---|
| Reuses T0 core as a library | ✅ directly | ❌ shell out or reimplement |
| Ingestion concurrency | goroutines, low memory | virtual threads (21+) — genuinely fine now |
| Deployment | ~20MB static binary, tiny container | JVM, heavier, slower cold start |
| pgvector support | `pgx` + raw SQL, straightforward | Spring AI `VectorStore` abstraction, nicer |
| Your velocity today | slower (learning) | fast |
| Enterprise procurement optics | neutral | slightly positive in Java shops |
| Solo-project ceremony | low | high |

**Recommendation:** default to **Go**, decide for real at B1. Two reasons the decision
can safely wait: B0 gates B2B behind OSS traction anyway, and by the time you reach
B1 you'll have written the entire core in Go and will actually know whether you like
it. Deciding now, before that information exists, is the mistake.

**One caveat worth naming:** if you're targeting Java/Spring roles specifically, a
Spring Boot service in the portfolio has direct interview value. But that's an argument
for keeping *a* Spring project, not for making this one Spring against its grain. The
polyglot version is the stronger story — see §8.

---

## 7. Where performance actually needs attention

Not everywhere. The profile is I/O-bound with three real hotspots:

| Hotspot | Design |
|---|---|
| **Parsing N source files** | worker pool sized `GOMAXPROCS`; cache by `path+mtime+size` in SQLite; only re-parse dirty files. This is where Go's parallelism earns its place. |
| **Similarity over all note pairs** | naive is O(n²) — at 1k notes that's 500k comparisons, still fast, but it won't hold at B2B scale. Use **MinHash + LSH banding** so you only compare candidate pairs. Design it right the first time; retrofitting LSH is annoying. |
| **Drift on codebase change** | **git-tree anchored** (rev 2): triggered by post-commit/merge/checkout, never by file saves. Diff `last_checked_sha..HEAD` with `--name-only` as the cheap gate; AST comparison only on files in that set. Verdicts are a pure function of tree state, so reverts/rollbacks restore demoted notes symmetrically (addendum §B.6). Never re-scan the vault on a hook. |

Explicitly *not* worth optimizing: report rendering, YAML parsing, the index rebuild.
They're weekly or per-session and already fast. Profile before touching them.

---

## 8. What this does to the CV story

The polyglot split is a better story than either monoglot version, and it's better
specifically because each choice has a *constraint* behind it rather than a taste:

> Go for the analysis core — it ships to developer machines as a static binary and
> runs in an edit hook, so startup time and zero-dependency install were hard
> constraints. Python only for the offline ML tooling, where it never touches a user
> machine. Postgres/pgvector for the server because self-hosting was a requirement.

That's three decisions, each traceable to a constraint. It's the shape of answer that
distinguishes someone who chose a stack from someone who defaulted to one — and it
survives the follow-up question, which "I used Spring because I know Spring" does not.

Keep §1's "Alternatives considered" table as `docs/adr/0002-go-for-static-core.md`.
The rejected-Rust paragraph does more for you than the accepted-Go one, because it
shows you can decline the more impressive option for a stated reason.

---

## 9. Roadmap deltas

| Phase | Change |
|---|---|
| 1 — Contract | validator + index rebuild move to Go. Migration script can stay a throwaway Python script — it runs **once**, ever, and doesn't ship. |
| **2b — Static core** | now Go. Add ~1 week for language ramp + the cgo/release toolchain. Biggest schedule change. |
| 3b — Engines | Go; HTTP clients for T2/T3 are stdlib. |
| 5 — Hooks | hooks call the Go binary directly. Measure the Edit-hook cost and publish the number in the README — "adds 40ms to file writes" is a claim that builds trust. |
| 6 — Package | + goreleaser, CI build matrix, checksum-pinned install shim, `CGO_ENABLED` hygiene. |
| 6b — Datasets | stays Python. Offline only. |
| B1 | ADR-002 decided for real, with a working Go core as evidence. |

Net: **+1 to 1.5 weeks** versus the Python plan. Worth it — the Python version would
have paid that back in install-failure issues within the first month of the OSS
release.

---

## 10. Prompt for Claude Code

Replaces the Phase 2b prompt in `CLAUDE-CODE-PROMPT.md`.

```
Read KNOWLEDGE-FORGE-ADDENDUM.md section B and KNOWLEDGE-FORGE-STACK.md in full.

Build the T0 static core in GO, not Python. Single static binary, shipped in the
plugin's bin/. Everything here is deterministic — no model calls anywhere in this
code. If you find yourself wanting one, stop and tell me instead.

Layout:
  cmd/forge/            CLI (cobra or stdlib flag)
  pkg/vault/            frontmatter + markdown AST (goldmark), mtime-cached
  pkg/similarity/       MinHash + LSH banding, hand-rolled. NO embeddings.
  pkg/graph/            note link graph: components, hubs, orphans, centrality
  pkg/codeindex/        go-tree-sitter for java/kotlin/python/typescript;
                        pom.xml / build.gradle / package.json -> dep + version map
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
- Drift is GIT-ANCHORED per addendum B.6: `forge drift --since-commit <sha>`,
  installed as post-commit/post-merge/post-checkout git hooks — never on file save,
  never on uncommitted working-tree changes. Diff sha..HEAD --name-only as the
  cheap gate; AST comparison only on files in that set. Record drift_checked_at
  per note. On revert/reset, restore demoted notes symmetrically — verdicts are a
  pure function of tree state. Never re-scan the vault.
- Table-driven tests with fixtures for every package. Benchmarks for
  parse, similarity, and drift.
- `forge --help` for every subcommand.

Performance targets — measure and report actuals, do not assume:
  forge drift --since-commit <sha> < 100ms   (this is the binding constraint)
  forge index                      < 200ms
  forge check (full vault + repo)  < 10s warm

Then generate all 10 reports from addendum B.4 into vault/reports/. Ranking is the
point: staleness.md sorts by (ask_frequency x days_overdue), not alphabetically.
Also generate vault/moc/codebase.md per addendum B.5 including the
"high churn + high complexity + zero notes" section.

Build/release:
- Makefile targets for local build and a cross-compile matrix
  (darwin/linux/windows x amd64/arm64). Use `zig cc` for local cgo
  cross-compilation; a GitHub Actions native-runner matrix for releases.
- goreleaser config, SHA-256 checksums published per artifact.
- bin/forge as a shell shim: resolve ~/.forge/bin/<version>/forge, download from
  the GitHub release for the detected platform if missing, VERIFY the pinned
  SHA-256 before executing. Honor FORGE_BIN override for airgapped installs.

Do NOT build the daemon from section 5. Ship the CLI, measure the hook path, and
report the number to me. We decide on a daemon from data.

Finish by running everything against my vault and repo and showing me all 10
reports plus the benchmark numbers. I especially want drift.md — how many of my
existing notes reference code that has already changed?
```

---

## Still open

1. **ADR-002** — Go vs Java for the B2B backend, decided at B1 with real evidence.
   No action now.
2. **Daemon** — decided by measurement after the CLI ships.
3. **Language coverage for `codeindex`** — start with Java + Kotlin only (your stack,
   and it keeps the cgo grammar set small). Add Python/TS when someone asks.
