# Project Blueprint

## Overview

This repository is a Go URL shortener service backed by PostgreSQL. It exposes HTTP endpoints for:

- user signup
- user login
- authenticated URL shortening
- public redirect by short code
- public stats lookup by short code

The application follows a lightweight layered structure:

1. `cmd/server/main.go` wires dependencies and starts the HTTP server.
2. `internal/http/handlers/...` translates HTTP requests into use-case calls.
3. `internal/usecase/...` contains business rules.
4. `internal/repository/postgres/...` persists and reads data from PostgreSQL.
5. `internal/domain/...` defines the shared contracts and data shapes.

There are no tests in the repository today, and sessions are stored in memory rather than in the database.

## High-Level Runtime Flow

At startup the server:

1. Loads configuration from `.env` and environment variables with Viper.
2. Builds a Postgres DSN.
3. Opens a pgx connection pool.
4. Pings the database.
5. Runs SQL migrations from `migrations/`.
6. Creates the in-memory session store.
7. Creates auth middleware, repositories, use cases, and handlers.
8. Registers routes on a `http.ServeMux`.
9. Wraps the mux with a per-IP token-bucket rate limiter.
10. Starts listening on `:8080`.

## Directory Map

### `cmd/`

#### `cmd/server/main.go`

The application entry point.

- loads config with `config.LoadConfig()`
- builds the DB pool with `db.NewPool()`
- constructs session/auth components
- creates URL and user repositories
- creates URL and user use cases
- creates HTTP handlers
- registers all routes
- enables the rate limiter
- starts `http.ListenAndServe(":8080", ...)`

## `internal/`

### `internal/config/`

#### `internal/config/config.go`

Defines the `Config` struct and `LoadConfig()`.

Expected settings:

- `DATABASE_HOST`
- `DATABASE_NAME`
- `DATABASE_USER`
- `DATABASE_PASSWORD`
- `SSL`
- `CHANNEL_BINDING`
- `BASE_URL`

Notes:

- the code expects a `.env` file in the repo root
- `viper.AutomaticEnv()` also allows process environment overrides
- config loading fails if `.env` cannot be read

### `internal/db/`

#### `internal/db/db.go`

Responsible for database setup.

- `DSN(conf)` formats the Postgres connection string
- `NewPool(ctx, dsn)` creates and validates the pgx pool
- `runMigrations(dsn)` applies migrations with `golang-migrate`

Important operational detail:

- migrations are loaded from `file://migrations`, so the server must run from the repository root or another location where that relative path resolves correctly

### `internal/domain/`

This package contains the shared core types and interfaces used across handlers, use cases, and repositories.

#### `internal/domain/types.go`

Defines request/response and entity shapes:

- `UrlReq`
  - used both for shorten input and repository reads
  - fields: `Url`, `Code`, `ExpiresAt`
- `StatsResp`
  - stats payload returned by `/stats/{code}`
  - includes summary fields and click detail records
- `Click`
  - one redirect event with IP and timestamp
- `User`
  - signup payload and repository entity
- `LoginReq`
  - login payload

#### `internal/domain/errors.go`

Defines shared sentinel errors:

- `ErrNotFound`
- `ErrConflict`
- `ErrExpiredCode`

These are used to coordinate behavior between repository, use case, and handler layers.

#### `internal/domain/url_store.go`

Defines the `UrlStore` interface implemented by the Postgres URL repository:

- `Save`
- `Get`
- `SaveClick`
- `GetStats`

#### `internal/domain/user_store.go`

Defines the `UserStore` interface implemented by the Postgres user repository:

- `CreateUser`
- `GetUser`

### `internal/repository/postgres/`

This package is the persistence layer.

#### `internal/repository/postgres/repository.go`

Implements URL persistence in `URLRepository`.

Key methods:

- `Save(ctx, userId, urlReq)`
  - inserts into `links`
  - stores `code`, `url`, `expires_at`, and `user_id`
  - maps Postgres unique-constraint error `23505` to `domain.ErrConflict`
- `Get(ctx, code)`
  - updates `links.clicks = clicks + 1`
  - returns `url`, `code`, `expires_at`
  - treats missing rows as `domain.ErrNotFound`
- `SaveClick(ctx, ipAddress, code)`
  - inserts one event into `url_clicks`
- `GetStats(ctx, code)`
  - reads the link row
  - computes `IsExpired` in Go
  - loads click history from `url_clicks`

