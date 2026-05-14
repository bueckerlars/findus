# Architecture

Findus is a **single Go binary** that embeds the Vue 3 SPA at compile time. The same process handles the JSON API, static file serving, binary endpoints (photos, QR codes, backup), and database access. There are no external runtime dependencies — no separate process, message queue, or cache service.

---

## High-level request flow

```
Browser / QR scanner
       │
       ▼
 ┌─────────────────────────────────────────────────────────┐
 │  HTTP server  (net/http, Go 1.22+ patterns)             │
 │                                                         │
 │  Middleware chain:                                       │
 │    logger → recovery → CORS → CSRF → auth (JWT)        │
 │                                                         │
 │  Route groups:                                          │
 │    /api/*        JSON handlers                          │
 │    /q/{token}    QR resolver (redirect)                 │
 │    /assets/*     Hashed JS/CSS from embedded dist/      │
 │    /{path...}    SPA shell (index.html fallthrough)     │
 └────────────────────────┬────────────────────────────────┘
                          │
                          ▼
              ┌───────────────────────┐
              │       Services        │
              │  inventory · auth     │
              │  admin · QR · backup  │
              └──────────┬────────────┘
                         │
          ┌──────────────┼──────────────┐
          ▼              ▼              ▼
     Repository      QR library    Image pipeline
     interfaces      (go-qrcode)   (imaging + WebP)
          │
          ▼
     SQLite (modernc.org/sqlite, CGO-free)
     + goose migrations (SQL, embedded)
```

---

## Package layout

| Package | Path | Responsibility |
|---|---|---|
| **config** | `backend/internal/config` | Reads environment variables into a typed struct |
| **domain** | `backend/internal/domain` | Core types, validation structures — no I/O |
| **repository** | `backend/internal/repository` | Persistence interfaces |
| **repository/sqlite** | `backend/internal/repository/sqlite` | SQLite implementations; migrations in `migrations/` |
| **search** | `backend/internal/search` | FTS5 full-text search with LIKE fallback |
| **service** | `backend/internal/service` | Application use-cases (inventory, auth, admin, QR, backup) |
| **transport/http** | `backend/internal/transport/http` | HTTP server, route registration, middleware, JSON handlers |
| **authjwt** | `backend/internal/authjwt` | JWT creation, parsing, cookie helpers |
| **secrets** | `backend/internal/secrets` | JWT secret bootstrap (env var or auto-generated file) |
| **platform/logger** | `backend/internal/platform/logger` | Structured logger (text/JSON) |
| **frontend** | `frontend/embed.go` | `go:embed all:dist` — bundles the Vite build into the binary |
| **main** | `backend/app/main.go` | Composition root: load config → open DB → wire services → start server |

---

## Storage

All data lives under `FINDUS_DATA_DIR` (default `./data`, `/data` inside the official container):

```
$FINDUS_DATA_DIR/
├── findus.db          SQLite database
├── findus.db-wal      Write-ahead log (auto-managed)
├── findus.db-shm      Shared memory (auto-managed)
├── images/            Uploaded photos, stored as WebP
└── .jwt_secret        Auto-generated JWT signing key (only when FINDUS_JWT_SECRET is unset)
```

The database is opened with WAL mode enabled. All schema changes go through goose migrations embedded in the binary — no manual schema management is needed on startup or upgrade.

---

## Frontend

The Vue 3 SPA is built with Vite and bundled into `frontend/dist/`. The Go binary embeds the entire `dist/` tree via `go:embed`. Tailwind is applied at build time — the running server has no Node.js dependency.

At runtime, the HTTP server:
1. Serves hashed assets from `/assets/*` (cache-safe).
2. Falls through all unmatched GET requests to `dist/index.html` so Vue Router handles client-side navigation.

---

## Authentication & security

### Authentication

Users authenticate with a username and password. On success, the server issues a **JWT** stored in an **HTTP-only cookie** (`findus_session`). The cookie is `SameSite=Lax`; set `FINDUS_COOKIE_SECURE=true` under HTTPS.

The JWT secret is loaded from `FINDUS_JWT_SECRET` or auto-generated and stored in `$DATA_DIR/.jwt_secret` on first run.

### CSRF protection

Mutating requests (`POST`, `PUT`, `PATCH`, `DELETE`) require the **`X-CSRF-Token`** header. The server issues a CSRF token as a non-HttpOnly cookie (`findus_csrf`); the Vue app reads it and echoes it in the header. The server validates the header matches the cookie (double-submit pattern).

### Roles and registration

| Role | Capabilities |
|---|---|
| `admin` | Full CRUD on all resources, user management, settings, templates, backup |
| `user` | Browse and search inventory, view QR codes and photos, edit own profile |

The **first registered account** is automatically assigned the `admin` role. Subsequent registrations follow the mode configured in admin settings:

| Mode | Behavior |
|---|---|
| `open` | Anyone with the URL can register |
| `invite` | Registration requires a valid invite token generated by an admin |
| `admin_only` | No self-service registration; admin creates accounts directly |
