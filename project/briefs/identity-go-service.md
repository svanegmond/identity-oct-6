---
description: "Land-ready contract for an Identity Go service used as a LoginID interview discussion object (DAO, JWT REST, IdP connector), not a production security product."
date: 2026-10-06
ref: ENG-561
---

# Brief: Identity Go service (interview discussion object)

## Problem

This workspace has no Identity Go service. The Lead needs a **discussable, runnable object** for LoginID-shaped interview conversation: local profile/credential persistence, authenticated REST profile search and retrieve, and a pluggable IdP connector that fetches PII. Oct-5 is library/layout inspiration only; cancelled ENG-560 / oct-5 product scope does not apply. Without a locked contract, Implementor cannot sequence design artifacts before code, or prove a composed IdP path beyond isolated httptest stubs.

## Premises

| # | Claim | Lead |
|---|-------|------|
| P1 | This deliverable is an **interview discussion object** aligned with LoginID docs spirit (local JWT issue/verify, separate auth vs identity/PII), not a production multi-tenant security product. | agree |
| P2 | Phase A contract artifacts must exist before Phase B Go code; Phase A→B→C may ship in one Implementor cycle with one commit per phase. | agree (SK-1) |
| P3 | REST callers authenticate with **JWT bearer** after local credential check (`Authorization: Bearer`). | agree (Auth D1=O1) |
| P4 | A composed callable path must invoke the IdP connector and **return PII to the caller** without merging IdP into the DAO; proof must exceed isolated httptest-only connector VCs. Persisting connector PII via DAO is optional latitude, not required for AC-10/VC-2. | agree (SK-2) |
| P5 | Credential-at-rest and JWT signing-key defaults are Implementor latitude for this mock (including weak/plain ideas); argon2id and non-default signing keys are not hard ACs. | agree (SK-3) |
| P6 | Profile search stays in scope. Passkeys/FIDO2, LoginID management/grant APIs, transaction confirmation, and a real LoginID SDK are out. | agree |

## Binary Forks

| Fork | Option A | Option B | Lead pick |
|------|----------|----------|-----------|
| REST auth scheme | **O1 JWT bearer** after credential check | O2 API key / O3 HTTP basic | **O1** — `project/decisions/rest-api-jwt-bearer.md` |
| Product framing | Interview mock / discussion object | Production security bar | **Interview mock** — SK-3 latitude |
| Profile search | In scope on authenticated REST | Out of scope (oct-5 constraint) | **In scope** |
| IdP PII composition | Composed REST enrich/lookup and/or thin `cmd` calling connector | Connector-only unit/httptest proof | **Composed callable path required** (SK-2) |
| Dual-DB DAO | PostgreSQL **and** SQLite behind one interface | Single backend only | **Both** behind one interface |

No open Lead forks remain.

## Approach

Sequenced shape. One Implementor cycle; one commit per phase; Phase A artifacts must exist before Phase B code starts.

### Phase A — Design contracts

After this phase, separate contract artifacts under `project/` define the DAO (profile/credential fields, store/retrieve/search, dual-DB selection without leaking into callers), the REST surface (routes, shapes, error model, JWT bearer auth), and the IdP connector (`POST /auth`, `POST /identity` wire shapes for ABC/XYC). Concepts/schemas/integrations (or OpenAPI under authorized taxonomy) match those contracts. Auth decision is recorded. No Phase B product Go packages yet.

### Phase B — Implement against contracts

After this phase, a Go service implements the accepted contracts: DAO adapters for PostgreSQL and SQLite, JWT issue after credential check and verify before protected routes, authenticated profile search/retrieve, IdP connector(s) free of DAO details, and a **composed callable path** (authenticated REST enrich/lookup route and/or thin `cmd`) that invokes the connector and **returns PII to the caller** without merging IdP into the DAO (persist-via-DAO optional latitude). Locked Dependency / Technology Decisions modules are in use. Package seams keep DAO free of HTTP/IdP and connectors free of DAO.

### Phase C — Prove

After this phase, `go test ./...` is green with named tests covering DAO on **both** PostgreSQL and SQLite backends (TP-2 + TP-3; no one-backend dialect-seam escape), auth gate, connector mapping, authenticated HTTP search/retrieve (TP-8), and the composed IdP path (behavioral TP-6). Proof Parade exhibits live request/response (or terminal) evidence for the VCs below.

