---
title: "OpenAPI Code Generation — Contract-First with openapi-generator-maven-plugin"
slug: openapi-code-generation-contract-first-with-openapi-generator-maven-plugin
type: concept
stack: [openapi, spring-boot, maven]
tags: [codegen, rest]
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

# OpenAPI Code Generation — Contract-First with openapi-generator-maven-plugin

Contract-first REST API development: define the OpenAPI spec first, generate server stubs,
implement only the delegate interfaces by hand.

## Plugin Configuration

```xml
<configOptions>
  <delegatePattern>true</delegatePattern>
  <dateLibrary>java8</dateLibrary>
</configOptions>
```

- `delegatePattern: true` — generates `*Api` controller + `*ApiDelegate` interface; only the
  delegate is hand-written
- `dateLibrary: java8` — maps `format: date` fields to `LocalDate`

## Workflow

1. Edit the OpenAPI YAML spec
2. Run `./mvnw generate-sources`
3. Generated code lands in `target/generated-sources/openapi/src/main/java`
4. IDE must be pointed at this path for import resolution — see [[notes/concept/build-helper-maven-plugin-adding-generated-sources-to-compile-path]]

## IDE Setup

IntelliJ does not auto-discover `target/generated-sources` paths. Run `generate-sources` once
after cloning or after spec changes before importing the project.

## Sources

- [[sources/daily/2026-04-13-local-ai-continue-rag-spring]]
- [[sources/daily/2026-04-14-spring-keycloak-postman]]

## Related

- [[notes/concept/build-helper-maven-plugin-adding-generated-sources-to-compile-path]]
- [[notes/concept/meterreadingsservice-spring-boot-4-x-project]]
- [[notes/concept/pom-xml-meterreadingsservice-maven-project-descriptor]]