Important implementation detail:

- redirect lookup increments the aggregate click counter before the use case checks whether the link is expired

#### `internal/repository/postgres/user_repository.go`

Implements user persistence in `UserRepository`.

Key methods:

- `CreateUser(ctx, u)`
  - inserts into `users(name, email, hashpass)`
  - returns the created user ID
- `GetUser(ctx, email)`
  - loads the user row by email

### `internal/usecase/`

This package contains application rules above the repository layer.

#### `internal/usecase/url/usecase.go`

`Usecase` is the URL-shortening business layer.

Methods:

- `Shorten(ctx, userId, urlReq)`
  - rejects `expires_at` values in the past
  - accepts a user-supplied code if present
  - otherwise generates a random 7-character code via `internal/shortid`
  - retries generation up to 5 times
- `Get(ctx, ipAddress, code)`
  - loads the link
  - rejects expired links
  - records click details in `url_clicks`
  - returns the destination URL
- `GetStats(ctx, code)`
  - thin pass-through to the repository

Behavior notes:

- generated codes are alphanumeric and 7 characters long
- the retry loop intends to recover from collisions
- user ownership is passed into storage on create, but there is no read path scoped by owner

#### `internal/usecase/user/usecase.go`

`UserUsecase` handles password management.

Methods:

- `CreateUser(ctx, user)`
  - hashes the password with bcrypt
  - stores the hashed password via the repository
- `VerifyPassword(ctx, loginReq)`
  - loads the user by email
  - compares the bcrypt hash
  - returns the user ID on success

### `internal/http/handlers/`

These packages are the HTTP adapters.

#### `internal/http/handlers/url/handlers.go`

`Handlers` serves the URL endpoints.

Methods:

- `Shorten`
  - requires session data in request context
  - decodes JSON into `domain.UrlReq`
  - validates that `url` is present
  - validates URL syntax with `url.ParseRequestURI`
  - calls the URL use case
  - returns `{"short_url":"<base>/<code>"}`
- `Redirect`
  - extracts the code from the path
  - resolves client IP
  - calls the URL use case
  - returns an HTTP 302 redirect
- `GetStats`
  - reads `{code}` from the route
  - returns the `StatsResp` payload as JSON

Helper:

- `GetIP(r)`
  - prefers `X-Forwarded-For`
  - then `X-Real-IP`
  - then falls back to `RemoteAddr`

#### `internal/http/handlers/user/handlers.go`

`UserHandler` serves authentication-related endpoints.

Methods:

- `CreateUserHandler`
  - decodes signup JSON into `domain.User`
  - creates the user via the user use case
  - creates an in-memory session
  - sets two cookies: `cookie` and `user-id`
- `LoginHandler`
  - decodes login JSON into `domain.LoginReq`
  - verifies credentials
  - creates an in-memory session
  - sets the same cookies

Cookie behavior:

- session ID is stored in the `cookie` cookie
- user ID is also sent as `user-id`
- the session cookie is `HttpOnly`, `Secure`, `Path=/`, `SameSite=Lax`

### `internal/http/middleware/`

#### `internal/http/middleware/rateLimiter.go`

Implements a local in-process token-bucket rate limiter.

Components:

- `TokenBucket`
  - keeps `capacity`, `tokens`, `refillRate`, and `lastRefill`
- `RateLimiter`
  - stores a bucket per client key in a map
  - creates buckets lazily
  - exposes `Middleware(next)`

Current wiring in `main.go`:

- rate: `5`
- capacity: `10`

Current keying behavior:

- middleware uses `r.RemoteAddr` directly, not the forwarded-IP helper used by redirect handling

#### `internal/http/middleware/auth/auth.go`

Implements session-based authentication middleware.

Behavior:

- reads the `cookie` cookie
- looks up the session ID in the in-memory session store
- rejects missing or unknown sessions with `401 Unauthorized`
- injects the session into request context under the `"session"` key

### `internal/auth/session/`

#### `internal/auth/session/session.go`

Implements a memory-backed session store.

Core parts:

- `Session`
  - `UserId`
  - `Id`
  - `UserEmail`
  - `CreatedAt`
  - `ExpiresAt`
- `SessionStore`
  - map of session ID to session
  - guarded by an `RWMutex`

Methods:

- `NewSessionStore()`
  - allocates the map
  - starts a cleanup goroutine
