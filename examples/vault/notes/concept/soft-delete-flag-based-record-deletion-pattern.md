---
title: "Soft Delete — Flag-Based Record Deletion Pattern"
slug: soft-delete-flag-based-record-deletion-pattern
type: concept
stack: [spring-boot, hibernate, jpa, sql]
tags: [patterns]
depth: 3
confidence: low
created: 2026-04-14
updated: 2026-04-14
verified: 2026-04-14
freshness_days: 365
sources:
  - url: sources/daily/2026-04-14-spring-keycloak-postman.md
    accessed: 2026-04-14
    kind: session
related: []
supersedes: []
forge_version: 2.0.0
origin: import
---

# Soft Delete — Flag-Based Record Deletion Pattern

Instead of issuing `DELETE FROM table WHERE id = ?`, a soft delete sets a flag (e.g.,
`deleted_at IS NOT NULL` or `deleted = true`) and keeps the row in the database.

## Hibernate 6 Implementation
```java
@SQLRestriction("deleted_at IS NULL")
public class Meter { ... }
```
`@SQLRestriction` is the Hibernate 6 replacement for the deprecated `@Where` annotation.
It transparently appends the filter to all queries without touching repository methods.

## Trade-offs
| Pro | Con |
|-----|-----|
| Full audit trail | Tables grow forever (need archiving strategy) |
| Un-delete is a flag flip | `findById` returns null for soft-deleted records |
| Preserves referential integrity | `@SQLRestriction` must be on every entity |
| GDPR / financial compliance | |

## MeterReadingsService Usage
`MeterService.delete()` sets `deletedAt` and calls `save()` rather than `repository.delete()`.

## Sources
- [[sources/daily/2026-04-14-spring-keycloak-postman]]

## Related
- [[notes/concept/meterreadingsservice-spring-boot-4-x-project]]
