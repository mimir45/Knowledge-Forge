---
title: "Issue: MapStruct Unmapped Target Properties Warning"
slug: issue-mapstruct-unmapped-target-properties-warning
type: pitfall
stack: [mapstruct, spring-boot, java]
tags: [warnings, decision-pending]
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

# Issue: MapStruct Unmapped Target Properties Warning

## Status

**Decision pending** — warning present, suppression strategy not chosen.

## Symptom

`MeterMapper` produces compile-time warnings about unmapped target properties:
`id`, `userId`, `deletedAt`, `createdAt`.

## Options

1. `@BeanMapping(ignoreByDefault = true)` on each mapping method — suppresses all unmapped
   warnings; requires explicit `@Mapping` for every field you DO want mapped
2. `@Mapper(unmappedTargetPolicy = ReportingPolicy.IGNORE)` — silences all warnings globally
   on that mapper class
3. Explicitly map all fields including `id`, `userId`, etc. — verbose but complete

## Sources

- [[sources/daily/2026-04-14-spring-keycloak-postman]]

## Related

- [[notes/concept/meterreadingsservice-spring-boot-4-x-project]]
- [[notes/concept/open-questions-unresolved-topics-and-open-threads]]
