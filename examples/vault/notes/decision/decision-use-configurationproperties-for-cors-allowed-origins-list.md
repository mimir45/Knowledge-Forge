---
title: "Decision: Use @ConfigurationProperties for CORS Allowed-Origins List"
slug: decision-use-configurationproperties-for-cors-allowed-origins-list
type: decision
stack: [spring-boot]
tags: [config, cors, yaml, decision]
depth: 3
confidence: high
created: 2026-04-13
updated: 2026-04-13
verified: 2026-04-13
freshness_days: 730
sources:
  - url: sources/daily/2026-04-13-local-ai-continue-rag-spring.md
    accessed: 2026-04-13
    kind: session
related: []
supersedes: []
forge_version: 2.0.0
origin: import
---

# Decision: Use @ConfigurationProperties for CORS Allowed-Origins List

**Decision:** Replaced `@Value("${cors.allowed-origins}")` with `@ConfigurationProperties(prefix = "cors")`
bean for injecting the CORS allowed-origins YAML list.

**Rationale:** `@Value` cannot resolve YAML list keys — Spring indexes them as `key[0]`, `key[1]`
so the bare key never exists. `@ConfigurationProperties` handles YAML lists natively via relaxed binding.

## Sources
- [[sources/daily/2026-04-13-local-ai-continue-rag-spring]]
- [[sources/daily/2026-04-14-spring-keycloak-postman]]

## Related
- [[notes/concept/configurationproperties-vs-value-yaml-list-binding]]
- [[notes/concept/meterreadingsservice-spring-boot-4-x-project]]
