# Development

This guide is for contributors and anyone running Findus **from source** without Docker.

---

## Prerequisites

| Tool | Minimum version | Purpose |
|---|---|---|
| [Go](https://go.dev/dl/) | 1.23 | Backend compiler and toolchain |
| [Node.js](https://nodejs.org/) | 20 | Frontend build (Vite + Vue + Tailwind) |
| npm | bundled with Node | Frontend package manager |

> A production binary embeds the pre-built `frontend/dist/`. You must run the frontend build at least once before `go build` or `go run` will succeed.

---

## First-time setup

```bash
git clone https://github.com/your-org/findus.git
cd findus

go mod tidy
make frontend-dist    # npm install + npm run build in frontend/

go run ./backend/app  # starts on :8080 with ./data as the data directory
```

Or build a standalone binary:

```bash
make build            # produces ./bin/findus
./bin/findus
```

---

## Makefile reference

All common tasks are wrapped in the [Makefile](../Makefile):

| Target | Command(s) run | Notes |
|---|---|---|
| `make tidy` | `go mod tidy` | Sync go.sum |
| `make frontend-dist` | `npm install && npm run build` (in `frontend/`) | Writes `frontend/dist/` — required before any Go build |
| `make run` | `go run ./backend/app` | Expects `frontend/dist/` to exist |
| `make dev` | Air hot reload with `FINDUS_LOG_LEVEL=debug` | See [Hot reload](#hot-reload) below |
| `make build` | `make frontend-dist` then `go build -o ./bin/findus` | Static binary (`CGO_ENABLED=0`) |
| `make test` | `go test ./...` | Run all backend tests |
| `make lint` | `go vet ./...` + golangci-lint | Static analysis |
| `make format` | `go fmt ./...` | Format all Go sources |
| `make docker-build` | `docker build -t findus:dev .` | Build local container image |
| `make db-reset` | Remove `findus.db*` and `images/` from `FINDUS_DATA_DIR` | **Destructive** — deletes all local data |
| `make hooks-install` | `lefthook install` | Install git pre-commit hooks |
| `make hooks-run` | `lefthook run pre-commit --all-files` | Run the same checks as the pre-commit hook |

---

## Hot reload

`make dev` starts [Air](https://github.com/air-verse/air) configured via [`.air.toml`](../.air.toml). Air watches Go sources, Vue/TS files, and SQL migrations, then rebuilds and restarts the binary automatically.

```bash
make dev
# Visit http://localhost:8080 — changes to backend code trigger a rebuild in ~1s.
# For frontend changes, run `make frontend-dist` (or `cd frontend && npm run build`) in a separate terminal.
```

Air re-embeds `frontend/dist/` on each Go rebuild, so you only need to rebuild the frontend manually when you change Vue/Tailwind sources.

---

## Tests

```bash
make test          # equivalent to go test ./...
```

Tests are integration-style where practical — the SQLite layer uses an in-memory database, not mocks. Run `make lint` to catch issues before pushing.

---

## Git hooks

Pre-commit hooks run `gofmt` and `golangci-lint` automatically:

```bash
make hooks-install    # one-time setup
```

To run the same checks manually:

```bash
make hooks-run
```

---

## Docker (from this repo)

For contributors testing the containerized build:

```bash
# Build a local image and start the full stack
docker compose -f docker-compose.yml.dev up --build
```

This uses the same environment and volume layout as the production Compose file, but builds the image from `Dockerfile` in this checkout (tagged `findus:dev`).

---

## Repository layout

```
findus/
├── backend/
│   ├── app/                  main.go — composition root
│   └── internal/
│       ├── config/           environment config struct
│       ├── domain/           core types (no I/O)
│       ├── repository/       persistence interfaces
│       │   └── sqlite/       SQLite implementations + migrations/
│       ├── search/           FTS5 + LIKE fallback search
│       ├── service/          application use-cases
│       ├── transport/http/   HTTP server, handlers, middleware
│       ├── authjwt/          JWT helpers
│       ├── secrets/          JWT secret bootstrap
│       └── platform/logger/  structured logger
├── frontend/
│   ├── src/                  Vue 3 SPA (views, components, composables)
│   ├── dist/                 Vite build output (embedded into binary)
│   └── embed.go              go:embed directive
├── docs/                     technical documentation (this directory)
├── Dockerfile
├── docker-compose.yml        production — pulls from GHCR
├── docker-compose.yml.dev    development — builds from Dockerfile
├── Makefile
└── .air.toml                 Air hot-reload config
```
