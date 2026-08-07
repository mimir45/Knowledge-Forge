---
title: "Soft Delete"
tags: [hibernate, jpa, soft-delete, java]
source: sources/daily/2026-04-13-local-ai-spring.md
date: 2026-04-13
status: active
---

# Soft Delete

Flag-based record deletion: rows are never removed, a `deleted_at` column is set instead,
and every query filters the flagged rows out automatically.

## How

Annotate the entity with `@SQLRestriction("deleted_at IS NULL")`. In Hibernate 6 this
replaces the deprecated `@Where` annotation. The repository needs no changes — the
restriction is applied to all generated queries, including collection loads.

## Gotcha

Native queries bypass `@SQLRestriction` entirely. Any `@Query(nativeQuery = true)` must
repeat the predicate by hand.

See [[concepts/hibernate]].
