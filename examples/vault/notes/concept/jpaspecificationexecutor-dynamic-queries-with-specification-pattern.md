---
title: "JpaSpecificationExecutor — Dynamic Queries with Specification Pattern"
slug: jpaspecificationexecutor-dynamic-queries-with-specification-pattern
type: concept
stack: [spring-boot, jpa, hibernate]
tags: [pagination, queries]
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

# JpaSpecificationExecutor — Dynamic Queries with Specification Pattern

`JpaSpecificationExecutor<T>` extends a Spring Data repository with the ability to accept
`Specification<T>` objects — composable predicate builders for dynamic queries.

## Repository Setup

```java
public interface MeterReadingRepository
    extends JpaRepository<MeterReading, UUID>,
            JpaSpecificationExecutor<MeterReading> {}
```

## Keyset Pagination Use Case

Used to build the compound OR predicate for cursor-based (keyset) pagination, where a
simple `WHERE id > cursor` is insufficient for compound sort keys:

```java
Specification<MeterReading> afterCursor(UUID id, LocalDate date) {
    return (root, query, cb) -> cb.or(
        cb.greaterThan(root.get("date"), date),
        cb.and(
            cb.equal(root.get("date"), date),
            cb.greaterThan(root.get("id"), id)
        )
    );
}
```

See [[notes/concept/keyset-pagination-compound-or-predicate]].

## Sources

- [[sources/daily/2026-04-15-spring-refactor-testcontainers]]

## Related

- [[notes/concept/keyset-pagination-compound-or-predicate]]
- [[notes/concept/meterreadingsservice-spring-boot-4-x-project]]
