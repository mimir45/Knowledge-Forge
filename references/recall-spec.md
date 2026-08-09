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
tags = |Q ∩ noteTags| / |Q ∩ tagVocab|
```

where `tagVocab` is the union of every tag in the vault. The denominator is the number
of query terms that *could* have matched some note's tags — not `|Q|`, and not
`|noteTags|`.

Both alternatives are wrong in a way worth recording:

- **`|Q ∩ noteTags| / |noteTags|` punishes good tagging.** A note tagged
  `[goroutines]` scores 1.0; a note tagged `[goroutines, concurrency, runtime,
  scheduler, go, parallelism]` scores 0.17 on the same match. The better-curated note
  ranks lower. That is backwards.
- **`|Q ∩ noteTags| / |Q|` caps the channel at the question's verbosity.** A two-term
  question that matches one tag exactly can never exceed 0.5, for reasons that have
  nothing to do with the note.

Dividing by `|Q ∩ tagVocab|` makes both notes above score 1.0 and keeps the channel
comparable across questions of different lengths.

### 2.4 Stack channel (0.2)

```
stack = |S ∩ noteStack| / |S|
```

`S` is the query's stack hints: the `--stack` values, plus any query term that appears
in the vault's stack vocabulary. Containment over `S` — a note whose `stack:` is a
superset of the hints is a full match, because listing extra technologies is not
evidence against relevance.

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
   └─ < 0.55                   → CREATE, then link the 0.3–0.55 neighbours
```

`answer_threshold: 0.85` and `update_threshold: 0.55` are DESIGN §10's config keys.
Phase 2 hardcodes the defaults in `pkg/recall`; Phase 3 wires them to the config chain
(AUDIT §8.4 D-7 — `DESIGN:516-518` joins the config union). Do not scatter literals:
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

Measured, nine adjacent-topic queries against the real 91-note vault — topics where a
closely related note exists and extending it is the right move:

| Query | Top score | Verdict |
|---|---|---|
| Redis caching in Spring Boot | 0.740 | UPDATE(extend) |
| Spring Boot 4 configuration properties binding | 0.700 | UPDATE(extend) |
| Storybook interaction testing with play functions | 0.617 | UPDATE(extend) |
| Java virtual threads with Spring Boot | 0.600 | UPDATE(extend) |
| Keycloak token exchange between clients | 0.529 | CREATE |
| React Server Components data fetching | 0.472 | CREATE |
| Kafka consumers with Testcontainers | 0.469 | CREATE |
| Docker multi-stage build cache optimization | 0.429 | CREATE |
| JPA entity graph to avoid N+1 | 0.333 | CREATE |

**Recommendation: leave both thresholds where they are.** The distribution has no gap
to cut at, and the miss and the false positive have the same cause, which moving a
threshold makes worse rather than better.

Look at the two ends side by side. Redis caching scored 0.740 against
`spring-cli-and-maven-commands-for-spring-boot` — a note about CLI invocations, not
caching:

```
title  0.476 x 0.4 = 0.190   boot, spring      <- the only discriminating channel
tags   1.000 x 0.3 = 0.300   spring
stack  1.000 x 0.2 = 0.200   boot, spring
```

Half the weight fired for "this note is in the Spring ecosystem" — which is nearly no
information in a vault where most notes are. Kafka/Testcontainers scored 0.469 against
the genuinely correct `testcontainers-docker-based-integration-testing`, because the
discriminating terms (`kafka`, `consumers`) are not in that note's title and its tags
did not overlap at all.

So the ranking error is not calibration. **The `tags` and `stack` channels have no
inverse-document-frequency weighting**: a term carried by every note scores identically
to one carried by three. Dropping `update_threshold` to 0.45 would admit
Kafka/Testcontainers and React Server Components, but it would also admit
`docker-compose-init-container-pattern` at 0.429 for a question about build caching,
and it would do nothing about the 0.740 false positive — which is the more damaging
error, because UPDATE(extend) writes into the wrong note.

Fixing the cause rather than the symptom is **BACKLOG B-008**, scoped to Phase 2b where
the nine reports re-measure these numbers anyway. Until then the thresholds stay at
DESIGN §5.3's values and the failure mode is documented rather than tuned around.

### 3.2 Neighbour band

On CREATE, candidates scoring `0.3 – 0.55` are the neighbours the new note links to.
They arrive pre-filtered in the `neighbours` array (§4), so the caller never applies a
threshold itself. The band is not a separate query — it is a slice of the same ranking.

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
  "neighbours": []
}
```

**The verdict ships in the payload, not in the caller.** §3's tree is implemented once,
in Go. A skill that restated the thresholds in prose would silently diverge the moment
AUDIT §8.4 D-7 moves them into Phase 3's config chain, and the divergence would be
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

keyset-pagination-compound-or-predicate              0.917
  title  0.833 x 0.4 = 0.333   keyset, pagination
  tags   1.000 x 0.3 = 0.300   pagination
  stack    inactive — the query supplied no stack input
  body   1.000 x 0.1 = 0.100   keyset, pagination
  sum   0.733 / 0.800 = 0.917

verdict: ANSWER_FROM_VAULT
```

The `sum` line prints the renormalizing denominator explicitly, because that is the
number a surprising verdict usually turns on. The `verdict:` line is printed on **every**
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
