# Proof Parade: Identity Go service (interview discussion object)

**Brief**: [project/briefs/identity-go-service.md](../briefs/identity-go-service.md)
**Task**: ENG-561
**Date**: 2026-10-06
**Scaffold drafted by**: Steward
**Evidence captured by**: Implementor

## Evidence Index

| Criterion | Planned evidence | Captured evidence | What it proves |
|---|---|---|---|
| AC-1 | What-landed: DAO contract path | `project/schemas.md` | DAO contract exists before code |
| AC-2 | What-landed: REST/OpenAPI + JWT decision cite | `project/openapi.yaml`, `project/decisions/rest-api-jwt-bearer.md` | REST + JWT bearer locked |
| AC-3 | What-landed: IdP integrations contract path | `project/integrations.md` | IdP `/auth` + `/identity` contracted |
| AC-4 | What-landed: separation statements in contracts | `internal/boundary_test.go:TestArchitecturalPackageBoundaries`, `project/schemas.md § 5`, `project/integrations.md § 4` | DAO/HTTP/IdP seams named and enforced |
| AC-5 | What-landed: commit order Phase A before B | Git commit `2c29cae` (Phase A) landed before `8c9dd67` (Phase B) | Sequencing held |
| AC-6 | TP-2, TP-3 / What-landed DAO packages | `internal/store/sqlite_test.go:TestSQLiteDAO_StoreRetrieveSearch`, `internal/store/postgres_test.go:TestPostgresDAO_StoreRetrieveSearch` | Dual-DB DAO landed |
| AC-7 | TP-4 / Exhibit Auth+profile | `internal/auth/jwt_test.go:TestAuthService_TokenIssueAndVerify`, `TestAuthService_CredentialCheckAndLogin`, `TestAuthService_AuthenticateBearer`, `internal/api/api_test.go:TestAPI_OpenAPIBearerAuthEnforcement`, Exhibit: Auth+profile round-trip | JWT issue + verify |
| AC-8 | TP-8 + Exhibit Auth+profile (search) | `internal/api/api_test.go:TestAPI_ProfileSearchAndRetrieve_TP8`, Exhibit: Auth+profile round-trip | Profile search/retrieve |
| AC-9 | TP-5 | `internal/idp/connector_test.go:TestIdPConnector_WireMappingAndBothConfigs` | Connector wire fidelity (httptest; no public enrich) |
| AC-11 | What-landed: go.mod modules | `go.mod` (`oapi-codegen/v2`, `nethttp-middleware`, `runtime`, `tool` directive, locked modules), `internal/tools/tools.go` | Locked deps present and honest |
| AC-12 | This Index complete for seams | Complete table covering AC-1–AC-9, AC-11–AC-15, VC-1, VC-3, VC-4 | Parade covers remaining boundaries |
| AC-13 | TP-7 battery / `go test ./...` | `go test ./...` exits 0 (all test packages pass) | Suite green |
| AC-14 | Exhibits VC-1, VC-3, VC-4 filled | Exhibits below with live command transcripts | Live VCs captured |
| AC-15 | What-landed: `make demo` + Exhibit Spec use-case demo walk | `Makefile` (`demo` target), `scripts/demo.sh`, Exhibit: Spec use-case demo walk | Use cases U1–U5 |
| TP-1 | Commit order / artifact paths | Commit `2c29cae` (Phase A contracts) before `8c9dd67` (Phase B implementation) | Phase A before B |
| TP-2 | Named DAO SQLite test | `internal/store/sqlite_test.go:TestSQLiteDAO_StoreRetrieveSearch` | SQLite DAO |
| TP-3 | Named PG testcontainers test | `internal/store/postgres_test.go:TestPostgresDAO_StoreRetrieveSearch` | Postgres DAO |
| TP-4 | Named JWT auth tests | `internal/auth/jwt_test.go:TestAuthService_TokenIssueAndVerify`, `TestAuthService_CredentialCheckAndLogin`, `TestAuthService_AuthenticateBearer`, `internal/api/api_test.go:TestAPI_OpenAPIBearerAuthEnforcement` | Auth gate |
| TP-5 | Named connector mapping tests | `internal/idp/connector_test.go:TestIdPConnector_WireMappingAndBothConfigs` | IdP shapes |
| TP-7 | `go test ./...` | `go test ./...` exits 0 | Full suite |
| TP-8 | Named HTTP search/retrieve tests | `internal/api/api_test.go:TestAPI_ProfileSearchAndRetrieve_TP8` | Authenticated REST profile handlers |
| VC-1 | Exhibit: Auth+profile round-trip | `make demo` U1–U4 output in Exhibit: Auth+profile round-trip | Live seed→login→bearer→profile |
| VC-3 | Exhibit: Auth gate refusal | `make demo` U5 output in Exhibit: Auth gate refusal | Live reject without/invalid token |
| VC-4 | Exhibit: Spec use-case demo walk (`make demo`) | Full transcript in Exhibit: Spec use-case demo walk | U1–U5 in one Lead-facing command |

