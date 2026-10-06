# Identity Go service

Interview discussion object: local profiles and credentials, JWT REST search/retrieve, SQLite or Postgres. Not production LoginID. No public IdP enrich route.

## Run

Needs Go 1.26+, `sqlite3` on PATH for the demo.

```bash
make demo    # seed SQLite, login, search/retrieve, 401 gate
go test ./...
```

Seed user: `alice` / `password123`.

```bash
go run ./cmd/server -db sqlite -dsn identity.db -port 8080 -seed
```

| Method | Path | Auth |
|--------|------|------|
| POST | `/auth/login` | none |
| POST | `/auth/register` | none |
| GET | `/profiles?name=` | Bearer |
| GET | `/profiles/{id}` | Bearer |

Postgres tests use Testcontainers (Docker). JWT secret default is a mock string; passwords stored as-is.

## Layout

| Path | What |
|------|------|
| [`cmd/server`](cmd/server) | HTTP process |
| [`internal/store`](internal/store) | DAO, SQLite + Postgres |
| [`internal/auth`](internal/auth) | credential check, JWT |
| [`internal/api`](internal/api) | OpenAPI handlers |
| [`internal/idp`](internal/idp) | vendor `/auth` + `/identity` client (httptest) |
| [`project/openapi.yaml`](project/openapi.yaml) | REST contract |

More: [`project/demos/ENG-561-engineering-handoff.md`](project/demos/ENG-561-engineering-handoff.md).
