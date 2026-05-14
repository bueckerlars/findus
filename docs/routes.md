# Routes

HTTP API reference for Findus. All routes follow Go 1.22+ `net/http` `ServeMux` pattern syntax. For exact registration order, see `backend/internal/transport/http/`.

---

## Public routes (no session required)

| Method | Path | Description |
|---|---|---|
| `GET` | `/healthz` | Liveness check — returns `200 ok` |
| `GET` | `/api/bootstrap` | Registration flags consumed by the login/register screen |
| `POST` | `/api/auth/login` | JSON body `{username, password}` — sets session cookie on success |
| `POST` | `/api/auth/register` | JSON body `{username, password}` — creates an account (subject to registration mode) |

---

## Static / SPA routes

| Method | Path | Description |
|---|---|---|
| `GET` | `/assets/*` | Hashed JS, CSS, and font files from the embedded `frontend/dist/assets/` tree — served with long cache headers |
| `GET` | `/{path...}` | SPA shell — returns `dist/index.html` for any path not matched by a more specific route, enabling client-side navigation via Vue Router |

---

## Authenticated routes (`/api/*`)

All routes below require a valid session cookie. The Vue app calls these with `fetch(..., { credentials: "same-origin" })`.

**CSRF:** Mutating requests (`POST`, `PUT`, `PATCH`, `DELETE`) must include the `X-CSRF-Token` header with the value of the `findus_csrf` cookie. Requests without this header are rejected with `403`.

### User & session

| Method | Path | Description |
|---|---|---|
| `GET` | `/api/me` | Current user info — `401` if not authenticated (used by router guards) |
| `POST` | `/api/auth/logout` | Clears the session cookie |
| `GET` | `/api/profile` | Current user's profile |
| `POST` | `/api/profile` | Update profile — `multipart/form-data` (supports avatar upload) |

### Home

| Method | Path | Description |
|---|---|---|
| `GET` | `/api/home` | Dashboard data (recent items, location summary) |

### Locations

| Method | Path | Notes |
|---|---|---|
| `GET` | `/api/locations` | List all locations |
| `GET` | `/api/locations/{id}` | Location detail with nested children and items |
| `POST` | `/api/locations` | Create location — **admin only** |
| `POST` | `/api/locations/{id}` | Update location — **admin only** |
| `DELETE` | `/api/locations/{id}` | Delete location — **admin only** |

### Items

| Method | Path | Notes |
|---|---|---|
| `GET` | `/api/items` | List items (supports filter params) |
| `GET` | `/api/items/{id}` | Item detail |
| `POST` | `/api/items` | Create item — `multipart/form-data` (photos + template fields) |
| `POST` | `/api/items/{id}` | Update item — `multipart/form-data` |
| `DELETE` | `/api/items/{id}` | Delete item — **admin only** |
| `GET` | `/api/items/new/fields` | Template field metadata — `?template_type=` query param |

### Search

| Method | Path | Description |
|---|---|---|
| `GET` | `/api/search` | Full-text search — `?q=` query param, returns matched items and locations |
| `GET` | `/api/command-search` | Command palette results — `?q=` param; returns `{"items":[]}` for empty query |

### Labels

| Method | Path | Notes |
|---|---|---|
| `GET` | `/api/labels` | List all labels |
| `GET` | `/api/labels/{id}` | Label detail |
| `POST` | `/api/labels` | Create label — **admin only** |
| `POST` | `/api/labels/{id}` | Update label — **admin only** |
| `DELETE` | `/api/labels/{id}` | Delete label — **admin only** |

---

## Authenticated binary / non-JSON routes

| Method | Path | Description |
|---|---|---|
| `GET` | `/q/{token}` | Resolve a QR token — redirects to the matching location or item page |
| `GET` | `/items/{id}/photo` | Item photo (WebP image) |
| `GET` | `/items/{id}/qr.png` | Item QR code as PNG |
| `GET` | `/locations/{id}/qr.png` | Location QR code as PNG |
| `GET` | `/profile/photo` | Current user's avatar |

---

## Admin-only routes

| Method | Path | Description |
|---|---|---|
| `GET` | `/admin/backup.zip` | Download full backup archive (see [Backup & Restore](backup-restore.md)) |
| `GET` | `/api/admin/users` | List all users |
| `POST` | `/api/admin/users` | Create a user |
| `DELETE` | `/api/admin/users/{id}` | Delete a user |
| `GET` | `/api/admin/invites` | List pending invites |
| `POST` | `/api/admin/invites` | Generate an invite token |
| `DELETE` | `/api/admin/invites/{id}` | Revoke an invite |
| `GET` | `/api/admin/settings` | Get registration mode and other settings |
| `POST` | `/api/admin/settings` | Update settings |
| `GET` | `/api/admin/templates` | List item templates |
| `POST` | `/api/admin/templates` | Create template |
| `POST` | `/api/admin/templates/{id}` | Update template |
| `DELETE` | `/api/admin/templates/{id}` | Delete template |

---

## Error responses

All JSON endpoints return errors in the form:

```json
{ "error": "human-readable message" }
```

Common status codes:

| Code | Meaning |
|---|---|
| `400` | Invalid request body or query parameter |
| `401` | Not authenticated — session cookie missing or expired |
| `403` | Authenticated but not authorized (wrong role, or missing CSRF token) |
| `404` | Resource not found |
| `500` | Internal server error (details in server logs) |
