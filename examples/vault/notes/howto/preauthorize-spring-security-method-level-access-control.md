---
title: "@PreAuthorize — Spring Security Method-Level Access Control"
slug: preauthorize-spring-security-method-level-access-control
type: howto
stack: [spring-boot, spring-security, java]
tags: [security]
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

# @PreAuthorize — Spring Security Method-Level Access Control

## What is it?

`@PreAuthorize` is a Spring Security annotation that evaluates a SpEL (Spring Expression Language) expression **before** a method executes. If the expression returns false, Spring throws `AccessDeniedException` → HTTP 403. The method body never runs.

## How it works

```java
@GetMapping("/admin/users")
@PreAuthorize("hasRole('ADMIN')")   // evaluated before method runs
public List<UserResponse> getAllUsers() {
    // only reaches here if caller has ROLE_ADMIN
}
```

Requires `@EnableMethodSecurity` on a `@Configuration` class:

```java
@Configuration
@EnableMethodSecurity   // Spring Boot 3.x+
public class SecurityConfig extends WebSecurityConfigurerAdapter { }
```

## Key implementation steps

**Common SpEL expressions:**
```java
@PreAuthorize("hasRole('ADMIN')")
@PreAuthorize("hasAnyRole('ADMIN', 'MODERATOR')")
@PreAuthorize("isAuthenticated()")
// Access own resource or be admin:
@PreAuthorize("hasRole('ADMIN') or #id == authentication.principal.id")
```

**Belt-and-suspenders pattern** — route-level + method-level:
```java
// SecurityFilterChain — broad protection
.requestMatchers("/api/admin/**").authenticated()

// Controller method — precise role check
@PostMapping("/api/admin/users")
@PreAuthorize("hasRole('ADMIN')")
public ResponseEntity<UserResponse> createAdmin(...) { }
```

**Never expose role via request body — use separate endpoints:**
```java
POST /api/users        // always creates USER role
POST /api/admin/users  // creates ADMIN role, @PreAuthorize("hasRole('ADMIN')")
```

## Common pitfalls

- Forgetting `@EnableMethodSecurity` — annotations are silently ignored without it
- In older Spring Security: `@EnableGlobalMethodSecurity(prePostEnabled = true)` was used instead
- Roles in Spring Security are prefixed with `ROLE_` internally — `hasRole('ADMIN')` checks for `ROLE_ADMIN` in the authority list

## When to use / not use

Use `@PreAuthorize` for per-method, fine-grained checks (especially when the check involves method parameters like `#id`). Use `SecurityFilterChain` for broad URL pattern protection. Both together = defense in depth.