## Seams

- **Existing seam used**: N/A — greenfield Identity service in this workspace.
- **New seam introduced**:
  - **DAO / persistence port** — `user_profile` and `user_credential` store/retrieve/search; PostgreSQL and SQLite adapters; callers never import DB drivers.
  - **HTTP / wire projection** — OpenAPI-generated (or equivalent) handlers + JWT middleware; auth vs identity/PII routes stay separable in contract language.
  - **IdP service connector** — client interface for ABC/XYC `/auth` and `/identity`; second adapter (or second config targeting the same interface) justifies the seam.
- **Not a seam here**: LoginID SDK wrapper; passkey/FIDO2 authenticator; merging IdP fetch into the DAO interface.

## Artifacts

- **Consulted**: `project/tmp/agent-prompt-identity-go.md` (source Phase A/B/C intent); `project/concepts.md`; `project/index.md`; `project/helm-config.yaml`; sibling `../agentic-eng-oct-5` go.mod / layout for library ideas only; https://docs.loginid.io (spirit: verify-token, auth vs identity separation).
- **Create/update**:
  - `project/briefs/identity-go-service.md` — this contract
  - `project/demos/ENG-561-identity-go-service.md` — Proof Parade scaffold
  - `project/decisions/rest-api-jwt-bearer.md` — Auth D1=O1
  - `project/decisions/interview-mock-security-latitude.md` — SK-3 latitude
  - `project/concepts.md` — JWT + interview-mock framing; clear auth ambiguity
  - Phase A (Implementor): `project/schemas.md` (or `schemas/`), `project/integrations.md` (or `integrations/`), OpenAPI under authorized `project/` taxonomy as needed

## Context Payload

- **Target Files**: Prefer `cmd/`, `internal/`, `api/` per `helm-config.yaml` `production_roots`. Do not edit Helm methodology plugin files. Do not copy oct-5 product scope (e.g. forbidding profile search, requiring passkey provider).
- **Required Context**: This brief; decisions above; source prompt; LoginID docs for discussion alignment only (no SDK).
- **Discussion Decisions**:
  - Auth = JWT bearer after credential check (`Authorization: Bearer`).
  - SK-1: Plan Review + Intend = Lead acceptance of contract requirements; sequenced A→B→C in one Implementor cycle OK; Phase A before Phase B code.
  - SK-2: Composed callable IdP path required; returns PII to caller; persist-via-DAO after connector is latitude; AC+VC beyond isolated httptest.
  - VC-1 credential bootstrap: DAO seed/fixture by default; REST register only if Phase A contracts it.
  - SK-3: Credential-at-rest and JWT signing-key defaults are latitude; argon2id / strong keys not hard ACs.
  - LoginID **in**: local JWT issue+verify before protected routes; separate auth vs identity/PII; optional JWT scopes OK. **Out**: passkeys/FIDO2, LoginID mgmt/grant APIs, transaction confirmation, real LoginID SDK.
  - Profile search remains in scope.
- **Dependency / Technology Decisions** (locked — not "pure latitude"; from oct-5 go.mod menu):
  - `github.com/golang-jwt/jwt/v5`
  - `github.com/jackc/pgx/v5`
  - `modernc.org/sqlite`
  - `github.com/pressly/goose/v3`
  - sqlc dual-packages pattern (PG + SQLite)
  - `github.com/oapi-codegen/oapi-codegen/v2` + `github.com/oapi-codegen/runtime` + `github.com/oapi-codegen/nethttp-middleware`
  - `github.com/getkin/kin-openapi`
  - `github.com/google/uuid`
  - `github.com/testcontainers/testcontainers-go/modules/postgres` for PostgreSQL tests
  - `github.com/alexedwards/argon2id` may appear only as optional Implementor choice under SK-3 latitude — **not** a required AC
- **External References**: https://docs.loginid.io — discussion alignment only.

## Demonstration Plan

