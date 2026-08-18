---
title: "Decision: Liquibase Migration Over @Column Alias for Column Rename"
slug: decision-liquibase-migration-over-column-alias-for-column-rename
type: decision
stack: [spring-boot, liquibase, hibernate]
tags: [decision]
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

# Decision: Liquibase Migration Over @Column Alias for Column Rename

**Decision:** Added Liquibase changeset 006 to `RENAME COLUMN note TO notes` rather than adding
`@Column(name = "note")` on the entity field.

**Rationale:** An `@Column` alias creates a hidden mapping mismatch between Java field names and
DB column names. Liquibase migration keeps entity fields and DB columns in sync with no hidden
surprises. Future developers reading the entity see the real column name.

## Sources
- [[sources/daily/2026-04-14-spring-keycloak-postman]]

## Related
- [[notes/pitfall/issue-hibernate-ddl-auto-validate-note-notes-column-mismatch]]
- [[notes/concept/meterreadingsservice-spring-boot-4-x-project]]
