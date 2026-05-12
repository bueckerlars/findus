# Development

This page is for contributors and anyone running Findus **without** Docker from source.

## Prerequisites

- **Go** 1.23 or newer
- **Node.js** 20+ and npm (for `frontend/` — Vite + Vue + Tailwind). A production binary embeds the built `frontend/dist` assets; you need a client build before `go build` succeeds.

## Common commands

The [Makefile](../Makefile) wraps the usual tasks:

| Target | Purpose |
|--------|---------|
| `make tidy` | `go mod tidy` |
| `make frontend-dist` | `npm install` and `npm run build` in `frontend/` (writes `frontend/dist/`) |
| `make run` | Run the app with `go run ./backend/app` (expects `frontend/dist` to exist) |
| `make dev` | Debug log level + [Air](https://github.com/air-verse/air) hot reload (rebuild on Go/Vue/TS/CSS/SQL changes; see `.air.toml`) |
| `make build` | Client build then compile static binary to `./bin/findus` |
| `make test` | `go test ./...` |
| `make docker-build` | Build container image `findus:dev` (same tag as `docker-compose.yml.dev`) |
| `make db-reset` | **Destructive**: remove local `findus.db*` and `images/` under `FINDUS_DATA_DIR` (default `./data`) |

## First-time local setup

```bash
go mod tidy
make frontend-dist
go run ./backend/app
```

Or: `make build` then run `./bin/findus`.

## Hot reload

`make dev` runs Air with `FINDUS_LOG_LEVEL=debug`. The config file `.air.toml` watches Go sources plus `vue` and `ts` under the repo. After changing the SPA, either let Air rebuild the Go binary (embed picks up `frontend/dist`) or run `make frontend-dist` manually if you skipped the client build.

## Tests and formatting

```bash
make test
go fmt ./...
```

## Repository layout (short)

| Path | Role |
|------|------|
| `backend/app` | `main`, wiring |
| `backend/internal/config` | Environment-based configuration |
| `backend/internal/domain` | Domain types |
| `backend/internal/repository` | Persistence interfaces and SQLite |
| `backend/internal/service` | Application services |
| `backend/internal/transport/http` | HTTP server, handlers, middleware |
| `frontend/src` | Vue 3 SPA (Vite, Vue Router, Tailwind) |
| `frontend/dist` | Production build output (embedded into the Go binary) |
| `frontend/embed.go` | `go:embed all:dist` |

## Docker from this repo

- **Dev stack (build):** `docker compose -f docker-compose.yml.dev up --build` — same env/volume layout as production Compose, image built locally as `findus:dev`.
- **Released image:** use root `docker-compose.yml` with `FINDUS_GHCR_IMAGE` / `FINDUS_IMAGE_TAG` (see [Configuration](configuration.md)).

See [Architecture](architecture.md) for a deeper overview.
