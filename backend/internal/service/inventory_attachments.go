package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"

	"findus/backend/internal/domain"
)

const maxItemAttachmentBytes = 32 << 20

// AttachmentPostProcessor runs after an attachment is stored (e.g. future OCR pipeline). Optional on Inventory.
type AttachmentPostProcessor interface {
	OnItemAttachmentCreated(ctx context.Context, att *domain.ItemAttachment) error
}

type noopAttachmentPostProcessor struct{}

func (noopAttachmentPostProcessor) OnItemAttachmentCreated(_ context.Context, _ *domain.ItemAttachment) error {
	return nil
}

// attachmentMimeToExt maps detected MIME to a safe file suffix (never use user-supplied names in paths).
var attachmentMimeToExt = map[string]string{
	"application/pdf": ".pdf",
	"image/jpeg":      ".jpg",
	"image/jpg":       ".jpg",
	"image/png":       ".png",
	"image/webp":      ".webp",
	"image/gif":       ".gif",
}

func sanitizeOriginalFilename(name string) string {
	name = filepath.Base(strings.TrimSpace(name))
	if name == "." || name == "" {
		return "upload"
	}
	var b strings.Builder
	for _, r := range name {
		if r < 32 || r == '/' || r == '\\' {
			continue
		}
		b.WriteRune(r)
	}
	out := strings.TrimSpace(b.String())
	if out == "" {
		return "upload"
	}
	if utf8.RuneCountInString(out) > 200 {
		out = string([]rune(out)[:200])
	}
	return out
}

func normalizeAttachmentTitle(title string) string {
	title = strings.TrimSpace(title)
	if utf8.RuneCountInString(title) > 120 {
		title = string([]rune(title)[:120])
	}
	return title
}

func validateMetadataJSON(raw json.RawMessage) (json.RawMessage, error) {
	if len(raw) == 0 {
		return json.RawMessage(`{}`), nil
	}
	var v any
	if err := json.Unmarshal(raw, &v); err != nil {
		return nil, fmt.Errorf("%w: metadata_json must be valid JSON", domain.ErrValidation)
	}
	b, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("%w: metadata_json", domain.ErrValidation)
	}
	return json.RawMessage(b), nil
}

// ListItemAttachments returns metadata rows for an item (no file I/O).
func (s *Inventory) ListItemAttachments(ctx context.Context, itemID string) ([]domain.ItemAttachment, error) {
	if s.ItemAttachments == nil {
		return nil, nil
	}
	if _, err := s.Items.GetByID(ctx, itemID); err != nil {
		return nil, err
	}
	return s.ItemAttachments.ListByItemID(ctx, itemID)
}

