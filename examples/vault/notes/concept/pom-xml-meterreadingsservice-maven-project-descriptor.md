---
title: "pom.xml — MeterReadingsService Maven Project Descriptor"
slug: pom-xml-meterreadingsservice-maven-project-descriptor
type: concept
stack: [maven, spring-boot, java, testcontainers]
tags: [openapi-generator]
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

# pom.xml — MeterReadingsService Maven Project Descriptor

Maven POM for the MeterReadingsService Spring Boot 4.x project.

## Notable Configurations
- `openapi-generator-maven-plugin` with `delegatePattern: true` and `dateLibrary: java8`
- `build-helper-maven-plugin` registers `target/generated-sources/openapi/src/main/java` as source root
  (path is TWO levels deep — generator places files under `src/main/java` subdirectory)
- `<build><testSourceDirectory>src/test/java</testSourceDirectory></build>` — standard declaration
  for IntelliJ to mark it correctly as Test Sources Root
- Testcontainers 2.x: artifact IDs prefixed `testcontainers-*` (e.g., `testcontainers-postgresql`)
- Removed `testng` dependency (was at compile scope, caused Surefire to switch to TestNG mode)
- Added `spring-boot-starter-webmvc-test` (test scope) for `@AutoConfigureMockMvc`

## Key Commands
```bash
./mvnw generate-sources   # generates OpenAPI classes
./mvnw test               # runs all tests
```

## Sources
- [[sources/daily/2026-04-13-local-ai-continue-rag-spring]]
- [[sources/daily/2026-04-14-spring-keycloak-postman]]
- [[sources/daily/2026-04-15-spring-refactor-testcontainers]]

## Related
- [[notes/concept/meterreadingsservice-spring-boot-4-x-project]]
- [[notes/pitfall/issue-spring-boot-4-x-testcontainers-2-x-artifact-renames]]
- [[notes/pitfall/issue-testcontainers-http-400-from-docker-desktop-cli-socket]]
