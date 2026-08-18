# 0001 — Lexical recall over embeddings

- **Status:** Accepted
- **Source:** `docs/KNOWLEDGE-FORGE-DESIGN.md` §8 ("Retrieval-before-research"), per
  `docs/AUDIT.md` §8.4 D-3.

## Context

`forge recall` (the "dedup engine") has to answer one question fast: given a new
"explain X" prompt, is there already a note for this, and if so is it stale? It runs
deterministically, with no model call, and needs to stay under a few hundred
milliseconds on a vault of a few thousand notes.

The obvious alternative is embeddings + a vector index: chunk every note, embed it,
store the vectors, do a nearest-neighbor search per query.

## Decision

Lexical recall — slug candidates, frontmatter scan, a ripgrep-style term hit-density
pass, then deterministic scoring — not embeddings. Implemented in `pkg/recall`, backed
by `pkg/similarity`'s hand-rolled MinHash + LSH for near-duplicate detection. No
embedding model, no vector store, anywhere in the pipeline.

Three reasons, in order of weight:

1. **A model already reads the question.** By the time recall runs, the calling agent
   has already parsed the user's intent. Lexical recall plus a model re-rank of the
   top ~20 candidates matches vector search at this scale, at zero infra cost.
2. **An embedding index is a second source of truth.** It can drift out of sync with
   the markdown files it was built from, which breaks this project's first
   invariant — plain markdown is the only source of truth; everything else (including
   the SQLite cache `pkg/store` builds) is a derived, disposable index that
   `forge reindex` can always rebuild from scratch.
3. **It's a deferred upgrade, not a foreclosed one.** The interface is designed so a
   `recall.strategy: lexical|hybrid` config flag could add embeddings later, behind a
   flag, if someone actually hits the lexical ceiling. Nothing here forecloses that;
   it just refuses to build it speculatively.

## Consequences

- Zero-model-call retrieval holds as an actual property of the system, not just a
  marketing claim — `forge recall`, `forge drift`, and `forge check` all run offline,
  with no API key and no network call, which is also what makes them safe to run from
  a git hook on every commit.
- Phase 2's own measurement work found DESIGN §8's literal scoring formula needed two
  corrections against real vault behavior (a weighted **mean** over active channels,
  not a literal weighted sum; **F₂**, not Dice, for the title measure) — see
  `references/recall-spec.md`. This ADR records *why lexical*, not the exact scoring
  formula; the spec is the source of truth for that.
- Recall quality is capped by lexical overlap. BACKLOG **B-008** tracks one open
  calibration gap this causes (a query whose meaning-bearing terms get filtered out of
  the denominator when no note carries them) — a known, accepted cost of this
  trade-off, not a defect in the decision itself.
