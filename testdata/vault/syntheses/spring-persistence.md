---
title: "Spring persistence — compiled view"
tags: [synthesis, spring-boot, jpa, hibernate]
source: multiple
date: 2026-04-15
status: active
---

# Spring persistence — compiled view

Compiled from the persistence-related notes in this vault.

Hibernate 6 patterns and their pitfalls are collected in [[concepts/hibernate]]. The
soft-delete pattern has two overlapping notes — [[concepts/soft-delete]] and
[[concepts/soft-deletion]] — which should be merged. Pagination lives outside the
`concepts/` tree entirely, in [[TIL/databases/keyset-cursor-pagination]], which is
itself part of why the topology needs migrating.

Schema changes go through Liquibase per [[decisions/liquibase-over-column-alias]].