Withdrawn (Lead 2026-10-06; not in this Parade): AC-10, VC-2, TP-6, demo U6.

---

## What landed

- **DAO contract / AC-1, AC-6:** `project/schemas.md` defines the Go-facing `store.DAO` interface and schemas. Implemented in `internal/store/dao.go`, `internal/store/sqlite.go`, and `internal/store/postgres.go` using `goose` embedded migrations and `sqlc` dual packages (`sqlc_sqlite` and `sqlc_postgres`). Callers select driver via `store.DBConfig` and never import DB drivers.
- **REST + JWT / AC-2, AC-7, AC-8:** `project/openapi.yaml` and `project/decisions/rest-api-jwt-bearer.md` lock OpenAPI 3.0 and JWT bearer authentication. Implemented via `internal/api/handler.go` (`oapi-codegen` generated `internal/api/api.gen.go`), `internal/auth/auth.go` (`golang-jwt/jwt/v5`), and `nethttp-middleware` delegating to a shared `auth.Service.AuthenticateBearer` path. Login issues token on credential match; protected `/profiles` routes reject unauthenticated requests with HTTP 401 and a stable unauthorized message without echoing jwt parse details. Public REST has no enrich route.
- **IdP connector / AC-3, AC-9:** `project/integrations.md` specifies external IdP wire contracts (`POST /auth` and `POST /identity` with address object). Implemented in `internal/idp/connector.go` with automatic authentication, in-memory token caching with skew safety, and pluggable provider configs for ABC and XYC. Proof is TP-5 httptest. `cmd/fake-idp` remains as a connector test helper; `make demo` does not start it.
- **Locked dependencies & middleware / AC-11:** All locked modules from Dependency / Technology Decisions are honest in `go.mod`: `github.com/golang-jwt/jwt/v5`, `github.com/jackc/pgx/v5`, `modernc.org/sqlite`, `github.com/pressly/goose/v3`, `github.com/oapi-codegen/oapi-codegen/v2` (locked via direct require, `tool` directive in `go.mod`, and `internal/tools/tools.go`), `github.com/oapi-codegen/runtime`, `github.com/oapi-codegen/nethttp-middleware` (direct require, wired in `internal/api/handler.go`), `github.com/getkin/kin-openapi`, `github.com/google/uuid`, and `github.com/testcontainers/testcontainers-go/modules/postgres`. OpenAPI `BearerAuth` security is enforced via `nethttp-middleware` rather than URL prefix matching.
- **Package boundary enforcement / AC-4:** `internal/boundary_test.go` (`TestArchitecturalPackageBoundaries`) resolves paths relative to the test file using `runtime.Caller`, validates non-empty `.go` file sets in `internal/store` and `internal/idp`, propagates walk errors, and permanently fails on illegal cross-package imports.
- **`make demo` / AC-15:** Target `demo` in `Makefile` and script `scripts/demo.sh` stands up the Identity service (default SQLite, docker-free) and walks U1 through U5 with inspectable request/response payloads. Fake IdP is not required for this walk.

```mermaid
sequenceDiagram
  participant Caller
  participant REST as REST Handler
  participant Auth as Auth Middleware
  participant DAO as store.DAO (SQLite/Postgres)

  Note over Caller,REST: U2: Login & JWT Issue
  Caller->>REST: POST /auth/login (username, password)
  REST->>DAO: GetCredentialByUsername("alice")
  DAO-->>REST: Credential
  REST-->>Caller: 200 OK + JWT Bearer Token

  Note over Caller,DAO: U3 & U4: Profile Retrieve & Search
  Caller->>Auth: GET /profiles/{id} with Bearer <token>
  Auth->>Auth: Verify JWT
  Auth->>DAO: GetProfileByID(id)
  DAO-->>Auth: UserProfile
  Auth-->>Caller: 200 OK UserProfile JSON

  Note over Caller,Auth: U5: Auth Gate Refusal
  Caller->>Auth: GET /profiles/{id} (no token / invalid token)
  Auth-->>Caller: 401 Unauthorized
```

---

## Exhibits

### Exhibit: Auth+profile round-trip

**What the Lead should see/feel:** A running server accepts credential login, returns a JWT, and serves authenticated profile search/retrieve with `Authorization: Bearer`.
**Maps to:** AC-7, AC-8, VC-1
**Captured:** Live output from `make demo` (U1–U4)

