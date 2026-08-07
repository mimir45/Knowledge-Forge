---
title: "Rename the column via Liquibase, not a @Column alias"
tags: [liquibase, hibernate, migration, decision]
date: 2026-04-13
status: active
---

# Rename the column via Liquibase, not a @Column alias

**Decision:** the `note` → `notes` column mismatch is fixed with a Liquibase changeset,
not by papering over it with `@Column(name = "note")` on the entity.

**Why:** the alias hides the divergence from anyone reading the schema, and the next
migration authored against the real schema would reintroduce the mismatch.

**Cost:** one immutable changeset; existing environments need the migration applied
before the next deploy.

Context: [[issues/hibernate-column-mismatch]], [[concepts/hibernate]].
