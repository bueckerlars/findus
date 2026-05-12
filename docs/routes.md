# Routes

This is an **overview** of HTTP surfaces. For implementation details, see `backend/internal/transport/http/handler`.

## Public

- `/login`, `/register` (when registration mode allows)
- `/healthz` — liveness text response (`ok`)
- `/static/*` — static files (CSS, etc.)

## Authenticated (signed-in users)

Typical inventory and read flows:

- `/` — home
- `/locations`, `/locations/{id}` — location tree and detail
- `/items`, `/items/{id}` — item list and detail
- `/search` — search (FTS5-backed where available, with LIKE fallback)
- `/command-search?q=` — JSON item hits for the signed-in command palette (empty `q` returns `{"items":[]}`)
- `/q/{token}` — resolve QR token to a location or item
- Authenticated GET endpoints for QR PNGs and item/location photos as implemented by handlers

## Admin

Admins can create, update, and delete locations and items, manage users and invites, adjust settings (including registration mode), and download backup archives:

- `/admin/*` — admin UI and actions
- `/admin/backup.zip` — ZIP containing database snapshot and images (see [Backup & restore](backup-restore.md))

Exact paths and method constraints follow the Go `ServeMux` patterns in the server package; treat this document as a map, not an OpenAPI spec.