- **Proof Parade path:** `project/demos/ENG-561-identity-go-service.md`
- [ ] **VC-1** Stand the HTTP server (SQLite or Postgres). Bootstrap a credential via **DAO seed/fixture** (or a REST register route **only if** Phase A contracts one — not required otherwise), obtain JWT via login, call authenticated profile search/retrieve with `Authorization: Bearer <token>`; capture request/response. Exhibit: Auth+profile round-trip.
- [ ] **VC-2** Drive the composed IdP path (authenticated REST enrich/lookup and/or thin `cmd`) against a fake/stub IdP implementing `/auth` + `/identity`; show PII **returned to the caller** without DAO importing IdP packages. Persist-after-connector is latitude, not this VC. Capture request/response or terminal session. Exhibit: Composed IdP path. (Isolated connector httptest alone does not satisfy this VC.)
- [ ] **VC-3** Show unauthenticated (or invalid-token) call to a protected profile route is rejected; capture status/body. Exhibit: Auth gate refusal.

## Test Plan

### Phase A

- [ ] **TP-1** Contract review gate (mechanical): Phase A artifacts exist at paths named in AC-1–AC-4 before any Phase B product package is introduced (enforced by Implementor commit order + IR; no prose-substring test).

### Phase B / C

- [ ] **TP-2** Unit/integration: DAO store/retrieve/search for `user_profile` and `user_credential` against SQLite.
- [ ] **TP-3** Integration: DAO against PostgreSQL via testcontainers-go postgres module. Both PostgreSQL and SQLite adapters must be implemented and exercised (SQLite via TP-2, PostgreSQL via this row). No dialect-seam-doc escape that greens with only one live backend.
- [ ] **TP-4** Unit/httptest: JWT issued after successful credential check; protected route rejects missing/invalid bearer; accepts valid bearer.
- [ ] **TP-5** Unit/httptest: IdP connector maps `/auth` and `/identity` request/response shapes (including address object) for ABC/XYC config targets.
- [ ] **TP-6** Unit/integration (behavioral): Composed path invokes the IdP connector and returns PII to the caller. An import/architecture or package-boundary check may only **supplement** the “DAO has no IdP dependency” half of AC-10; it cannot be the sole proof that the path calls the connector or returns PII.
- [ ] **TP-7** Suite: `go test ./...` exits 0.
- [ ] **TP-8** Unit/HTTP: Authenticated profile **search** and **retrieve** handlers (valid Bearer) return contracted profile shapes and error model per AC-8 / REST contract. DAO-only or JWT-gate-only tests do not satisfy this row.

## Acceptance Criteria

### Phase A

- [ ] **AC-1** DAO contract artifact exists under authorized `project/` taxonomy: Go-facing interfaces; `user_profile` (name, address, phone) and `user_credential` (username, method, password) fields; store/retrieve/search; how PostgreSQL vs SQLite is selected without leaking into callers.
- [ ] **AC-2** REST contract (OpenAPI or equivalent under `project/`) defines profile search and retrieve routes, request/response shapes, error model, and JWT bearer authentication (`Authorization: Bearer`) per `project/decisions/rest-api-jwt-bearer.md`.
- [ ] **AC-3** IdP connector contract exists (`project/integrations.md` or equivalent): ABC/XYC client interface; `POST /auth` and `POST /identity` types; error and token-handling rules; address shape `{street_address, locality, region, postal_code, country}`.
- [ ] **AC-4** Contracts state separation: DAO must not import HTTP or IdP packages; connector must not own DAO persistence; auth (credential/JWT) vs identity/PII concerns are named separately (LoginID spirit).
- [ ] **AC-5** Phase A commit(s) land contract artifacts before Phase B introduces product Go implementation packages under `cmd/`, `internal/`, or `api/`.

### Phase B

- [ ] **AC-6** Go DAO implements the Phase A interface for both PostgreSQL and SQLite (goose migrations + sqlc dual packages as locked).
- [ ] **AC-7** After successful local credential check, service issues a JWT; protected REST routes verify JWT before handler logic (verify-token spirit).
- [ ] **AC-8** Authenticated REST supports profile **search** and retrieve per REST contract.
- [ ] **AC-9** IdP connector implements `/auth` and `/identity` against the Phase A contract; ABC and XYC differ by config/base URL unless Phase A documents a justified fork.
- [ ] **AC-10** A composed callable path (authenticated REST enrich/lookup route and/or thin `cmd`) invokes the IdP connector and **returns PII to the caller** without merging IdP into the DAO interface (SK-2). Persisting that PII via the DAO after the connector call is **Allowed latitude**, not required to meet AC-10 or VC-2.
- [ ] **AC-11** Locked Dependency / Technology Decisions modules are present in `go.mod` (argon2id optional only).
- [ ] **AC-12** Proof Parade demonstrates each new boundary/seam with inspectable evidence (Evidence Index rows for AC-6–AC-10 and VC-1–VC-3).

