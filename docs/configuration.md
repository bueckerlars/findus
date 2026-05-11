# Configuration

All settings are read from environment variables (optional `.env` for local or Compose overrides).

## Environment variables

| Variable | Default | Description |
|----------|---------|-------------|
| `FINDUS_PORT` | `8080` | TCP port the server listens on |
| `FINDUS_DATA_DIR` | `./data` | Root directory for SQLite and `images/` |
| `FINDUS_BASE_URL` | `http://localhost:8080` | Canonical public origin (no trailing slash); embedded in QR URLs as `{BASE_URL}/q/{token}` |
| `FINDUS_JWT_SECRET` | *(empty)* | HS256 secret for JWTs. If empty, the app creates or loads `$DATA_DIR/.jwt_secret` |
| `FINDUS_COOKIE_SECURE` | `false` | If `true`, auth cookies are marked `Secure` (required for HTTPS deployments) |
| `FINDUS_LOG_LEVEL` | `info` | One of `debug`, `info`, `warn`, `error` |

## HTTPS and reverse proxies

1. Terminate TLS at your reverse proxy (Caddy, nginx, Traefik, etc.) and forward HTTP to Findus.
2. Set **`FINDUS_BASE_URL`** to the **https** URL users type in the browser (including scheme and host, no path unless you mount the app under a subpath—if you do, ensure routing matches your proxy).
3. Set **`FINDUS_COOKIE_SECURE=true`** so session cookies are only sent over HTTPS.

## QR codes and base URL

QR payloads point at `{FINDUS_BASE_URL}/q/{token}`. If `FINDUS_BASE_URL` is wrong, scans will hit the wrong host or scheme. This is the most common misconfiguration behind TLS or LAN access.

## Docker Compose

The sample `docker-compose.yml` sets `FINDUS_BASE_URL` and mounts a named volume at `/data` inside the container (`FINDUS_DATA_DIR=/data` in the image). Adjust published ports and environment to match your deployment.