- `Create(userId, email)`
  - generates a random 32-byte hex session ID
  - stores a session expiring in 7 days
- `Delete(id)`
  - removes a session
- `Get(id)`
  - returns the session if found and unexpired
- `cleanUpLoop()`
  - periodically removes expired sessions

Operational implication:

- all sessions are lost on process restart
- sessions are not shared across multiple server instances

### `internal/shortid/`

#### `internal/shortid/generator.go`

Contains the random short-code generator.

- alphabet: lowercase letters, uppercase letters, digits
- length: 7
- randomness source: `crypto/rand`

## Routing Surface

Routes are registered in `cmd/server/main.go`.

### `POST /signup`

Handled by `UserHandler.CreateUserHandler`.

Purpose:

- create a user account
- issue an in-memory session immediately

### `POST /login`

Handled by `UserHandler.LoginHandler`.

Purpose:

- verify credentials
- issue an in-memory session

### `POST /shorten`

Handled by `Handlers.Shorten`.

Middleware:

- auth middleware
- rate limiter

Purpose:

- create a short URL for the authenticated user

Expected payload:

```json
{
  "url": "https://example.com",
  "code": "optionalCustomCode",
  "expires_at": "2026-12-31T00:00:00Z"
}
```

### `GET /{code}`

Handled by `Handlers.Redirect`.

Purpose:

- increment aggregate click count
- record click details
- redirect to the original URL

### `GET /stats/{code}`

Handled by `Handlers.GetStats`.

Purpose:

- fetch URL metadata and click history

## Database Blueprint

Migrations live in `migrations/` and are applied automatically on startup.

### Migration Order

#### `migrations/000001_create_urls.up.sql`

Creates `links`:

- `code TEXT PRIMARY KEY`
- `url TEXT NOT NULL`
- `clicks INT NOT NULL DEFAULT 0`
- `created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`

#### `migrations/000002_create_url_clicks.up.sql`

Creates `url_clicks`:

- `id BIGSERIAL PRIMARY KEY`
- `code TEXT NOT NULL REFERENCES links(code) ON DELETE CASCADE`
- `created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`
- `ip_address TEXT`

Also creates:

- `idx_url_clicks_code`

#### `migrations/000003_add_expires_at_column.up.sql`

Adds nullable `expires_at TIMESTAMPTZ` to `links`.

#### `migrations/000004_create_user_link_join.up.sql`

Creates `users`:

- `id SERIAL PRIMARY KEY`
- `name TEXT NOT NULL`
- `email TEXT NOT NULL UNIQUE`
- `hashpass TEXT NOT NULL`
- `created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`

Then adds to `links`:

- `user_id INTEGER REFERENCES users(id) ON DELETE CASCADE`

### Effective Schema Relationships

- one `users` row can own many `links`
- one `links` row can have many `url_clicks`
- deleting a link deletes its click records
- deleting a user deletes the owned links, which also deletes dependent click rows through cascade behavior

## End-to-End Request Flows

### Signup Flow

1. Client sends JSON to `POST /signup`.
2. Handler decodes the request body into `domain.User`.
3. User use case hashes the plaintext password with bcrypt.
4. User repository inserts the row into `users`.
5. Session store creates an in-memory session.
6. Handler sets cookies on the response.

### Login Flow

1. Client sends JSON to `POST /login`.
2. Handler decodes into `domain.LoginReq`.
3. User use case loads the user by email.
4. bcrypt compares the stored hash with the supplied password.
5. Session store creates an in-memory session.
6. Handler sets cookies on the response.

### Shorten Flow

1. Client sends JSON to `POST /shorten`.
2. Auth middleware checks the `cookie` cookie and loads the in-memory session.
3. Handler reads the authenticated user ID from context.
4. Handler validates request JSON and URL syntax.
5. URL use case validates `expires_at`.
6. If `code` is empty, the use case generates a random code.
7. Repository inserts the link into `links`.
8. Handler returns the final shortened URL using `BASE_URL`.

### Redirect Flow

1. Client requests `GET /{code}`.
2. Handler extracts the short code.
3. Repository updates `links.clicks = clicks + 1` and returns the link.
4. Use case checks whether the link is expired.
5. Use case stores the click event in `url_clicks`.
6. Handler issues `302 Found` to the destination URL.

### Stats Flow

