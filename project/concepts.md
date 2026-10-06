---
description: "Core domain vocabulary and mental models for this project."
---

# Concepts

Vocabulary for this project's prose. Technical surfaces (APIs, schemas, CLIs) may use other words; briefs, decisions, and status reports follow this language. Seeded at `/helm-init` from the Identity scope; extend when work surfaces a new term worth pinning.

## Framing

**Interview discussion object**:
A runnable Identity Go service shaped for LoginID interview conversation (local JWT issue/verify, separate auth vs identity/PII), not a production multi-tenant security product. Credential-at-rest and JWT signing-key defaults may be weak by design; see `project/decisions/interview-mock-security-latitude.md`.
_Avoid_: Treating ENG-561 acceptance as a production hardening bar; copying oct-5 product scope constraints.

## Language

**User profile**:
Stored personal record for a person: name, address, and phone. Owned by the persistence layer; retrieved by the REST API and enriched from identity providers when needed.
_Avoid_: Treating profile fields as authentication secrets; credentials are a separate concept.

**User credential**:
Authentication material for a user: username, method, and password (or equivalent). Stored via the DAO; not returned as public profile PII. At-rest representation is Implementor latitude under interview-mock framing.
_Avoid_: Calling this "identity" when you mean vendor IdP personal data.

**DAO**:
Data-access boundary that can store and retrieve user profiles and credentials against either PostgreSQL or SQLite behind one interface.
_Avoid_: Putting HTTP or IdP logic inside the DAO.

**JWT bearer auth**:
Caller authentication for the REST API: after a successful local credential check, the service issues a JWT; protected routes require `Authorization: Bearer <token>` and verify the token before handler logic (LoginID verify-token spirit, local only).
_Avoid_: Equating local JWT auth with LoginID cloud, passkeys/FIDO2, or management/grant APIs. Decision: `project/decisions/rest-api-jwt-bearer.md`.

**Identity provider (IdP)**:
A third-party vendor (ABC or XYC) that exposes `/auth` (access token from username/password) and `/identity` (PII lookup by phone and name).
_Avoid_: Confusing local stored credentials or local JWTs with vendor IdP tokens.

**Service connector**:
Client that talks to an IdP's `/auth` and `/identity` endpoints and returns structured personal data to the rest of the system. Invoked from a composed callable path (authenticated REST enrich/lookup and/or thin `cmd`), not merged into the DAO.
_Avoid_: Embedding vendor-specific HTTP details in REST handlers or DAOs.

## Relationships

- The DAO owns durable local storage of user profile and user credential.
- The REST API authenticates callers with JWT bearer tokens and searches/retrieves user profile data; it does not own IdP wire formats.
- Auth (credential check + JWT) is separate from identity/PII (profile + IdP enrich), even when both appear on one server.
- The service connector owns IdP `/auth` and `/identity` contracts; profiles may be populated or refreshed from connector results via a composed path outside the DAO interface.

## Flagged ambiguities

- Vendor names ABC and XYC are placeholders; treat their endpoint shapes as the contract until real vendor docs replace them.
- Optional JWT scopes are allowed for discussion; whether scopes are required on any route is left to Phase A REST contract detail unless the brief hardens them later.
