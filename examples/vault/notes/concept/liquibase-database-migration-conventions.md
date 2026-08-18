---
title: "Liquibase — Database Migration Conventions"
slug: liquibase-database-migration-conventions
type: concept
stack: [liquibase, spring-boot]
tags: [databases, migration, yaml]
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

# Liquibase — Database Migration Conventions

Liquibase manages incremental database schema changes via versioned changelogs.

## Changeset Immutability

Once a changeset has been applied to any environment, it must never be modified. Liquibase
checksums the changeset content and will refuse to run if it detects a modification.
Add a new changeset instead of editing existing ones.

## YAML Comma Parsing

YAML flow mappings treat commas as delimiters. Values containing commas (e.g. addresses,
lists of roles) must be quoted:

```yaml
# WRONG — comma splits into multiple tokens
defaultValue: "123 Main St, Apt 4"

# CORRECT
defaultValue: "'123 Main St, Apt 4'"
```

See [[notes/pitfall/issue-liquibase-yaml-flow-mapping-comma-parse-error]].

## OSS-Compatible Change Types Only

Avoid Pro-only change types like `createSchema` — these silently fail or throw on OSS builds.
Use only change types documented in the OSS Liquibase reference.

## Column Rename Strategy

Prefer a migration (`renameColumn`) over `@Column(name = "...")` alias annotations. The alias
approach leaks implementation detail into the domain model.

See [[notes/decision/decision-liquibase-migration-over-column-alias-for-column-rename]].

## Sources

- [[sources/daily/2026-04-13-local-ai-continue-rag-spring]]
- [[sources/daily/2026-04-14-spring-keycloak-postman]]

## Related

- [[notes/pitfall/issue-liquibase-yaml-flow-mapping-comma-parse-error]]
- [[notes/decision/decision-liquibase-migration-over-column-alias-for-column-rename]]
- [[notes/concept/meterreadingsservice-spring-boot-4-x-project]]
- [[notes/concept/pom-xml-meterreadingsservice-maven-project-descriptor]]
