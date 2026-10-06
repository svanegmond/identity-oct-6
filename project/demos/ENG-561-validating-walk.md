---
description: "Steward Validating walk at the no-enrich SHA: DAO, JWT REST search/retrieve, connector as library plus 404 on public enrich."
date: 2026-10-06
ref: ENG-561
---

# Validating walk: Identity scope (post-enrich withdrawal)

**Task:** ENG-561  
**Product SHA:** `b817112a4b16106f92b76baae3aef1f2fb279dbf`  
**Worktree:** `_worktrees/ENG-561`  
**Related:** [engineering handoff](ENG-561-engineering-handoff.md) · [Proof Parade](ENG-561-identity-go-service.md) · [brief](../briefs/identity-go-service.md)

Steward Validating artifact. Not the Implementor Parade. Traffic below was captured live on 2026-10-06 against SHA `b817112` after Lead withdrew AC-10 / VC-2 / demo U6.

| Process | Bind |
|---------|------|
| Identity API ([`cmd/server`](../../cmd/server/main.go), sqlite, `-seed`) | `localhost:8098` |
| Fake vendor IdP ([`cmd/fake-idp`](../../cmd/fake-idp/main.go)) — **optional helper, not started by [`make demo`](../../Makefile)** | `localhost:8099` |
| SQLite file | `/tmp/eng561-scope.db` |

Replay the packaged walk: `cd _worktrees/ENG-561 && make demo` ([`Makefile`](../../Makefile), [`scripts/demo.sh`](../../scripts/demo.sh); needs `sqlite3` on PATH). Replay this exact bind: [`server`](../../cmd/server/main.go) `-db sqlite -dsn /tmp/eng561-scope.db -port 8098 -seed`. Fake IdP is only for the vendor-wire exhibit in §3; Identity REST does not call it.

---

## 1. Database DAO

**Original:** store and retrieve `user_profile` (name, address, phone) and `user_credential` (username, method, password) against PostgreSQL or SQLite behind one interface.

**Landed:** port `store.DAO` in [`internal/store/dao.go`](../../internal/store/dao.go). `store.Open(ctx, DBConfig{Driver, DSN})` selects **`sqlite`** or **`postgres`**. Callers never import a dialect. Goose + sqlc per backend. Operations: `CreateProfile`, `GetProfileByID`, `SearchProfiles`, `CreateCredential`, `GetCredentialByUsername`.

**Gap vs older wording:** Cockroach is not a third adapter. Brief locked PG + SQLite. PG is proven by [`internal/store/postgres_test.go`](../../internal/store/postgres_test.go) (Testcontainers), not by this sqlite live bind.

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

Local profile `address` is one string. Vendor PII `address` (structured object) exists only on connector types — not on this REST surface.

---

## 2. REST API + authentication

**Original:** RESTful search/retrieve of profile data, plus API authentication.

**Landed:** OpenAPI [`project/openapi.yaml`](../openapi.yaml). Login issues HS256 JWT (`POST /auth/login`). Protected routes use `Authorization: Bearer`. Validator: `nethttp-middleware` → `auth.Service.AuthenticateBearer` ([`internal/api/handler.go`](../../internal/api/handler.go)). Search: `GET /profiles?name=` / `phone=`. Retrieve: `GET /profiles/{id}`. `api.NewRouter(dao, authSvc)` — no IdP client.

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

