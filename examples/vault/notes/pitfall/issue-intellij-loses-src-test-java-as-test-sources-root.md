---
title: "Issue: IntelliJ Loses src/test/java as Test Sources Root"
slug: issue-intellij-loses-src-test-java-as-test-sources-root
type: pitfall
stack: [intellij, maven, spring-boot]
tags: [issue]
depth: 3
confidence: low
created: 2026-04-14
updated: 2026-04-14
verified: 2026-04-14
freshness_days: 365
sources:
  - url: sources/daily/2026-04-14-spring-keycloak-postman.md
    accessed: 2026-04-14
    kind: session
related: []
supersedes: []
forge_version: 2.0.0
origin: import
---

# Issue: IntelliJ Loses src/test/java as Test Sources Root

## Symptom
IntelliJ shows `org.junit.jupiter.api does not exist`. Maven CLI tests work fine.

## Root Cause
After a Maven reimport, IntelliJ dropped `src/test/java` from the module's test source roots.
`build-helper-maven-plugin add-test-source` registers the directory but IntelliJ categorizes it
as "generated sources" (wrong icon, wrong semantics) because it feeds `getTestSourceRoots()`,
not `getTestSourceDirectory()`.

IntelliJ's `AdditionalModuleElements` in `.iml` stores manual "Mark Directory As" overrides —
wiped on every Maven reload.

## Fix
Use Maven's standard `<build><testSourceDirectory>src/test/java</testSourceDirectory></build>`
declaration. IntelliJ's model reader maps this directly to a "Test Sources Root" (green folder).

## Affected Files
- `pom.xml` — added `<testSourceDirectory>`
- Deleted orphaned `MeterReadingsService.iml` (not in `modules.xml`, had stale entries)

## Sources
- [[sources/daily/2026-04-14-spring-keycloak-postman]]

## Related
- [[notes/concept/pom-xml-meterreadingsservice-maven-project-descriptor]]
- [[notes/concept/meterreadingsservice-spring-boot-4-x-project]]
