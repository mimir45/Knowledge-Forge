---
title: IntelliJ Test Sources Root — Why It Gets Lost and the build-helper Fix
slug: intellij-test-sources-root-why-it-gets-lost-and-the-build-helper-fix
type: howto
stack: [intellij, maven, junit, spring-boot]
tags: [build-helper]
depth: 3
confidence: low
created: 2026-04-14
updated: 2026-04-14
verified: 2026-04-14
freshness_days: 180
sources: []
related: []
supersedes: []
forge_version: 2.0.0
origin: import
---

# IntelliJ Test Sources Root — Why It Gets Lost and the build-helper Fix

## What is it?

IntelliJ tracks source roots per-module in a `.iml` file. The key distinction is `isTestSource="true"` — without it, `src/test/java` is treated as regular source and JUnit symbols from `spring-boot-starter-test` are invisible to the IDE compiler, producing `package org.junit.jupiter.api does not exist`.

## How it works

IntelliJ maintains two separate sections in the `.iml`:

- **Maven model section** — regenerated on every Maven reimport from the effective POM
- **`AdditionalModuleElements`** — manual IDE-only overrides (right-click → Mark Directory As)

`AdditionalModuleElements` is **wiped on Maven reload**. So manually marking `src/test/java` as Test Sources Root fixes the error until the next reimport, then it vanishes.

`src/test/java` is a Maven convention — implicitly known to Maven CLI, but IntelliJ's Maven model reader only writes it into the `.iml` if it can cleanly resolve the model. Any disruption (compile-scope test dependency, plugin resolution order issue) causes the implicit registration to be dropped.

## Key implementation steps

Add an explicit `add-test-source` execution to the `build-helper-maven-plugin` block already in `pom.xml`:

```xml
<execution>
    <id>add-test-sources</id>
    <phase>generate-test-sources</phase>
    <goals>
        <goal>add-test-source</goal>
    </goals>
    <configuration>
        <sources>
            <source>${project.basedir}/src/test/java</source>
        </sources>
    </configuration>
</execution>
```

Then trigger **Maven panel → Reload All Maven Projects** in IntelliJ. IntelliJ processes `build-helper` executions as first-class model contributions (`MavenProject.addTestSourceDirectory()`), so it writes `src/test/java` into the Maven model section of the `.iml` — not `AdditionalModuleElements` — surviving every future reload.

## Common pitfalls

- **Phase must be `generate-test-sources`** not `generate-sources` — build-helper routes `add-test-source` vs `add-source` based on phase
- **Compile-scope test dependency** — `spring-boot-starter-test` without `<scope>test</scope>` is the most common trigger; Surefire and IntelliJ both misread the classpath
- **Orphaned `.iml` files** — IntelliJ only loads the `.iml` registered in `.idea/modules.xml`; a stale `ProjectName.iml` in the root doesn't interfere but is confusing noise
- **"Invalidate Caches" won't help** — the module model is the root cause, not the IDE cache; reload the Maven project instead

## When to use / not use

Use `add-test-source` whenever:
- Your project already has `build-helper-maven-plugin` for other source roots (OpenAPI codegen, etc.)
- The team has recurrent "mark directory as test sources root" complaints

Skip it if you have zero `build-helper` usage — adding the plugin solely for this is overkill; fixing the compile-scope test dep is usually enough.

## References

- [build-helper-maven-plugin: add-test-source goal](https://www.mojohaus.org/build-helper-maven-plugin/add-test-source-mojo.html)
- [IntelliJ IDEA: Content roots](https://www.jetbrains.com/help/idea/content-roots.html)
