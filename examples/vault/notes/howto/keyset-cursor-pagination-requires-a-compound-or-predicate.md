---
title: "Keyset Cursor Pagination Requires a Compound OR Predicate"
slug: keyset-cursor-pagination-requires-a-compound-or-predicate
type: howto
stack: [java, spring-boot, sql, jpa]
tags: [pagination]
depth: 3
confidence: low
created: 2026-04-17
updated: 2026-04-17
verified: 2026-04-17
freshness_days: 180
sources: []
related: []
supersedes: []
forge_version: 2.0.0
origin: import
---

# Keyset Cursor Pagination Requires a Compound OR Predicate

## What is it?

Keyset (cursor-based) pagination using a single `createdAt > cursor` predicate silently skips rows when multiple records share the same `createdAt` timestamp. The correct predicate is a compound OR that uses both the sort field and a unique tiebreaker (usually `id`).

## How it works

**Broken:** `WHERE createdAt > :cursor` — if three rows have the same timestamp and the page boundary falls mid-group, 1-2 rows are silently dropped.

**Correct:**
```sql
WHERE (createdAt > :cursorDate)
   OR (createdAt = :cursorDate AND id > :cursorId)
ORDER BY createdAt ASC, id ASC
```

This guarantees every row appears exactly once regardless of timestamp ties.

## Key implementation steps

```java
// Cursor encodes both fields
record Cursor(LocalDateTime createdAt, UUID id) {
    static Cursor decode(String encoded) { /* base64 decode */ }
    String encode() { /* base64 encode */ }
}

// JPA Specification
static Specification<Meter> afterCursor(Cursor c) {
    return (root, query, cb) -> cb.or(
        cb.greaterThan(root.get("createdAt"), c.createdAt()),
        cb.and(
            cb.equal(root.get("createdAt"), c.createdAt()),
            cb.greaterThan(root.get("id"), c.id())
        )
    );
}
```

```java
// Malformed cursor must throw, not fall back to full scan
if (encodedCursor != null) {
    try {
        Cursor c = Cursor.decode(encodedCursor);
        spec = spec.and(afterCursor(c));
    } catch (IllegalArgumentException e) {
        throw new BadRequestException("Invalid cursor");
    }
}
```

## Common pitfalls

- Single-field cursor (`createdAt > X`) is broken on timestamp ties — silent data loss
- Silent fallback to full scan on malformed cursor is a correctness and performance bug
- `GenerationType.UUID` in Hibernate 6 produces UUIDv4 (random), not sequential — the compound OR predicate is still correct for `(createdAt ASC, id ASC)` ordering

## When to use / not use

Use keyset pagination (vs. offset) for large datasets or infinite scroll — `OFFSET N` degrades with table size, keyset stays O(log N) with an index. Always index the sort columns: `(createdAt, id)`.
