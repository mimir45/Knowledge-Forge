---
title: 'TIL: Demoing Keycloak + Google Login in a Presentation'
slug: til-demoing-keycloak-plus-google-login-in-a-presentation
type: howto
stack: [keycloak]
tags: [auth, presentation, google-idp, jwt]
depth: 3
confidence: low
created: 2026-04-14
updated: 2026-08-09
verified: 2026-08-09
freshness_days: 180
sources: []
related: []
supersedes: []
forge_version: 2.0.0
origin: import
---

# TIL: Demoing Keycloak + Google Login in a Presentation

How to show the complete auth flow live: user logs in with Google → Keycloak issues a JWT → API call succeeds.

## Setup (from realm-export.json + docker-compose-local.yaml)

| Component        | Value                                |
| ---------------- | ------------------------------------ |
| Keycloak URL     | `http://localhost:9090`              |
| Realm            | `meter-realm`                        |
| Public client    | `postman`                            |
| Redirect URI     | `https://oauth.pstmn.io/v1/callback` |
| Google IdP alias | `google`                             |
| Token lifespan   | 3600 s                               |
| API base         | `http://localhost:8080/api/v1`       |

## The Flow

```
Browser ──► Keycloak (auth endpoint)
         ──► Google login page  ← user signs in with Google account
         ◄── Google callback to Keycloak (authorization code)
         ──► Keycloak token endpoint (code exchange)
         ◄── Keycloak JWT  ← this is what the API validates
         ──► API  Bearer {{keycloak_jwt}}
```

**Key talking point:** Google's token never reaches your API. Keycloak acts as an identity broker and re-issues its own JWT. Your `JwtAuthFilter` only ever sees Keycloak tokens.

## Act 1 — Open the Authorization URL in a Browser

Paste this URL into the browser address bar:

```
http://localhost:9090/realms/meter-realm/protocol/openid-connect/auth?client_id=postman&response_type=code&scope=openid%20email%20profile&redirect_uri=https%3A%2F%2Foauth.pstmn.io%2Fv1%2Fcallback&state=demo123
```

Keycloak shows its login page with a **Sign in with Google** button. Clicking it opens Google's consent screen.

## Act 2 — Exchange the Code for a Token

After Google redirects back, Keycloak appends `?code=...` to the callback URL. Exchange it:

```bash
curl -s -X POST http://localhost:9090/realms/meter-realm/protocol/openid-connect/token \
  -d "grant_type=authorization_code" \
  -d "client_id=postman" \
  -d "code=<code_from_callback>" \
  -d "redirect_uri=https://oauth.pstmn.io/v1/callback" | jq .
```

Response shape:

```json
{
  "access_token": "[REDACTED-KEY]...",
  "token_type": "Bearer",
  "expires_in": 3600,
  "refresh_token": "eyJ..."
}
```

## Act 3 — Inspect the JWT at jwt.io

Paste the `access_token` at [jwt.io](https://jwt.io). The decoded payload contains:

```json
{
  "sub": "f47ac10b-58cc-4372-a567-0e02b2c3d479",
  "preferred_username": "[REDACTED-EMAIL]",
  "email": "[REDACTED-EMAIL]",
  "email_verified": true,
  "identity_provider": "google",
  "iss": "http://localhost:9090/realms/meter-realm",
  "exp": 1713106800
}
```

`JwtAuthFilter` reads `sub` and sets it as `AuthenticatedUser.userId`. The `identity_provider: "google"` claim proves the user authenticated via Google — useful talking point.

## Act 4 — Call the Protected API

```bash
TOKEN="eyJhbGci..."

# Happy path
curl http://localhost:8080/api/v1/meters \
  -H "Authorization: Bearer $TOKEN"

# Without token → 401
curl http://localhost:8080/api/v1/meters
# {"code":"UNAUTHORIZED","message":"Missing or invalid token"}
```

## Postman OAuth 2.0 Tab (Fastest Live Demo Path)

In Postman → collection → Authorization tab → Type: **OAuth 2.0**:

| Field            | Value                                                                    |
| ---------------- | ------------------------------------------------------------------------ |
| Grant type       | Authorization Code                                                       |
| Auth URL         | `http://localhost:9090/realms/meter-realm/protocol/openid-connect/auth`  |
| Access Token URL | `http://localhost:9090/realms/meter-realm/protocol/openid-connect/token` |
| Client ID        | `postman`                                                                |
| Scope            | `openid email profile`                                                   |
| State            | `demo`                                                                   |

Click **Get New Access Token** → browser opens → sign in with Google → token appears automatically. No curl needed for the demo.

## Discovery Endpoint (Good Talking Point)

```bash
curl -s http://localhost:9090/realms/meter-realm/.well-known/openid-configuration \
  | jq '{issuer, authorization_endpoint, token_endpoint, jwks_uri}'
```

Shows the four URLs the application uses. Keycloak exposes this standard endpoint so clients can auto-configure — no hardcoded URLs needed in production.

## Realm Config Notes (from realm-export.json)

- `trustEmail: true` — Keycloak trusts Google-verified emails without a second verification step
- `syncMode: IMPORT` — user attributes are copied from Google on first login only (not re-synced on each login)
- Token lifespan 3600 s — tokens expire after 1 hour; use `refresh_token` to get a new one silently