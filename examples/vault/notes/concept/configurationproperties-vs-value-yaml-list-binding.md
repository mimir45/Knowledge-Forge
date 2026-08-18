---
title: "ConfigurationProperties vs @Value — YAML List Binding"
slug: configurationproperties-vs-value-yaml-list-binding
type: concept
stack: [spring-boot]
tags: [yaml, config, cors]
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

# ConfigurationProperties vs @Value — YAML List Binding

## The Problem
`@Value("${cors.allowed-origins}")` will always throw `PlaceholderResolutionException` for YAML
sequences. Spring Boot's `YamlPropertySourceLoader` indexes list entries as `key[0]`, `key[1]`,
etc., so the bare key `cors.allowed-origins` does not exist as a resolvable placeholder.

## The Fix
Use `@ConfigurationProperties` which uses relaxed binding and handles YAML lists natively:

```java
@ConfigurationProperties(prefix = "cors")
public class CorsProperties {
    private List<String> allowedOrigins;
}
```

## Rule
Always use `@ConfigurationProperties` for list/map injection from YAML.
`@Value` is suitable only for scalar (single) values.

## Sources
- [[sources/daily/2026-04-13-local-ai-continue-rag-spring]]
- [[sources/daily/2026-04-14-spring-keycloak-postman]]

## Related
- [[notes/decision/decision-use-configurationproperties-for-cors-allowed-origins-list]]
- [[notes/concept/meterreadingsservice-spring-boot-4-x-project]]
