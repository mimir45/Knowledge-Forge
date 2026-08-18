---
title: "Decision: Set tc.host to /var/run/docker.sock"
slug: decision-set-tc-host-to-var-run-docker-sock
type: decision
stack: [testcontainers, docker]
tags: [testing, decision]
depth: 3
confidence: high
created: 2026-04-15
updated: 2026-04-15
verified: 2026-04-15
freshness_days: 730
sources:
  - url: sources/daily/2026-04-15-spring-refactor-testcontainers.md
    accessed: 2026-04-15
    kind: session
related: []
supersedes: []
forge_version: 2.0.0
origin: import
---

# Decision: Set tc.host to /var/run/docker.sock

**Decision:** Changed `tc.host` in `~/.testcontainers.properties` from
`unix://[REDACTED-PATH]` to `unix:///var/run/docker.sock`.

**Rationale:** Docker Desktop has two socket types:
- `~/.docker/desktop/docker.sock` — CLI socket; rejects API negotiation version `v1.24` with HTTP 400
- `/var/run/docker.sock` — daemon socket; works correctly

The docker-java API version system properties (`api.version`, `DOCKER_API_VERSION`) are ineffective
in the shaded Testcontainers bundle — neither overrides the hardcoded `v1.24` minimum.
Socket selection is the correct lever.

## Sources
- [[sources/daily/2026-04-15-spring-refactor-testcontainers]]

## Related
- [[notes/pitfall/issue-testcontainers-http-400-from-docker-desktop-cli-socket]]
- [[notes/concept/pom-xml-meterreadingsservice-maven-project-descriptor]]