{"expires_in":86400,"token":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiMTExMTExMTEtMTExMS0xMTExLTExMTEtMTExMTExMTExMTExIiwidXNlcm5hbWUiOiJhbGljZSIsImlzcyI6ImlkZW50aXR5LWdvLXNlcnZpY2UiLCJzdWIiOiIxMTExMTExMS0xMTExLTExMTEtMTExMS0xMTExMTExMTExMTEiLCJleHAiOjE3OTEzNjA5MTksImlhdCI6MTc5MTI3NDUxOX0.w6Z35dcq6kSIaFUAd37O7M-tpvH3bLoQFK_l8GFyWrg","token_type":"Bearer"}
```

### Retrieve

```http
GET /profiles/11111111-1111-1111-1111-111111111111 HTTP/1.1
Host: localhost:8098
Authorization: Bearer <token from login>
```

```http
HTTP/1.1 200 OK
Content-Type: application/json
Content-Length: 227

{"address":"123 Market St, San Francisco, CA 94105","created_at":"2026-10-06T08:15:18.808589Z","id":"11111111-1111-1111-1111-111111111111","name":"Alice Smith","phone":"+15551234567","updated_at":"2026-10-06T08:15:18.808589Z"}
```

### Search

```http
GET /profiles?name=Smith HTTP/1.1
Host: localhost:8098
Authorization: Bearer <token from login>
```

```http
HTTP/1.1 200 OK
Content-Type: application/json
Content-Length: 454

[{"address":"123 Market St, San Francisco, CA 94105","created_at":"2026-10-06T08:15:18.808589Z","id":"11111111-1111-1111-1111-111111111111","name":"Alice Smith","phone":"+15551234567","updated_at":"2026-10-06T08:15:18.808589Z"},{"address":"456 Castro St, Mountain View, CA 94041","created_at":"2026-10-06T08:15:18.808589Z","id":"22222222-2222-2222-2222-222222222222","name":"Bob Smith","phone":"+15559876543","updated_at":"2026-10-06T08:15:18.808589Z"}]
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

Invalid Bearer yields the same 401 body.

### Public enrich is gone

```http
POST /profiles/enrich HTTP/1.1
Host: localhost:8098
Authorization: Bearer <token from login>
Content-Type: application/json

{"phone":"+15552345678","name":"Robert Taylor"}
```

```http
HTTP/1.1 404 Not Found
Content-Type: application/json
Content-Length: 53

{"error":"not_found","message":"method not allowed"}
```

Observed: no enrich handler. Body wording is the OpenAPI catch-all, not a dedicated enrich error.

---

## 3. Service connector to ABC / XYC (library, not public REST)

**Original:** connector to third-party IdPs with:

- `POST /auth` body `{"username":"<string>","password":"<string>"}` → access token
- `POST /identity` body `{"phone":"<string>","name":"<string>"}` → PII name, phone, address `{street_address, locality, region, postal_code, country}`

**Landed:** `idp.Connector` in [`internal/idp/connector.go`](../../internal/idp/connector.go). ABC vs XYC differ by `ProviderConfig.BaseURL`. Proof of the Identity service is TP-5 httptest ([`internal/idp/connector_test.go`](../../internal/idp/connector_test.go)). [`cmd/fake-idp`](../../cmd/fake-idp/main.go) still implements the two vendor routes as a helper. Identity [`cmd/server`](../../cmd/server/main.go) does **not** take an IdP URL and does **not** compose those hops onto `/profiles/*`.

The captures below are against **fake-idp :8099**, labeled simulator. They are not Identity REST.

### Vendor `/auth` (simulator)

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

### Vendor `/identity` (simulator)

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

---

## Request flow

```mermaid
sequenceDiagram
  participant C as Caller
  participant API as Identity REST :8098
  participant DAO as store.DAO sqlite
  C->>API: POST /auth/login
  API->>DAO: GetCredentialByUsername
  API-->>C: JWT
  C->>API: GET /profiles… Bearer
  API->>DAO: GetProfileByID / SearchProfiles
  DAO-->>C: profile JSON one-string address
  C->>API: POST /profiles/enrich Bearer
  API-->>C: 404 not_found
```

---

## Limits

- Interview mock: stored passwords as-is; default JWT secret is a flag string.
- Cockroach not implemented.
- [`make demo`](../../Makefile) U1 uses the `sqlite3` CLI.
- Register exists on OpenAPI ([`project/openapi.yaml`](../openapi.yaml)); this walk bootstraps via `-seed`.
- Fake vendor only; no LoginID cloud/SDK/passkeys.
- Fake IdP is not part of [`make demo`](../../Makefile). Connector fidelity for the Identity module is httptest ([`internal/idp/connector_test.go`](../../internal/idp/connector_test.go)), not this simulator bind.