// AddItemAttachment validates MIME/size, writes under dataDir/item-attachments/{itemID}/, inserts DB row, then optional post-processor.
func (s *Inventory) AddItemAttachment(ctx context.Context, dataDir, itemID, title, originalFilename string, r io.Reader) (*domain.ItemAttachment, error) {
	if s.ItemAttachments == nil {
		return nil, fmt.Errorf("%w: attachments not configured", domain.ErrValidation)
	}
	if _, err := s.Items.GetByID(ctx, itemID); err != nil {
		return nil, err
	}
	head := make([]byte, 512)
	n, err := io.ReadFull(r, head)
	if err == io.ErrUnexpectedEOF {
		err = nil
	}
	if err != nil {
		return nil, err
	}
	if n == 0 {
		return nil, fmt.Errorf("%w: empty file", domain.ErrValidation)
	}
	head = head[:n]
	detected := http.DetectContentType(head)
	ext, ok := attachmentMimeToExt[detected]
	if !ok {
		return nil, fmt.Errorf("%w: unsupported file type", domain.ErrValidation)
	}
	orig := sanitizeOriginalFilename(originalFilename)
	tit := normalizeAttachmentTitle(title)
	now := time.Now().UTC()
	id := newID()
	rel := filepath.ToSlash(filepath.Join("item-attachments", itemID, id+ext))
	absDir := filepath.Join(dataDir, "item-attachments", itemID)
	if err := os.MkdirAll(absDir, 0o755); err != nil {
		return nil, err
	}
	absPath := filepath.Join(dataDir, filepath.FromSlash(rel))
	tmp := absPath + ".tmp"
	if err := writeLimitedAttachmentFile(tmp, head, r, maxItemAttachmentBytes); err != nil {
		_ = os.Remove(tmp)
		return nil, err
	}
	if err := os.Rename(tmp, absPath); err != nil {
		_ = os.Remove(tmp)
		return nil, err
	}
	fi, err := os.Stat(absPath)
	if err != nil {
		_ = os.Remove(absPath)
		return nil, err
	}
	meta := json.RawMessage(`{}`)
	meta, err = validateMetadataJSON(meta)
	if err != nil {
		_ = os.Remove(absPath)
		return nil, err
	}
	att := &domain.ItemAttachment{
		ID:               id,
		ItemID:           itemID,
		OriginalFilename: orig,
		StoragePath:      rel,
		MimeType:         detected,
		SizeBytes:        fi.Size(),
		Title:            tit,
		MetadataJSON:     meta,
		CreatedAt:        now,
	}
	if err := s.ItemAttachments.Create(ctx, att); err != nil {
		_ = os.Remove(absPath)
		return nil, err
	}
	proc := s.AttachmentPostProcessor
	if proc == nil {
		proc = noopAttachmentPostProcessor{}
	}
	if err := proc.OnItemAttachmentCreated(ctx, att); err != nil {
		_ = s.ItemAttachments.DeleteByID(ctx, att.ID)
		_ = os.Remove(absPath)
		return nil, err
	}
	return att, nil
}

func writeLimitedAttachmentFile(path string, head []byte, r io.Reader, max int64) error {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	var written int64
	if len(head) > 0 {
		if int64(len(head)) > max {
			return fmt.Errorf("%w: file too large", domain.ErrValidation)
		}
		nw, err := f.Write(head)
		if err != nil {
			return err
		}
		written += int64(nw)
	}
	buf := make([]byte, 32*1024)
	for {
		nr, er := r.Read(buf)
		if nr > 0 {
			if written+int64(nr) > max {
				return fmt.Errorf("%w: file too large", domain.ErrValidation)
			}
			nw, ew := f.Write(buf[:nr])
			written += int64(nw)
			if ew != nil {
				return ew
			}
			if nw != nr {
				return io.ErrShortWrite
			}
		}
		if er == io.EOF {
			break
		}
		if er != nil {
			return er
		}
	}
	return nil
}

// UpdateItemAttachmentTitle updates the display title (admin workflow); empty title is allowed.
func (s *Inventory) UpdateItemAttachmentTitle(ctx context.Context, itemID, attachmentID, title string) error {
	if s.ItemAttachments == nil {
		return fmt.Errorf("%w: attachments not configured", domain.ErrValidation)
	}
	a, err := s.ItemAttachments.GetByID(ctx, attachmentID)
	if err != nil {
		return err
	}
	if a.ItemID != itemID {
		return domain.ErrNotFound
	}
	tit := normalizeAttachmentTitle(title)
	return s.ItemAttachments.UpdateTitle(ctx, attachmentID, tit)
}

// DeleteItemAttachment removes one attachment if it belongs to itemID.
func (s *Inventory) DeleteItemAttachment(ctx context.Context, dataDir, itemID, attachmentID string) error {
	if s.ItemAttachments == nil {
		return fmt.Errorf("%w: attachments not configured", domain.ErrValidation)
	}
	a, err := s.ItemAttachments.GetByID(ctx, attachmentID)
	if err != nil {
		return err
	}
	if a.ItemID != itemID {
		return domain.ErrNotFound
	}
	if err := s.ItemAttachments.DeleteByID(ctx, attachmentID); err != nil {
		return err
	}
	_ = os.Remove(filepath.Join(dataDir, filepath.FromSlash(a.StoragePath)))
	return nil
}

// RemoveItemAttachmentDir deletes the on-disk folder for all attachments of an item (call before deleting the item row).
func RemoveItemAttachmentDir(dataDir, itemID string) error {
	dir := filepath.Join(dataDir, "item-attachments", itemID)
	return os.RemoveAll(dir)
}
