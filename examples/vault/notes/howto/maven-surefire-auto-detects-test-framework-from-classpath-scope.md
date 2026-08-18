---
title: "Maven Surefire Auto-Detects Test Framework from Classpath Scope"
slug: maven-surefire-auto-detects-test-framework-from-classpath-scope
type: howto
stack: [maven, junit, testng, java]
tags: [testing]
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

# Maven Surefire Auto-Detects Test Framework from Classpath Scope

## What is it?

Maven Surefire 3.x automatically detects which test framework to use by scanning the **compile classpath**. If it finds a non-JUnit framework (like TestNG) at `compile` scope instead of `test` scope, it silently switches to that framework's provider — making JUnit 5 symbols appear unresolvable even when the JUnit jars are present transitively.

## How it works

Surefire inspects the classpath for known test framework markers. TestNG at `compile` scope means it appears in the compile classpath, which triggers the TestNG provider. The JUnit Platform Launcher is never initialized, so `org.junit.jupiter.api` is reported as missing despite being on the classpath.

The error looks like a missing dependency:
```
[ERROR] package org.junit.jupiter.api does not exist
```
But the real cause is a misplaced TestNG dependency.

## Key implementation steps

Check your `pom.xml` for test framework dependencies at the wrong scope:

```xml
<!-- WRONG — TestNG at compile scope flips Surefire to TestNG mode -->
<dependency>
    <groupId>org.testng</groupId>
    <artifactId>testng</artifactId>
    <version>7.x</version>
    <!-- no <scope> = defaults to compile -->
</dependency>

<!-- If you need TestNG, use test scope -->
<dependency>
    <groupId>org.testng</groupId>
    <artifactId>testng</artifactId>
    <version>7.x</version>
    <scope>test</scope>
</dependency>
```

If you're using Spring Boot, `spring-boot-starter-test` already provides the full JUnit 5 stack — remove TestNG entirely unless you specifically need it.

## Common pitfalls

- The error message says "package does not exist" — looks like a missing dep, not a framework detection issue
- Stray TestNG dependencies from copy-paste or old templates
- The fix is usually to delete the TestNG block, not add more JUnit deps

## When to use / not use

If your project uses JUnit 5 exclusively (which Spring Boot projects do by default), never include TestNG at compile scope. If you genuinely need both frameworks, scope TestNG to `test` and be aware that Surefire configuration may need adjustment.