```text
=================================================================
[U1] Durable local persistence: credential + profile stored
=================================================================
Identity service started with SQLite storage at demo_identity.db and seeded data.
Querying SQLite directly for seeded profile (Alice Smith)...
SQLite profile row: 11111111-1111-1111-1111-111111111111|Alice Smith|+15551234567
Querying SQLite directly for seeded credential (alice)...
SQLite credential row: 11111111-1111-1111-1111-111111111111|alice|password
Verified: Real SQLite readback confirmed seeded profile and credential persistence.

=================================================================
[U2] Credential check -> API auth: login and obtain JWT
=================================================================
POST http://localhost:8088/auth/login
Payload: {"username": "alice", "password": "password123"}
Response:
{"expires_in":86400,"token":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiMTExMTExMTEtMTExMS0xMTExLTExMTEtMTExMTExMTExMTExIiwidXNlcm5hbWUiOiJhbGljZSIsImlzcyI6ImlkZW50aXR5LWdvLXNlcnZpY2UiLCJzdWIiOiIxMTExMTExMS0xMTExLTExMTEtMTExMS0xMTExMTExMTExMTEiLCJleHAiOjE3OTEzNjA1OTksImlhdCI6MTc5MTI3NDE5OX0.82Zk2tmHCkyD-H7jWbWGi1fsTggSsrC67T7Ej7OM46I","token_type":"Bearer"}
Extracted Bearer Token: eyJhbGciOiJIUzI1NiIsInR5cCI6Ik...

=================================================================
[U3] Authenticated profile retrieve: Bearer + retrieve profile
=================================================================
GET http://localhost:8088/profiles/11111111-1111-1111-1111-111111111111
Header: Authorization: Bearer <token>
Response:
{"address":"123 Market St, San Francisco, CA 94105","created_at":"2026-10-06T08:09:59.834549Z","id":"11111111-1111-1111-1111-111111111111","name":"Alice Smith","phone":"+15551234567","updated_at":"2026-10-06T08:09:59.834549Z"}

=================================================================
[U4] Authenticated profile search: Bearer + search profiles
=================================================================
GET http://localhost:8088/profiles?name=Smith
Header: Authorization: Bearer <token>
Response:
[{"address":"123 Market St, San Francisco, CA 94105","created_at":"2026-10-06T08:09:59.834549Z","id":"11111111-1111-1111-1111-111111111111","name":"Alice Smith","phone":"+15551234567","updated_at":"2026-10-06T08:09:59.834549Z"},{"address":"456 Castro St, Mountain View, CA 94041","created_at":"2026-10-06T08:09:59.834549Z","id":"22222222-2222-2222-2222-222222222222","name":"Bob Smith","phone":"+15559876543","updated_at":"2026-10-06T08:09:59.834549Z"}]
```

### Exhibit: Auth gate refusal

**What the Lead should see/feel:** A protected profile route rejects missing or invalid bearer without leaking profile PII.
**Maps to:** AC-7, VC-3
**Captured:** Live output from `make demo` (U5)

```text
=================================================================
[U5] Auth gate refusal: missing/invalid bearer rejected (401)
=================================================================
GET http://localhost:8088/profiles/11111111-1111-1111-1111-111111111111 (no Authorization header)
HTTP Status: 401
Response: {"error":"unauthorized","message":"Unauthorized: missing or invalid bearer token"}

GET http://localhost:8088/profiles/11111111-1111-1111-1111-111111111111 (invalid Bearer token)
HTTP Status: 401
Response: {"error":"unauthorized","message":"Unauthorized: missing or invalid bearer token"}
```

### Exhibit: Spec use-case demo walk

**What the Lead should see/feel:** From the worktree, `make demo` stands the service (SQLite default) and walks U1–U5: persist, login→JWT, retrieve, search, auth-gate refusal — with inspectable output, not canned stubs. Fake IdP is not required.
**Maps to:** AC-15, VC-4
**Captured:** Full terminal session transcript of `make demo`

