---
title: "Spring Boot 4 Breaking Changes — Artifact Renames and Test Module Split"
slug: spring-boot-4-breaking-changes-artifact-renames-and-test-module-split
type: howto
stack: [spring-boot, java, maven]
tags: [migration]
depth: 3
confidence: low
created: 2026-04-17
updated: 2026-04-17
verified: 2026-04-17
freshness_days: 180
sources: []
related: []
supersedes: []
forge_version: 2.0.0
origin: import
---

# Spring Boot 4 Breaking Changes — Artifact Renames and Test Module Split

## What is it?

Spring Boot 4.x (with Spring Framework 7) renamed several starter artifacts and split the test autoconfiguration module. Projects migrating from SB3 will see build failures for these dependencies even though the functionality still exists under new names.

## How it works

### Artifact renames

| SB3 artifact | SB4 artifact |
|---|---|
| `spring-boot-starter-aop` | `spring-boot-starter-aspectj` |
| `testcontainers` (group `org.testcontainers`) modules | All prefixed with `testcontainers-` |

Testcontainers 2.x example:
```xml
<!-- SB3 -->
<dependency>
    <groupId>org.testcontainers</groupId>
    <artifactId>postgresql</artifactId>
</dependency>

<!-- SB4 / TC 2.x -->
<dependency>
    <groupId>org.testcontainers</groupId>
    <artifactId>testcontainers-postgresql</artifactId>
</dependency>
```

### Test module split

`@AutoConfigureMockMvc` moved to a dedicated starter:

```xml
<dependency>
    <groupId>org.springframework.boot</groupId>
    <artifactId>spring-boot-starter-webmvc-test</artifactId>
    <scope>test</scope>
</dependency>
```

Import also changed:
```java
// SB3
import org.springframework.boot.test.autoconfigure.web.servlet.AutoConfigureMockMvc;

// SB4
import org.springframework.boot.webmvc.test.autoconfigure.AutoConfigureMockMvc;
```

### Hibernate annotation changes

`@Where` is deprecated → replaced by `@SQLRestriction`:
```java
// SB3 / Hibernate 5
@Where(clause = "deleted_at IS NULL")

// SB4 / Hibernate 6
@SQLRestriction("deleted_at IS NULL")
```

## Common pitfalls

- Old artifact names silently fail to resolve — no obvious "renamed" message in the error
- The BOM at `~/.m2/.../spring-boot-dependencies-X.X.X.pom` is the ground truth for current artifact names

## References

- [Spring Boot 4.0 Migration Guide](https://github.com/spring-projects/spring-boot/wiki/Spring-Boot-4.0-Migration-Guide)
