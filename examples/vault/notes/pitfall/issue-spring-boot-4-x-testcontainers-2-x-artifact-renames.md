---
title: "Issue: Spring Boot 4.x / Testcontainers 2.x Artifact Renames"
slug: issue-spring-boot-4-x-testcontainers-2-x-artifact-renames
type: pitfall
stack: [spring-boot, maven, testcontainers]
tags: [issue]
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

# Issue: Spring Boot 4.x / Testcontainers 2.x Artifact Renames

## Symptom
Build fails after scaffolding Spring Boot 4.x project — missing dependencies or wrong artifact IDs.

## Root Cause
Spring Boot 4.x / Spring Framework 7 renamed and reorganized several artifacts:
- `spring-boot-starter-aop` → `spring-boot-starter-aspectj`
- `@AutoConfigureMockMvc` moved to `spring-boot-starter-webmvc-test` module

Testcontainers 2.x prefixed all module artifact IDs:
- Old: `testcontainers` (just the core)
- New: `testcontainers-postgresql`, `testcontainers-junit-jupiter`, etc.

## Fix
Update pom.xml with correct artifact IDs. Use the Spring Boot BOM at
`~/.m2/.../spring-boot-dependencies-4.x.pom` as ground truth for current artifact names.

## Affected Files
- `pom.xml`
- `IntegrationTestBase.java` (import updated for `@AutoConfigureMockMvc`)

## Sources
- [[sources/daily/2026-04-13-local-ai-continue-rag-spring]]
- [[sources/daily/2026-04-15-spring-refactor-testcontainers]]

## Related
- [[notes/concept/pom-xml-meterreadingsservice-maven-project-descriptor]]
- [[notes/concept/meterreadingsservice-spring-boot-4-x-project]]
