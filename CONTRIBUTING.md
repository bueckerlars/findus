# Contributing to Findus

Thank you for taking the time to contribute. This document covers everything you need to go from zero to a mergeable pull request.

---

## Table of contents

- [Getting started](#getting-started)
- [Development workflow](#development-workflow)
- [Code style](#code-style)
- [Commit messages](#commit-messages)
- [Opening a pull request](#opening-a-pull-request)
- [Reporting bugs](#reporting-bugs)
- [Suggesting features](#suggesting-features)

---

## Getting started

### Prerequisites

| Tool | Minimum version |
|---|---|
| Go | 1.23 |
| Node.js | 20 |
| npm | bundled with Node |
| Docker + Compose | v2 (optional, for container testing) |

### First-time setup

```bash
git clone https://github.com/bueckerlars/Findus.git
cd Findus

go mod tidy
make frontend-dist   # build the Vue SPA (required before any Go build)
make hooks-install   # install git pre-commit hooks
```

Start the development server with hot reload:

```bash
make dev             # starts on http://localhost:8080 with debug logs
```

The backend rebuilds automatically when you change Go or SQL files. For frontend changes, run `make frontend-dist` (or `cd frontend && npm run build`) in a separate terminal — Air picks up the new `dist/` on the next Go rebuild.

See [docs/development.md](docs/development.md) for the full Makefile reference and repository layout.

---

## Development workflow

### 1 — Create a branch

```bash
git checkout -b feat/short-description   # new feature
git checkout -b fix/short-description    # bug fix
```

### 2 — Make your changes

- Backend code lives in `backend/internal/`.
- Vue SPA lives in `frontend/src/`.
- Database migrations go in `backend/internal/repository/sqlite/migrations/` as numbered SQL files managed by [goose](https://github.com/pressly/goose).

### 3 — Run tests and linting

```bash
make test      # go test ./...
make lint      # go vet + golangci-lint
go fmt ./...   # format all Go sources
```

All three must pass before opening a PR. The pre-commit hooks run `gofmt` and `golangci-lint` automatically on staged files — install them once with `make hooks-install`.

### 4 — Reset local data if needed

```bash
make db-reset   # removes findus.db* and images/ from ./data — destructive
```

---

## Code style

- **Go:** follow standard Go conventions (`gofmt`, `go vet`). golangci-lint enforces additional rules — run `make lint` to check locally.
- **Vue / TypeScript:** follow the existing patterns in `frontend/src/`. Vite + ESLint handle formatting on build.
- **SQL migrations:** one file per schema change, numbered sequentially (`00XX_description.sql`). Never edit an existing migration that has already been committed — add a new one.
- **Comments:** only when the *why* is non-obvious. Don't describe what the code does; name your variables and functions clearly instead.
- **No new dependencies** without discussion in an issue first — especially on the backend, keeping the binary small and CGO-free matters.

---

## Commit messages

Use the conventional commits format:

```
<type>: <short summary in present tense>

Optional longer explanation if needed.
```

Common types: `feat`, `fix`, `refactor`, `docs`, `test`, `chore`.

Examples:

```
feat: add invite-based registration mode
fix: correct QR redirect for items with special characters in name
docs: expand backup restore procedure
```

Keep the subject line under 72 characters. No trailing period.

---

## Opening a pull request

1. Push your branch and open a PR against `main`.
2. Fill in the PR description — what changed and why.
3. Make sure `make test` and `make lint` pass locally before requesting a review.
4. Keep PRs focused: one logical change per PR. If you spot something unrelated worth fixing, open a separate PR or issue.

---

## Reporting bugs

Open an issue and include:

- What you did (steps to reproduce)
- What you expected
- What actually happened
- Findus version or commit hash
- Relevant log output (`FINDUS_LOG_LEVEL=debug`)

---

## Suggesting features

Open an issue describing the use case — what problem you're trying to solve and why the current behaviour doesn't address it. Implementation ideas are welcome but not required.
