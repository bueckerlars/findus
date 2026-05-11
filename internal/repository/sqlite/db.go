package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/pressly/goose/v3"

	_ "modernc.org/sqlite"
)

func OpenDB(ctx context.Context, dataDir string) (*sql.DB, error) {
	if err := os.MkdirAll(filepath.Join(dataDir, "images"), 0o755); err != nil {
		return nil, fmt.Errorf("mkdir images: %w", err)
	}
	dbPath := filepath.Join(dataDir, "findus.db")
	dsn := fmt.Sprintf("file:%s?_pragma=foreign_keys(1)&_pragma=busy_timeout(5000)", dbPath)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	db.SetConnMaxLifetime(time.Hour)
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}
	goose.SetBaseFS(embeddedMigrations)
	defer func() { goose.SetBaseFS(nil) }()
	if err := goose.SetDialect("sqlite3"); err != nil {
		_ = db.Close()
		return nil, err
	}
	if err := goose.UpContext(ctx, db, "migrations"); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("migrate: %w", err)
	}
	return db, nil
}

func parseTime(s string) (time.Time, error) {
	if s == "" {
		return time.Time{}, fmt.Errorf("empty time")
	}
	return time.Parse(time.RFC3339Nano, s)
}

func formatTime(t time.Time) string {
	return t.UTC().Format(time.RFC3339Nano)
}

func sqlNullString(ns sql.NullString) *string {
	if !ns.Valid {
		return nil
	}
	v := ns.String
	return &v
}

func strPtr(ns *string) sql.NullString {
	if ns == nil {
		return sql.NullString{}
	}
	return sql.NullString{String: *ns, Valid: true}
}
