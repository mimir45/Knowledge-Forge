---
title: "Issue: Testcontainers HTTP 400 from Docker Desktop CLI Socket"
slug: issue-testcontainers-http-400-from-docker-desktop-cli-socket
type: pitfall
stack: [testcontainers, docker]
tags: [testing, issue]
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

# Issue: Testcontainers HTTP 400 from Docker Desktop CLI Socket

## Symptom
Integration tests fail with HTTP 400 error when Testcontainers tries to connect to Docker.

## Root Cause
`tc.host` pointed at `~/.docker/desktop/docker.sock` which is the Docker Desktop **CLI socket**.
This socket requires minimum API version `v1.40` and rejects the default docker-java negotiation
version `v1.24` with HTTP 400.

The docker-java API version system properties (`api.version`, `DOCKER_API_VERSION`) do not work
in the shaded Testcontainers bundle — hardcoded `VERSION = "1.24"` minimum in `DockerConfigFile`.

## Fix
Change `tc.host` in `~/.testcontainers.properties` to:
```
tc.host=unix:///var/run/docker.sock
```

## Affected Files
- `~/.testcontainers.properties`

## Sources
- [[sources/daily/2026-04-15-spring-refactor-testcontainers]]

## Related
- [[notes/decision/decision-set-tc-host-to-var-run-docker-sock]]
- [[notes/concept/pom-xml-meterreadingsservice-maven-project-descriptor]]
