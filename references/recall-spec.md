# Recall specification

How `forge recall` scores a question against the vault, and what the caller does with
the number. This is the reference `skills/forge/SKILL.md` points at so the skill can
stay short.

Recall is **T0**: pure Go, zero model calls, deterministic. The same question against
the same tree always produces the same ranking. Nothing here may consult a model —
DESIGN §5.2 locks stage 1 to engine `none`, and `forge recall` refuses to start
otherwise.

Source of truth: DESIGN §5.3 (decision tree, thresholds) and §8 (the five-step
pipeline, the scoring blend).

---

## 1. The pipeline

DESIGN §8's five steps, as built:

| # | Step | Where |
|---|---|---|
| 1 | Normalize the question into content terms; derive slug candidates | `pkg/recall/normalize.go` |
| 2 | Frontmatter scan — title/slug/tags/stack/updated/verified, cached | `pkg/store` + `cmd/forge/recall.go` |
| 3 | Body term-hit density, **top 20 candidates only** | `pkg/recall/score.go` |
| 4 | Score and rank | `pkg/recall/score.go` |
| 5 | Emit JSON, top 10 | `cmd/forge/recall.go` |

Two deliberate departures from §8's wording:

- **§8 says the frontmatter cache lives in `.forge/state.json`.** It lives in SQLite
  (`<vault>/.forge/cache/index.db`), which is what Phase 1 built and what the Phase 2
  brief asks for. Same contract either way: markdown is the only source of truth, the
  cache is derived, `forge reindex` rebuilds it entirely.
- **§8 says "ripgrep pass".** We scan with `pkg/vault` instead. ripgrep would be an
  external runtime dependency in a binary that is deliberately static
  (`CGO_ENABLED=0`, no PATH assumptions). The brief permits either: *"ripgrep or
  pkg/vault scan"*. The top-20 restriction from §8 is kept exactly — it is what makes
  the body pass cheap.

### Why there is no term-frequency table

The obvious optimization is a `note_terms` table in SQLite. Measured first, per
CLAUDE.md: the real vault is 91 notes / 370 KB, and reading every body costs ~5 ms.
Step 3 only reads 20 of them. A term table would be speculative infrastructure with
no demonstrated need, and it would add write cost to `forge index`. Revisit when a
vault makes the body pass measurable — the interface (`recall.Doc`) does not change.

---

## 2. Scoring

DESIGN §8's blend, unchanged:

| Channel | Weight | Compares |
|---|---|---|
| `title` | **0.4** | query terms vs the note's title + slug tokens |
| `tags` | **0.3** | query terms vs the note's `tags:` |
| `stack` | **0.2** | query stack hints vs the note's `stack:` |
| `body` | **0.1** | query terms vs the note body |

Every channel returns a value in `[0,1]`. The final score is a weighted mean over the
**active** channels (§2.5) — not a raw weighted sum. That distinction is the one place
this spec makes a decision the design docs left implicit; §2.5 argues it.

### 2.1 Normalization

Lowercase, NFC, strip punctuation, split on non-alphanumerics, drop tokens of length 1.
Then drop **question stopwords** — the scaffolding of the phrasings that trigger the
skill in the first place (`how`, `does`, `what`, `is`, `explain`, `difference`,
`between`, `best`, `practices`, `work`, `the`, `a`, …; full list in
`pkg/recall/normalize.go`).

`"how does spring boot transaction propagation work"` → `{spring, boot, transaction,
propagation}`.

Stopword removal is what makes the title channel behave. Without it, every question
carries 3–5 tokens of noise that no title can match, and the ceiling on a perfect hit
drops by a third.

### 2.2 Title channel (0.4)

An **F₂ measure** over the query terms `Q` and the note's title-and-slug token set `T`.
Recall is the share of the question the title covers; precision is the share of the
title the question accounts for:

```
r = |Q ∩ T| / |Q|      p = |Q ∩ T| / |T|
title = 5·p·r / (4·p + r)
```

Both sides are stopword-filtered (§2.1). Filtering the note side matters as much as
filtering the query side: the slug `hexagonal-architecture-ports-and-adapters` carries
an "and" that no question will ever match, and every such token dilutes `p`.

