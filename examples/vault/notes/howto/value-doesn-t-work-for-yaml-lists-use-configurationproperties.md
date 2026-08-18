---
title: "@Value Doesn't Work for YAML Lists — Use @ConfigurationProperties"
slug: value-doesn-t-work-for-yaml-lists-use-configurationproperties
type: howto
stack: [spring-boot, java]
tags: [yaml, config]
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

# @Value Doesn't Work for YAML Lists — Use @ConfigurationProperties

## What is it?

`@Value("${some.list}")` will always throw `PlaceholderResolutionException` when the property is a YAML sequence. Spring's `YamlPropertySourceLoader` indexes list entries as `key[0]`, `key[1]`, etc., so the bare key `some.list` never exists as a resolvable flat string.

## How it works

Given this YAML:
```yaml
cors:
  allowed-origins:
    - http://localhost:3000
    - https://myapp.com
```

Spring registers these keys internally:
- `cors.allowed-origins[0]` = `http://localhost:3000`
- `cors.allowed-origins[1]` = `https://myapp.com`
- `cors.allowed-origins` → **does not exist**

So `@Value("${cors.allowed-origins}")` throws at startup.

## Key implementation steps

```java
@ConfigurationProperties(prefix = "cors")
@Component
public class CorsProperties {
    private List<String> allowedOrigins;
    // getter + setter or use Lombok @Data
}
```

```java
@EnableConfigurationProperties(CorsProperties.class)
@Configuration
public class AppConfig { }
```

## Common pitfalls

- Using `@Value` for any YAML list or map — always fails
- Forgetting `@EnableConfigurationProperties` or missing the `@Component` annotation
- Using `@Value` works fine for scalar strings and integers — the problem is specific to sequences

## When to use / not use

- **Use `@ConfigurationProperties`** — any time you bind a YAML list, map, or nested object
- **Use `@Value`** — only for simple scalar properties (strings, numbers, booleans)

## References

- [Spring Boot Config Properties Docs](https://docs.spring.io/spring-boot/docs/current/reference/html/configuration-metadata.html)
