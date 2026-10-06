---
description: "Steward Validating walk: original three scope items with live HTTP, plus engineer handoff of enrich/auth/DAO."
date: 2026-10-06
ref: ENG-561
---

# Validating walk: original Identity scope

**Superseded for public REST:** Lead 2026-10-06 withdrew `POST /profiles/enrich` / composed caller path (AC-10, VC-2, TP-6, U6). Current contract: [brief](../briefs/identity-go-service.md). Current product surface: [engineering handoff](ENG-561-engineering-handoff.md). The traffic below is a historical capture at SHA `45ad2eb` when enrich still existed.

**Task:** ENG-561  
**Historical product SHA:** `45ad2eb60214505467a3d778b74764cb3c8817bc`  
**This note committed after that SHA.**  
**Worktree:** `_worktrees/ENG-561`  
**Related:** [engineering handoff](ENG-561-engineering-handoff.md) · [Proof Parade](ENG-561-identity-go-service.md) · [brief](../briefs/identity-go-service.md)

Steward Validating artifact. Not the Implementor Parade. Traffic below was captured live on 2026-10-06 against:

| Process | Bind |
|---------|------|
| Identity API (`cmd/server`, sqlite, `-seed`) | `localhost:8098` |
| Fake vendor IdP (`cmd/fake-idp`) | `localhost:8099` |
| SQLite file | `/tmp/eng561-scope.db` |

Replay the packaged walk: `cd _worktrees/ENG-561 && make demo` (needs `sqlite3` on PATH). Replay this exact bind: start `fake-idp -port 8099` and `server -db sqlite -port 8098 -idp-url http://localhost:8099 -seed`.

---

## What “enrich” means

Senior-eng, no IdP jargon: **enrich** is “ask the vendor directory for a person, return that payload.” It is **not** merge, update, or save a local profile.

- Local **retrieve/search** reads *our* tables (`user_profiles`).
- **Enrich** is an outbound hop to ABC/XYC-shaped `/auth` then `/identity`, then the JSON comes back on `POST /profiles/enrich`.
- Address shapes differ: local profile `address` is one string; vendor PII `address` is `{street_address, locality, region, postal_code, country}`.
- Current code does not upsert vendor PII into the DAO (`EnrichProfile` only calls `idpConn.FetchIdentity`).

Analog: `GET /users/123` vs a Salesforce lookup you do not write into `users`.

---

## 1. Database DAO

**Original:** store and retrieve `user_profile` (name, address, phone) and `user_credential` (username, method, password) against multiple databases (PostgreSQL or Cockroach, SQLite, etc.).

**Landed:** one port `store.DAO` in `internal/store/dao.go`. `store.Open(ctx, DBConfig{Driver, DSN})` selects **`sqlite`** or **`postgres`**. Callers never import a dialect. Goose + sqlc per backend. Operations: `CreateProfile`, `GetProfileByID`, `SearchProfiles`, `CreateCredential`, `GetCredentialByUsername`.

**Gap vs original wording:** Cockroach is **not** a third adapter. Brief locked PG + SQLite. Cockroach would be Postgres protocol, unverified here. PG is proven by `internal/store/postgres_test.go` (Testcontainers), not by this sqlite live bind.

**Live sqlite schema and rows** (after `-seed` on `/tmp/eng561-scope.db`):

```sql
CREATE TABLE user_profiles (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    address TEXT NOT NULL,
    phone TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE user_credentials (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    username TEXT NOT NULL UNIQUE,
    method TEXT NOT NULL DEFAULT 'password',
    password TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

```text
SELECT id, name, address, phone FROM user_profiles;
11111111-1111-1111-1111-111111111111|Alice Smith|123 Market St, San Francisco, CA 94105|+15551234567
22222222-2222-2222-2222-222222222222|Bob Smith|456 Castro St, Mountain View, CA 94041|+15559876543

