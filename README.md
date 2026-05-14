<div align="center">

# 📦 Findus

**Remember where everything lives.**

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Go](https://img.shields.io/badge/Go-1.23+-00ADD8?logo=go&logoColor=white)](https://go.dev/)
[![Vue 3](https://img.shields.io/badge/Vue-3-4FC08D?logo=vue.js&logoColor=white)](https://vuejs.org/)
[![SQLite](https://img.shields.io/badge/SQLite-embedded-003B57?logo=sqlite&logoColor=white)](https://www.sqlite.org/)
[![Docker](https://img.shields.io/badge/Docker-ready-2496ED?logo=docker&logoColor=white)](https://hub.docker.com/)

Findus is a **self-hosted inventory manager** that organizes your physical world — rooms, shelves, boxes, and the items inside them — with a hierarchical structure that mirrors how real storage works.  
Scan a QR code on any box and land directly on its page. No SaaS, no subscriptions, your data stays on your machine.

[**Documentation**](docs/README.md) · [**Quick Start**](#quick-start) · [**Configuration**](#configuration) · [**Contributing**](#contributing)

</div>

---

## Why Findus?

| Problem | Findus solution |
|---|---|
| "Where did I put the camping gear?" | Nested locations: *Basement → Metal shelf → Box 3* |
| Printing labels for boxes | Built-in QR codes — stick, scan, done |
| Shared space with a partner or team | Multi-user accounts with role-based access |
| Afraid of cloud lock-in | SQLite on disk, ZIP backup from the admin panel |
| Complex self-hosted setup | Single binary, one Docker Compose file |

---

## Features

- **Hierarchical locations** — Model your actual storage: rooms contain shelves, shelves contain boxes, boxes contain items. No artificial depth limit.
- **Item templates** — Pre-defined field sets (electronics, clothing, tools…) so you only fill in what matters for each category.
- **Photos** — Attach images to locations and items; Findus auto-converts and compresses to WebP.
- **QR codes** — Every location and item gets a scannable code. Scanning redirects straight to its page using your configured base URL.
- **Full-text search** — Find anything by name, description, or custom field value (FTS5-powered, with a LIKE fallback).
- **Command palette** — Keyboard-driven search overlay for quick navigation.
- **Role-based access** — First registered user becomes admin. Admins manage users, templates, labels, and settings; regular users browse and edit inventory.
- **Registration modes** — Open, invite-only, or admin-only: you control who can sign up.
- **One-click backup** — Admins download a ZIP containing the SQLite snapshot and all images.
- **No external dependencies** — Single Go binary, SQLite, embedded Vue SPA. No Redis, no Postgres, no cloud services.

---

## Quick Start

> **Requirements:** [Docker](https://docs.docker.com/get-docker/) with the Compose plugin (v2).

### 1 — Create your configuration

```bash
cp .env.example .env
```

### 2 — Start the stack

```bash
docker compose pull
docker compose up -d
```

### 3 — Open the app

Navigate to **http://localhost:8080**.  
The **first account you register becomes the admin**. Subsequent registrations follow the mode configured in admin settings (default: open).

### Health check

```
GET http://localhost:8080/healthz  →  200 ok
```

### Stopping

```bash
docker compose down        # stop containers, keep data volume
docker compose down -v     # stop containers AND delete all data
```

Data is persisted in a Docker named volume (`findus_data`), mapped to `/data` inside the container.

---

## Configuration

All settings are environment variables. Copy [`.env.example`](.env.example) as a starting point.

| Variable | Default | Required | Description |
|---|---|---|---|
| `FINDUS_PORT` | `8080` | No | TCP port the server listens on |
| `FINDUS_DATA_DIR` | `./data` (`/data` in Docker) | No | Root directory for the SQLite database and `images/` folder |
| `FINDUS_BASE_URL` | `http://localhost:8080` | No | **Public** URL users type in the browser — embedded in every QR code |
| `FINDUS_JWT_SECRET` | *(auto-generated)* | No | HS256 signing secret. If empty, Findus creates or loads `$DATA_DIR/.jwt_secret` |
| `FINDUS_COOKIE_SECURE` | `false` | No | Set `true` when serving over HTTPS so session cookies are Secure |
| `FINDUS_LOG_LEVEL` | `info` | No | `debug` · `info` · `warn` · `error` |
| `FINDUS_LOG_FORMAT` | `text` | No | `text` (key=value) or `json` (structured — for log aggregators) |

### HTTPS / reverse proxy

1. Terminate TLS at Caddy, nginx, Traefik, etc. and forward plain HTTP to Findus on its port.
2. Set `FINDUS_BASE_URL` to the **https** URL — e.g. `https://findus.example.com`. QR codes embed this URL.
3. Set `FINDUS_COOKIE_SECURE=true` so session cookies are only sent over HTTPS.

> **Common issue:** scanning a QR code opens the wrong host or shows a browser error?  
> The most likely cause is `FINDUS_BASE_URL` pointing at the wrong scheme or hostname.

See [Configuration](docs/configuration.md) for the full reference.

---

## Deployment

### Docker Compose (recommended)

The [docker-compose.yml](docker-compose.yml) in this repo pulls `ghcr.io/bueckerlars/findus:latest` from GitHub Container Registry.

### Build from source

```bash
git clone https://github.com/your-org/findus.git && cd findus
make build          # builds frontend assets, then compiles ./bin/findus
./bin/findus        # runs on :8080 with ./data as the data directory
```

Full development workflow: [Development](docs/development.md).

---

## Backup & Restore

Admins can download a **ZIP snapshot** (SQLite database + all images) from the admin panel, or by visiting `/admin/backup.zip`.

The ZIP contains:
- `findus.db` — consistent SQLite snapshot
- `images/` — all uploaded photos

To restore: stop the container, extract the ZIP into your data directory, restart.  
Full procedure: [Backup & Restore](docs/backup-restore.md).

---

## Documentation

| Page | What it covers |
|---|---|
| [Architecture](docs/architecture.md) | Go layers, Vue SPA, SQLite, security model |
| [Configuration](docs/configuration.md) | All environment variables, HTTPS setup, Docker Compose variants |
| [Development](docs/development.md) | Local toolchain, Makefile targets, hot reload, tests |
| [Routes](docs/routes.md) | HTTP API surface (public, authenticated, admin) |
| [Backup & Restore](docs/backup-restore.md) | Backup contents and step-by-step restore |

---

## Contributing

Contributions are welcome — bug reports, feature suggestions, and pull requests alike. See [CONTRIBUTING.md](CONTRIBUTING.md) for the full guide including setup, code style, and PR process.

---

## License

MIT — see [LICENSE](LICENSE).
