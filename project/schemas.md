---
description: "Data shapes, schemas, and Go-facing DAO contracts for SQLite and PostgreSQL persistence."
date: 2026-10-06
---

# Schemas and DAO Contract

This document specifies the persistence schemas and Go-facing DAO contracts for the Identity Go service. It addresses **AC-1** and **AC-4** of ENG-561.

## 1. Domain Entities and Fields

### `user_profile`
Stores personal identification and contact information.
- `id`: unique identifier (UUID v4 string)
- `name`: user's full name (string)
- `address`: user's formatted physical address (string)
- `phone`: user's contact telephone number (string)
- `created_at`: timestamp of creation (RFC 3339 timestamp)
- `updated_at`: timestamp of last modification (RFC 3339 timestamp)

### `user_credential`
Stores authentication material. Separated from `user_profile` to enforce the distinction between auth credentials and identity PII (LoginID spirit).
- `id`: unique credential identifier (UUID v4 string)
- `user_id`: foreign reference to associated user/profile (UUID string or unique identifier)
- `username`: unique login username (string)
- `method`: authentication method (string, e.g., `"password"`)
- `password`: password representation (string; mock hash or plaintext per SK-3 interview-mock latitude)
- `created_at`: timestamp of creation (RFC 3339 timestamp)
- `updated_at`: timestamp of last modification (RFC 3339 timestamp)

---

## 2. Go-Facing DAO Interface

Callers interact solely with the public DAO interface. Callers **never** import database drivers (`pgx`, `modernc.org/sqlite`) or database-specific query packages.

```go
package store

import (
	"context"
	"time"
)

type UserProfile struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Address   string    `json:"address"`
	Phone     string    `json:"phone"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type UserCredential struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Username  string    `json:"username"`
	Method    string    `json:"method"`
	Password  string    `json:"-"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ProfileSearchQuery struct {
	Name  string
	Phone string
}

// DAO defines data access for user profiles and credentials.
type DAO interface {
	// Profile operations
	CreateProfile(ctx context.Context, p *UserProfile) error
	GetProfileByID(ctx context.Context, id string) (*UserProfile, error)
	SearchProfiles(ctx context.Context, q ProfileSearchQuery) ([]*UserProfile, error)

	// Credential operations
	CreateCredential(ctx context.Context, c *UserCredential) error
	GetCredentialByUsername(ctx context.Context, username string) (*UserCredential, error)

	// Lifecycle
	Close() error
	Ping(ctx context.Context) error
}
```

---

## 3. Dual-Database Selection and Factory

Backend selection is controlled purely via configuration without leaking dialect or driver details to consumers.

```go
type DBConfig struct {
	Driver string // "sqlite" or "postgres"
	DSN    string // connection string or file path
}

// Open initializes the configured backend, applies goose migrations, and returns a DAO.
func Open(ctx context.Context, cfg DBConfig) (DAO, error)
```

- When `Driver == "sqlite"`: opens local file or `:memory:` via `modernc.org/sqlite`, applies SQLite migrations.
- When `Driver == "postgres"`: opens connection pool via `github.com/jackc/pgx/v5`, applies PostgreSQL migrations.
- Callers (HTTP handlers, services, CLI) hold a `store.DAO` reference and have zero awareness of the underlying SQL engine.

---

## 4. SQL Schema Definitions

### SQLite (`internal/store/sqlite/migrations/00001_init.sql`)

```sql
-- +goose Up
CREATE TABLE IF NOT EXISTS user_profiles (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    address TEXT NOT NULL,
    phone TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS user_credentials (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    username TEXT NOT NULL UNIQUE,
    method TEXT NOT NULL DEFAULT 'password',
    password TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_user_profiles_name ON user_profiles(name);
CREATE INDEX IF NOT EXISTS idx_user_profiles_phone ON user_profiles(phone);

-- +goose Down
DROP TABLE IF EXISTS user_credentials;
DROP TABLE IF EXISTS user_profiles;
```

### PostgreSQL (`internal/store/postgres/migrations/00001_init.sql`)

```sql
-- +goose Up
CREATE TABLE IF NOT EXISTS user_profiles (
    id VARCHAR(64) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    address TEXT NOT NULL,
    phone VARCHAR(64) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS user_credentials (
    id VARCHAR(64) PRIMARY KEY,
    user_id VARCHAR(64) NOT NULL,
    username VARCHAR(255) NOT NULL UNIQUE,
    method VARCHAR(64) NOT NULL DEFAULT 'password',
    password TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_user_profiles_name ON user_profiles(name);
CREATE INDEX IF NOT EXISTS idx_user_profiles_phone ON user_profiles(phone);

-- +goose Down
DROP TABLE IF EXISTS user_credentials;
DROP TABLE IF EXISTS user_profiles;
```

---

## 5. Seam Separation Invariants (AC-4)

1. **DAO does not import HTTP or IdP packages**:
   - The `store` package does not import `net/http`, OpenAPI-generated packages, or IdP connector packages.
   - It deals strictly with local storage, entities, and SQL execution.
2. **Connector does not own persistence**:
   - Third-party IdP clients never read from or write to the DAO.
3. **Auth vs. Identity/PII separation**:
   - Authentication secrets (`user_credential`) and personal profile information (`user_profile`) are kept distinct. Credentials are never returned by profile search/retrieve endpoints.