SELECT username, method, user_id FROM user_credentials;
alice|password|11111111-1111-1111-1111-111111111111
```

---

## 2. REST API + authentication

**Original:** RESTful search/retrieve of profile data, plus API authentication.

**Landed:** OpenAPI `project/openapi.yaml`. Login issues HS256 JWT (`POST /auth/login`). Protected routes use `Authorization: Bearer`. Validator: `nethttp-middleware` → `auth.Service.AuthenticateBearer` (`internal/api/handler.go`). Search: `GET /profiles?name=` / `phone=`. Retrieve: `GET /profiles/{id}`.

JWT signing default is interview-mock (`-jwt-secret`, default `dev-jwt-secret-interview-mock-long-enough`). Passwords compared as stored.

### Login

```http
POST /auth/login HTTP/1.1
Host: localhost:8098
Content-Type: application/json

{"username":"alice","password":"password123"}
```

```http
HTTP/1.1 200 OK
Content-Type: application/json
Content-Length: 370

{"expires_in":86400,"token":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiMTExMTExMTEtMTExMS0xMTExLTExMTEtMTExMTExMTExMTExIiwidXNlcm5hbWUiOiJhbGljZSIsImlzcyI6ImlkZW50aXR5LWdvLXNlcnZpY2UiLCJzdWIiOiIxMTExMTExMS0xMTExLTExMTEtMTExMS0xMTExMTExMTExMTEiLCJleHAiOjE3OTEzNTkwNzgsImlhdCI6MTc5MTI3MjY3OH0.CAkTPn2zTwrGufaRClFHfEUAzoschhg9ClEuWDrLnr4","token_type":"Bearer"}
```

### Retrieve

```http
GET /profiles/11111111-1111-1111-1111-111111111111 HTTP/1.1
Host: localhost:8098
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiMTExMTExMTEtMTExMS0xMTExLTExMTEtMTExMTExMTExMTExIiwidXNlcm5hbWUiOiJhbGljZSIsImlzcyI6ImlkZW50aXR5LWdvLXNlcnZpY2UiLCJzdWIiOiIxMTExMTExMS0xMTExLTExMTEtMTExMS0xMTExMTExMTExMTEiLCJleHAiOjE3OTEzNTkwNzgsImlhdCI6MTc5MTI3MjY3OH0.CAkTPn2zTwrGufaRClFHfEUAzoschhg9ClEuWDrLnr4
```

```http
HTTP/1.1 200 OK
Content-Type: application/json
Content-Length: 227

{"address":"123 Market St, San Francisco, CA 94105","created_at":"2026-10-06T07:44:38.660288Z","id":"11111111-1111-1111-1111-111111111111","name":"Alice Smith","phone":"+15551234567","updated_at":"2026-10-06T07:44:38.660288Z"}
```

### Search

```http
GET /profiles?name=Smith HTTP/1.1
Host: localhost:8098
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiMTExMTExMTEtMTExMS0xMTExLTExMTEtMTExMTExMTExMTExIiwidXNlcm5hbWUiOiJhbGljZSIsImlzcyI6ImlkZW50aXR5LWdvLXNlcnZpY2UiLCJzdWIiOiIxMTExMTExMS0xMTExLTExMTEtMTExMS0xMTExMTExMTExMTEiLCJleHAiOjE3OTEzNTkwNzgsImlhdCI6MTc5MTI3MjY3OH0.CAkTPn2zTwrGufaRClFHfEUAzoschhg9ClEuWDrLnr4
```

```http
HTTP/1.1 200 OK
Content-Type: application/json
Content-Length: 454

[{"address":"123 Market St, San Francisco, CA 94105","created_at":"2026-10-06T07:44:38.660288Z","id":"11111111-1111-1111-1111-111111111111","name":"Alice Smith","phone":"+15551234567","updated_at":"2026-10-06T07:44:38.660288Z"},{"address":"456 Castro St, Mountain View, CA 94041","created_at":"2026-10-06T07:44:38.660288Z","id":"22222222-2222-2222-2222-222222222222","name":"Bob Smith","phone":"+15559876543","updated_at":"2026-10-06T07:44:38.660288Z"}]
```

### Auth gate (no Bearer)

```http
GET /profiles/11111111-1111-1111-1111-111111111111 HTTP/1.1
Host: localhost:8098
```

```http
HTTP/1.1 401 Unauthorized
Content-Type: application/json
Content-Length: 83

{"error":"unauthorized","message":"Unauthorized: missing or invalid bearer token"}
```

---

## 3. Service connector to ABC / XYC

**Original:** connector to third-party IdPs with:

- `POST /auth` body `{"username":"<string>","password":"<string>"}` → access token
- `POST /identity` body `{"phone":"<string>","name":"<string>"}` → PII name, phone, address `{street_address, locality, region, postal_code, country}`

**Landed:** `idp.Connector` in `internal/idp/connector.go`. ABC vs XYC differ by `ProviderConfig.BaseURL` (and credentials), one client. Fake process `cmd/fake-idp` implements those two vendor routes. Our REST composes them on `POST /profiles/enrich` (JWT in; vendor hop; PII out; no DAO write).

### Vendor `/auth`

```http
POST /auth HTTP/1.1
Host: localhost:8099
Content-Type: application/json

{"username":"vendor_user","password":"vendor_secret"}
```

```http
HTTP/1.1 200 OK
Content-Type: application/json
Content-Length: 91

{"access_token":"vendor-jwt-token-fake-abc-12345","expires_in":3600,"token_type":"Bearer"}
```

### Vendor `/identity`

```http
POST /identity HTTP/1.1
Host: localhost:8099
Authorization: Bearer vendor-jwt-token-fake-abc-12345
Content-Type: application/json

{"phone":"+15552345678","name":"Robert Taylor"}
```

```http
HTTP/1.1 200 OK
Content-Type: application/json
Content-Length: 187

{"address":{"country":"USA","locality":"San Francisco","postal_code":"94103","region":"CA","street_address":"789 Market Street, Suite 400"},"name":"Robert Taylor","phone":"+15552345678"}
```

### Composed path (our API; connector does `/auth` then `/identity` internally)

```http
POST /profiles/enrich HTTP/1.1
Host: localhost:8098
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiMTExMTExMTEtMTExMS0xMTExLTExMTEtMTExMTExMTExMTExIiwidXNlcm5hbWUiOiJhbGljZSIsImlzcyI6ImlkZW50aXR5LWdvLXNlcnZpY2UiLCJzdWIiOiIxMTExMTExMS0xMTExLTExMTEtMTExMS0xMTExMTExMTExMTEiLCJleHAiOjE3OTEzNTkwNzgsImlhdCI6MTc5MTI3MjY3OH0.CAkTPn2zTwrGufaRClFHfEUAzoschhg9ClEuWDrLnr4
Content-Type: application/json

{"phone":"+15552345678","name":"Robert Taylor"}
```

```http
HTTP/1.1 200 OK
Content-Type: application/json
Content-Length: 187

{"address":{"country":"USA","locality":"San Francisco","postal_code":"94103","region":"CA","street_address":"789 Market Street, Suite 400"},"name":"Robert Taylor","phone":"+15552345678"}
```

---

## Request flow

```mermaid
sequenceDiagram
  participant C as Caller
  participant API as Identity REST :8098
  participant DAO as store.DAO sqlite
  participant IdP as idp.Client
  participant V as fake-idp :8099
  C->>API: POST /auth/login
  API->>DAO: GetCredentialByUsername
  API-->>C: JWT
  C->>API: GET /profiles… Bearer
  API->>DAO: GetProfileByID / SearchProfiles
  DAO-->>C: profile JSON one-string address
  C->>API: POST /profiles/enrich Bearer
  API->>IdP: FetchIdentity
  IdP->>V: POST /auth
  V-->>IdP: access_token
  IdP->>V: POST /identity
  V-->>C: structured PII, no DAO write
```

---

## Limits

- Interview mock: stored passwords as-is; default JWT secret is a flag string.
- Cockroach not implemented.
- `make demo` U1 uses the `sqlite3` CLI.
- `AuthenticateRequest` wraps `AuthenticateBearer` and has no production caller.
- Register exists on OpenAPI; this walk bootstraps via `-seed`.
- Fake vendor only; no LoginID cloud/SDK/passkeys.
