---
description: "Third-party identity provider (IdP) connector wire contracts, client interfaces, and token-handling rules."
date: 2026-10-06
---

# IdP Integrations Contract

This document specifies the wire contracts, client interfaces, and token-handling rules for connecting to third-party Identity Providers (IdPs, e.g., ABC and XYC). It addresses **AC-3** and **AC-4** of ENG-561.

## 1. External IdP Wire Contracts

External IdP providers ABC and XYC share an identical wire protocol and differ only by configuration (Base URL, client credentials).

### `POST /auth` (Vendor Authentication)
Used to obtain an access token from the IdP.

- **Request Headers**: `Content-Type: application/json`
- **Request Body**:
  ```json
  {
    "username": "<string>",
    "password": "<string>"
  }
  ```
- **Success Response (`200 OK`)**:
  ```json
  {
    "access_token": "<string>",
    "token_type": "Bearer",
    "expires_in": 3600
  }
  ```
- **Error Response (`401 Unauthorized`)**:
  ```json
  {
    "error": "invalid_credentials",
    "message": "Vendor authentication failed"
  }
  ```

### `POST /identity` (Vendor PII Lookup)
Used to fetch detailed PII for an identity given their phone and name.

- **Request Headers**:
  - `Content-Type: application/json`
  - `Authorization: Bearer <access_token>`
- **Request Body**:
  ```json
  {
    "phone": "<string>",
    "name": "<string>"
  }
  ```
- **Success Response (`200 OK`)**:
  ```json
  {
    "name": "<string>",
    "phone": "<string>",
    "address": {
      "street_address": "<string>",
      "locality": "<string>",
      "region": "<string>",
      "postal_code": "<string>",
      "country": "<string>"
    }
  }
  ```
- **Error Responses**:
  - `401 Unauthorized`: Token expired or invalid. Triggers token re-fetch and retry.
  - `404 Not Found`: No identity record found for phone and name.
  - `500 Internal Server Error`: Vendor service error.

---

## 2. Go-Facing Client Interface

The Go client contract isolates the rest of the application from HTTP wire details, serialization, and vendor specifics.

```go
package idp

import (
	"context"
	"time"
)

type Address struct {
	StreetAddress string `json:"street_address"`
	Locality      string `json:"locality"`
	Region        string `json:"region"`
	PostalCode    string `json:"postal_code"`
	Country       string `json:"country"`
}

type IdentityPII struct {
	Name    string  `json:"name"`
	Phone   string  `json:"phone"`
	Address Address `json:"address"`
}

type ProviderConfig struct {
	Name     string        // "ABC" or "XYC"
	BaseURL  string        // e.g. "https://api.abc-idp.example.com"
	Username string        // vendor client credentials
	Password string
	Timeout  time.Duration
}

// Connector is the public interface for third-party IdP interaction.
type Connector interface {
	// Authenticate explicitly acquires an access token from the IdP /auth endpoint.
	Authenticate(ctx context.Context, username, password string) (string, error)

	// GetIdentity fetches PII from the IdP /identity endpoint using a given bearer token.
	GetIdentity(ctx context.Context, token, name, phone string) (*IdentityPII, error)

	// FetchIdentity manages authentication and token caching transparently to fetch PII.
	FetchIdentity(ctx context.Context, name, phone string) (*IdentityPII, error)
}
```

---

## 3. Token-Handling and Lifecycle Rules

1. **Automatic Acquisition**: When `FetchIdentity` is called, the connector checks if a valid cached token exists. If not, it calls `POST /auth`.
2. **Caching**: The token is cached in-memory with an expiration timestamp derived from `expires_in` (with a 30-second skew safety window).
3. **Retry on 401**: If `POST /identity` returns `401 Unauthorized`, the connector invalidates its cached token, re-authenticates via `POST /auth`, and retries `POST /identity` once before failing.
4. **Transport Safety**: Secrets (passwords and bearer tokens) are never logged.

---

## 4. Seam Separation Invariants (AC-4)

1. **Connector does not own persistence**:
   - The IdP connector packages (`internal/idp`) do not import or reference `store.DAO`.
   - The connector is an egress client that produces `IdentityPII`.
2. **DAO does not import IdP**:
   - The persistence layer does not import or know about IdP connectors or vendor types.
3. **No public REST enrich path**:
   - The connector is a library seam proved by httptest (AC-9 / TP-5). Public REST does not expose `POST /profiles/enrich` or any composed caller that returns vendor PII to API clients (Lead 2026-10-06; AC-10 / VC-2 withdrawn).
