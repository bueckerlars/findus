package handler

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"findus/backend/internal/config"
	"findus/backend/internal/domain"
	"findus/backend/internal/repository/sqlite"
	"findus/backend/internal/service"
	"findus/backend/internal/transport/http/middleware"
)

type Server struct {
	Log    *slog.Logger
	Config config.Config

	DB *sql.DB

	Users        *sqlite.UserRepo
	Groups       *sqlite.GroupRepo
	Locs         *sqlite.LocationRepo
	Items        *sqlite.ItemRepo
	ItemQRTokens *sqlite.ItemQRTokenReservationRepo
	Labels       *sqlite.LabelRepo
	Templates    *sqlite.ItemTemplateRepo
	Invites      *sqlite.InviteRepo
	Settings     *sqlite.SettingsRepo

	Auth      *service.Auth
	Admin     *service.Admin
	Inventory *service.Inventory
	QR        *service.QR
	Backup    *service.Backup

	JWTSecret []byte
}

type labelRow struct {
	LB                   domain.Label
	ChipHref             string
	DefaultTemplateTitle string
}

func (s *Server) labelRows(canEditLabels bool, list []domain.Label, templateTitles map[string]string) []labelRow {
	rows := make([]labelRow, 0, len(list))
	for _, lb := range list {
		row := labelRow{LB: lb}
		if canEditLabels {
			row.ChipHref = "/labels/" + lb.ID + "/edit"
		} else {
			row.ChipHref = "/search?q=" + url.QueryEscape(lb.Name)
		}
		if lb.DefaultTemplateType != nil {
			tid := *lb.DefaultTemplateType
			if templateTitles != nil {
				if title := templateTitles[tid]; title != "" {
					row.DefaultTemplateTitle = title
				} else {
					row.DefaultTemplateTitle = tid
				}
			} else {
				row.DefaultTemplateTitle = tid
			}
		}
		rows = append(rows, row)
	}
	return rows
}

type locOpt struct {
	ID    string
	Label string
}

func recentlyCreated(createdAt, updatedAt time.Time) bool {
	d := updatedAt.Sub(createdAt)
	return d >= 0 && d < 2*time.Second
}

func mergedAdditionalForEdit(it *domain.Item) json.RawMessage {
	merged, err := service.MergeTemplateIntoAdditional(it.AdditionalData, it.TemplateData)
	if err != nil {
		return it.AdditionalData
	}
	return merged
}

func (s *Server) Health(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = io.WriteString(w, "ok\n")
}

func (s *Server) AuthSettingsMode(ctx context.Context) (domain.RegistrationMode, bool, error) {
	v, ok, err := s.Settings.Get(ctx, domain.SettingRegistrationMode)
	if err != nil || !ok {
		return "", ok, err
	}
	m, valid := domain.ParseRegistrationMode(v)
	if !valid {
		return "", false, nil
	}
	return m, true, nil
}

func (s *Server) ProfilePhoto(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	u, ok := middleware.User(ctx)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	if u.AvatarPath == nil || *u.AvatarPath == "" {
		http.NotFound(w, r)
		return
	}
	p := filepath.Join(s.Config.DataDir, filepath.FromSlash(*u.AvatarPath))
	b, err := os.ReadFile(p)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "image/webp")
	_, _ = w.Write(b)
}

func (s *Server) QRedirect(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	token := strings.TrimSpace(r.PathValue("token"))
	if token == "" {
		http.NotFound(w, r)
		return
	}
	if loc, err := s.Locs.GetByQRToken(ctx, token); err == nil {
		http.Redirect(w, r, "/locations/"+loc.ID, http.StatusFound)
		return
	}
	if it, err := s.Items.GetByQRToken(ctx, token); err == nil {
		http.Redirect(w, r, "/items/"+it.ID, http.StatusFound)
		return
	}
	if s.ItemQRTokens != nil {
		itemID, ok, err := s.ItemQRTokens.GetItemIDByToken(ctx, token)
		if err == nil && ok {
			if it, itemErr := s.Items.GetByID(ctx, itemID); itemErr == nil {
				http.Redirect(w, r, "/items/"+it.ID, http.StatusFound)
				return
			}
		}
	}
	http.NotFound(w, r)
}