### Phase C

- [ ] **AC-13** `go test ./...` passes; TP-2–TP-8 named tests exist and guard the contracts above.
- [ ] **AC-14** VC-1–VC-3 evidence is captured in `project/demos/ENG-561-identity-go-service.md` (live surface, not test-name-as-VC-proof).

## Validation Criteria

- [ ] **VC-1** Shape: single-surface. Medium: request/response transcript against a running server. Lead sees credential bootstrap (DAO seed/fixture, or Phase-A-contracted register if present) → login → bearer → authenticated profile search/retrieve succeed. → Parade Exhibit: Auth+profile round-trip.
- [ ] **VC-2** Shape: single-surface (integration-shaped if fake IdP is a separate process). Medium: request/response and/or terminal session. Lead sees composed path call IdP `/auth`+`/identity` and **return PII to the caller**; DAO remains free of IdP. Persist-after-connector is latitude, not this VC. Fake-green resistance: connector unit test alone must not satisfy this VC. → Parade Exhibit: Composed IdP path.
- [ ] **VC-3** Shape: single-surface. Medium: request/response transcript. Lead sees protected profile route refuse missing/invalid bearer. → Parade Exhibit: Auth gate refusal.

## Tracking

- **Task ref**: ENG-561
- **Task body (tracker)**: pointer only — `brief_path: project/briefs/identity-go-service.md`

## NOT in Scope

- Production multi-tenant SaaS, hardening for real credential theft, or security audit bar — interview discussion object only.
- Passkeys / FIDO2 / WebAuthn.
- LoginID management, grant, or admin APIs; transaction confirmation flows.
- Real LoginID SDK or live LoginID cloud dependency.
- UI / frontend.
- Helm methodology plugin file edits.
- Copying oct-5 product constraints (e.g. dropping profile search; requiring a passkey provider).
- Inventing new `$ARTIFACT_DIR` top-level taxonomy directories without Lead approval.

**Allowed latitude** (Implementor may choose without escalating):

- Credential-at-rest representation and hashing, including weak or plain storage ideas appropriate to a mock (SK-3). **argon2id is not a hard AC.**
- JWT signing-key material and defaults, including weak/dev defaults (SK-3). Non-default/rotated production key management is **not** required.
- Optional JWT scopes claims if useful for discussion; not required for AC-7.
- Exact package names under `internal/`, HTTP router wiring details, and whether the composed IdP path is REST-only, `cmd`-only, or both — provided AC-10 and VC-2 are met.
- Whether to persist IdP-returned PII through the DAO after the composed call (optional; not required for AC-10/VC-2).
- Optional REST register/signup route for credential bootstrap if Phase A contracts it; otherwise VC-1 uses DAO seed/fixture.
- ABC vs XYC as two configs on one connector interface when wire shapes match.

## Error & Failure Map

| Codepath | Failure | Caller-visible outcome |
|----------|---------|------------------------|
| Login / credential check | Unknown user or bad password | Auth error response; no JWT issued (exact status/body per REST contract) |
| Protected route | Missing/invalid/expired bearer | Reject before handler; no profile PII leak |
| Profile search/retrieve | Not found | Documented not-found error; no 500 for empty result |
| DAO / DB | Connection or migration failure | Fail closed at startup or request; error logged without password/token secrets |
| IdP connector `/auth` | Vendor auth failure / transport error | Mapped connector error to composed path; no panic; no secret echo in logs |
| IdP connector `/identity` | Not found / malformed PII / transport error | Mapped error to caller; if Implementor chooses persist-after-connector latitude, partial PII must not be silently written as success |
| Composed path | IdP up, DAO down (or reverse) | Error identifies failing seam; no merge of IdP types into DAO API |

## References

- `project/tmp/agent-prompt-identity-go.md`
- `project/decisions/rest-api-jwt-bearer.md`
- `project/decisions/interview-mock-security-latitude.md`
- `project/concepts.md`
- `project/demos/ENG-561-identity-go-service.md`
- https://docs.loginid.io
- Sibling `../agentic-eng-oct-5` — library/layout ideas only
