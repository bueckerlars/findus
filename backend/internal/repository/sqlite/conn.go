package sqlite

import (
	"context"
	"database/sql"
)

// DBConn is implemented by *sql.DB and *sql.Tx for transactional admin import.
type DBConn interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}
