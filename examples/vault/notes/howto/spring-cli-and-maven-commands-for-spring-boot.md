---
title: Spring CLI & Maven Commands for Spring Boot
slug: spring-cli-and-maven-commands-for-spring-boot
type: howto
stack: [spring-boot, maven, java]
tags: [spring-cli]
depth: 3
confidence: low
created: 2026-04-30
updated: 2026-04-30
verified: 2026-04-30
freshness_days: 180
sources: []
related: []
supersedes: []
forge_version: 2.0.0
origin: import
---

# Spring CLI & Maven Commands for Spring Boot

## What is Spring CLI?

Spring CLI is a command-line tool that scaffolds Spring Boot projects via `spring init`, which calls [start.spring.io](https://start.spring.io) to generate a project zip and unpack it locally — faster than the web UI for repeatable setups.

---

## Installing Spring CLI

```zsh
brew install spring-io/tap/spring-boot
spring --version
```

---

## Creating a Project with `spring init`

### Common Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--dependencies` | `-d` | Comma-separated list of starters |
| `--java-version` | `-j` | Java version (17, 21, etc.) |
| `--packaging` | `-p` | `jar` (default) or `war` |
| `--build` | | `maven` (default) or `gradle` |
| `--group` | `-g` | Group ID (e.g. `com.samir`) |
| `--artifact` | `-a` | Artifact ID / project name |
| `--name` | `-n` | Display name of the project |
| `--extract` | `-x` | Extract into current directory |

### Common Dependency Shorthands

| Shorthand | What it adds |
|-----------|-------------|
| `web` | Spring MVC + Tomcat |
| `data-jpa` | Spring Data JPA + Hibernate |
| `postgresql` | PostgreSQL JDBC driver |
| `validation` | Bean Validation (Jakarta) |
| `actuator` | Health/metrics endpoints |
| `security` | Spring Security |
| `devtools` | Hot reload during development |
| `lombok` | Lombok annotation processor |
| `flyway` | Flyway database migrations |
| `testcontainers` | Testcontainers integration |

```zsh
# List all available dependencies
spring init --list
```

### Full Example (matches your typical stack)

```zsh
spring init \
  --group=com.samir \
  --artifact=my-api \
  --name=MyApi \
  --java-version=21 \
  --packaging=jar \
  --build=maven \
  --dependencies=web,data-jpa,postgresql,validation,actuator,devtools,lombok \
  my-api
```

---

## Maven Commands for Spring Boot

### Run the App

```zsh
mvn spring-boot:run

# With a profile
mvn spring-boot:run -Dspring-boot.run.profiles=dev

# With JVM args
mvn spring-boot:run -Dspring-boot.run.jvmArguments="-Xmx512m"
```

### Build

```zsh
mvn clean install
mvn clean install -DskipTests
mvn clean package -DskipTests
```

### Test

```zsh
mvn test
mvn test -Dtest=UserServiceTest
mvn test -Dtest=UserServiceTest#shouldReturnUserById
mvn verify
```

### Dependencies

```zsh
mvn dependency:tree
mvn dependency:tree -Dincludes=org.postgresql:postgresql
mvn versions:display-dependency-updates
mvn versions:display-plugin-updates
mvn dependency:sources
```

---

## Suggested zsh Aliases

```zsh
alias sbr='mvn spring-boot:run'
alias sbrd='mvn spring-boot:run -Dspring-boot.run.profiles=dev'
alias mci='mvn clean install'
alias mcis='mvn clean install -DskipTests'
alias mcp='mvn clean package -DskipTests'
alias mt='mvn test'
alias mdep='mvn dependency:tree'
alias mvu='mvn versions:display-dependency-updates'
```

---

## Common Pitfalls

- **Java version mismatch**: `--java-version=21` sets `pom.xml`, but `JAVA_HOME` must also point to a matching JDK.
- **Dependency name typos**: Use `spring init --list` for exact names. It's `data-jpa` not `jpa`, `postgresql` not `postgres`.
- **Wrong directory**: Always run Maven commands from the directory containing `pom.xml`.
- **Silent overwrites**: `spring init` in a non-empty directory overwrites without warning — always target a fresh directory.

---

## References

- [Spring Initializr](https://start.spring.io)
- [Spring CLI Reference](https://docs.spring.io/spring-boot/docs/current/reference/html/cli.html)
- [Spring Boot Maven Plugin](https://docs.spring.io/spring-boot/docs/current/maven-plugin/reference/htmlsingle/)
