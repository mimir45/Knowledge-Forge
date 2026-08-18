---
title: "Configure Kafka consumer timeouts"
slug: configure-kafka-consumer-timeouts
type: howto
stack: [java, spring-boot]
tags: [messaging, consumer-group]
depth: 2
confidence: high
created: 2026-08-01
updated: 2026-08-01
verified: 2026-08-01
freshness_days: 365
sources:
  - url: https://kafka.apache.org/documentation/#consumerconfigs
    accessed: 2026-08-01
    kind: official
related: ["[[kafka-consumer-group-rebalancing]]"]
supersedes: []
forge_version: 2.0.0
origin: ask
---

# Configure Kafka consumer timeouts

`session.timeout.ms` and `max.poll.interval.ms` control how fast a dead consumer
triggers [[kafka-consumer-group-rebalancing]].
