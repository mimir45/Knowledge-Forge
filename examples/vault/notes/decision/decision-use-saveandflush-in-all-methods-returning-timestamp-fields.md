---
title: "Decision: Use saveAndFlush() in All Methods Returning Timestamp Fields"
slug: decision-use-saveandflush-in-all-methods-returning-timestamp-fields
type: decision
stack: [spring-boot, hibernate, jpa]
tags: [timestamps, decision]
depth: 3
confidence: high
created: 2026-04-14
updated: 2026-04-14
verified: 2026-04-14
freshness_days: 730
sources:
  - url: sources/daily/2026-04-14-spring-keycloak-postman.md
    accessed: 2026-04-14
    kind: session
related: []
supersedes: []
forge_version: 2.0.0
origin: import
---

# Decision: Use saveAndFlush() in All Methods Returning Timestamp Fields

**Decision:** All service methods that map a just-saved entity to a response DTO containing
`@CreationTimestamp` or `@UpdateTimestamp` fields must use `saveAndFlush()`, not `save()`.

**Rationale:** `save()` in a `@Transactional` method schedules the SQL but doesn't execute it.
Timestamps are populated at flush time (SQL execution). Using `save()` returns the entity with
null timestamps before the transaction commits.

## Affected Services
- MeterService.create(), update()
- MeterReadingService.create(), update()
- UserService.create(), update()

## Sources
- [[sources/daily/2026-04-14-spring-keycloak-postman]]

## Related
- [[notes/concept/saveandflush-vs-save-hibernate-timestamp-flush-timing]]
- [[notes/concept/meterreadingsservice-spring-boot-4-x-project]]
