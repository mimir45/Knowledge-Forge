---
title: "@CreationTimestamp Is Null Until Flush"
date: 2026-04-17
tags: [java, spring-boot, hibernate, jpa]
---

# @CreationTimestamp Is Null Until Flush

## What is it?

`save()` returns an entity whose `@CreationTimestamp` field is still null. The value is
generated when the persistence context flushes, which by default happens at transaction
commit — after the method has already built its response DTO.

## How it works

```java
// Broken: dto.createdAt == null
var saved = repository.save(entity);
return mapper.toDto(saved);

// Correct
var saved = repository.saveAndFlush(entity);
return mapper.toDto(saved);
```

## Why it matters

The null only shows up in DTO-returning methods, so an integration test that re-reads
the row in a fresh transaction passes while the API returns nulls to real clients.
