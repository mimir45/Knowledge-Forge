---
title: "saveAndFlush vs save — Hibernate Timestamp Flush Timing"
slug: saveandflush-vs-save-hibernate-timestamp-flush-timing
type: concept
stack: [spring-boot, hibernate, jpa]
tags: [timestamps, patterns]
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

# saveAndFlush vs save — Hibernate Timestamp Flush Timing

## The Problem
`@CreationTimestamp` and `@UpdateTimestamp` are Hibernate flush-phase hooks. They fire at the
moment SQL is executed, not when `persist()` / `save()` is called.

In a `@Transactional` method:
- `save()` schedules the INSERT/UPDATE but does NOT execute it
- Flush happens at transaction commit — AFTER the method returns
- So the entity returned from `save()` still has **null** timestamps

## The Fix
Use `saveAndFlush()` in any method that maps the just-saved entity to a response DTO:

```java
Meter saved = repository.saveAndFlush(meter);
return mapper.toResponse(saved); // timestamps are populated
```

## Rule
Every service method that returns a DTO containing `@CreationTimestamp` or `@UpdateTimestamp`
fields must use `saveAndFlush()`, not `save()`.

## Affected Services (MeterReadingsService)
- `MeterService.create()` and `update()`
- `MeterReadingService.create()` and `update()`
- `UserService.create()` and `update()`

## Sources
- [[sources/daily/2026-04-14-spring-keycloak-postman]]

## Related
- [[notes/decision/decision-use-saveandflush-in-all-methods-returning-timestamp-fields]]
- [[notes/concept/meterreadingsservice-spring-boot-4-x-project]]