func (s *Server) LocationQR(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := r.PathValue("id")
	loc, err := s.Locs.GetByID(ctx, id)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	png, err := s.QR.PNG(loc.QRToken)
	if err != nil {
		http.Error(w, "qr", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "image/png")
	_, _ = w.Write(png)
}

func (s *Server) flatLocations(ctx context.Context) ([]locOpt, error) {
	var out []locOpt
	var walk func(parent *string, prefix string) error
	walk = func(parent *string, prefix string) error {
		ls, err := s.Locs.ListChildren(ctx, parent)
		if err != nil {
			return err
		}
		for _, l := range ls {
			label := prefix + l.Name
			out = append(out, locOpt{ID: l.ID, Label: label})
			p := l.ID
			if err := walk(&p, label+" / "); err != nil {
				return err
			}
		}
		return nil
	}
	if err := walk(nil, ""); err != nil {
		return nil, err
	}
	return out, nil
}

// locationDisplayNamesForIDs returns a full path label ("A / B / C") per location id (deduplicated).
func (s *Server) locationDisplayNamesForIDs(ctx context.Context, ids []string) map[string]string {
	seen := make(map[string]struct{}, len(ids))
	uniq := make([]string, 0, len(ids))
	for _, id := range ids {
		if id == "" {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		uniq = append(uniq, id)
	}
	out := make(map[string]string, len(uniq))
	for _, id := range uniq {
		path, err := s.Locs.PathFromRoot(ctx, id)
		if err != nil || len(path) == 0 {
			if loc, err2 := s.Locs.GetByID(ctx, id); err2 == nil {
				out[id] = loc.Name
			} else {
				out[id] = "Unknown"
			}
			continue
		}
		parts := make([]string, len(path))
		for i, e := range path {
			parts[i] = e.Name
		}
		out[id] = strings.Join(parts, " / ")
	}
	return out
}

func (s *Server) parentLocationOptions(ctx context.Context, excludeID string) ([]locOpt, error) {
	all, err := s.flatLocations(ctx)
	if err != nil {
		return nil, err
	}
	out := []locOpt{{ID: "", Label: "— Top level —"}}
	for _, o := range all {
		if o.ID != excludeID {
			out = append(out, o)
		}
	}
	return out, nil
}

type additionalKV struct {
	K, V string
}

func additionalFieldPairs(raw json.RawMessage) []additionalKV {
	if len(raw) == 0 || string(raw) == "null" || string(raw) == "{}" {
		return nil
	}
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil || len(m) == 0 {
		return nil
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	rows := make([]additionalKV, 0, len(keys))
	for _, k := range keys {
		rows = append(rows, additionalKV{K: k, V: fmt.Sprint(m[k])})
	}
	return rows
}

func parseLabelIDs(r *http.Request) []string {
	seen := make(map[string]struct{})
	var out []string
	for _, id := range r.Form["label_id"] {
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out
}

func (s *Server) saveItemPhoto(_ context.Context, itemID string, r io.Reader) error {
	b, err := service.EncodeWebPFromImage(r)
	if err != nil {
		return err
	}
	dir := filepath.Join(s.Config.DataDir, "images")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	dst := filepath.Join(dir, itemID+".webp")
	tmp := dst + ".tmp"
	if err := os.WriteFile(tmp, b, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, dst)
}

func (s *Server) saveUserAvatar(_ context.Context, userID string, r io.Reader) error {
	b, err := service.EncodeWebPFromImage(r)
	if err != nil {
		return err
	}
	dir := filepath.Join(s.Config.DataDir, "images")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	dst := filepath.Join(dir, "user-"+userID+".webp")
	tmp := dst + ".tmp"
	if err := os.WriteFile(tmp, b, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, dst)
}

func (s *Server) ItemPhoto(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := r.PathValue("id")
	it, err := s.Items.GetByID(ctx, id)
	if err != nil || it.PhotoPath == nil {
		http.NotFound(w, r)
		return
	}
	p := filepath.Join(s.Config.DataDir, filepath.FromSlash(*it.PhotoPath))
	b, err := os.ReadFile(p)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "image/webp")
	_, _ = w.Write(b)
}

func (s *Server) ItemQR(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := r.PathValue("id")
	it, err := s.Items.GetByID(ctx, id)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	png, err := s.QR.PNG(it.QRToken)
	if err != nil {
		http.Error(w, "qr", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "image/png")
	_, _ = w.Write(png)
}

func (s *Server) CommandSearchGet(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if q == "" {
		_, _ = w.Write([]byte(`{"items":[]}`))
		return
	}
	res, err := s.Items.Search(ctx, q, 12)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	type row struct {
		ID           string `json:"id"`
		Name         string `json:"name"`
		LocationName string `json:"location_name"`
	}
	ids := make([]string, 0, len(res))
	for _, it := range res {
		ids = append(ids, it.LocationID)
	}
	locNames := s.locationDisplayNamesForIDs(ctx, ids)
	out := struct {
		Items []row `json:"items"`
	}{Items: make([]row, 0, len(res))}
	for _, it := range res {
		out.Items = append(out.Items, row{ID: it.ID, Name: it.Name, LocationName: locNames[it.LocationID]})
	}
	_ = json.NewEncoder(w).Encode(out)
}

func (s *Server) AdminBackup(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="findus-backup-%d.zip"`, time.Now().Unix()))
	b := &service.Backup{DataDir: s.Config.DataDir}
	if err := b.StreamZIP(ctx, w, s.DB); err != nil {
		s.Log.Error("backup", "err", err)
	}
}
