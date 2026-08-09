---
# java-backend — for a JVM service developer: Spring Boot, Hibernate, Kafka, Postgres.
#
# Tunes the static core for a codebase where the interesting churn is in service and
# repository classes. Code indexing covers Java and Kotlin; drift watches method
# signatures, so a renamed `@Transactional` method demotes the note that cites it.
#
# api notes go stale fastest here (90 days) because a library upgrade invalidates them
# silently — that is the default and this preset keeps it.

static:
  code_index: true
  languages: [java, kotlin]
  git_signals: true
  drift: {enabled: true, auto_repair_line_numbers: true}

research:
  prefer: [official-docs, source-code, rfc]

verify:
  run_code: auto
  require_citation_for: [version-specific, performance-claim, security-claim]

write:
  diagrams: mermaid
---
