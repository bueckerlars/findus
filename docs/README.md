# Findus — Documentation

This is the technical reference for Findus. For a user-facing overview and quick start, see the [root README](../README.md).

---

## Pages

| Page | What it covers |
|---|---|
| [Architecture](architecture.md) | Go layer structure, Vue SPA, SQLite storage, security model |
| [Configuration](configuration.md) | All environment variables, HTTPS, JWT secrets, Docker Compose variants |
| [Development](development.md) | Prerequisites, Makefile targets, hot reload, tests, repo layout |
| [Routes](routes.md) | Full HTTP API surface — public, authenticated, and admin routes |
| [Backup & Restore](backup-restore.md) | What the backup ZIP contains and how to restore it |

---

## Architecture in one sentence

Findus is a **single Go binary** that serves an embedded Vue 3 SPA, a JSON API under `/api/*`, and binary endpoints (photos, QR PNGs, backup ZIP). It stores everything in SQLite on disk — no external services required.

## Stack at a glance

| Layer | Technology |
|---|---|
| Backend language | Go 1.23 |
| HTTP router | `net/http` (Go 1.22+ route patterns) |
| Database | SQLite via `modernc.org/sqlite` (CGO-free) |
| Migrations | [goose](https://github.com/pressly/goose) SQL files, embedded |
| Frontend | Vue 3 + Vite + Tailwind CSS (embedded `dist/`) |
| Image processing | Resize → WebP via `disintegration/imaging` + `HugoSmits86/nativewebp` |
| Auth | JWT in HTTP-only cookie; CSRF via double-submit cookie |
