---
title: "Testcontainers — Docker-based Integration Testing"
slug: testcontainers-docker-based-integration-testing
type: concept
stack: [testcontainers, spring-boot, docker, java]
tags: [testing]
depth: 3
confidence: low
created: 2026-04-15
updated: 2026-04-15
verified: 2026-04-15
freshness_days: 365
sources:
  - url: sources/daily/2026-04-15-spring-refactor-testcontainers.md
    accessed: 2026-04-15
    kind: session
related: []
supersedes: []
forge_version: 2.0.0
origin: import
---

# Testcontainers — Docker-based Integration Testing

Testcontainers spins up real Docker containers (Postgres, Keycloak, etc.) for integration tests,
replacing mocks with actual services.

## Spring Boot 4 Changes

- Prefix changed from `org.testcontainers` to `testcontainers` in several artifact IDs
- `spring-boot-testcontainers` auto-configuration replaces manual `@DynamicPropertySource` setup

## Docker Socket Configuration

Docker Desktop CLI socket returns HTTP 400 on macOS. Fix:

```java
// in test application.properties or @TestConfiguration
tc.host=/var/run/docker.sock
```

See [[notes/pitfall/issue-testcontainers-http-400-from-docker-desktop-cli-socket]] and [[notes/decision/decision-set-tc-host-to-var-run-docker-sock]].

## Sources

- [[sources/daily/2026-04-15-spring-refactor-testcontainers]]
- [[sources/daily/2026-04-13-local-ai-continue-rag-spring]]

## Related

- [[notes/pitfall/issue-testcontainers-http-400-from-docker-desktop-cli-socket]]
- [[notes/decision/decision-set-tc-host-to-var-run-docker-sock]]
- [[notes/concept/meterreadingsservice-spring-boot-4-x-project]]