```text
$ make demo
mkdir -p bin
go build -o bin/server ./cmd/server
./scripts/demo.sh
=================================================================
 ENG-561 Identity Go Service: Use Case Demo Walk (U1 - U5)
=================================================================

=================================================================
[U1] Durable local persistence: credential + profile stored
=================================================================
Identity service started with SQLite storage at demo_identity.db and seeded data.
Querying SQLite directly for seeded profile (Alice Smith)...
SQLite profile row: 11111111-1111-1111-1111-111111111111|Alice Smith|+15551234567
Querying SQLite directly for seeded credential (alice)...
SQLite credential row: 11111111-1111-1111-1111-111111111111|alice|password
Verified: Real SQLite readback confirmed seeded profile and credential persistence.

=================================================================
[U2] Credential check -> API auth: login and obtain JWT
=================================================================
POST http://localhost:8088/auth/login
Payload: {"username": "alice", "password": "password123"}
Response:
{"expires_in":86400,"token":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiMTExMTExMTEtMTExMS0xMTExLTExMTEtMTExMTExMTExMTExIiwidXNlcm5hbWUiOiJhbGljZSIsImlzcyI6ImlkZW50aXR5LWdvLXNlcnZpY2UiLCJzdWIiOiIxMTExMTExMS0xMTExLTExMTEtMTExMS0xMTExMTExMTExMTEiLCJleHAiOjE3OTEzNjA1OTksImlhdCI6MTc5MTI3NDE5OX0.82Zk2tmHCkyD-H7jWbWGi1fsTggSsrC67T7Ej7OM46I","token_type":"Bearer"}
Extracted Bearer Token: eyJhbGciOiJIUzI1NiIsInR5cCI6Ik...

=================================================================
[U3] Authenticated profile retrieve: Bearer + retrieve profile
=================================================================
GET http://localhost:8088/profiles/11111111-1111-1111-1111-111111111111
Header: Authorization: Bearer <token>
Response:
{"address":"123 Market St, San Francisco, CA 94105","created_at":"2026-10-06T08:09:59.834549Z","id":"11111111-1111-1111-1111-111111111111","name":"Alice Smith","phone":"+15551234567","updated_at":"2026-10-06T08:09:59.834549Z"}

=================================================================
[U4] Authenticated profile search: Bearer + search profiles
=================================================================
GET http://localhost:8088/profiles?name=Smith
Header: Authorization: Bearer <token>
Response:
[{"address":"123 Market St, San Francisco, CA 94105","created_at":"2026-10-06T08:09:59.834549Z","id":"11111111-1111-1111-1111-111111111111","name":"Alice Smith","phone":"+15551234567","updated_at":"2026-10-06T08:09:59.834549Z"},{"address":"456 Castro St, Mountain View, CA 94041","created_at":"2026-10-06T08:09:59.834549Z","id":"22222222-2222-2222-2222-222222222222","name":"Bob Smith","phone":"+15559876543","updated_at":"2026-10-06T08:09:59.834549Z"}]

=================================================================
[U5] Auth gate refusal: missing/invalid bearer rejected (401)
=================================================================
GET http://localhost:8088/profiles/11111111-1111-1111-1111-111111111111 (no Authorization header)
HTTP Status: 401
Response: {"error":"unauthorized","message":"Unauthorized: missing or invalid bearer token"}

GET http://localhost:8088/profiles/11111111-1111-1111-1111-111111111111 (invalid Bearer token)
HTTP Status: 401
Response: {"error":"unauthorized","message":"Unauthorized: missing or invalid bearer token"}

=================================================================
 All use cases U1 through U5 successfully demonstrated!
=================================================================

--- Cleaning up background demo processes ---
Demo completed and cleaned up.
```

---

## Reproduction appendix

- **Battery report:** `/tmp/helm-ir-battery/ENG-561/<sha>.json` (generated via `make ir-instrument`)
- **Named tests:**
  - `internal/store/sqlite_test.go:TestSQLiteDAO_StoreRetrieveSearch` (TP-2)
  - `internal/store/postgres_test.go:TestPostgresDAO_StoreRetrieveSearch` (TP-3)
  - `internal/auth/jwt_test.go:TestAuthService_TokenIssueAndVerify`, `TestAuthService_CredentialCheckAndLogin`, `TestAuthService_AuthenticateBearer` (TP-4)
  - `internal/api/api_test.go:TestAPI_OpenAPIBearerAuthEnforcement` (TP-4 live router gate)
  - `internal/idp/connector_test.go:TestIdPConnector_WireMappingAndBothConfigs` (TP-5)
  - `internal/boundary_test.go:TestArchitecturalPackageBoundaries` (AC-4)
  - `internal/api/api_test.go:TestAPI_ProfileSearchAndRetrieve_TP8` (TP-8)
- **VC drivers:**
  - `VC-1`: `make demo` (drives U1–U4: seed -> login -> bearer -> profile retrieve/search)
  - `VC-3`: `make demo` (drives U5: auth gate refusal with 401)
  - `VC-4`: `make demo` (full walk U1–U5)
- **Commands:**
  - `go test ./...` — executes the full unit and integration test suite
  - `make demo` — runs live service, exercising U1 through U5

---

## Known Gaps

- Interview mock: credential-at-rest and JWT signing defaults use mock/dev secrets by design per `project/decisions/interview-mock-security-latitude.md`; Parade does not prove production hardening.
- Real LoginID cloud / SDK / passkeys are out of scope per brief; exhibits use local JWT only. Connector httptest covers vendor `/auth` + `/identity` without a public enrich route.
