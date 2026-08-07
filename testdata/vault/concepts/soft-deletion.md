---
title: "Soft Deletion (flag-based deletes)"
tags: [hibernate, jpa, soft-delete, orm]
source: sources/daily/2026-04-14-spring-keycloak.md
date: 2026-04-14
---

# Soft Deletion (flag-based deletes)

Flag-based record deletion: rows are never physically removed, a `deleted_at` column is
set instead, and queries filter the flagged rows out automatically.

## How

Put `@SQLRestriction("deleted_at IS NULL")` on the entity. Hibernate 6 introduced this
as the replacement for the deprecated `@Where` annotation. Repositories are untouched —
the restriction applies to every generated query.

## Gotcha

Native queries ignore `@SQLRestriction`. Every `nativeQuery = true` must repeat the
predicate manually.

See [[concepts/hibernate]].