Neither half works alone. Both alternatives were tried against the real 91-note vault
and both had a named failure:

- **Pure containment-over-query** (recall only) rates *"Spring Boot 4 Breaking Changes
  — Artifact Renames and Test Module Split"* a perfect 1.00 for "how does spring boot
  work". A long title contains everything.
- **Symmetric Dice** punishes a title for being *more specific* than the question.
  *"Keyset Pagination — Compound OR Predicate"* is exactly the note wanted for "how
  does keyset pagination work", and Dice scored it 0.67 — below the answer threshold,
  in the dedup engine's headline case.

β = 2 leans on coverage, because the question is what is being looked up, while
precision still pulls an over-broad title down:

| Query terms | Title tokens | ∩ | Dice | **F₂** |
|---|---|---|---|---|
| `{keyset, pagination}` | `{keyset, pagination, compound, predicate}` | 2 | 0.67 | **0.83** |
| `{boot, spring}` | the 9-token Breaking Changes slug | 2 | 0.36 | **0.59** |
| `{goroutines}` | `{goroutines}` | 1 | 1.00 | **1.00** |

The two named cases, measured end to end after the blend: keyset **0.786 → 0.917**
(clears 0.85), hexagonal architecture **0.411 and third → 0.867 and first**.

### 2.3 Tags channel (0.3)

```
tags = Σ w(t) over t ∈ Q ∩ noteTags  /  Σ w(t) over t ∈ Q
```

`w(t)` is the term weight of §2.3.1. The denominator is **every** query term — not
`|noteTags|`, and no longer the query terms the vault's tag vocabulary
happens to carry.

Two alternatives are wrong in ways worth recording, and the second is what this spec
prescribed originally:

- **Dividing by `|noteTags|` punishes good tagging.** A note tagged `[goroutines]`
  scores 1.0; a note tagged `[goroutines, concurrency, runtime, scheduler, go,
  parallelism]` scores 0.17 on the same match. The better-curated note ranks lower.
  That is backwards, and it is why the note side is not the denominator.
- **Dividing by `|Q ∩ tagVocab|` collapses when the vocabulary carries one query term.**
  It was chosen so a verbose question could not cap the channel, and it does that. But
  "Redis caching in Spring Boot" against a Spring CLI note left `spring` as the only
  surviving term, and one-of-one reads 1.000 however it is weighted — so half the blend
  fired for "this note is in the Spring ecosystem", which in this vault is nearly no
  information. Worse, it made the channel structurally unable to say the thing the caller
  most needs to hear: *the vault has no Redis note*. Measured, that scored the wrong note
  0.740 and put it in the UPDATE(extend) band, where extending **writes**.

Dividing by all of `Q` does cap the channel at the question's verbosity, and that is now
deliberate: a five-term question a note answers on one tag *should* read about 0.2 there.
Two things keep the cap from being crude — §2.3.1's weights, so the terms are not a flat
count, and §2.5's two-sided activation, which drops an untagged note out of the
comparison instead of scoring it zero.

#### 2.3.1 Term weights

Both frontmatter channels weigh their terms by smoothed inverse document frequency,
`log(1 + N/df)`, capped at 3.5 and counted over the vault's own notes in the pass that
already walks every one of them. Unweighted, a tag half the vault carries counted for
exactly as much as the one carrying the question's meaning.

The cap is a guard, not the mechanism: because a universal term always weighs `log(2)`,
capping fixes the widest spread between the rarest and the commonest term at about 5:1
whatever the vault's size, so a hapax tag cannot decide a verdict alone. The smoothed
form is used rather than `log(N/df)` because the unsmoothed one is exactly zero for a
term on every note, which would empty the denominator of an active channel.

**A term no note carries weighs the mean of the terms that do.** This half of the fix
took two attempts. The weighting above shipped first and moved neither case it
was meant to fix, because the terms carrying a question's meaning were being filtered out
before any weight was computed — `redis` and `caching` were not weighted lightly, they
were absent. Admitting them is the fix; giving them a weight is what makes admitting them
do anything, since a raw IDF of 0 adds nothing to the denominator and would simply
deactivate the channel.

