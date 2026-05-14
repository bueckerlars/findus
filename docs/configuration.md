# Configuration

All Findus settings are read from **environment variables** at startup. You can supply them directly, via a `.env` file beside `docker-compose.yml`, or through your container orchestrator's secret management.

Copy [`.env.example`](../.env.example) as a starting point.

---

## Environment variables

| Variable | Default | Description |
|---|---|---|
| `FINDUS_PORT` | `8080` | TCP port the HTTP server listens on |
| `FINDUS_DATA_DIR` | `./data` | Root directory for the SQLite database and `images/` folder. Set to `/data` inside the official container. |
| `FINDUS_BASE_URL` | `http://localhost:8080` | **Canonical public URL** — no trailing slash. Embedded verbatim in every QR code as `{BASE_URL}/q/{token}`. Must match what users type in the browser. |
| `FINDUS_JWT_SECRET` | *(empty)* | HS256 signing secret for session JWTs. If empty, Findus auto-generates a secret and persists it as `$DATA_DIR/.jwt_secret`. Rotating this value invalidates all active sessions. |
| `FINDUS_COOKIE_SECURE` | `false` | Set `true` to mark session cookies as `Secure`. Required for HTTPS deployments; breaks login on plain HTTP. |
| `FINDUS_LOG_LEVEL` | `info` | Log verbosity: `debug`, `info`, `warn`, or `error` |
| `FINDUS_LOG_FORMAT` | `text` | `text` — human-readable `key=value` lines. `json` — structured JSON (recommended for log aggregators and Docker logging drivers). |

---

## HTTPS and reverse proxies

Findus does not handle TLS itself. Terminate HTTPS at a reverse proxy and forward plain HTTP:

**Step 1** — Configure your proxy (Caddy, nginx, Traefik…) to forward to `http://localhost:8080` (or the container's hostname/port).

**Step 2** — Set `FINDUS_BASE_URL` to the **https** URL:

```env
FINDUS_BASE_URL=https://findus.example.com
```

**Step 3** — Enable the Secure cookie flag:

```env
FINDUS_COOKIE_SECURE=true
```

> **Why `FINDUS_BASE_URL` matters:** every QR code encodes the URL `{FINDUS_BASE_URL}/q/{token}`. If this is wrong, scanning a code opens the wrong host or scheme. This is the most common misconfiguration for LAN or HTTPS deployments.

### Subpath deployment

If you mount Findus at a subpath (e.g. `https://example.com/findus/`), set `FINDUS_BASE_URL` to include that path and ensure your reverse proxy strips or preserves it consistently. The SPA router and QR resolver both rely on this prefix.

---

## JWT secret handling

| Scenario | Behavior |
|---|---|
| `FINDUS_JWT_SECRET` is set | Used directly as the HS256 signing key |
| `FINDUS_JWT_SECRET` is empty | On first startup, a random 32-byte secret is generated and written to `$DATA_DIR/.jwt_secret`. Subsequent startups load it from that file. |

**Backup note:** if you rely on the auto-generated secret, include `$DATA_DIR/.jwt_secret` in your backup process. Restoring the database without restoring the secret file will invalidate all existing sessions (users can simply log in again — no data is lost).

---

## Docker Compose variants

### Production (`docker-compose.yml`)

Pulls `ghcr.io/bueckerlars/findus:latest` from GitHub Container Registry.

Includes a one-shot `findus-data-init` service that `chown`s the volume to `65532:65532` (the nonroot user in the distroless image) before the app container starts.

### Development (`docker-compose.yml.dev`)

Builds the image from `Dockerfile` in this checkout. Useful for testing changes before tagging a release:

```bash
docker compose -f docker-compose.yml.dev up --build
```

Both Compose files mount a named Docker volume at `/data` and use the same environment variable names.

---

## Minimal production `.env`

```env
# Public URL — must match what users type in the browser
FINDUS_BASE_URL=https://findus.example.com

# Required for HTTPS
FINDUS_COOKIE_SECURE=true

# Recommended: set an explicit secret so it survives container recreations
FINDUS_JWT_SECRET=change-me-to-a-long-random-string

# Structured logs for Docker log driver / aggregator
FINDUS_LOG_FORMAT=json
```
