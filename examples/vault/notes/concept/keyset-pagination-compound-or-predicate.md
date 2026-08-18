---
title: "Keyset Pagination — Compound OR Predicate"
slug: keyset-pagination-compound-or-predicate
type: concept
stack: [spring-boot, jpa, hibernate]
tags: [pagination, performance]
depth: 3
confidence: low
created: 2026-04-15
updated: 2026-04-15
verified: 2026-04-15
freshness_days: 365
sources:
  - url: sources/daily/2026-04-15-spring-refactor-testcontainers.md
    accessed: 2026-04-15
    kind: session
related: []
supersedes: []
forge_version: 2.0.0
origin: import
---

# Keyset Pagination — Compound OR Predicate

Cursor-based (keyset) pagination using `(createdAt, id)` as the composite cursor.

## Correct Predicate
Simple `createdAt > X` is **broken** — it skips rows when timestamps tie.

Correct compound OR predicate:
```
(createdAt > cursorDate) OR (createdAt = cursorDate AND id > cursorId)
```

This handles timestamp ties correctly by falling back to the ID tiebreaker.

## Null Guard
`cb.conjunction()` (SQL `1=1`) is used as a null-safe guard for optional filter parameters,
allowing `Specification` objects to compose cleanly without service-layer null branching.

## Hibernate 6 UUID Note
`GenerationType.UUID` produces UUIDv4 (random) — not sequential. The compound OR predicate
is still correct for the declared `(createdAt ASC, id ASC)` sort order.

## Termination
Malformed cursors must throw `IllegalArgumentException` explicitly — never silently fall back
to a full table scan (correctness and performance bug).

## Sources
- [[sources/daily/2026-04-15-spring-refactor-testcontainers]]

## Related
- [[notes/concept/meterreadingsservice-spring-boot-4-x-project]]