The mean is chosen over the alternatives for reasons that are not tuning:

- It is **parameterless**. It calibrates against whatever the query's present terms
  weigh, so there is no constant to fit — and fitting a constant against these nine
  queries is exactly what §3.1 forbids.
- It preserves the ratio's meaning. A query whose `m` channel terms the vault carries
  `k` of scores about `k/m` when the present weights are equal.
- It stays capped, being a mean of capped values. Flooring document frequency at 1 —
  the obvious alternative — hands an absent term the **maximum** weight and inverts the
  cap's purpose: absence would outweigh presence.
- The degenerate case falls out instead of being special-cased. With no present term
  there is no mean, every weight stays 0, the denominator is empty, and §2.5 leaves the
  channel inactive — which is the correct reading of a query the vault's controlled
  vocabulary cannot speak to at all.

`--explain` prints each term's weight and its raw `df`, because a weight of 0.00 alone
cannot distinguish a term on every note from a term on none, and those two now have
opposite consequences.

### 2.4 Stack channel (0.2)

```
stack = Σ w(t) over t ∈ S ∩ noteStack  /  Σ w(t) over t ∈ S
```

`S` is every query term, plus the `--stack` values **the vault's stack vocabulary
carries**. Containment over `S` — a note whose `stack:` is a superset is still a full
match on the terms it covers, because listing extra technologies is not evidence against
relevance.

The vocabulary filter applies to the hints and not to the question, which is the reverse
of the naive reading, and the asymmetry is the point. **A hint is a user filter; a
question term is evidence.** Narrowing a search by `--stack kotlin` in a vault that has
never seen Kotlin must not thereby make every note match less well — harmless while
unknown terms weighed nothing, a real regression the moment absent terms carry weight.
The vault holding no note about `redis`, by contrast, is exactly what the caller needs
the score to reflect.

### 2.5 Active channels and renormalization

**A channel with nothing to compare is undefined, not zero.** The score is the
weighted mean over the channels that are active for this query:

| Channel | Active when |
|---|---|
| `title` | `T ≠ ∅` — every note has a title, so in practice always |
| `tags` | `Q ∩ tagVocab ≠ ∅` **and** the note has `tags:` |
| `stack` | (`--stack` given or a query term is in the stack vocabulary) **and** the note has `stack:` |
| `body` | always, for the top-20 candidates that reach the body pass |

Activation is **two-sided**: the query must have supplied input *and* the note must
carry the field. Tag *mismatch* is evidence against relevance and scores zero on an
active channel; tag *absence* is no evidence either way and deactivates it. The
distinction is not academic — 31 of this vault's 91 notes have no `tags:` or `stack:`
after the Phase 1 migration, and zeroing them ranked a correct but under-curated note
below a well-tagged irrelevant one.

**The absent-term admission above does not weaken this.** The two rules answer different
questions and are easy to confuse. Admission is about the *query* side: a term the vault
tags nowhere still counts in the denominator, because the vault having no such note is
information. Activation is about the *note* side: a note with no `tags:` at all is not
thereby a worse answer, so it leaves the comparison rather than losing it. The query-side
rule cannot resurrect a channel the note-side rule switched off — with no query term in
the vocabulary there is no weight to take a mean of, every weight stays zero, and the
channel deactivates exactly as it did before.

**The asymmetry above was real and is fixed.** An
untagged note escaped the absent-term penalty entirely — its tags channel went inactive
and dropped out of the denominator — while a tagged note with no relevant tags paid the
same penalty in full, active at a hard 0.000. The note that carried nothing relevant was
worse off than the note that carried nothing at all, which is the effect this section was
written to prevent, running the other way round.

Activation is now decided on the **hit**, not on field presence: `tagsChannel` and
`stackChannel` activate on `len(hits) > 0` rather than `len(tags) > 0` / `len(stack) > 0`.
A note whose tags don't overlap the query is now inactive exactly like a note with no tags
at all — parity, not a new exemption. The table above still holds as the *note-carries-
the-field* row: what changed is that "carries the field" now means "carries something the
query could match." This is the same principle §2.5 already argues for the corpus-wide
case (a query outside the vault's vocabulary entirely must not activate a channel and
score every note 0.0) generalized to the per-note case.

