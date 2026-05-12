package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"findus/backend/internal/repository/sqlite"
	"findus/backend/internal/service"
)

// APIAdminInventoryExport streams JSON or a CSV bundle ZIP (format=json|csv).
func (s *Server) APIAdminInventoryExport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	ctx := r.Context()
	format := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("format")))
	bundle, err := s.Inventory.BuildInventoryExportBundle(ctx)
	if err != nil {
		s.writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	switch format {
	case "json":
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="findus-inventory-%d.json"`, time.Now().Unix()))
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		if err := enc.Encode(bundle); err != nil {
			s.Log.Error("inventory export json", "err", err)
		}
	case "csv":
		zb, err := service.EncodeInventoryCSVZIP(bundle)
		if err != nil {
			s.writeJSONError(w, http.StatusBadRequest, err.Error())
			return
		}
		w.Header().Set("Content-Type", "application/zip")
		w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="findus-inventory-%d.zip"`, time.Now().Unix()))
		if _, err := w.Write(zb); err != nil {
			s.Log.Error("inventory export zip", "err", err)
		}
	default:
		s.writeJSONError(w, http.StatusBadRequest, "bad format (use json or csv)")
	}
}

// APIAdminInventoryImport accepts JSON body (full bundle) or multipart file (ZIP from csv export).
func (s *Server) APIAdminInventoryImport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	ctx := r.Context()
	const maxBody = 32 << 20
	r.Body = http.MaxBytesReader(w, r.Body, maxBody)

	ct := r.Header.Get("Content-Type")
	var bundle *service.InventoryExportBundle
	switch {
	case strings.HasPrefix(ct, "application/json"):
		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()
		var b service.InventoryExportBundle
		if err := dec.Decode(&b); err != nil {
			s.writeJSONError(w, http.StatusBadRequest, "invalid json")
			return
		}
		bundle = &b
	case strings.HasPrefix(ct, "multipart/form-data"):
		if err := r.ParseMultipartForm(maxBody); err != nil {
			s.writeJSONError(w, http.StatusBadRequest, "bad multipart form")
			return
		}
		f, _, err := r.FormFile("file")
		if err != nil {
			s.writeJSONError(w, http.StatusBadRequest, "missing file field")
			return
		}
		defer func() { _ = f.Close() }()
		data, err := io.ReadAll(f)
		if err != nil {
			s.writeJSONError(w, http.StatusBadRequest, "read failed")
			return
		}
		bundle, err = service.DecodeInventoryCSVZIP(data)
		if err != nil {
			s.writeJSONError(w, http.StatusBadRequest, err.Error())
			return
		}
	default:
		s.writeJSONError(w, http.StatusUnsupportedMediaType, "use application/json or multipart/form-data with file field")
		return
	}

	if bundle == nil {
		s.writeJSONError(w, http.StatusBadRequest, "empty import")
		return
	}
	if err := service.ValidateInventoryImportBundle(bundle); err != nil {
		s.writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		s.writeJSONError(w, http.StatusInternalServerError, "tx begin failed")
		return
	}
	defer func() { _ = tx.Rollback() }()

	invTx := &service.Inventory{
		Locations: sqlite.NewLocationRepoConn(tx),
		Items:     sqlite.NewItemRepoConn(tx),
		Labels:    sqlite.NewLabelRepoConn(tx),
		Templates: sqlite.NewItemTemplateRepoConn(tx),
		Settings:  sqlite.NewSettingsRepoConn(tx),
	}
	res, err := invTx.ImportInventoryBundle(ctx, bundle)
	if err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, service.ErrLabelNameConflict) {
			status = http.StatusConflict
		}
		s.writeJSONError(w, status, err.Error())
		return
	}
	if err := tx.Commit(); err != nil {
		s.writeJSONError(w, http.StatusInternalServerError, "commit failed")
		return
	}
	if err := s.Inventory.PostInventoryImportSearchRefresh(ctx, bundle); err != nil {
		s.Log.Error("inventory import search refresh", "err", err)
	}
	if err := s.Inventory.ReconcileSequentialNextSeqAfterImport(ctx); err != nil {
		s.Log.Error("inventory import reconcile item ids", "err", err)
	}
	s.writeJSON(w, http.StatusOK, res)
}
