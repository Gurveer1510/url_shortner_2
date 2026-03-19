# URL Shortener (Go)

Lightweight URL shortener service with optional custom codes, expiration, and click tracking. Built with Go and PostgreSQL.

## Features

- Shorten any URL; optionally supply your own `code`
- Optional `expires_at` per short link
- Click tracking with IP capture and per-code statistics
- Simple, per-IP token-bucket rate limiting
- Auto-applied SQL migrations at startup

## Tech Stack

- Go 1.24+
- PostgreSQL (pgx v5)
- Viper for configuration
- golang-migrate for migrations

## Quickstart

1) Install prerequisites
- Go 1.24+
- PostgreSQL running and reachable

2) Configure environment
Create a `.env` file in the repo root (the app loads it via Viper):

```
DATABASE_HOST=localhost:5432
DATABASE_NAME=urlshortner
DATABASE_USER=postgres
DATABASE_PASSWORD=postgres
SSL=disable
CHANNEL_BINDING=
BASE_URL=http://localhost:8080
```

3) Run the server

```
go run ./cmd/server
```

- The server listens on `:8080`.
- SQL migrations in `migrations/` apply automatically on startup. Run from the repo root so `file://migrations` resolves correctly.
- The base URL returned in responses is currently set in `cmd/server/main.go` (default `http://localhost:8080`). Update as needed for your environment.

## API

- POST `/shorten`
  - Body:
    ```json
    {
      "url": "https://example.com/very/long/path",
      "code": "myalias",         // optional; server generates if omitted
      "expires_at": "2026-12-31T00:00:00Z" // optional (RFC3339)
    }
    ```
  - Response: `{"short_url": "http://localhost:8080/Abc1234"}`
  - Errors: `400` invalid input or expired, `409` code conflict (if custom code already used)

- GET `/{code}`
  - 302 Redirects to the original URL (or `404` if not found, `400` if expired)

- GET `/stats/{code}`
  - Returns aggregated stats:
    ```json
    {
      "code": "Abc1234",
      "url": "https://example.com/very/long/path",
      "clicks": 12,
      "created_at": "2026-03-12T15:04:05Z",
      "expires_at": null,
      "is_expired": false,
      "data": [
        {"ip_address": "203.0.113.42", "created_at": "2026-03-12T15:10:02Z"}
      ]
    }
    ```

## Rate Limiting

- Configured in `cmd/server/main.go` via `middleware.NewRateLimiter(5, 10)`
  - `rate = 5` tokens/second, `capacity = 10` tokens per IP
  - Keyed by client address

## Project Structure

```
cmd/
  server/
    main.go
internal/
  auth/
    session/
      session.go
  config/
    config.go
  db/
    db.go
  domain/
    errors.go
    store.go
    types.go
  http/
    handlers/
      url/
        handlers.go
      user/
        handlers.go
    middleware/
      rateLimiter.go
      auth/
        auth.go
  repository/
    postgres/
      repository.go
  shortid/
    generator.go
migrations/
  ...
```

## Notes

- The `users` table and related code are scaffolded for future expansion (auth not yet wired to endpoints).
- If you prefer configuring the base URL via env, we can move it into `config.Config`.
- Contributions welcome — issues and PRs encouraged.
