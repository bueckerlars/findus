# Backup & Restore

Findus stores all data in two places: an SQLite database file and an `images/` directory. The built-in backup bundles both into a single ZIP file that can be used to restore a complete instance.

---

## Creating a backup

### From the UI

1. Sign in as an **admin**.
2. Open the admin panel and use the backup action.

### Direct download

Any admin session can download the backup directly:

```
GET /admin/backup.zip
```

The browser (or `curl`) receives a file named like **`findus-backup-<unix_timestamp>.zip`**.

### What's inside the ZIP

```
findus-backup-1715123456.zip
├── findus.db       consistent SQLite snapshot (exported via VACUUM INTO)
└── images/         all uploaded photos as WebP files
```

The SQLite snapshot is generated with `VACUUM INTO`, which produces a clean, consistent copy of the database at that point in time without locking out concurrent reads.

---

## Restoring from a backup

### 1 — Stop the running instance

```bash
docker compose down
# or: kill the findus process if running directly
```

The database must not be open for writes during the restore.

### 2 — Extract the ZIP

```bash
unzip findus-backup-1715123456.zip -d /tmp/findus-restore
```

### 3 — Copy files into the data directory

The data directory is `FINDUS_DATA_DIR` — `/data` inside the Docker volume by default.

```bash
# Place the database at the root of the data directory
cp /tmp/findus-restore/findus.db $DATA_DIR/findus.db

# Restore images alongside it
cp -r /tmp/findus-restore/images/ $DATA_DIR/images/
```

For Docker volumes, copy via a temporary container:

```bash
docker run --rm \
  -v findus_data:/data \
  -v /tmp/findus-restore:/restore \
  busybox \
  sh -c "cp /restore/findus.db /data/findus.db && cp -r /restore/images /data/images"
```

### 4 — Start the instance

```bash
docker compose up -d
```

Findus will start with the restored data. Migrations run automatically if needed.

---

## JWT secret

If you use the **auto-generated** JWT secret (i.e. `FINDUS_JWT_SECRET` is not set), the secret is stored as `$DATA_DIR/.jwt_secret`. It is **not** included in the backup ZIP.

- If you restore the database to the same machine/volume with the `.jwt_secret` file intact, existing sessions continue to work.
- If the `.jwt_secret` file is missing or different from when the backup was taken, existing sessions are invalidated. Users can simply log in again — no data is lost.

**Best practice:** set `FINDUS_JWT_SECRET` explicitly in your environment so it's managed outside the data directory and survives backups, volume replacements, and container recreations.

---

## Automated backups

For automated off-site backups, schedule a cron job that downloads the ZIP and copies it to your backup destination:

```bash
# Example: daily backup to a local directory
0 2 * * * curl -s -b "findus_session=<token>" http://localhost:8080/admin/backup.zip \
  -o /backups/findus-$(date +%Y%m%d).zip
```

Or use your container host's volume snapshot / backup tooling to snapshot the entire `findus_data` volume.
