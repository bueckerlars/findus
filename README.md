# Findus

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Go](https://img.shields.io/badge/Go-1.23+-00ADD8?logo=go)](https://go.dev/)

**Findus** helps you remember *where things live*: nested places (room → shelf → box), items with simple templates, photos, and QR labels you can scan to jump straight to a location or item. It runs on your own machine or server—one process, one database folder, no external SaaS.

---

## Why use it

- **Hierarchy that matches real life** — Organize storage the way you actually stack and label it, not a flat list.
- **QR codes** — Print or stick a code on a box; scanning opens the right page in Findus (based on your public base URL).
- **Households or small teams** — Multiple accounts with clear roles: full access for admins, read-focused access for everyone else (see [documentation](docs/README.md)).
- **Your data stays yours** — SQLite on disk, images beside it, optional ZIP backup from the admin area.

---

## Run it (recommended)

You need [Docker](https://docs.docker.com/get-docker/) with Compose.

Pre-built images are published to **GitHub Container Registry** on each [git tag](https://docs.github.com/en/repositories/releasing-a-project-on-github/managing-releases-in-a-repository). Image reference: `ghcr.io/<github-owner>/<repo>:<tag>` (repository name is lowercased in the registry).

```bash
cp .env.example .env
# Set FINDUS_GHCR_IMAGE to your owner/repo (e.g. myorg/findus) and FINDUS_IMAGE_TAG to a release tag or latest
docker compose pull
docker compose up -d
```

For **local development** with a container built from this checkout, use the dev override file:

```bash
docker compose -f docker-compose.yml.dev up --build
```

Open **http://localhost:8080**, create the **first account** — that user becomes the **admin**. After that you can add locations, items, and photos from the web UI.

**Check that the app is up:** open http://localhost:8080/healthz — you should see `ok`.

Data is stored in a Docker volume (`findus_data` by default), mapped inside the container to `/data` (database and images).

---

## Configure for your environment

Set environment variables (or use a `.env` file next to `docker-compose.yml`). For the default Compose file you must set **`FINDUS_GHCR_IMAGE`** (and optionally **`FINDUS_IMAGE_TAG`**) so Docker can pull `ghcr.io/<owner>/<repo>:<tag>`. The most important app setting behind a reverse proxy or HTTPS is **`FINDUS_BASE_URL`** — it must match the URL people use in the browser, because QR codes embed that host.

| Variable | Typical local value | Purpose |
|----------|---------------------|---------|
| `FINDUS_GHCR_IMAGE` | *(required for default Compose)* | GitHub Container Registry path `owner/repo` (lowercase), e.g. `myorg/findus` |
| `FINDUS_IMAGE_TAG` | `latest` or e.g. `v1.2.0` | Image tag to pull (matches the git tag used for releases) |
| `FINDUS_PORT` | `8080` | HTTP port (host mapping in Compose) |
| `FINDUS_DATA_DIR` | *(Docker: `/data`)* | Where the database and `images/` live |
| `FINDUS_BASE_URL` | `http://localhost:8080` | Public URL prefix for QR links |
| `FINDUS_JWT_SECRET` | *(optional)* | Sign-in tokens; if unset, a file is created under the data directory |
| `FINDUS_COOKIE_SECURE` | `false` (`true` with HTTPS) | Secure cookie flag |
| `FINDUS_LOG_LEVEL` | `info` | `debug`, `info`, `warn`, `error` |
| `FINDUS_LOG_FORMAT` | `text` | `text` (human-readable) or `json` (structured) |

Copy [`.env.example`](.env.example) as a starting point. Full detail: [Configuration](docs/configuration.md).

---

## Backup and restore

Admins can download a **ZIP** snapshot (database + images) from the app. For restore steps and path notes, see [Backup & restore](docs/backup-restore.md).

---

## Documentation

Technical and contributor-focused material lives in the **[docs/](docs/README.md)** wiki-style pages (architecture, development workflow, routes, configuration depth).

---

## Contributing

Issues and pull requests are welcome. Before you open a PR, run `make test` (or `go test ./...`) and format with `go fmt ./...`. See [Development](docs/development.md) for the usual edit–run loop.

---

## License

MIT — see [LICENSE](LICENSE).
