---
title: "build-helper-maven-plugin — Adding Generated Sources to Compile Path"
slug: build-helper-maven-plugin-adding-generated-sources-to-compile-path
type: concept
stack: [maven, spring-boot, openapi]
tags: [codegen, build]
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

# build-helper-maven-plugin — Adding Generated Sources to Compile Path

Maven only compiles `src/main/java` by default. When `openapi-generator-maven-plugin` writes
output to `target/generated-sources/openapi/...`, the IDE and `javac` won't see it without an
explicit path declaration.

## Configuration

```xml
<plugin>
  <groupId>org.codehaus.mojo</groupId>
  <artifactId>build-helper-maven-plugin</artifactId>
  <executions>
    <execution>
      <phase>generate-sources</phase>
      <goals><goal>add-source</goal></goals>
      <configuration>
        <sources>
          <source>
            target/generated-sources/openapi/src/main/java
          </source>
        </sources>
      </configuration>
    </execution>
  </executions>
</plugin>
```

## Two-Level-Deep Path Quirk

The generated path is `target/generated-sources/openapi/src/main/java` — note the two extra
levels (`src/main/java`) inside `openapi/`. Omitting these causes silent compile failures where
the classes appear to exist on disk but are invisible to the compiler.

## Sources

- [[sources/daily/2026-04-13-local-ai-continue-rag-spring]]

## Related

- [[notes/concept/openapi-code-generation-contract-first-with-openapi-generator-maven-plugin]]
- [[notes/concept/pom-xml-meterreadingsservice-maven-project-descriptor]]
- [[notes/concept/dev-tools-continue-dev-and-intellij-integration-summary]]
