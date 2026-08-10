# Savings Tracker API

Backend for the Frontend Mentor Savings Tracker challenge. A Go HTTP API for
users to track savings goals and deposits toward them, with JWT auth, email
password resets, and rate limiting.

```
Client
  │  HTTP request
  ▼
router.NewRouter (internal/router)
  │  wraps handlers in middleware: Log → RequireAuth (protected routes)
  ▼
handlers (api/users, api/goals, api/health)
  │  decode + validate input, map domain errors to HTTP responses
  ▼
sqlc generated queries (internal/database)
  │
  ▼
PostgreSQL
```

## Tech stack

| Layer      | Choice                                   |
| ---------- | ---------------------------------------- |
| Language   | Go 1.24 (stdlib `net/http`)              |
| Database   | PostgreSQL                               |
| Migrations | goose (SQL files in `sql/schema`)        |
| Queries    | sqlc (SQL in `sql/queries`)              |
| Auth       | JWT (HS256) + argon2id password hashing  |
| Email      | Resend                                   |
| Routing    | stdlib `http.ServeMux` method patterns   |

## Prerequisites

- Go 1.24+
- PostgreSQL (running locally)
- [goose](https://github.com/pressly/goose) CLI
- [sqlc](https://sqlc.dev) CLI (only needed when changing queries)

## Setup

1. Clone the repo and move into `backend/`:

   ```sh
   cd backend
   ```

2. Create your environment file from the template:

   ```sh
   cp .env.example .env
   ```

3. Fill in the values in `.env`. See the [Configuration](#configuration)
   section for each variable.

4. Create the database (and a test database for integration tests):

   ```sh
   createdb savings_tracker
   createdb savings_tracker_test
   ```

5. Apply the migrations:

   ```sh
   goose postgres "$DB_URL" up
   ```

   `DB_URL` is the value from your `.env`, e.g.:

   ```sh
   goose postgres "postgres://USER:PASSWORD@localhost:5432/savings_tracker?sslmode=disable" up
   ```

6. Start the server:

   ```sh
   go run .
   ```

   The server listens on `:8080` and serves static files (the `index.html`
   in the working directory) at `/app`.

## Configuration

All configuration is read from environment variables loaded via
[godotenv](https://github.com/joho/godotenv) from `.env` in the working
directory. See `.env.example` for a documented template.

| Variable         | Required | Purpose                                                       |
| ---------------- | -------- | ------------------------------------------------------------- |
| `DB_URL`         | yes      | Postgres connection string used by the server                 |
| `DB_URL_TEST`    | yes*     | Postgres connection string used by integration tests          |
| `RESEND_API_KEY` | yes      | Resend API key for password reset emails                      |
| `FROM_EMAIL`     | yes      | From address for outgoing emails                              |
| `JWT_SECRET`     | yes      | Secret used to sign and verify JWTs                           |
| `BASE_URL`       | yes      | Base URL used to build password reset links; server exits if unset |

\* `DB_URL_TEST` is only required when running the integration tests.

## Testing

Unit tests run without any external services:

```sh
go test ./...
```

The integration tests in `api/` (`*_integration_test.go`) hit a real
PostgreSQL database and are skipped when `DB_URL_TEST` is not set:

```sh
DB_URL_TEST=postgres://USER:PASSWORD@localhost:5432/savings_tracker_test?sslmode=disable go test ./...
```

## Continuous integration

`.github/workflows/ci.yml` runs on pull requests to `main`:

- `go test ./...` against a `postgres:16` service container
- `gofmt` format check
- `staticcheck` static analysis

## Development workflow

Migrations and queries are SQL-first:

1. Add a migration in `sql/schema/NNN_description.sql`.
2. Add/update queries in `sql/queries/*.sql`.
3. Regenerate the Go code: `sqlc generate` (config in `sqlc.yaml`).
4. Use the generated code from `internal/database` in your handlers.

See [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) for a deeper look, and
[docs/API.md](docs/API.md) for the full API reference.
