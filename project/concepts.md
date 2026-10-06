---
description: "Core domain vocabulary and mental models for this project."
---

# Concepts

Vocabulary for this project's prose. Technical surfaces (APIs, schemas, CLIs) may use other words; briefs, decisions, and status reports follow this language. Seeded at `/helm-init` from the Identity scope; extend when work surfaces a new term worth pinning.

## Language

**User profile**:
Stored personal record for a person: name, address, and phone. Owned by the persistence layer; retrieved by the REST API and enriched from identity providers when needed.
_Avoid_: Treating profile fields as authentication secrets; credentials are a separate concept.

**User credential**:
Authentication material for a user: username, method, and password (or equivalent). Stored via the DAO; not returned as public profile PII.
_Avoid_: Calling this "identity" when you mean vendor IdP personal data.

**DAO**:
Data-access boundary that can store and retrieve user profiles and credentials against either PostgreSQL or SQLite behind one interface.
_Avoid_: Putting HTTP or IdP logic inside the DAO.

**Identity provider (IdP)**:
A third-party vendor (ABC or XYC) that exposes `/auth` (access token from username/password) and `/identity` (PII lookup by phone and name).
_Avoid_: Confusing local stored credentials with vendor IdP tokens.

**Service connector**:
Client that talks to an IdP's `/auth` and `/identity` endpoints and returns structured personal data to the rest of the system.
_Avoid_: Embedding vendor-specific HTTP details in REST handlers or DAOs.

## Relationships

- The DAO owns durable local storage of user profile and user credential.
- The REST API authenticates callers and searches/retrieves user profile data; it does not own IdP wire formats.
- The service connector owns IdP `/auth` and `/identity` contracts; profiles may be populated or refreshed from connector results.

## Flagged ambiguities

- Vendor names ABC and XYC are placeholders; treat their endpoint shapes as the contract until real vendor docs replace them.
- "Security API authentication" for the REST surface is not yet locked (scheme, token source, middleware).