**Measured on `examples/vault`, "Redis caching in Spring Boot":** `meterreadingsservice-
spring-boot-4-x-project` (still untagged) stays at 0.500, and `spring-cli-and-maven-
commands-for-spring-boot` (the one note with a genuinely matching tag) stays at 0.415 —
neither of the two notes moves, and the earlier false positive does not
return to first place. What moves is every note in between that carried an *irrelevant*
tag: across the corpus, 128 active tags/stack channels drop to 84, and all 43 that scored
a hard 0.000 are gone. One calibration row's winner changes as a result — the Docker
query's top-1 moves from `docker-compose-init-container-pattern…` (0.163, one real
`docker` tag hit) to `docker-compose-local-yaml…` (0.170, no relevant tags at all) — which
is the same shape of trade recurring one level down: a channel that can only ever manage
a low value (three of six query terms have vault-wide df 0 here) still drags an active
note below where exclusion would leave it. That is `weighted`'s documented behavior
operating as designed, not a new mechanism, and it is why this fix is scoped to
activation and does not touch `weightsOver`'s weighting formula.

**Consequence for §3.2's floor and `forge intent`'s gate, both re-measured rather than
re-tuned.** Every score in the corpus moved by some amount, so both of the neighbour
floor's and intent gate's derivations
were re-run against their original, unedited label files. The neighbour floor's F1 peak
moved from 0.125 to **0.150** — §3.2 below carries the new sweep. `forge intent`'s gate
derivation (`cmd/forge/testdata/intent-gate.golden`) still holds mechanically — the gate
stays 0.50, no QUIET prompt is admitted, and 8 of 10 FIRE prompts still are — but the
measured separation margin between the two classes went from +0.005 to **-0.036**: the
lowest admitted-at-gate FIRE prompt and the highest QUIET prompt now overlap in score
between 0.407 and 0.443. The gate itself sits safely above both, so nothing is broken
today, but the classes are no longer cleanly separable everywhere, which is a real finding
and not something this fix corrects, and remains an open item.

```
score = Σ(w_c · v_c) / Σ(w_c)   over active c
```

Activation is decided **per query, not per candidate**. Every candidate in one run is
scored over the same channel set, so renormalization cannot reorder results — it only
sets the scale. Ranking is unaffected; thresholds are what change.

#### Why this is the right reading

Three reasons, in increasing order of force.

**1. The raw sum cannot reach the threshold it is paired with.** Take a question the
vault answers perfectly: `"how do goroutines work"` against a note titled
"Goroutines". Title 1.0, body ~0.9, no tag hit, no `--stack`. Raw sum:

```
0.4·1.0 + 0.3·0 + 0.2·0 + 0.1·0.9 = 0.49
```

DESIGN §5.3 routes that to **CREATE** — it would write a second note about
goroutines next to the one that already exists. The dedup engine's headline case fails
under the literal arithmetic. Under renormalization the active set is `{title, body}`:

```
(0.4·1.0 + 0.1·0.9) / 0.5 = 0.98  →  ANSWER_FROM_VAULT
```

**2. Under the raw sum, a CLI flag decides the branch.** Same question, with and
without `--stack go`, against the same matching note:

| | raw sum | renormalized |
|---|---|---|
| no `--stack` | 0.49 → CREATE | 0.98 → ANSWER |
| `--stack go` | 0.69 → UPDATE | 0.99 → ANSWER |

The raw sum moves the verdict two branches on a flag that added no new information
about the note. That is not a threshold, it is a coin flip. Renormalization moves it
by 0.01.

**3. DESIGN's own telemetry example expects the renormalized range.** §5.3's sample
log line reads `"decision":"ANSWER_FROM_VAULT","recall_top_score":0.94`. 0.94 is not
reachable as a raw weighted sum unless all four channels fire near 1.0 — which
requires the user to have typed `--stack` *and* the question to have hit the tag
vocabulary. The doc's own worked example is a renormalized score.

