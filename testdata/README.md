# testdata/

`testdata/vault/` is a deliberately small, deliberately **broken** fixture vault. It
exists so vault-touching work — above all Phase 1's topology migration, the one
irreversible step in the plan — can be rehearsed without going anywhere near the real
vault at `/Users/mimir45/Documents/Base` (108 notes, personal, no backups).

`testdata/` is ignored by the Go toolchain, so these files never enter a build.

## This is NOT `examples/vault/`

`examples/vault/` is a named Phase 6 deliverable (ROADMAP row 6): a *clean, exemplary*
vault shipped with the plugin so a stranger can see what good output looks like. This
fixture is the opposite — it is full of the defects the pipeline has to survive. Do not
merge the two, and do not treat this directory as evidence that Phase 6 has started.

## Topology

Mirrors the real vault's **current** layout (`concepts/ decisions/ entities/ issues/
raw/ sources/ syntheses/ archive/ TIL/<topic>/`), *not* DESIGN §7's target layout
(`notes/{concept,howto,…}/ moc/ _inbox/ _archive/ profiles/`). That is the point: the
fixture is the migration's input.

## What is deliberately wrong in it

| # | Defect | Where | Exercises |
|---|---|---|---|
| F1 | Full 5-key frontmatter (`title`/`tags`/`source`/`date`/`status`) | `concepts/hibernate.md`, `concepts/soft-delete.md` (**not** `soft-deletion.md`, which omits `status:`) | the baseline shape, plus one near-baseline variant |
| F2 | 3-key frontmatter — no `source:`, no `status:` | `TIL/**/*.md` | `forge validate` on the real asymmetry between vault regions |
| F3 | `source:` missing on a non-TIL note | `decisions/liquibase-over-column-alias.md` | required-field detection outside `TIL/` |
| F4 | No frontmatter at all | `issues/hibernate-column-mismatch.md`, `raw/daily/2026-04-13.md` | the migration must add, not just rewrite |
| F5 | Dangling wikilink | `entities/meter-readings-service.md` → `[[entities/does-not-exist]]` | link check / `forge check` |
| F6 | Orphan (nothing links to it) | `archive/old-rag-notes.md` | `pkg/graph` orphan report — exactly one *content* note; `index.md` and `log.md` are roots and link each other, so they must not count |
| F7 | Near-duplicate pair | `concepts/soft-delete.md` vs `concepts/soft-deletion.md` | `pkg/similarity` MinHash+LSH and the duplicate gate |
| F8 | Path-bearing wikilinks (`[[concepts/hibernate]]`) | throughout | **link rewriting** — every one of these paths moves during migration |
| F9 | `tags` as a YAML flow list, unquoted | everywhere | frontmatter parsing |
| F10 | `status:` values that are not in any schema yet — exactly three in use: `active`, `archived`, `processed` | mixed | schema design pressure before Phase 1 fixes the vocabulary |
| F11 | Status carried as **body prose**, not frontmatter (`Status: resolved` on line 3) | `issues/hibernate-column-mismatch.md` | migration must lift ad-hoc body metadata into frontmatter rather than dropping it |
| F12 | `source:` points at a file that does not exist (`raw/daily/2026-04-14.md`; only the 04-13 raw note is present) | `sources/daily/2026-04-14-spring-keycloak.md` | dangling *path* reference — a distinct check from F5's dangling wikilink |

## Git

**`testdata/vault/` deliberately has no `.git`.** It briefly did; that was wrong. Once
this repo is itself `git init`-ed, a nested repo here is recorded as a gitlink (mode
160000) with no `.gitmodules`, so the fixture's files fall out of version control and a
fresh clone gets an empty directory.

The fixture is therefore committed as plain files, and **the harness creates the repo**:
copy `testdata/vault/` into `t.TempDir()` and `git init` the copy at test setup. That is
what exercises Phase 1's "refuses to run on a dirty tree" precondition and drift's
`--since-commit <sha>` / post-commit hooks — and it gets a pristine tree per test rather
than depending on a shared working copy being clean.

Reset after a destructive run is then just re-copying; no `git reset --hard` is involved.
Never `git init` this directory in place.
