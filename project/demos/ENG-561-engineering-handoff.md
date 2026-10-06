---
description: "Steward Validating engineering handoff for ENG-561 at 45ad2eb."
date: 2026-10-06
ref: ENG-561
---

# Engineering handoff: Identity Go service

**SHA:** `45ad2eb60214505467a3d778b74764cb3c8817bc`  
**Worktree:** `_worktrees/ENG-561` (`feat/ENG-561`)  
**Parade (captures):** [ENG-561-identity-go-service.md](ENG-561-identity-go-service.md)  
**Replay:** from worktree root, `make demo` (needs `sqlite3` on PATH for U1)

Steward reading of landed code plus a live `make demo` at this SHA (not a paste of the Implementor Parade). HTTP bodies below are from that drive (`/tmp/eng561-validating-c3-demo.txt`).

## What this is

A local Identity service for interview discussion. Callers log in against stored credentials, get an HS256 JWT, then search/retrieve profiles. **Enrich** means: POST name+phone, this service asks a fake vendor IdP, returns that PII. It does not write the vendor payload into SQLite/Postgres.

Not LoginID production: plaintext-or-as-stored passwords, JWT secret is a flag default, no passkeys/SDK.

## Layout

| Package | Role |
|---------|------|
| `internal/store` | DAO port + SQLite/Postgres adapters (goose + sqlc) |
| `internal/auth` | Credential check, JWT issue/verify, `AuthenticateBearer` |
| `internal/idp` | HTTP client for vendor `/auth` + `/identity` |
| `internal/api` | OpenAPI-generated handlers + `nethttp-middleware` validator |
| `cmd/server` | Flags, `store.Open`, seed, listen |
| `cmd/fake-idp` | Simulator used by `make demo` |

`store.Open` switches on `DBConfig.Driver`. Callers never import a dialect.

## Data model

- `user_profiles`: `id`, `name`, **one-string** `address`, `phone`, timestamps
- `user_credentials`: `id`, `user_id`, unique `username`, `method`, `password`, timestamps

IdP PII `address` is structured (`street_address`, `locality`, `region`, `postal_code`, `country`). Enrich keeps that shape in the HTTP response only.

Seed: Alice `11111111-…` / `alice` / `password123`; Bob `22222222-…` for search.

## HTTP surface

| Method | Path | Auth | Meaning |
|--------|------|------|---------|
| POST | `/auth/login` | none | credential → JWT |
| POST | `/auth/register` | none | create profile + credential |
| GET | `/profiles` | Bearer | search `name` / `phone` |
| GET | `/profiles/{id}` | Bearer | retrieve |
| POST | `/profiles/enrich` | Bearer | vendor lookup; no DAO write |

Live gate: OpenAPI `BearerAuth` → `AuthenticationFunc` → `auth.Service.AuthenticateBearer` (`internal/api/handler.go`). `Middleware` is gone.

## Observed traffic (this SHA)

U1 is not HTTP. Direct SQLite:

```text
SQLite profile row: 11111111-1111-1111-1111-111111111111|Alice Smith|+15551234567
SQLite credential row: 11111111-1111-1111-1111-111111111111|alice|password
```

Login:

```http
POST /auth/login HTTP/1.1
Host: localhost:8088
Content-Type: application/json

{"username":"alice","password":"password123"}
```

```http
HTTP/1.1 200 OK
Content-Type: application/json

{"expires_in":86400,"token":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiMTExMTExMTEtMTExMS0xMTExLTExMTEtMTExMTExMTExMTExIiwidXNlcm5hbWUiOiJhbGljZSIsImlzcyI6ImlkZW50aXR5LWdvLXNlcnZpY2UiLCJzdWIiOiIxMTExMTExMS0xMTExLTExMTEtMTExMS0xMTExMTExMTExMTEiLCJleHAiOjE3OTEzNTg5MTMsImlhdCI6MTc5MTI3MjUxM30.zkSaS7YQ8SwZzWCdA3jJB871n3I4janBDqzuFpwlWss","token_type":"Bearer"}
```

