# Contributing

## Read this first

1. [`docs/ROADMAP.md`](docs/ROADMAP.md) — condensed index over the design. Always start
   here.
2. `CLAUDE.md` — the project's layout, invariants, config chain, and commands.
3. [`docs/KNOWLEDGE-FORGE-DESIGN.md`](docs/KNOWLEDGE-FORGE-DESIGN.md) and
   [`docs/KNOWLEDGE-FORGE-ADDENDUM.md`](docs/KNOWLEDGE-FORGE-ADDENDUM.md) — the full
   spec: schema, pipeline, gates, engine tiers, datasets, config.

## Build & test

Two build lanes, because `pkg/codeindex` (go-tree-sitter) is the one package that needs
cgo:

```sh
CGO_ENABLED=0 go build ./...   # portable lane — what ships in releases
go test ./...                  # or: make test (runs both cgo lanes)
make lint                      # gofmt -l + go vet
```

`make full` builds this host's binary with the tree-sitter code index compiled in
(needs a C toolchain). `make dist` produces the six-target release matrix from the
portable lane. `make bench` runs the parse/similarity/drift benchmarks.

## Project invariants

Stated once here; each is also called out in `CLAUDE.md` because it's easy to violate
by accident:

- The T0 static core makes **zero model calls**. If a change seems to need one for
  `recall`, `write`, or `index`, stop — those stages accept engine `none` only, and code
  that says otherwise should refuse to start with a clear error, never silently
  override.
- Drift is git-anchored (`--since-commit`), never against the uncommitted working tree.
- Markdown is the only source of truth; SQLite (`pkg/store`) is a derived cache that
  `forge reindex` must be able to rebuild from scratch.
- `pkg/similarity` is hand-rolled MinHash + LSH — no embeddings (see
  `docs/adr/0001-lexical-recall-vs-embeddings.md`).
- Never auto-mutate the vault on a schedule. A quality-gate failure goes to `_inbox/`
  with `confidence: low`, never a silent publish.
- Code verification (`forge verify-code`) compiles in a throwaway directory, never in
  the user's project.
- Telemetry logs a topic and a hash — never raw question text, code, or file contents.

## Style

- Every function stays small enough to read at a glance — this repo's own convention,
  visible throughout `pkg/` and `cmd/forge/`, keeps most functions under ~20 lines by
  splitting orchestration from the piece doing the actual work (see `cmd/forge/gate.go`
  or `pkg/scrub` for the pattern).
- Match the surrounding file's comment density and naming — comments here tend to
  explain *why* a choice was made (a trade-off, a measured number, a rejected
  alternative), not just restate what the code does.

## Fixture vaults — don't touch by accident

`testdata/vault/` is a 13-note fixture with twelve **deliberate** defects (F1–F12) that
other packages' tests exercise. Do not fix the defects; see `testdata/README.md`.
`examples/vault/` (Phase 6) is a different thing — real vault content run through
`forge scrub` and human-reviewed before commit — don't confuse the two.
