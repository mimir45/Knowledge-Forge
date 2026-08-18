---
title: "MeterReadingsService — Spring Boot 4.x Project"
slug: meterreadingsservice-spring-boot-4-x-project
type: concept
stack: [spring-boot, java, maven, openapi, keycloak, liquibase, mapstruct]
tags: []
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

# MeterReadingsService — Spring Boot 4.x Project

Spring Boot 4.x REST API for meter readings management. Contract-first design using
openapi-generator-maven-plugin with `delegatePattern: true`.

## Key Architecture Choices
- OpenAPI contract-first: `*ApiDelegate` interfaces generated; only delegates are hand-implemented
- `@SQLRestriction("deleted_at IS NULL")` for soft delete (Hibernate 6 replacement for `@Where`)
- `dateLibrary: java8` maps `format: date` → `LocalDate`
- Cursor-based pagination with compound OR predicate (keyset)
- Keycloak JWT authentication; Google social login via broker pattern

## Project Path
`[REDACTED-PATH]`

## Generated Sources
Run `./mvnw generate-sources` to populate `target/generated-sources/openapi/src/main/java`.
Must run before IDE can resolve generated imports.

## Sources
- [[sources/daily/2026-04-13-local-ai-continue-rag-spring]]
- [[sources/daily/2026-04-14-spring-keycloak-postman]]
- [[sources/daily/2026-04-15-spring-refactor-testcontainers]]

## Related
- [[notes/concept/pom-xml-meterreadingsservice-maven-project-descriptor]]
- [[notes/concept/docker-compose-local-yaml-local-development-docker-compose]]
- [[notes/concept/keycloak-realm-export-json-keycloak-realm-definition]]
- [[notes/concept/soft-delete-flag-based-record-deletion-pattern]]
- [[notes/concept/keyset-pagination-compound-or-predicate]]
- [[notes/concept/saveandflush-vs-save-hibernate-timestamp-flush-timing]]
- [[notes/concept/hibernate-orm-patterns-and-gotchas]]
- [[notes/concept/liquibase-database-migration-conventions]]
- [[notes/concept/openapi-code-generation-contract-first-with-openapi-generator-maven-plugin]]
- [[notes/decision/decision-use-configurationproperties-for-cors-allowed-origins-list]]
- [[notes/decision/decision-liquibase-migration-over-column-alias-for-column-rename]]
- [[notes/decision/decision-use-saveandflush-in-all-methods-returning-timestamp-fields]]
- [[notes/decision/decision-separate-endpoint-for-admin-user-creation]]
