---
title: "@CreationTimestamp and @UpdateTimestamp Require a Database Flush"
slug: creationtimestamp-and-updatetimestamp-require-a-database-flush
type: howto
stack: [hibernate, spring-boot, jpa, java]
tags: []
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

# @CreationTimestamp and @UpdateTimestamp Require a Database Flush

## What is it?

`@CreationTimestamp` and `@UpdateTimestamp` are Hibernate lifecycle hooks — they fire when an SQL INSERT or UPDATE is **executed**, not when `save()` is called. In a `@Transactional` method, `save()` schedules the SQL but commits it only at transaction end — after your method returns. This means the entity you return still has `null` timestamps.

## How it works

```
save()  →  entity scheduled  →  method returns  →  transaction commits  →  SQL fires  →  timestamps set
                                 ↑
                           you map to DTO here — timestamps are null
```

`saveAndFlush()` forces immediate SQL execution, so Hibernate fires its lifecycle listeners before you map the entity:

```
saveAndFlush()  →  SQL fires immediately  →  timestamps set  →  you map to DTO  →  timestamps populated
```

## Key implementation steps

```java
// WRONG — timestamps null in response
public MeterResponse create(CreateMeterRequest request) {
    Meter meter = mapper.toEntity(request);
    return mapper.toResponse(repository.save(meter)); // timestamps null
}

// CORRECT — timestamps populated
public MeterResponse create(CreateMeterRequest request) {
    Meter meter = mapper.toEntity(request);
    return mapper.toResponse(repository.saveAndFlush(meter));
}
```

## Common pitfalls

- Affects every service method that maps a just-saved entity to a response DTO
- Also affects `update()` — `@UpdateTimestamp` has the same flush requirement
- Easy to miss: the field exists on the entity, just contains `null` at map time

## When to use / not use

Use `saveAndFlush()` when you need to read back lifecycle-computed values (timestamps, DB-generated IDs, trigger-set columns) in the same request. Use plain `save()` when you don't need the persisted state back immediately (batch processing, fire-and-forget).
