# Backup & restore

## Backup (admin)

1. Sign in as an **admin**.
2. Use the in-app backup action, or open **`GET /admin/backup.zip`**.

The browser download is named like **`findus-backup-<unix_timestamp>.zip`**. Inside the archive:

- **`findus.db`** — consistent SQLite snapshot (via `VACUUM INTO` at export time)
- **`images/`** — image files from the data directory (flat files under `images/`, not nested subfolders in the current implementation)

## Restore

1. **Stop** the Findus process so the database is not open for writes.
2. Extract the ZIP and copy into your **`FINDUS_DATA_DIR`**:
   - Place **`findus.db`** at the root of the data directory (same filename the running instance expects: `findus.db`).
   - Copy the **`images/`** directory next to it so paths match a normal install (`$FINDUS_DATA_DIR/images/...`).
3. Start Findus again with the **same** `FINDUS_DATA_DIR` pointing at that folder.

If you also rely on a persisted JWT secret file (`.jwt_secret` under the data directory when not using `FINDUS_JWT_SECRET`), keep or restore that file consistently with your security practices; rotating secrets invalidates existing sessions.

## Docker volume

When using Compose, the data directory is usually the **`/data`** volume mount. Restore into that volume while the container is stopped, or copy files into the named volume using your preferred volume tooling.
