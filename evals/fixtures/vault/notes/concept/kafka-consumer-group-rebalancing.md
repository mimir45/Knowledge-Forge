---
title: "Kafka consumer group rebalancing"
slug: kafka-consumer-group-rebalancing
type: concept
stack: [java, spring-boot]
tags: [messaging, consumer-group]
depth: 3
confidence: high
created: 2026-08-01
updated: 2026-08-01
verified: 2026-08-01
freshness_days: 365
sources:
  - url: https://kafka.apache.org/documentation/
    accessed: 2026-08-01
    kind: official
related: ["[[configure-kafka-consumer-timeouts]]"]
supersedes: []
forge_version: 2.0.0
origin: ask
---

# Kafka consumer group rebalancing

A consumer group rebalances when a member joins, leaves, or is judged dead. See
[[configure-kafka-consumer-timeouts]] for the timeout knobs that control how fast a
dead member is detected.
