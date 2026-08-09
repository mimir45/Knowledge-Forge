# Duplicate detection: measured deviations from ADDENDUM §B.4

`duplicates.md` is specified in ADDENDUM §B.4 as "note pairs >0.85 similar". Three things
about that line are wrong for this corpus, and all three were found by measurement against
the real 91-note vault and the fixture vault's deliberate F7 near-duplicate
(`testdata/vault/concepts/soft-delete.md` ↔ `soft-deletion.md`). This file records what
replaced it and why, so a later phase does not restore the original numbers on the strength
of the doc alone. Where this file and §B.4 disagree, this file is what the code does.

## 1. The threshold is 0.40, not 0.85

At 0.85 the report renders **zero rows** on the real vault, and F7 is never nominated at
any shingle width — the widest measured was w=2 at 0.322. 0.85 detects copy-paste. This
vault contains none: its duplicates are notes written months apart about one behaviour,
independently reworded.

`similarity.DuplicateThreshold = 0.40` sits four standard errors below F7's exact 0.575
(the 128-hash sketch's standard error at that score is 0.044), so sampling noise cannot drop
the pair the fixture exists to catch. Counts at candidate thresholds, same-type pairs only:

| threshold | fixture | real vault |
|---|---|---|
| 0.30 | 1 | 18 |
| 0.35 | 1 | 10 |
| **0.40** | **1** | **3** |
| 0.50 | 1 | 1 |
| 0.85 | 0 | 0 |

## 2. Shingles are one word wide, not five

Five is the conventional width for prose and it is wrong here: word order carries no signal
between two independent rewordings, so every extra word of context costs overlap. F7 against
its nearest same-type non-duplicate, by width — w=1 0.575/0.196, w=2 0.322/0.067,
w=3 0.214/0.032, w=5 0.096/0.006. The margin shrinks monotonically; w=1 (a bag of words) is
the widest separation and the only width at which F7 outranks every real-vault same-type
pair.

## 3. Only same-type pairs are scored

This is the load-bearing constraint. Scored across all types, the real vault's top pairs are
0.609 / 0.593 / 0.531 / 0.529 — all `decision`↔`pitfall` or `concept`↔`decision`, and none of
them a duplicate. A decision note and the pitfall that caused it share nearly all their
vocabulary *by design*; that separation is exactly what BACKLOG B-005's seven-type taxonomy
exists to create. Cross-type scoring ranked five vault non-duplicates above F7 and made the
report unusable at every threshold. Restricted to same-type pairs the real vault's ceiling is
0.504 and F7 (0.575) is top of its corpus.

`Index.Add` takes the group explicitly. The group is derived from the note's **directory**,
not its `type:` frontmatter key — F7's two notes have no frontmatter at all, and the fixture
deliberately contains notes with none. Directory works in both topologies (`concepts/`
pre-migration, `notes/concept/` after).

## 4. Consequence for band tuning

Banding is a recall device; `Estimate` decides. So Rows is chosen for P ≥ 0.999 *at* the
threshold, not for a steep curve there: 64×2 gives P(0.40) = 0.99998. Two earlier tunings
were wrong and both failed **silently** — 16×8 gave P(0.575) ≈ 0.11 and 32×4 gave
P(0.40) = 0.56, each returning an empty report while `Estimate` agreed the pair was a
duplicate. `TestFixtureNearDuplicateIsNominated` is the regression: it asserts the F7 pair
survives the candidate stage end to end.

The price is candidate volume — at s=0.10, 64×2 nominates about half of all pairs. That is
1142 signature comparisons on the real vault. Not worth optimising until a vault is large
enough to measure it; `BenchmarkPairs500Notes` bounds the no-pruning case at 90ms.

## 5. What the report must say

`duplicates.md` states in its header that **no pair in the vault crosses §B.4's 0.85**. That
is the honest headline, not a footnote: a reader who knows the spec must not read three rows
at 0.40 as three rows at 0.85.
