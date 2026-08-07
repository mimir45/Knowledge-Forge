---
title: "Keyset Cursor Pagination Requires a Compound OR Predicate"
date: 2026-04-17
tags: [java, spring-boot, sql, pagination, jpa]
---

# Keyset Cursor Pagination Requires a Compound OR Predicate

## What is it?

Keyset (cursor-based) pagination driven by a single `createdAt > :cursor` predicate
silently skips rows whenever several records share the same `createdAt` timestamp.

## How it works

**Broken:** `WHERE createdAt > :cursorDate` — every row sharing the cursor's timestamp
is dropped, including ones the previous page never returned.

**Correct:**

```sql
WHERE (createdAt > :cursorDate)
   OR (createdAt = :cursorDate AND id > :cursorId)
```

The cursor must therefore carry both fields, and `ORDER BY` must match the predicate
exactly: `ORDER BY created_at, id`.

## Why it matters

The bug is invisible in test data with unique timestamps and only appears under bulk
inserts, where many rows land in the same millisecond.