The weights themselves are untouched. What changes is the denominator, and only when a
channel had nothing to weigh.

### 2.6 Body channel (0.1)

Computed for the top 20 candidates by the other three channels (§8 step 3). Per query
term, count occurrences in the body, saturate at 3, average across terms:

```
body = ( Σ_t min(count(t), 3) / 3 ) / |Q|
```

Saturation stops one term repeated forty times from standing in for coverage. The
channel is 0.1 for a reason — it breaks ties between frontmatter-equivalent notes, it
does not decide matches.

### 2.7 Determinism

Ties break on `rel` ascending. Two runs over the same tree return byte-identical JSON.
This matters because Phase 2b re-measures against these numbers and Phase 6b exports
them.

---

## 3. Thresholds and the decision tree

DESIGN §5.3, unchanged:

```
top_score
   ├─ ≥ 0.85  and note fresh   → ANSWER_FROM_VAULT
   ├─ ≥ 0.85  and note stale   → UPDATE(refresh)
   ├─ 0.55 – 0.85              → UPDATE(extend)
   └─ < 0.55                   → CREATE, then link the 0.125–0.55 neighbours
```

`answer_threshold: 0.85` and `update_threshold: 0.55` are DESIGN §10's config keys.
`pkg/recall` hardcodes the defaults; the config chain
(`DESIGN:516-518` joins the config union) can override them. Do not scatter literals:
they live in `recall.DefaultThresholds`.

**Stale** means the note's `verified` date — **falling back to `updated`**, in that
order — is older than its `freshness_days`, which comes from the note's own frontmatter
or the type default in `references/schema.yaml`.

