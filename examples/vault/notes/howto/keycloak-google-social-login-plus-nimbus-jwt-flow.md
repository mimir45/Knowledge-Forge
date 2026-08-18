---
title: Keycloak Google Social Login + Nimbus JWT Flow
slug: keycloak-google-social-login-plus-nimbus-jwt-flow
type: howto
stack: [keycloak, oauth2, spring-boot, docker]
tags: [jwt, nimbus-jose]
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

# Keycloak Google Social Login + Nimbus JWT Flow

## What is it?

Keycloak acts as an **identity broker**: it delegates login to Google via OAuth 2.0 / OIDC, then issues its **own signed JWT** to your application. The app never sees a Google token — it only validates a Keycloak JWT, so the auth stack is unchanged.

Key mental model: Google authenticates the user, Keycloak is the issuer your service trusts.

---

## How it works

```
Browser           Keycloak              Google        Spring Boot App
  │                  │                    │                 │
  │── GET /login ──> │                    │                 │
  │<─ redirect ──────│ (to Google)        │                 │
  │── Google OAuth ──────────────────────>│                 │
  │<─ auth code ─────────────────────────│                 │
  │── code ─────────>│                    │                 │
  │                  │── exchange code ──>│                 │
  │                  │<─ Google id_token ─│                 │
  │                  │  (provisions KC user)               │
  │<─ Keycloak JWT ──│                    │                 │
  │── Bearer <KC JWT> ────────────────────────────────────>│
  │                  (JwksProvider fetches KC public keys)  │
  │                  (JwtValidator: sig + iss + sub + exp)  │
  │<──────────────────────────────────────────────── 200 ──│
```

### What Keycloak does internally

1. Exchanges Google auth code → Google `id_token` + `access_token`
2. Reads `sub`, `email`, `name` claims to find or **create a local Keycloak user** (federated identity)
3. Issues a fresh JWT signed with the **realm's own RSA key** — this is what `JwksProvider` fetches
4. The `sub` in the Keycloak JWT is the **Keycloak UUID**, not Google's numeric sub

---

## Full from-scratch setup

### Step 1 — Google Cloud Console

1. Go to **APIs & Services → Credentials → Create OAuth 2.0 Client ID**
2. Application type: **Web application**
3. Add Authorized redirect URI:
   ```
   http://localhost:9090/realms/meter-realm/broker/google/endpoint
   ```
   Pattern: `/realms/<realm>/broker/<idp-alias>/endpoint`
4. Copy the **Client ID** and **Client Secret**

---

### Step 2 — Fill in the realm JSON

Edit `keycloak/realm-export.json` and replace the two placeholders:

```json
"config": {
  "clientId": "REPLACE_WITH_GOOGLE_CLIENT_ID",
  "clientSecret": "REPLACE_WITH_GOOGLE_CLIENT_SECRET"
}
```

The realm JSON (`keycloak/realm-export.json`) pre-configures:

| Resource | Details |
|---|---|
| Realm | `meter-realm`, SSL disabled (dev) |
| Client `postman` | Public, Standard flow, redirect: `http://localhost/*` + Postman callback |
| Client `meter-readings-service` | Confidential, service account only |
| Identity Provider | Google, `trustEmail: true`, `syncMode: IMPORT` |

---

### Step 3 — docker-compose setup

```yaml
keycloak:
  image: quay.io/keycloak/keycloak:25.0.5
  volumes:
    - keycloak_data:/opt/keycloak/data
    - ./keycloak:/opt/keycloak/data/import   # realm JSON lives here
  command:
    - 'start-dev'
    - '--import-realm'                        # imports any realm not yet in DB
```

**Import behaviour:**
- Keycloak imports realm JSON files from `/opt/keycloak/data/import/` on startup
- Skips import if the realm already exists in the database (safe to restart)
- `keycloak_data` volume persists the H2 database between restarts

**If Keycloak was already started before the realm JSON was in place:**
```bash
docker-compose down -v   # deletes keycloak_data volume — wipes existing data
docker-compose up -d keycloak
```

---

### Step 4 — Start and verify

```bash
docker-compose -f docker-compose-local.yaml up -d keycloak
```

Check the logs:
```bash
docker logs keycloak_demo | grep -E "import|meter-realm|ERROR"
```

You should see:
```
Imported realm meter-realm from file ...
```

Then visit `http://localhost:9090/admin` → login with `admin / admin1234` → confirm `meter-realm` exists.

---

### Step 5 — Disable "Review Profile" on first Google login

By default Keycloak asks the user to review their profile on first Google login. To skip this:

1. Admin Console → `meter-realm` → **Authentication**
2. Click **First Broker Login** flow
3. Find **Review Profile** → click the action → set to **Disabled**
4. Save

---

### Step 6 — Spring Boot app: zero changes

`JwtValidator` and `JwksProvider` are untouched. The issuer and JWKS URI in `application.yaml` already point at the right realm:

```yaml
keycloak:
  issuer: ${KEYCLOAK_ISSUER:http://localhost:9090/realms/meter-realm}
  jwks-uri: ${keycloak.issuer}/protocol/openid-connect/certs
```

---

### Step 7 — Get a JWT in Postman

1. Open the collection → **Authorization** tab → Type: **OAuth 2.0**
2. Click **Get New Access Token** and fill in:
   ```
   Auth URL:    http://localhost:9090/realms/meter-realm/protocol/openid-connect/auth
   Token URL:   http://localhost:9090/realms/meter-realm/protocol/openid-connect/token
   Client ID:   postman
   Scope:       openid
   ```
3. Click **Request Token** → browser opens Keycloak login page → **Sign in with Google**
4. Copy the returned `access_token` → paste into the `bearer_token` environment variable

---

## Common pitfalls

- **`sub` is the Keycloak UUID, not Google's sub** — `claims.getSubject()` returns Keycloak's internal UUID. This is what your `findByIdAndUserId()` queries use — consistent across all login methods.
- **Import skipped silently if realm already exists** — if you change the realm JSON, you must `docker-compose down -v` to force re-import.
- **JWKS cache staleness on key rotation** — `JwksProvider` caches keys for 1 hour. On "Unknown key ID" errors, evict the cache and retry once.
- **Redirect URI mismatch** — the URI in Google Cloud Console must exactly match Keycloak's broker endpoint (scheme, host, realm name, no trailing slash).
- **`preferred_username` is the Google email** — Keycloak defaults to `[REDACTED-EMAIL]` as the username. Change via an Identity Provider mapper if needed.
- **Google requires HTTPS for production** — `http://localhost` is allowed for dev only. Add the `https://` production URI before going live.

---

## When to use / not use

**Use when:**
- You want multiple login methods (password + Google + GitHub) without changing your service's token validation.
- Multiple services share the same Keycloak realm.
- You need centralized session management and RBAC.

**Don't use when:**
- Pure SPA — direct Google OIDC via Auth.js or Firebase Auth is simpler.
- Minimal operational overhead — Auth0, Supabase Auth, or Clerk are lighter than self-hosted Keycloak.

---

## References

- [Keycloak docs — Identity Brokering](https://www.keycloak.org/docs/latest/server_admin/#_identity_broker)
- [Keycloak docs — Social Identity Providers](https://www.keycloak.org/docs/latest/server_admin/#social-identity-providers)
- [Google Cloud Console — OAuth 2.0 Credentials](https://console.cloud.google.com/apis/credentials)
- [RFC 7519 — JWT specification](https://datatracker.ietf.org/doc/html/rfc7519)