---
title: "Open Questions — Unresolved Topics and Open Threads"
slug: open-questions-unresolved-topics-and-open-threads
type: concept
tags: [open-threads, todo, research, issue]
depth: 4
confidence: low
created: 2026-04-17
updated: 2026-04-17
verified: 2026-04-17
freshness_days: 365
sources:
  - url: sources/daily/2026-04-13-local-ai-continue-rag-spring.md
    accessed: 2026-04-17
    kind: session
related: []
supersedes: []
forge_version: 2.0.0
origin: import
---

# Open Questions — Unresolved Topics and Open Threads

Compiled from all "Open Threads" sections across all ingested daily logs.

---

## Continue.dev Sends No Requests to llama.cpp

**Status:** Unresolved  
**Source:** [[sources/daily/2026-04-13-local-ai-continue-rag-spring]]  

Even with valid `config.json`, correct `apiBase`, and reachable llama.cpp, Continue.dev in
IntelliJ sends zero requests. Config loads without error. Root cause unknown.

**Possible causes:**
- Hub/auth UI intercepting requests
- Plugin-level initialization issue specific to IntelliJ (not VS Code)

**Investigation needed:**
- Check Continue plugin logs via IntelliJ Help → Show Log in Finder
- Test with VS Code Continue to isolate IntelliJ-specific vs general issue
- Check if Continue hub authentication intercepts local model requests

---

## Continue.dev IntelliJ Dropdown Shows Only Cloud Models

**Status:** Unresolved  
**Source:** [[sources/daily/2026-04-13-local-ai-continue-rag-spring]]  

The "Qwen3 4B (local)" label appears in the top-left corner but clicking the dropdown shows
only Claude and Gemini cloud models. The custom local model is not selectable.

**Possible cause:** Hub-based model selector in 0.9.264; models from `config.json` may not
surface in the dropdown UI even when config is valid.

---

## Spring Boot MeterReadingsService — Receipt Feature

**Status:** Not started  
**Source:** [[sources/daily/2026-04-15-spring-refactor-testcontainers]]  

Receipt feature requested but not yet implemented:
- OpenAPI spec additions
- Liquibase migration
- Receipt module (6 files)

---

## Keycloak Google Social Login — Production Credentials

**Status:** Partial (IDP configured, credentials not set)  
**Source:** [[sources/daily/2026-04-14-spring-keycloak-postman]]  

Need from Google Cloud Console:
- OAuth 2.0 Client ID + Secret
- Redirect URI: `http://localhost:9090/realms/meter-readings-realm/broker/google/endpoint`
- Set in Keycloak Admin → meter-readings-realm → Identity Providers → Google

---

## MapStruct Unmapped Target Properties Warning

**Status:** Decision pending  
**Source:** [[sources/daily/2026-04-14-spring-keycloak-postman]]  

`MeterMapper` warns about unmapped target properties: `id`, `userId`, `deletedAt`, `createdAt`.
Decision: use `@BeanMapping(ignoreByDefault=true)` to suppress, or explicitly map all fields.

---

## Frontend: WCAG Gaps in US1.7 (ForgotPassword)

**Status:** Not implemented  
**Source:** [[sources/daily/2026-04-17-storybook-llm-wiki]]  

Three WCAG accessibility gaps identified:
- `aria-live` debounced password strength indicator
- Per-step focus management
- `aria-disabled` on cooldown resend button

---

## Frontend: 11 Missing Storybook Stories

**Status:** Not started  
**Source:** [[sources/daily/2026-04-17-storybook-llm-wiki]]  

11 components listed in the status doc still need Storybook stories.

---

## Frontend: CheckEmail Success Toast on Resend

**Status:** Not implemented  
**Source:** [[sources/daily/2026-04-17-storybook-llm-wiki]]  

Missing success toast in `CheckEmail` resend path — one-liner + i18n key.

---

## llm-wiki Skill — End-to-End Validation

**Status:** In progress  
**Source:** [[sources/daily/2026-04-17-storybook-llm-wiki]]  

Drop first source into `raw/` and run ingest to validate the skill end-to-end.

## Sources
- [[sources/daily/2026-04-13-local-ai-continue-rag-spring]]
- [[sources/daily/2026-04-14-spring-keycloak-postman]]
- [[sources/daily/2026-04-15-spring-refactor-testcontainers]]
- [[sources/daily/2026-04-16-frontend-code-review]]
- [[sources/daily/2026-04-17-storybook-llm-wiki]]

## Related
- [[notes/concept/llm-wiki-pattern-llm-maintained-knowledge-archive]]
