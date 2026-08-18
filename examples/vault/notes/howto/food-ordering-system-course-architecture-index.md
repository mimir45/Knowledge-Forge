---
title: Food Ordering System — Course Architecture Index
slug: food-ordering-system-course-architecture-index
type: howto
depth: 3
confidence: low
created: 2026-08-09
updated: 2026-08-09
verified: 2026-08-09
freshness_days: 180
sources: []
related: []
supersedes: []
forge_version: 2.0.0
origin: import
---
# Food Ordering System — Course Architecture Index

> **Source:** Udemy — Microservices: Clean Architecture, DDD, SAGA, Outbox & Kafka  
> **Stack:** Java 17, Spring Boot 3, Kafka, PostgreSQL, Avro  
> **Notes created:** 2026-05-13

---

## Overview

4 microservices that communicate **only through Kafka** (never REST-to-REST):

```
HTTP Client
    │
    ▼
┌─────────────────┐
│  order-service  │──[payment-request topic]──────▶ payment-service
│                 │◀─[payment-response topic]───────┤
│                 │──[restaurant-approval-request]──▶ restaurant-service
│                 │◀─[restaurant-approval-response]──┤
└─────────────────┘
        ▲
        │ [customer topic — read model sync]
customer-service
```

Each service is structured the same way (Hexagonal + DDD):

```
service/
├── domain/
│   ├── domain-core/            ← Pure Java. ZERO Spring. Business rules only.
│   └── application-service/    ← Orchestrates domain + defines ports (interfaces).
├── application/                ← REST controllers
├── dataaccess/                 ← JPA entities + adapters (implements output ports)
├── messaging/                  ← Kafka producers + consumers (implements output ports)
└── container/                  ← Spring Boot @SpringBootApplication + @Bean wiring
```

---

## Notes in This Series

| Note                                | Pattern                      | Key Question Answered                                             |
| ----------------------------------- | ---------------------------- | ----------------------------------------------------------------- |
| [[hexagonal-architecture-ports-and-adapters]]       | Hexagonal / Ports & Adapters | Why no @Service in domain-core? How do layers connect?            |
| [[ddd-aggregates-value-objects-domain-events-domain-services]] | Domain-Driven Design         | What are aggregates, value objects, domain services?              |
| [[saga-pattern-choreography-based]]                 | SAGA (Choreography)          | How do you handle distributed transactions without 2PC?           |
| [[transactional-outbox-pattern]]               | Transactional Outbox         | How do you guarantee Kafka publish without dual-write?            |
| [[cqrs-and-event-driven-messaging]]           | CQRS + Event-Driven          | How are commands/queries separated? How does CQRS tie with Kafka? |

---

## Key Differences From Usual Spring Boot

| Aspect | Usual Spring Boot | This Architecture |
|--------|-----------------|-------------------|
| Business logic | In `@Service` | In domain entities and domain services |
| Database coupling | `@Service` imports JPA | Domain never imports JPA |
| Kafka publish | `KafkaTemplate.send()` in service | Written to outbox DB table first, scheduler publishes later |
| Service-to-service | REST calls | Only Kafka topics |
| Failure handling | try/catch | SAGA compensating transactions |
| Idempotency | Usually ignored | SagaId + status check on every incoming event |

---

## Why It Matters

- **Swap PostgreSQL → MongoDB?** Write a new adapter. Zero domain changes.  
- **Kafka goes down at commit time?** Outbox guarantees the message will be sent when it recovers.  
- **Duplicate Kafka message?** SagaId idempotency check silently drops it.  
- **Restaurant rejects after payment succeeds?** SAGA rollback sends a payment cancel request automatically.