The order is deliberate and it is the reverse of the intuitive one. An edit that fixes
a typo bumps `updated` without anyone re-checking the claims against their sources;
reading `updated` first would treat that as a re-verification, which is how a vault
quietly starts lying. A note with neither date is **stale**: recall cannot vouch for it,
so it must not answer from it. `freshness_days: 0` means never stale and outranks even
the undatable case — that is how `decision` notes behave (DESIGN §10: *"decisions never
go stale, they get superseded"*).

### 3.1 Are 0.85 / 0.55 right for this vault?

**Yes, and they have not moved.** What moved is the scale underneath them, which is why
this section is now generated rather than transcribed.

Nine adjacent-topic queries — topics where a closely related note exists and extending it
is the plausible move, deliberately the hardest band for the decision tree. Regenerate
with:

```
go test ./cmd/forge -run TestCalibration -update   # rewrites the golden
git diff cmd/forge/testdata/calibration.golden     # before -> after, reviewable
```

The golden carries a fifth column the table below does not restate: the neighbour set
each row emits. Verdict says a note gets created; that column says
whether it gets created linked or orphaned, and it is the only column a floor change may
move — score and verdict are functions of `Rank` and `Decide`, which the floor does not
enter. A golden diff that moves them is a leak into scoring.

The corpus is `examples/vault` (92 scored docs), staged into a temp dir per run so the
SQLite cache cannot warm one column and not the other. It is git-tracked on purpose: the
first version of this table was measured against a live vault that then drifted, so the
"before" column became unreproducible and any "after" compared with it proved nothing.

| Query | Top-1 before → after | Score | Verdict |
|---|---|---|---|
| Redis caching in Spring Boot | `spring-cli-and-maven…` → `meterreadingsservice-spring-boot-4-x-project` | 0.740 → **0.500** | UPDATE(extend) → CREATE |
| Spring Boot 4 configuration properties binding | `spring-cli-and-maven…` → `meterreadingsservice-spring-boot-4-x-project` | 0.700 → **0.410** | UPDATE(extend) → CREATE |
| Storybook interaction testing with play functions | `storybook-isolated-component-development…` (unchanged) | 0.617 → **0.217** | UPDATE(extend) → CREATE |
| Java virtual threads with Spring Boot | `meterreadingsservice-spring-boot-4-x-project` (unchanged) | 0.600 → **0.486** | UPDATE(extend) → CREATE |
| Keycloak token exchange between clients | `til-demoing-keycloak-plus-google-login…` (unchanged) | 0.529 → **0.315** | CREATE |
| Kafka consumers with Testcontainers | `testcontainers-docker-based-integration-testing` (unchanged) | 0.501 → **0.311** | CREATE |
| React Server Components data fetching | `loader-component-pattern-react-frontend-conventions` (unchanged) | 0.472 → **0.300** | CREATE |
| Docker multi-stage build cache optimization | `docker-compose-init-container-pattern…` (unchanged) | 0.429 → **0.163** | CREATE |
| JPA entity graph to avoid N+1 | `creationtimestamp-and-updatetimestamp…` (unchanged) | 0.333 → **0.119** | CREATE |

Seven of nine winners are unchanged. The two that moved are the same note —
`spring-cli-and-maven-commands-for-spring-boot`, the original false positive that motivated the IDF weighting — losing
the same two inflated channels in two different queries.

**What the fix removed.** Every UPDATE verdict in the "before" column came from the same
artifact: a frontmatter channel reading 1.000 off a denominator of one surviving term.
Case 1's `tags` denominator was `{spring}` alone, at the 3.5 cap, because `setOf` folds
the tag `spring-cli` into `{spring, cli}` and exactly one note in the vault is tagged
that way. One-of-one is 1.000 whatever it weighs, so half the blend fired for ecosystem
membership. `redis` and `caching` — the terms carrying the question — were not in the
arithmetic at all.

**The artifact was sometimes right, and that cost is real.** The Storybook row is the
honest case against this change. `storybook-isolated-component-development-and-visual-
documentation` genuinely is the note to extend, and it fell 0.617 → 0.217. But its 0.617
was the same artifact: `tags` 1.000 off `{testing}` alone, `stack` 1.000 off `{storybook}`
alone, while `play`, `interaction` and `functions` sat outside both vocabularies. The
fix cannot keep the artifact where it happened to be right and drop it where it was
wrong — that distinction does not exist in the data. Four narrower admission rules were
measured against this row and none restores it (0.242 tags-only-with-vault-wide-absence,
0.377 tags-only, 0.392, 0.217 as shipped); the only knob that would is the threshold,
which is what this section refuses to move.

**Why nine-of-nine CREATE is not a dead decision tree.** These queries are adjacent-topic
by construction — the note is a neighbour, not an answer. Measured on the same corpus,
queries that name a note's actual subject still clear the bands:

| Query | before | after |
|---|---|---|
| what is the transactional outbox pattern | 1.000 ANSWER_FROM_VAULT | 1.000 ANSWER_FROM_VAULT |
| hexagonal architecture ports and adapters | 1.000 ANSWER_FROM_VAULT | 1.000 ANSWER_FROM_VAULT |
| how does keyset pagination work | 0.917 ANSWER_FROM_VAULT | 0.729 UPDATE(extend) |
| Storybook decorator pattern for Redux providers | 0.823 UPDATE(extend) | 0.652 UPDATE(extend) |
| Testcontainers Docker based integration testing | 0.814 UPDATE(extend) | 0.626 UPDATE(extend) |

So the tree still answers and still extends; it stopped extending into notes that merely
share an ecosystem. The keyset row is a demotion from ANSWER to UPDATE against a note
narrower than the question (`keyset-pagination-compound-or-predicate`), which is arguably
the better reading of it — but it is a demotion, and it is recorded here rather than
argued away.

**The residual cost — recorded here, closed 2026-08-23.** Adjacent-topic queries briefly
lost their neighbour links as well as their UPDATE verdict. The Storybook query verdicted
CREATE with **zero** neighbours: both Storybook notes landed at 0.217 and 0.201, under
§3.2's then-0.30 floor, so the new note would have been written unlinked to the two notes
obviously related to it. That floor was calibrated against the old scale; re-deriving it
against the same nine queries used to validate the fix would have been circular, so the
floor was re-derived to **0.125**
against a separate labelled query set. The same row now emits seven neighbours, five of
them the Storybook family. §3.2 carries the derivation. Two further consequences worth
recording separately: the Kafka/Testcontainers miss is a coverage defect, not a precision
one, and §2.5's untagged-note asymmetry (addressed above).

**The thresholds stay at DESIGN §5.3's 0.85 / 0.55.** Lowering `update_threshold` to 0.45
was the tempting move when this section was first written, and it remains wrong: on the
"before" column it admitted `docker-compose-init-container-pattern…` at 0.429 for a
question about build caching and did nothing about the 0.740 false positive. Fixing the
cause moved that note to 0.163 — the argument against the threshold change is now a
measurement rather than a projection.

### 3.2 Neighbour band

On CREATE, candidates scoring `0.150 – 0.55` are the neighbours the new note links to.
They arrive pre-filtered in the `neighbours` array (§4), so the caller never applies a
threshold itself. The band is not a separate query — it is a slice of the same ranking.

**The floor is 0.150, re-derived on top of an earlier re-derivation to 0.125.**
0.30 was chosen before §2.3.1's IDF change moved the scale under it; 0.125
was F1's maximum on that scale. The activation fix above (§2.5) then moved every note whose
tags or stack didn't overlap the query — the exact population sitting in the neighbour
band — which shifted F1's peak again.

The same label file, unedited, was re-swept: `cmd/forge/testdata/neighbour-labels.txt`,
fifteen adjacent-topic questions with 58 expected neighbours, written before any score was
measured and re-used as-is — re-labelling after seeing new scores is how a derivation
becomes a fit. `TestNeighbourFloorSweep` re-runs the sweep and records
`testdata/neighbour-sweep.golden`; the number without that sweep is tuning.

| Floor | Precision | Recall | F1 | Median links/query | Queries with none |
|---|---|---|---|---|---|
| 0.100 | 0.406 | 0.741 | 0.524 | 7 | 0 |
| 0.125 | 0.451 | 0.707 | 0.550 | 5 | 0 |
| **0.150** | **0.506** | **0.672** | **0.578** | **5** | **0** |
| 0.175 | 0.531 | 0.586 | 0.557 | 3 | 0 |
| 0.200 | 0.558 | 0.414 | 0.475 | 2 | 1 |
| 0.300 | 0.857 | 0.207 | 0.333 | 1 | 5 |

F1 peaks at 0.150 (0.578, up from 0.125's now-0.550) and no query is left empty until
0.200 — the same shape of decision as the earlier re-derivation, re-run on the new scale rather than
re-argued from scratch.

Precision 0.506 is, as before, a **lower bound**: several counted false positives are
defensible links the labels simply did not name. Left uncorrected on purpose, for the
same reason as the earlier derivation.

A known limitation: §3.1's broadest queries now emit ten neighbours, because
two general Spring notes score on every Spring question. That is a corpus property no
floor can separate, and a cap on the neighbour count is a different change — re-measured
after the activation fix above rather than responded to by raising the
floor: still three of nine queries at the ten-neighbour cap, unchanged in kind.

---

## 4. Output contract

```
forge recall --question "..." [--stack java,spring-boot] [--vault PATH] [--explain]
```

One JSON object on stdout. `candidates` is sorted by `score` descending, at most 10
entries:

```json
{
  "question": "how does spring transaction propagation work",
  "verdict": "ANSWER_FROM_VAULT",
  "top_score": 0.93,
  "candidates": [
    {
      "slug": "spring-transaction-propagation",
      "path": "notes/concept/spring-transaction-propagation.md",
      "title": "Spring Transaction Propagation",
      "score": 0.93,
      "updated": "2026-05-02",
      "verified": "2026-05-02",
      "stale": false,
      "matched_on": ["title", "tags"]
    }
  ],
  "neighbours": [],
  "run_id": "3f9a1c2e7b4d6081a5c3e9f0b2d4a716"
}
```

**`run_id`** is the D1 outcome-joining correlation key, purely additive to
this contract. It is minted fresh per call (`telemetry.NewRunID`, 128 random bits, no
ordering or wall-clock semantics), never repeats, and carries no meaning on its own — a
caller threads it back through `forge gate --run-id <id>` to join this routing decision
to whether the note write it led to was actually published. A caller that ignores the
field loses nothing else; the join is optional on both ends.

**The verdict ships in the payload, not in the caller.** §3's tree is implemented once,
in Go. A skill that restated the thresholds in prose would silently diverge the moment a
config change moves them, and the divergence would be
invisible — both copies keep producing plausible numbers.

`matched_on` lists the active channels that scored above zero, in weight order. It is
the cheap explanation; `--explain` is the full one.

`neighbours` is the §3.2 band, and is populated **on a `CREATE` verdict only**. On
`ANSWER_FROM_VAULT` or either `UPDATE` the same notes are ones the caller was just told
not to write to, and emitting them invites it to link them anyway.

### 4.1 `--explain`

Prints the per-candidate breakdown to **stderr**, so stdout stays parseable JSON:

```
query terms: keyset, pagination

keyset-pagination-compound-or-predicate              0.729
  title  0.833 x 0.4 = 0.333   keyset, pagination
  tags   0.500 x 0.3 = 0.150   pagination
         idf keyset 3.46 (df 0), pagination 3.46 (df 3)
  stack    inactive — the query supplied no stack input
  body   1.000 x 0.1 = 0.100   keyset, pagination
  sum   0.583 / 0.800 = 0.729

verdict: UPDATE(extend)
```

The `sum` line prints the renormalizing denominator explicitly, because that is the
number a surprising verdict usually turns on. The `idf` line under a frontmatter channel
prints §2.3.1's per-term weight and the raw document frequency behind it, because
the hit list alone no longer explains the value under IDF weighting — here the channel reads 0.500
rather than 1.000 because `keyset` is tagged by no note in the vault (`df 0`) and still
counts, at the mean of the present weights. Reading a weight without its `df` cannot
distinguish a term on every note from a term on none. The `verdict:` line is printed on **every**
path including the empty one — a caller reading stderr must not get silence on exactly
the case the verdict matters most. On a `CREATE` verdict the neighbour links follow it.

### 4.2 The zero floor

Candidates that matched on no channel at all are dropped rather than padded out to ten.
Without the floor, "how does Rust ownership and borrowing work" returned `index.md` and
`log.md` at `0.000` as the top two rows of a CREATE verdict — noise a caller has to know
to ignore.

So `candidates` may be shorter than 10 and may be empty. A question the vault has
nothing on returns `[]` for both arrays with `"verdict": "CREATE"` and
`"top_score": 0` — not an error, not `null`. That is also the honest CREATE case with
**no neighbours to link**: a genuinely new topic in an ecosystem the vault does not
cover has nothing to attach to, and inventing links for it is worse than leaving none.

---

## 5. Caching and latency

Budget: **< 200 ms warm** on a few-thousand-note vault (DESIGN §8). Measured on the
real 91-note / 370 KB vault: **~5 ms warm, ~20 ms cold** (cache deleted, every note
re-parsed). Two orders of magnitude of headroom, which is why §1's "no term-frequency
table" holds — the table would buy nothing and cost a write on every `forge index`.

Warm path per note: one `stat`, one lookup in the batch-loaded SQLite row set. A row is
reused when `store.Fresh(rel, mtime, size)` holds; otherwise the markdown is re-parsed
and the row upserted, so `forge recall` self-heals a cold or partial cache without
requiring `forge index` to have run.

The cache is derived and disposable. `forge reindex` drops and rebuilds it from
markdown, and a deleted `index.db` costs one slow run, never data.

---

## 6. What recall does not do

- **No embeddings.** DESIGN §8 gives three reasons; the short one is that a model
  already read the question, so lexical plus a model re-rank of the top 20 matches
  vectors at this scale. `recall.strategy: hybrid` is a v2.2 config value with no
  implementation behind it — the interface exists, the vectors do not.
- **No model call, ever.** Including for query expansion, synonym lookup, or re-ranking.
  The re-rank in §8's reason 1 happens in the *caller*, on recall's output.
- **No writes.** Recall reads the vault and writes only the derived cache. Deciding
  and writing is stage 2's job, in `skills/forge/SKILL.md`.
- **No cross-vault search.** One `--vault` per run.
