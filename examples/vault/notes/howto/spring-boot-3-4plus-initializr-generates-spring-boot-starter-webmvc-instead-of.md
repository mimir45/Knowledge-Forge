---
title: "Spring Boot 3.4+: Initializr generates spring-boot-starter-webmvc instead of spring-boot-starter-web"
slug: spring-boot-3-4plus-initializr-generates-spring-boot-starter-webmvc-instead-of
type: howto
stack: [spring-boot]
tags: [initializr, starter]
depth: 3
confidence: low
created: 2026-04-30
updated: 2026-04-30
verified: 2026-04-30
freshness_days: 180
sources: []
related: []
supersedes: []
forge_version: 2.0.0
origin: import
---

## What Changed

Spring Boot 3.4+ Initializr generates `spring-boot-starter-webmvc` (explicit artifact) instead of `spring-boot-starter-web` when using `-d=web`.

Both starters work in 3.4+ — `spring-boot-starter-webmvc` is just more explicit about including Spring MVC.

## Workarounds

**Pin older Boot version:**
```bash
spring init -d=web --boot-version=3.3.6 myapp
```

**Shell function to swap after generation:**
```bash
springinit() {
  spring init "$@" && find . -name "pom.xml" -exec \
    sed -i '' 's/spring-boot-starter-webmvc/spring-boot-starter-web/g' {} +
}
```