Retrieve:

```http
GET /profiles/11111111-1111-1111-1111-111111111111 HTTP/1.1
Host: localhost:8088
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiMTExMTExMTEtMTExMS0xMTExLTExMTEtMTExMTExMTExMTExIiwidXNlcm5hbWUiOiJhbGljZSIsImlzcyI6ImlkZW50aXR5LWdvLXNlcnZpY2UiLCJzdWIiOiIxMTExMTExMS0xMTExLTExMTEtMTExMS0xMTExMTExMTExMTEiLCJleHAiOjE3OTEzNTg5MTMsImlhdCI6MTc5MTI3MjUxM30.zkSaS7YQ8SwZzWCdA3jJB871n3I4janBDqzuFpwlWss
```

```http
HTTP/1.1 200 OK
Content-Type: application/json

{"address":"123 Market St, San Francisco, CA 94105","created_at":"2026-10-06T07:41:53.628299Z","id":"11111111-1111-1111-1111-111111111111","name":"Alice Smith","phone":"+15551234567","updated_at":"2026-10-06T07:41:53.628299Z"}
```

Search `GET /profiles?name=Smith` with the same Bearer returns Alice and Bob (full array in Parade / demo transcript).

Auth gate, no header / `Bearer bad-invalid-token`:

```http
HTTP/1.1 401 Unauthorized
Content-Type: application/json

{"error":"unauthorized","message":"Unauthorized: missing or invalid bearer token"}
```

No jwt library parse text in that body at this SHA.

Enrich (vendor lookup, not a local upsert):

```http
POST /profiles/enrich HTTP/1.1
Host: localhost:8088
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiMTExMTExMTEtMTExMS0xMTExLTExMTEtMTExMTExMTExMTExIiwidXNlcm5hbWUiOiJhbGljZSIsImlzcyI6ImlkZW50aXR5LWdvLXNlcnZpY2UiLCJzdWIiOiIxMTExMTExMS0xMTExLTExMTEtMTExMS0xMTExMTExMTExMTEiLCJleHAiOjE3OTEzNTg5MTMsImlhdCI6MTc5MTI3MjUxM30.zkSaS7YQ8SwZzWCdA3jJB871n3I4janBDqzuFpwlWss
Content-Type: application/json

{"name":"Robert Taylor","phone":"+15552345678"}
```

```http
HTTP/1.1 200 OK
Content-Type: application/json

{"address":{"country":"USA","locality":"San Francisco","postal_code":"94103","region":"CA","street_address":"789 Market Street, Suite 400"},"name":"Robert Taylor","phone":"+15552345678"}
```

## Request flow

```mermaid
sequenceDiagram
  participant C as Caller
  participant API as api.NewRouter
  participant Auth as AuthenticateBearer
  participant DAO as store.DAO
  participant IdP as idp.Client
  C->>API: POST /auth/login
  API->>DAO: GetCredentialByUsername
  API-->>C: HS256 JWT
  C->>API: GET /profiles + Bearer
  API->>Auth: AuthenticateBearer
  API->>DAO: search/retrieve
  C->>API: POST /profiles/enrich
  API->>IdP: FetchIdentity
  IdP-->>C: structured PII, no DAO write
```

## Limits

- JWT secret default `dev-jwt-secret-interview-mock-long-enough`. Passwords compared as stored.
- `make demo` U1 shells out to `sqlite3`; host without it fails closed (`command not found`), not a silent green.
- `AuthenticateRequest(*http.Request)` wraps `AuthenticateBearer` and has no production caller (leftover helper).
- Register exists; VC-1 bootstrap is seed + SQLite readback.

## Replay

```bash
cd _worktrees/ENG-561
make demo
go test ./...
```

Instrument: `/tmp/helm-ir-battery/ENG-561/45ad2eb60214505467a3d778b74764cb3c8817bc.json`
