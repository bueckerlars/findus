package service

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"findus/backend/internal/domain"
)

// GetItemIDPolicy returns the stored policy or the default.
func (s *Inventory) GetItemIDPolicy(ctx context.Context) (domain.ItemIDPolicy, error) {
	if s.Settings == nil {
		return domain.DefaultItemIDPolicy(), nil
	}
	raw, ok, err := s.Settings.Get(ctx, domain.SettingItemIDPolicy)
	if err != nil {
		return domain.ItemIDPolicy{}, err
	}
	if !ok || strings.TrimSpace(raw) == "" {
		return domain.DefaultItemIDPolicy(), nil
	}
	p, err := domain.ParseItemIDPolicyJSON(raw)
	if err != nil {
		return domain.DefaultItemIDPolicy(), nil
	}
	return p, nil
}

func (s *Inventory) saveItemIDPolicy(ctx context.Context, p domain.ItemIDPolicy) error {
	if s.Settings == nil {
		return fmt.Errorf("%w: settings", domain.ErrValidation)
	}
	b, err := domain.ItemIDPolicyJSON(p)
	if err != nil {
		return err
	}
	return s.Settings.Set(ctx, domain.SettingItemIDPolicy, string(b))
}

// SetItemIDPolicy validates, migrates all items when the effective scheme changes, then persists policy.
func (s *Inventory) SetItemIDPolicy(ctx context.Context, dataDir string, want domain.ItemIDPolicy) error {
	if err := want.Validate(); err != nil {
		return err
	}
	want = want.Normalize()
	cur, err := s.GetItemIDPolicy(ctx)
	if err != nil {
		return err
	}
	items, err := s.Items.ListAll(ctx, 2000)
	if err != nil {
		return err
	}
	if cur.EffectiveKey() == want.EffectiveKey() && allItemIDsMatchPolicy(items, want) {
		return s.saveItemIDPolicy(ctx, want)
	}

	sort.SliceStable(items, func(i, j int) bool {
		if !items[i].CreatedAt.Equal(items[j].CreatedAt) {
			return items[i].CreatedAt.Before(items[j].CreatedAt)
		}
		return items[i].ID < items[j].ID
	})

	rows, nextSeq, err := buildItemIDMigrationRows(items, want)
	if err != nil {
		return fmt.Errorf("build migration: %w", err)
	}

	var createdFiles []string
	success := false
	defer func() {
		if !success {
			for _, p := range createdFiles {
				_ = os.Remove(p)
			}
		}
	}()

	createdFiles, err = copyItemImagesForMigration(dataDir, rows)
	if err != nil {
		return fmt.Errorf("copy images: %w", err)
	}

	if err = s.Items.MigrateItemPrimaryKeys(ctx, rows); err != nil {
		return fmt.Errorf("migrate pk: %w", err)
	}

	for _, rw := range rows {
		if rw.OldID == rw.NewID {
			continue
		}
		oldPath := filepath.Join(dataDir, "images", rw.OldID+".webp")
		if _, statErr := os.Stat(oldPath); statErr == nil {
			_ = os.Remove(oldPath)
		}
	}

	want.NextSeq = nextSeq
	if err = s.saveItemIDPolicy(ctx, want); err != nil {
		return fmt.Errorf("save policy: %w", err)
	}

	for _, rw := range rows {
		if rw.OldID == rw.NewID {
			continue
		}
		if err = s.refreshItemSearchDenorm(ctx, rw.NewID); err != nil {
			return fmt.Errorf("refresh search %q: %w", rw.NewID, err)
		}
	}
	success = true
	return nil
}

func allItemIDsMatchPolicy(items []domain.Item, pol domain.ItemIDPolicy) bool {
	for i := range items {
		if !domain.ItemIDMatchesPolicy(items[i].ID, pol) {
			return false
		}
	}
	return true
}

func buildItemIDMigrationRows(items []domain.Item, pol domain.ItemIDPolicy) ([]domain.ItemIDMigration, int64, error) {
	pol = pol.Normalize()
	n := len(items)
	rows := make([]domain.ItemIDMigration, 0, n)
	if pol.Kind != domain.ItemIDKindSequential {
		return nil, 0, fmt.Errorf("%w: kind", domain.ErrValidation)
	}
	for i, it := range items {
		seq := int64(i) + 1
		newID := domain.FormatSequentialID(pol.Prefix, pol.Width, seq)
		rows = append(rows, itemMigrationRow(it, newID))
	}
	return rows, int64(n) + 1, nil
}

func itemMigrationRow(it domain.Item, newID string) domain.ItemIDMigration {
	var photo *string
	if it.PhotoPath != nil && strings.TrimSpace(*it.PhotoPath) != "" {
		rel := filepath.ToSlash(filepath.Join("images", newID+".webp"))
		photo = &rel
	}
	return domain.ItemIDMigration{OldID: it.ID, NewID: newID, NewPhotoPath: photo}
}

func copyItemImagesForMigration(dataDir string, rows []domain.ItemIDMigration) (created []string, err error) {
	imgDir := filepath.Join(dataDir, "images")
	if err := os.MkdirAll(imgDir, 0o755); err != nil {
		return nil, err
	}
	for _, rw := range rows {
		if rw.OldID == rw.NewID {
			continue
		}
		src := filepath.Join(dataDir, "images", rw.OldID+".webp")
		dst := filepath.Join(dataDir, "images", rw.NewID+".webp")
		if _, statErr := os.Stat(src); statErr != nil {
			continue
		}
		in, err := os.Open(src)
		if err != nil {
			return created, err
		}
		tmp := dst + ".tmp"
		out, err := os.Create(tmp)
		if err != nil {
			_ = in.Close()
			return created, err
		}
		if _, err := io.Copy(out, in); err != nil {
			_ = in.Close()
			_ = out.Close()
			_ = os.Remove(tmp)
			return created, err
		}
		_ = in.Close()
		if err := out.Close(); err != nil {
			_ = os.Remove(tmp)
			return created, err
		}
		if err := os.Rename(tmp, dst); err != nil {
			_ = os.Remove(tmp)
			return created, err
		}
		created = append(created, dst)
	}
	return created, nil
}

func (s *Inventory) allocateNextItemID(ctx context.Context) (string, error) {
	pol, err := s.GetItemIDPolicy(ctx)
	if err != nil {
		return "", err
	}
	pol = pol.Normalize()
	if pol.Kind != domain.ItemIDKindSequential {
		return "", fmt.Errorf("%w: kind", domain.ErrValidation)
	}
	return domain.FormatSequentialID(pol.Prefix, pol.Width, pol.NextSeq), nil
}

func (s *Inventory) sequentialBumpAfterSuccessfulCreate(ctx context.Context, issuedID string) error {
	pol, err := s.GetItemIDPolicy(ctx)
	if err != nil {
		return err
	}
	pol = pol.Normalize()
	if pol.Kind != domain.ItemIDKindSequential {
		return fmt.Errorf("%w: kind", domain.ErrValidation)
	}
	expected := domain.FormatSequentialID(pol.Prefix, pol.Width, pol.NextSeq)
	if issuedID != expected {
		return nil
	}
	pol.NextSeq++
	return s.saveItemIDPolicy(ctx, pol)
}

func isSQLiteUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "UNIQUE constraint failed") || strings.Contains(msg, "constraint failed")
}
