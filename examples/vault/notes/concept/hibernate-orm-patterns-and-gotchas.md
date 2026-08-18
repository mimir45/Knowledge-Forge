---
title: "Hibernate — ORM Patterns and Gotchas"
slug: hibernate-orm-patterns-and-gotchas
type: concept
stack: [hibernate, spring-boot, jpa, java]
tags: [orm]
depth: 3
confidence: low
created: 2026-04-13
updated: 2026-04-13
verified: 2026-04-13
freshness_days: 365
sources:
  - url: sources/daily/2026-04-13-local-ai-continue-rag-spring.md
    accessed: 2026-04-13
    kind: session
related: []
supersedes: []
forge_version: 2.0.0
origin: import
---

# Hibernate — ORM Patterns and Gotchas

Key patterns and known pitfalls when using Hibernate 6 with Spring Boot.

## Soft Delete

Use `@SQLRestriction("deleted_at IS NULL")` (Hibernate 6 replacement for the deprecated
`@Where` annotation) to filter soft-deleted rows automatically on all queries.

See [[notes/concept/soft-delete-flag-based-record-deletion-pattern]].

## Flush Timing and @CreationTimestamp

`save()` may not flush immediately, so `createdAt` / `updatedAt` fields can be null in the
returned DTO. Use `saveAndFlush()` in any method that returns a DTO relying on these fields.

See [[notes/concept/saveandflush-vs-save-hibernate-timestamp-flush-timing]] and [[notes/decision/decision-use-saveandflush-in-all-methods-returning-timestamp-fields]].

## Column Naming

Hibernate derives column names from field names using camelCase → snake_case conversion.
Field `note` maps to column `note`; field `notes` maps to `notes` — a one-character difference
breaks schema validation silently.

See [[notes/pitfall/issue-hibernate-ddl-auto-validate-note-notes-column-mismatch]].

## Sources

- [[sources/daily/2026-04-13-local-ai-continue-rag-spring]]
- [[sources/daily/2026-04-14-spring-keycloak-postman]]
- [[sources/daily/2026-04-15-spring-refactor-testcontainers]]

## Related

- [[notes/concept/soft-delete-flag-based-record-deletion-pattern]]
- [[notes/concept/saveandflush-vs-save-hibernate-timestamp-flush-timing]]
- [[notes/pitfall/issue-hibernate-ddl-auto-validate-note-notes-column-mismatch]]
- [[notes/concept/meterreadingsservice-spring-boot-4-x-project]]
