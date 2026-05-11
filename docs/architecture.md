# Architecture

Findus is a single Go binary that serves HTML (server-rendered templates plus [HTMX](https://htmx.org/)), static assets, and a small set of JSON/binary endpoints (e.g. QR images, photos). There is no separate frontend build at runtime in production: Tailwind CSS is compiled ahead of time into `web/static/css/output.css`.

## High-level flow

```mermaid
flowchart LR
  HTTP[HTTP handlers and middleware] --> Services[Services]
  Services --> Domain[Domain model]
  Services --> Repos[Repository interfaces]
  Repos --> SQLite[SQLite implementation]
  Services --> Ports[QR, images, backup ZIP]
```

- **Transport** (`internal/transport/http`): routing, middleware (logging, recovery, auth, CSRF), and handlers that call services and render templates.
- **Services** (`internal/service`): application use cases (inventory, auth, admin, QR, backup).
- **Domain** (`internal/domain`): core types and validation-oriented structures.
- **Repositories** (`internal/repository`): persistence interfaces; **SQLite** implementations live under `internal/repository/sqlite`.
- **Supporting packages**: JWT helpers (`internal/authjwt`, `internal/secrets`), configuration (`internal/config`), logging (`internal/platform/logger`).

## Stack

| Layer | Choice |
|-------|--------|
| Language | Go 1.23 |
| HTTP | `net/http` with Go 1.22+ route patterns |
| Database | SQLite via `modernc.org/sqlite` (CGO-free); item search uses **FTS5** where available with a **LIKE** fallback |
| Migrations | [goose](https://github.com/pressly/goose) SQL, embedded under `internal/repository/sqlite/migrations` |
| HTML | `html/template`, HTMX, Tailwind CSS (build-time) |
| Images | Resize/compress pipeline; WebP storage under the data directory |

## Data on disk

Under `FINDUS_DATA_DIR` (default `./data`, `/data` in the official container):

- SQLite database file(s) for users, locations, items, settings, invites, etc.
- `images/` for processed uploads (WebP).

The JWT signing secret may be persisted as a dotfile under the data directory when `FINDUS_JWT_SECRET` is not set (see [Configuration](configuration.md)).

## Security-related behavior (summary)

- **Authentication**: JWT stored in an HTTP-only cookie after login/register.
- **CSRF**: Double-submit cookie pattern plus `X-CSRF-Token` header for mutating HTMX requests.
- **Roles**: First registered user is `admin`; RBAC distinguishes admin capabilities (full CRUD, users, settings, backup) from read-oriented `user` access to inventory views and QR/photo reads.
- **Registration modes** (admin-configurable): `admin_only`, `invite`, `open`.

For exact route groups and admin surfaces, see [Routes](routes.md).

## Composition root

`cmd/findus/main.go` loads configuration, opens the database, wires repositories into services, parses templates from the embedded `web` filesystem, and starts the HTTP server. This is the only entrypoint for the application binary.
