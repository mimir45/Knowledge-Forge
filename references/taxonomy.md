# Taxonomy

How `stack` and `tags` in `references/schema.yaml` were derived, and how to extend them.

Nothing here was invented. Every value traces to a tag that actually appears in a real
personal vault, measured 2026-08-09 by walking the vault and parsing YAML frontmatter:
**128 distinct tags across 434 assignments**, of which **81 (63%) are used exactly
once**. (An earlier count with a slightly different parser reported 131/431; the delta is
inline-vs-block `tags:` handling and does not change any decision below.)

---

## 1. Why two axes instead of one

The existing vault has a single `tags:` list carrying two different kinds of fact at
once:

- **what the note is written in** — `spring-boot`, `react`, `docker`, `keycloak`
- **what the note is about** — `testing`, `pagination`, `security`, `accessibility`

Collapsing these is why the audit found the tag vocabulary uncontrolled: `spring-boot`
(32 uses) and `wcag` (1 use) are not the same kind of token and cannot be governed by
the same policy. So the contract splits them:

| Field | Axis | Policy | Rationale |
|---|---|---|---|
| `stack` | technology | **closed** | Small, slow-moving, and the thing `forge recall` filters on. An unknown value is almost always a typo. |
| `tags` | concept | **open**, kebab-case, alias-normalized | The long tail is real signal, not noise. 63% singletons is a property of a personal vault, not a defect to legislate away. |

A closed `tags` list would reject 81 of the vault's own tags on day one. That fails the
Phase 1 instruction — *"don't invent a taxonomy I'm not using."*

---

## 2. `stack` — the closed list, and where each value came from

41 values. A tag was promoted to `stack` when it names a language, framework, library,
tool, or runtime. Counts are vault occurrences.

| Group | Values (count) |
|---|---|
| JVM / backend | `spring-boot` (32), `java` (14), `hibernate` (10), `maven` (10), `jpa` (8), `liquibase` (7), `mapstruct` (3), `spring-security` (2), `openapi` (3), `testng` (2), `junit` (2 after alias merge), `kotlin` (0) |
| Frontend | `react` (18), `storybook` (11), `nextjs` (5), `mui` (4), `typescript` (4), `styled-components` (2), `redux` (1), `jest` (1), `jsdom` (1) |
| Infra | `docker` (10), `testcontainers` (8), `sql` (2), `postgresql` (1) |
| Auth | `keycloak` (9), `oauth2` (1) |
| Local AI | `continue-dev` (17), `llama-cpp` (9), `litellm` (5), `qwen3` (4), `bge` (1) |
| Languages / tooling | `intellij` (10), `go` (2 + 1 alias), `python` (3), `shell` (2), `claude-code` (2), `obsidian` (4), `git` (1), `postman` (1), `stripe` (2) |

Two deliberate exceptions to "measured only":

- **`kotlin`** has zero vault uses. It is in the list because `pkg/codeindex` is
  specified (target layout, `CLAUDE.md`) to start with *Java + Kotlin only*. Admitting it
  now avoids a schema edit the first time a Kotlin note is written.
- **`junit`** absorbs `junit5`, so its count is the merged 2, not the raw 1.

### Not promoted, on purpose

`spring` (3) — an ambiguous truncation of `spring-boot` / `spring-security`; it is an
alias (§3), not a value. `frontend` (8) and `backend` (1) — architectural layers, not
technologies; they stay in `tags`. `local-ai` (7) — the *name of a stack combination*,
not a technology; the individual components (`llama-cpp`, `litellm`, `continue-dev`) are
what a recall query should filter on, so `local-ai` stays a tag. `leprecoin` (8) and
`gitnexus` (1) — project names, which belong in `profiles/projects/<project>.md`
(DESIGN §7), not in a technology vocabulary.

---

## 3. Alias map — rejected values and their canonical form

`forge validate` reports these as errors. `forge validate --fix` rewrites them.
The first five are the collisions the initial audit named explicitly; the
rest were found in the same measurement pass.

| Written | Canonical | Field | Why |
|---|---|---|---|
| `configuration` (4) | `config` (6) | tags | Same concept, majority form wins |
| `golang` (1) | `go` (2) | stack | Majority form; also the module's own word |
| `junit5` (1) | `junit` (1) | stack | Version in the tag defeats the purpose |
| `issues` (1) | `issue` (9) | tags | Plural drift |
| `migrations` (1) | `migration` (1) | tags | Plural drift; tie broken toward singular, matching every other tag |
| `spring` (3) | `spring-boot` | stack | Every one of the three notes is Spring Boot |
| `database` (1) | `databases` | tags | Matches the `TIL/databases/` folder the user already keeps |
| `local-model` (1) | `local-ai` | tags | Same concept, majority form wins |
| `google-oauth` (1) | `oauth2` | stack | Provider-specific truncation of the protocol |

**Not aliased, though they look like it.** `google-idp` (2) is Keycloak identity-provider
configuration and is genuinely distinct from `oauth2`. `llm` (3) and `ai` (1) sit at
different levels of generality and both are used correctly. `api` (3) and `rest` (1) are
not synonyms.

---

## 4. `type` — seven values, and an arity note

`concept | howto | pattern | pitfall | decision | api | incident`

The `type` enum in DESIGN §6.1 has seven values, but two other places imply fewer:
`docs/CLAUDE-CODE-PROMPT.md` item 2 lists six templates (no `incident`), and DESIGN §7's
topology sketch shows five `notes/` subdirs. Phase 1 treats the **enum as authoritative**
and builds seven templates and seven subdirs, because the alternative is a schema-valid
`type` with no template and nowhere to live.

Nothing in the existing vault constrains this choice: the initial audit found `type` used
**zero** times vault-wide, in its baseline metrics. Every value is assigned
for the first time by the migration.

### Mapping the old topology onto `type`

The migration infers `type` from a note's current directory, then from body cues. Every
inferred value is written with `confidence: low`.

| Old location | Inferred `type` | Confidence |
|---|---|---|
| `decisions/` | `decision` | high — the directory *is* the assertion |
| `issues/` | `pitfall` if the body records a fix; `incident` if it records an outage | low |
| `concepts/` | `concept` | low |
| `syntheses/` | `concept`, `depth: 4` | low |
| `entities/` | `concept` — these are service/system descriptions | low |
| `TIL/**` | `howto` if the body has imperative steps or a fenced command; else `concept` | low |
| `sources/daily/`, `raw/daily/` | not notes — stay outside `notes/`, see §5 | — |

---

## 5. What the vocabulary deliberately does not cover

- **`raw/daily/`** (5 notes) are symlinks to an external daily-notes tool's output. They
  are ingest input, not vault notes. They are excluded from the contract and from
  `forge validate --all`.
- **`sources/daily/`** (9 notes) are compiled digests of the above. Same treatment.
- **Non-note root files** — `CLAUDE.md`, `lint-report.md`, `CCFA_…_TR.md` — are excluded.
- **`status`** is not a schema field. The audit found it carries no information: 72 of
  73 values are the identical string `active`, and the two values documented in the
  vault's own `CLAUDE.md` (`draft`, `archived`) are never used while an undocumented
  `design-complete-no-code` is. `confidence` plus location (`_inbox/`, `_archive/`)
  replaces it. The migration drops the key.

---

## 6. Extending this

1. Edit `references/schema.yaml` — it is the only enforcement point.
2. Add the value to the right table above with its justification.
3. Adding to `stack` requires a reason a `tag` would not serve; the closed list is
   only useful while it stays small.
4. Adding an alias requires the count of both forms, so the majority rule stays checkable.
