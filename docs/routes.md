# Routes

This is an **overview** of HTTP surfaces. For implementation details, see `backend/internal/transport/http/handler` (notably `spa.go` for mount order).

## Public (no session required)

- `GET /healthz` — liveness text response (`ok`)
- `GET /api/bootstrap` — registration flags for the register screen
- `POST /api/auth/login`, `POST /api/auth/register` — JSON body; set session cookie on success

## Static SPA (built with Vite)

- `GET /assets/*` — hashed JS/CSS from the embedded `frontend/dist` tree
- `GET /{path...}` — SPA shell (`dist/index.html`) for browser navigation on app routes (only when no more specific route matches)

## Authenticated JSON (`/api/*`)

The Vue app calls these with `fetch(..., { credentials: "same-origin" })`. Mutating requests require the `X-CSRF-Token` header (double-submit cookie `findus_csrf`). Highlights:

- `GET /api/me` — current user or 401 (router guards)
- `GET /api/home`, `GET /api/profile`, `POST /api/profile` (JSON or multipart for avatar)
- Locations: `GET /api/locations`, `GET /api/locations/{id}`, create/update/delete (admin for writes)
- Items: list/detail/edit payloads, `POST /api/items` and `POST /api/items/{id}` as `multipart/form-data` for creates/updates (photos + template fields)
- `GET /api/items/new/fields?template_type=` — template field metadata for dynamic forms
- `GET /api/search?q=` — JSON search results
- `GET /api/command-search?q=` — JSON hits for the command palette (empty `q` returns `{"items":[]}`)
- Labels CRUD under `/api/labels...` (admin for writes)
- Admin: `/api/admin/users`, invites, registration mode, item templates, etc.

## Authenticated binary / non-JSON

- `GET /q/{token}` — resolve QR token to a location or item (redirect)
- `GET /items/{id}/photo`, `GET /items/{id}/qr.png`, `GET /locations/{id}/qr.png`, `GET /profile/photo`
- `GET /admin/backup.zip` — admin-only backup archive (see [Backup & restore](backup-restore.md))

Exact patterns follow Go 1.22+ `ServeMux` registration order in `Server.Handler()`.
