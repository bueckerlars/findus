# Development

This page is for contributors and anyone running Findus **without** Docker from source.

## Prerequisites

- **Go** 1.23 or newer
- **Node.js** 20+ (only for building Tailwind CSS in `frontend/`; not required at runtime for `go run` if `output.css` is already present)

## Common commands

The [Makefile](../Makefile) wraps the usual tasks:

| Target | Purpose |
|--------|---------|
| `make tidy` | `go mod tidy` |
| `make css` | `npm install` in `frontend/` and build minified Tailwind to `frontend/static/css/output.css` |
| `make run` | Run the app with `go run ./backend/app` |
| `make dev` | Debug log level + [Air](https://github.com/air-verse/air) hot reload (rebuild on Go/HTML/CSS/SQL changes; see `.air.toml`) |
| `make build` | Build CSS then compile static binary to `./bin/findus` |
| `make test` | `go test ./...` |
| `make docker-build` | Build container image `findus:dev` (same tag as `docker-compose.yml.dev`) |
| `make db-reset` | **Destructive**: remove local `findus.db*` and `images/` under `FINDUS_DATA_DIR` (default `./data`) |

## First-time local setup

```bash
go mod tidy
( cd frontend && npm install \
  && npx tailwindcss -i ./static/css/input.css -o ./static/css/output.css --minify )
go run ./backend/app
```

Or: `make css` then `make run`.

## Hot reload

`make dev` runs Air with `FINDUS_LOG_LEVEL=debug`. The config file `.air.toml` watches `go`, `html`, `tmpl`, `tpl`, `css`, and `sql` files and rebuilds `./tmp/findus`. Template or static changes may still need a browser refresh.

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
| `frontend/templates` | HTML templates |
| `frontend/static` | CSS and static files (Tailwind input/output) |
| `frontend/embed.go` | `go:embed` of templates and static assets |

## Docker from this repo

- **Dev stack (build):** `docker compose -f docker-compose.yml.dev up --build` — same env/volume layout as production Compose, image built locally as `findus:dev`.
- **Released image:** use root `docker-compose.yml` with `FINDUS_GHCR_IMAGE` / `FINDUS_IMAGE_TAG` (see [Configuration](configuration.md)).

See [Architecture](architecture.md) for a deeper overview.
