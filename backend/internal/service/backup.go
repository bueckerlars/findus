package service

import (
	"archive/zip"
	"context"
	"database/sql"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Backup struct {
	DataDir string
}

// StreamZIP writes a zip containing a consistent SQLite snapshot and all images.
func (b *Backup) StreamZIP(ctx context.Context, w io.Writer, db *sql.DB) error {
	if err := os.MkdirAll(filepath.Join(b.DataDir, "images"), 0o755); err != nil {
		return err
	}
	snap := filepath.Join(b.DataDir, fmt.Sprintf("backup-snapshot-%d.db", time.Now().UnixNano()))
	snap = filepath.Clean(snap)
	if !strings.HasPrefix(snap, filepath.Clean(b.DataDir)) {
		return fmt.Errorf("invalid snapshot path")
	}
	_ = os.Remove(snap)
	escaped := strings.ReplaceAll(snap, "'", "''")
	if _, err := db.ExecContext(ctx, "VACUUM INTO '"+escaped+"'"); err != nil {
		return fmt.Errorf("vacuum into: %w", err)
	}
	defer func() { _ = os.Remove(snap) }()

	zw := zip.NewWriter(w)
	writeErr := func() error {
		f, err := os.Open(snap)
		if err != nil {
			return err
		}
		defer f.Close()
		wz, err := zw.Create("findus.db")
		if err != nil {
			return err
		}
		if _, err := io.Copy(wz, f); err != nil {
			return err
		}

		imgDir := filepath.Join(b.DataDir, "images")
		entries, err := os.ReadDir(imgDir)
		if err != nil {
			return err
		}
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			name := e.Name()
			lp := filepath.Join(imgDir, name)
			zf, err := zw.Create(filepath.Join("images", name))
			if err != nil {
				return err
			}
			bf, err := os.Open(lp)
			if err != nil {
				return err
			}
			if _, err := io.Copy(zf, bf); err != nil {
				_ = bf.Close()
				return err
			}
			_ = bf.Close()
		}
		return nil
	}()
	closeErr := zw.Close()
	if writeErr != nil {
		return writeErr
	}
	return closeErr
}