1. Client requests `GET /stats/{code}`.
2. Handler reads the code from the route.
3. Repository loads the link row.
4. Repository loads all click rows ordered by `created_at`.
5. Handler returns the combined `StatsResp` JSON.

## Dependency Summary

Main external libraries and how they are used:

- `github.com/jackc/pgx/v5`
  - Postgres connection pooling and queries
- `github.com/golang-migrate/migrate/v4`
  - schema migration runner
- `github.com/spf13/viper`
  - config loading from `.env` and environment variables
- `golang.org/x/crypto/bcrypt`
  - password hashing and verification

## Current Design Characteristics

### Strengths

- clear separation between HTTP, business logic, and persistence
- schema migrations are versioned and auto-applied
- redirect stats combine aggregate counters with raw click event history
- random short-code generation uses `crypto/rand`

### Constraints

- auth is stateful and process-local because sessions live only in memory
- there is no logout endpoint
- there is no authorization layer around stats or link ownership
- configuration is minimal and fixed around a single `:8080` listener
- rate limiting is process-local and non-distributed

## Important Caveats In The Current Code

These are not documentation issues; they are implementation details worth knowing before modifying the project.

### Collision retry logic is probably incorrect

In `internal/usecase/url/usecase.go`, generated-code retries only continue when the repository returns `domain.ErrNotFound`, but the save path returns `domain.ErrConflict` for duplicate codes. That means a collision would currently fail early instead of retrying.

### Redirect increments `clicks` before expiration is checked

`URLRepository.Get()` increments the aggregate counter as part of fetching the URL. The expiration check happens later in the use case, so an expired link request can still increase the stored click count.

### Stats handler does not map not-found cleanly

`GetStats` returns `500` for repository errors, including missing codes, rather than turning `domain.ErrNotFound` into `404`.

### Session cleanup has a locking bug

`SessionStore.cleanUpLoop()` takes the write lock and then calls `s.Delete(id)`, which tries to take the same write lock again. That can deadlock once cleanup runs.

### Session cookie is always `Secure`

The login and signup handlers set `Secure: true`, which means browsers will not send the session cookie over plain HTTP during local development unless the app is behind HTTPS.

### Rate limiting keys by `RemoteAddr`

The rate limiter uses the raw `host:port` remote address, which can create separate buckets for the same client across different source ports and does not align with the proxy-aware IP extraction used elsewhere.

### Logging leaks sensitive data

`internal/usecase/user/usecase.go` logs the stored hash and incoming password on bcrypt comparison failure. That should be treated as sensitive.

## Suggested Mental Model For Future Work

When changing the project, treat it as five core slices:

1. bootstrapping and wiring
2. transport and middleware
3. business rules
4. persistence and schema
5. session/auth behavior

A feature usually touches those layers in this order:

1. add or update domain types/interfaces
2. update repository queries or migrations
3. adjust use-case rules
4. expose or adapt the HTTP handler
5. wire it in `main.go`

## File Inventory

For quick reference, these are the tracked project files that make up the application codebase:

- `cmd/server/main.go`
- `internal/auth/session/session.go`
- `internal/config/config.go`
- `internal/db/db.go`
- `internal/domain/errors.go`
- `internal/domain/types.go`
- `internal/domain/url_store.go`
- `internal/domain/user_store.go`
- `internal/http/handlers/url/handlers.go`
- `internal/http/handlers/user/handlers.go`
- `internal/http/middleware/rateLimiter.go`
- `internal/http/middleware/auth/auth.go`
- `internal/repository/postgres/repository.go`
- `internal/repository/postgres/user_repository.go`
- `internal/shortid/generator.go`
- `internal/usecase/url/usecase.go`
- `internal/usecase/user/usecase.go`
- `migrations/000001_create_urls.up.sql`
- `migrations/000001_create_urls.down.sql`
- `migrations/000002_create_url_clicks.up.sql`
- `migrations/000002_create_url_clicks.down.sql`
- `migrations/000003_add_expires_at_column.up.sql`
- `migrations/000003_add_expires_at_column.down.sql`
- `migrations/000004_create_user_link_join.up.sql`
- `migrations/000004_create_user_link_join.down.sql`
- `README.md`
- `go.mod`
- `go.sum`

## Bottom Line

This codebase is a small layered Go service with a straightforward request path and a simple schema. The two main domains are link management and user/session management. The URL feature is functional end to end, while auth is intentionally minimal and currently implemented as in-memory session state attached to HTTP cookies.
