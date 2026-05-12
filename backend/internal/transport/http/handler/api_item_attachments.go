package handler

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"findus/backend/internal/domain"
)

func (s *Server) itemAttachmentsPayload(ctx context.Context, itemID string) ([]map[string]any, error) {
	if s.Inventory == nil || s.Inventory.ItemAttachments == nil {
		return []map[string]any{}, nil
	}
	list, err := s.Inventory.ListItemAttachments(ctx, itemID)
	if err != nil {
		return nil, err
	}
	out := make([]map[string]any, 0, len(list))
	for _, a := range list {
		out = append(out, map[string]any{
			"ID":               a.ID,
			"ItemID":           a.ItemID,
			"OriginalFilename": a.OriginalFilename,
			"Title":            a.Title,
			"MimeType":         a.MimeType,
			"SizeBytes":        a.SizeBytes,
			"CreatedAt":        a.CreatedAt.UTC().Format(time.RFC3339Nano),
			"DownloadURL":      "/items/" + itemID + "/attachments/" + a.ID,
		})
	}
	return out, nil
}

// APIItemAttachmentsList returns attachment metadata for an item (authenticated users).
func (s *Server) APIItemAttachmentsList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	ctx := r.Context()
	itemID := r.PathValue("id")
	atts, err := s.itemAttachmentsPayload(ctx, itemID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			s.writeJSONError(w, http.StatusNotFound, "not found")
			return
		}
		s.writeJSONError(w, http.StatusInternalServerError, "server error")
		return
	}
	s.writeJSON(w, http.StatusOK, map[string]any{"attachments": atts})
}

// APIItemAttachmentCreate accepts multipart form: file (required), title (optional).
func (s *Server) APIItemAttachmentCreate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	itemID := r.PathValue("id")
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		s.writeJSONError(w, http.StatusBadRequest, "bad form")
		return
	}
	f, hdr, err := r.FormFile("file")
	if err != nil {
		s.writeJSONError(w, http.StatusBadRequest, "missing file field")
		return
	}
	defer f.Close()
	title := strings.TrimSpace(r.FormValue("title"))
	orig := hdr.Filename
	att, err := s.Inventory.AddItemAttachment(ctx, s.Config.DataDir, itemID, title, orig, f)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrNotFound):
			s.writeJSONError(w, http.StatusNotFound, "not found")
		case errors.Is(err, domain.ErrValidation):
			s.writeJSONError(w, http.StatusBadRequest, err.Error())
		default:
			s.Log.Error("attachment upload", "err", err)
			s.writeJSONError(w, http.StatusBadRequest, err.Error())
		}
		return
	}
	s.writeJSON(w, http.StatusOK, map[string]any{"attachment": att})
}

type apiItemAttachmentPatchReq struct {
	Title string `json:"title"`
}

// APIItemAttachmentPatch updates attachment metadata (admin only); body: {"title":"..."}.
func (s *Server) APIItemAttachmentPatch(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	itemID := r.PathValue("id")
	attachmentID := r.PathValue("attachmentId")
	var req apiItemAttachmentPatchReq
	if err := readJSON(r, &req); err != nil {
		s.writeJSONError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if err := s.Inventory.UpdateItemAttachmentTitle(ctx, itemID, attachmentID, req.Title); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			s.writeJSONError(w, http.StatusNotFound, "not found")
			return
		}
		if errors.Is(err, domain.ErrValidation) {
			s.writeJSONError(w, http.StatusBadRequest, err.Error())
			return
		}
		s.writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	s.writeJSON(w, http.StatusOK, map[string]string{"ok": "1"})
}

// APIItemAttachmentDelete removes one attachment (admin only).
func (s *Server) APIItemAttachmentDelete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	itemID := r.PathValue("id")
	attachmentID := r.PathValue("attachmentId")
	if err := s.Inventory.DeleteItemAttachment(ctx, s.Config.DataDir, itemID, attachmentID); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			s.writeJSONError(w, http.StatusNotFound, "not found")
			return
		}
		s.writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	s.writeJSON(w, http.StatusOK, map[string]string{"ok": "1"})
}

// ItemAttachmentDownload serves the binary file (same auth as item photo).
func (s *Server) ItemAttachmentDownload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	ctx := r.Context()
	itemID := r.PathValue("id")
	attachmentID := r.PathValue("attachmentId")
	if s.Inventory == nil || s.Inventory.ItemAttachments == nil {
		http.NotFound(w, r)
		return
	}
	a, err := s.Inventory.ItemAttachments.GetByID(ctx, attachmentID)
	if err != nil || a.ItemID != itemID {
		http.NotFound(w, r)
		return
	}
	p := filepath.Join(s.Config.DataDir, filepath.FromSlash(a.StoragePath))
	f, err := os.Open(p)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	defer f.Close()
	w.Header().Set("Content-Type", a.MimeType)
	w.Header().Set("Content-Disposition", attachmentContentDisposition(a))
	_, _ = io.Copy(w, f)
}

func attachmentContentDisposition(att *domain.ItemAttachment) string {
	name := strings.TrimSpace(att.OriginalFilename)
	if att.Title != "" {
		ext := filepath.Ext(att.OriginalFilename)
		if ext == "" {
			ext = filepath.Ext(att.StoragePath)
		}
		name = att.Title + ext
	}
	if name == "" {
		name = "download"
	}
	ascii := strings.Map(func(r rune) rune {
		if r >= 32 && r <= 126 && r != '"' && r != '\\' {
			return r
		}
		return '_'
	}, name)
	if strings.Trim(ascii, "_") == "" {
		ascii = "download"
	}
	star := url.PathEscape(name)
	return fmt.Sprintf(`attachment; filename="%s"; filename*=UTF-8''%s`, ascii, star)
}
