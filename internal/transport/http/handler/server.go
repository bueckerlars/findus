package handler

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"findus/internal/config"
	"findus/internal/domain"
	"findus/internal/repository/sqlite"
	"findus/internal/service"
	"findus/internal/transport/http/middleware"
	"findus/web"
)

type Server struct {
	Log    *slog.Logger
	Config config.Config

	DB *sql.DB

	Users     *sqlite.UserRepo
	Locs      *sqlite.LocationRepo
	Items     *sqlite.ItemRepo
	Labels    *sqlite.LabelRepo
	Templates *sqlite.ItemTemplateRepo
	Invites   *sqlite.InviteRepo
	Settings  *sqlite.SettingsRepo

	Auth      *service.Auth
	Admin     *service.Admin
	Inventory *service.Inventory
	QR        *service.QR
	Backup    *service.Backup

	JWTSecret []byte
	Tpl       *template.Template
}

type labelRow struct {
	LB                   domain.Label
	ChipHref             string
	DefaultTemplateTitle string
}

// homeRecentItem is used on the dashboard for recently changed items.
type homeRecentItem struct {
	Item           domain.Item
	LocationName   string
	RecentlyAdded  bool
}

// homeLocationRow is used on the dashboard for the locations list/gallery.
type homeLocationRow struct {
	Location         domain.Location
	SubLocationCount int64
	RecentlyAdded    bool
}

type shell struct {
	Title   string
	CSRF    string
	User    *domain.User
	IsAdmin bool
	Path    string
}

func (s *Server) shell(r *http.Request, title string) shell {
	u, _ := middleware.User(r.Context())
	return shell{
		Title:   title,
		CSRF:    middleware.CSRFToken(r.Context()),
		User:    u,
		IsAdmin: u != nil && u.Role.IsAdmin(),
		Path:    r.URL.Path,
	}
}

func (s *Server) labelRows(sh shell, list []domain.Label, templateTitles map[string]string) []labelRow {
	rows := make([]labelRow, 0, len(list))
	for _, lb := range list {
		row := labelRow{LB: lb}
		if sh.IsAdmin {
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

func (s *Server) render(w http.ResponseWriter, status int, name string, data any) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	_ = s.Tpl.ExecuteTemplate(w, name, data)
}

func (s *Server) Health(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = io.WriteString(w, "ok\n")
}

func (s *Server) MountStatic(mux *http.ServeMux) error {
	sub, err := fs.Sub(web.Assets, "static")
	if err != nil {
		return err
	}
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.FS(sub))))
	return nil
}

func (s *Server) LoginGet(w http.ResponseWriter, r *http.Request) {
	sh := s.shell(r, "Login")
	data := struct {
		shell
		Error string
		Next  string
	}{shell: sh, Next: r.URL.Query().Get("next")}
	s.render(w, http.StatusOK, "page_login", data)
}

func (s *Server) LoginPost(w http.ResponseWriter, r *http.Request) {
	sh := s.shell(r, "Login")
	u, err := s.Auth.Login(r.Context(), r.FormValue("username"), r.FormValue("password"))
	if err != nil {
		st := http.StatusOK
		if errors.Is(err, domain.ErrUnauthorized) {
			data := struct {
				shell
				Error string
				Next  string
			}{shell: sh, Error: "Invalid credentials", Next: r.FormValue("next")}
			s.render(w, st, "page_login", data)
			return
		}
		s.Log.Error("login", "err", err)
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	if err := middleware.IssueSessionCookie(w, s.JWTSecret, u, s.Config.CookieSecure); err != nil {
		s.Log.Error("issue cookie", "err", err)
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	next := r.FormValue("next")
	if next == "" || !strings.HasPrefix(next, "/") || strings.HasPrefix(next, "//") {
		next = "/"
	}
	http.Redirect(w, r, next, http.StatusSeeOther)
}

func (s *Server) LogoutPost(w http.ResponseWriter, r *http.Request) {
	middleware.ClearSessionCookie(w, s.Config.CookieSecure)
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func (s *Server) RegisterGet(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	n, err := s.Users.Count(ctx)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	invite := strings.TrimSpace(r.URL.Query().Get("token"))
	mode, hasMode, _ := s.AuthSettingsMode(ctx)
	if n > 0 {
		if !hasMode {
			mode = domain.RegistrationAdminOnly
		}
		switch mode {
		case domain.RegistrationAdminOnly:
			http.Error(w, "registration closed", http.StatusForbidden)
			return
		case domain.RegistrationInvite:
			if invite == "" {
				http.Error(w, "invite token required", http.StatusForbidden)
				return
			}
			if _, err := s.Invites.GetByToken(ctx, invite); err != nil {
				http.Error(w, "invalid invite", http.StatusForbidden)
				return
			}
		}
	}
	sh := s.shell(r, "Register")
	help := "Create your Findus account."
	if n == 0 {
		help = "You are creating the first admin account."
	} else if mode == domain.RegistrationInvite {
		help = "Register using your invite link."
	} else if mode == domain.RegistrationOpen {
		help = "Open registration is enabled."
	}
	data := struct {
		shell
		Error  string
		Help   string
		Invite string
	}{shell: sh, Help: help, Invite: invite}
	s.render(w, http.StatusOK, "page_register", data)
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

func (s *Server) RegisterPost(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	sh := s.shell(r, "Register")
	fail := func(msg string) {
		n, _ := s.Users.Count(ctx)
		mode, _, _ := s.AuthSettingsMode(ctx)
		help := "Create your Findus account."
		if n == 0 {
			help = "You are creating the first admin account."
		} else if mode == domain.RegistrationInvite {
			help = "Register using your invite link."
		} else if mode == domain.RegistrationOpen {
			help = "Open registration is enabled."
		}
		data := struct {
			shell
			Error  string
			Help   string
			Invite string
		}{shell: sh, Error: msg, Help: help, Invite: r.FormValue("invite")}
		s.render(w, http.StatusOK, "page_register", data)
	}
	u, err := s.Auth.Register(ctx, r.FormValue("username"), r.FormValue("email"), r.FormValue("password"), r.FormValue("invite"))
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrRegistrationClosed):
			fail("Registration is closed.")
		case errors.Is(err, domain.ErrInvalidInvite):
			fail("Invalid or expired invite.")
		default:
			if errors.Is(err, domain.ErrValidation) {
				fail(err.Error())
				return
			}
			s.Log.Error("register", "err", err)
			fail("Could not register.")
		}
		return
	}
	if err := middleware.IssueSessionCookie(w, s.JWTSecret, u, s.Config.CookieSecure); err != nil {
		s.Log.Error("issue cookie", "err", err)
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func recentlyCreated(createdAt, updatedAt time.Time) bool {
	d := updatedAt.Sub(createdAt)
	return d >= 0 && d < 2*time.Second
}

func (s *Server) Home(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	sh := s.shell(r, "Home")
	u, ok := middleware.User(ctx)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	const recentItemsLimit = 10

	var itemCount, locCount, labelCount int64
	recentItems, err := s.Items.ListRecentByUpdated(ctx, recentItemsLimit)
	if err != nil {
		s.Log.Error("home list recent items", "err", err)
		recentItems = nil
	}
	rootLocs, err := s.Locs.ListChildren(ctx, nil)
	if err != nil {
		s.Log.Error("home list root locations", "err", err)
		rootLocs = nil
	}
	childCounts, err := s.Locs.ChildCountsByParentID(ctx)
	if err != nil {
		s.Log.Error("home location child counts", "err", err)
		childCounts = nil
	}
	if n, err := s.Items.Count(ctx); err != nil {
		s.Log.Error("home item count", "err", err)
	} else {
		itemCount = n
	}
	if n, err := s.Locs.Count(ctx); err != nil {
		s.Log.Error("home location count", "err", err)
	} else {
		locCount = n
	}
	if s.Labels != nil {
		if n, err := s.Labels.Count(ctx); err != nil {
			s.Log.Error("home label count", "err", err)
		} else {
			labelCount = n
		}
	}

	var allLabels []domain.Label
	if s.Labels != nil {
		if lbs, err := s.Labels.ListAll(ctx); err != nil {
			s.Log.Error("home list labels", "err", err)
		} else {
			allLabels = lbs
		}
	}

	locNames := make(map[string]string)
	itemRows := make([]homeRecentItem, 0, len(recentItems))
	for _, it := range recentItems {
		name, ok := locNames[it.LocationID]
		if !ok {
			loc, err := s.Locs.GetByID(ctx, it.LocationID)
			if err != nil {
				name = "Unknown"
			} else {
				name = loc.Name
			}
			locNames[it.LocationID] = name
		}
		itemRows = append(itemRows, homeRecentItem{
			Item:          it,
			LocationName: name,
			RecentlyAdded: recentlyCreated(it.CreatedAt, it.UpdatedAt),
		})
	}

	locRows := make([]homeLocationRow, 0, len(rootLocs))
	for _, loc := range rootLocs {
		var subN int64
		if childCounts != nil {
			subN = childCounts[loc.ID]
		}
		locRows = append(locRows, homeLocationRow{
			Location:         loc,
			SubLocationCount: subN,
			RecentlyAdded:    recentlyCreated(loc.CreatedAt, loc.UpdatedAt),
		})
	}

	data := struct {
		shell
		User            *domain.User
		ItemCount       int64
		LocationCount   int64
		LabelCount      int64
		RecentItems     []homeRecentItem
		HomeLocations   []homeLocationRow
		AllLabels       []domain.Label
	}{
		shell:           sh,
		User:            u,
		ItemCount:       itemCount,
		LocationCount:   locCount,
		LabelCount:      labelCount,
		RecentItems:     itemRows,
		HomeLocations:   locRows,
		AllLabels:       allLabels,
	}
	s.render(w, http.StatusOK, "page_home", data)
}

func (s *Server) ProfileGet(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	u, ok := middleware.User(ctx)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	fresh, err := s.Users.GetByID(ctx, u.ID)
	if err == nil {
		u = fresh
	}
	sh := s.shell(r, "Profile")
	data := struct {
		shell
		Error string
		User  *domain.User
	}{shell: sh, User: u}
	s.render(w, http.StatusOK, "page_profile", data)
}

func (s *Server) profileRenderError(w http.ResponseWriter, r *http.Request, msg string) {
	ctx := r.Context()
	u, ok := middleware.User(ctx)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	fresh, err := s.Users.GetByID(ctx, u.ID)
	if err == nil {
		u = fresh
	}
	sh := s.shell(r, "Profile")
	data := struct {
		shell
		Error string
		User  *domain.User
	}{shell: sh, Error: msg, User: u}
	s.render(w, http.StatusOK, "page_profile", data)
}

func (s *Server) ProfilePost(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	u, ok := middleware.User(ctx)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	if r.MultipartForm == nil {
		if err := r.ParseMultipartForm(32 << 20); err != nil {
			s.profileRenderError(w, r, "Invalid form.")
			return
		}
	}
	u2, err := s.Auth.UpdateProfile(ctx, u.ID, r.FormValue("username"), r.FormValue("email"), r.FormValue("current_password"), r.FormValue("new_password"))
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidCurrentPassword):
			s.profileRenderError(w, r, "Current password is incorrect.")
		case errors.Is(err, domain.ErrValidation):
			s.profileRenderError(w, r, err.Error())
		case errors.Is(err, domain.ErrConflict):
			s.profileRenderError(w, r, "That username or email is already in use.")
		default:
			s.Log.Error("profile update", "err", err)
			s.profileRenderError(w, r, "Could not save profile.")
		}
		return
	}
	rel := filepath.ToSlash(filepath.Join("images", "user-"+u.ID+".webp"))
	if f, _, err := r.FormFile("avatar"); err == nil {
		defer f.Close()
		if err := s.saveUserAvatar(ctx, u.ID, f); err != nil {
			s.Log.Error("avatar", "err", err)
			s.profileRenderError(w, r, "Could not process profile image. Other changes were saved.")
			return
		}
		u2.AvatarPath = &rel
		u2.UpdatedAt = time.Now().UTC()
		if err := s.Users.Update(ctx, u2); err != nil {
			s.Log.Error("avatar db", "err", err)
			s.profileRenderError(w, r, "Could not save profile image path.")
			return
		}
	} else if r.FormValue("remove_avatar") != "" && u2.AvatarPath != nil && *u2.AvatarPath != "" {
		p := filepath.Join(s.Config.DataDir, filepath.FromSlash(*u2.AvatarPath))
		_ = os.Remove(p)
		u2.AvatarPath = nil
		u2.UpdatedAt = time.Now().UTC()
		if err := s.Users.Update(ctx, u2); err != nil {
			s.Log.Error("avatar remove", "err", err)
			s.profileRenderError(w, r, "Could not remove profile image.")
			return
		}
	}
	http.Redirect(w, r, "/profile", http.StatusSeeOther)
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
	http.NotFound(w, r)
}

func (s *Server) LocationsList(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	sh := s.shell(r, "Locations")
	roots, err := s.Locs.ListChildren(ctx, nil)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	data := struct {
		shell
		Roots []domain.Location
	}{shell: sh, Roots: roots}
	s.render(w, http.StatusOK, "page_locations", data)
}

func (s *Server) LocationNewGet(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	sh := s.shell(r, "New location")
	opts, err := s.parentLocationOptions(ctx, "")
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	pid := strings.TrimSpace(r.URL.Query().Get("parent_id"))
	data := struct {
		shell
		Error          string
		ParentOptions  []locOpt
		SelectedParent string
	}{shell: sh, ParentOptions: opts, SelectedParent: pid}
	s.render(w, http.StatusOK, "page_location_new", data)
}

func (s *Server) LocationCreatePost(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	parent := strings.TrimSpace(r.FormValue("parent_id"))
	var p *string
	if parent != "" {
		p = &parent
	}
	loc, err := s.Inventory.CreateLocation(ctx, r.FormValue("name"), r.FormValue("description"), p)
	if err != nil {
		sh := s.shell(r, "New location")
		opts, optsErr := s.parentLocationOptions(ctx, "")
		if optsErr != nil {
			http.Error(w, "server error", http.StatusInternalServerError)
			return
		}
		data := struct {
			shell
			Error          string
			ParentOptions  []locOpt
			SelectedParent string
		}{shell: sh, Error: err.Error(), ParentOptions: opts, SelectedParent: parent}
		s.render(w, http.StatusOK, "page_location_new", data)
		return
	}
	http.Redirect(w, r, "/locations/"+loc.ID, http.StatusSeeOther)
}

func (s *Server) LocationDetail(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := r.PathValue("id")
	loc, err := s.Locs.GetByID(ctx, id)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	path, err := s.Locs.PathFromRoot(ctx, id)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	var crumb []domain.LocationPathElement
	if len(path) > 0 {
		crumb = path[:len(path)-1]
	}
	backHref := "/locations"
	backLabel := "All locations"
	if len(crumb) > 0 {
		p := crumb[len(crumb)-1]
		backHref = "/locations/" + p.ID
		backLabel = "← " + p.Name
	}
	ch, err := s.Locs.ListChildren(ctx, &id)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	items, err := s.Items.ListByLocation(ctx, id)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	sh := s.shell(r, loc.Name)
	data := struct {
		shell
		Location   *domain.Location
		Breadcrumb []domain.LocationPathElement
		Children   []domain.Location
		Items      []domain.Item
		BackHref   string
		BackLabel  string
	}{shell: sh, Location: loc, Breadcrumb: crumb, Children: ch, Items: items, BackHref: backHref, BackLabel: backLabel}
	s.render(w, http.StatusOK, "page_location_detail", data)
}

func (s *Server) LocationEditGet(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := r.PathValue("id")
	loc, err := s.Locs.GetByID(ctx, id)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	parent := ""
	if loc.ParentID != nil {
		parent = *loc.ParentID
	}
	opts, err := s.parentLocationOptions(ctx, id)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	sh := s.shell(r, "Edit location")
	data := struct {
		shell
		Error          string
		Location       *domain.Location
		ParentOptions  []locOpt
		SelectedParent string
	}{shell: sh, Location: loc, ParentOptions: opts, SelectedParent: parent}
	s.render(w, http.StatusOK, "page_location_edit", data)
}

func (s *Server) LocationUpdatePost(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := r.PathValue("id")
	parent := strings.TrimSpace(r.FormValue("parent_id"))
	var p *string
	if parent != "" {
		p = &parent
	}
	_, err := s.Inventory.UpdateLocation(ctx, id, r.FormValue("name"), r.FormValue("description"), p)
	if err != nil {
		loc, _ := s.Locs.GetByID(ctx, id)
		opts, optsErr := s.parentLocationOptions(ctx, id)
		if optsErr != nil {
			http.Error(w, "server error", http.StatusInternalServerError)
			return
		}
		sh := s.shell(r, "Edit location")
		data := struct {
			shell
			Error          string
			Location       *domain.Location
			ParentOptions  []locOpt
			SelectedParent string
		}{shell: sh, Error: err.Error(), Location: loc, ParentOptions: opts, SelectedParent: parent}
		s.render(w, http.StatusOK, "page_location_edit", data)
		return
	}
	http.Redirect(w, r, "/locations/"+id, http.StatusSeeOther)
}

func (s *Server) LocationDeletePost(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := r.PathValue("id")
	if err := s.Inventory.DeleteLocation(ctx, id); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	http.Redirect(w, r, "/locations", http.StatusSeeOther)
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

func (s *Server) ItemsList(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	items, err := s.Items.ListAll(ctx, 500)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	sh := s.shell(r, "Items")
	data := struct {
		shell
		Items []domain.Item
	}{shell: sh, Items: items}
	s.render(w, http.StatusOK, "page_items", data)
}

type locOpt struct {
	ID    string
	Label string
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

// itemDetailSystemRow is a read-only row in the item detail table (IDs, timestamps, preset caption).
type itemDetailSystemRow struct {
	Label string
	Value string
}

type itemDetailPageData struct {
	shell
	Item            *domain.Item
	Labels          []domain.Label
	PhotoURL        string
	LocationPath    []domain.LocationPathElement
	EditMode        bool
	Error           string
	AttrRows        []service.ItemPropertyDetailRow
	SystemRows      []itemDetailSystemRow
	Locations       []locOpt
	AllLabels       []domain.Label
	SelectedLabels  map[string]bool
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

func (s *Server) renderItemNewNoLocation(w http.ResponseWriter, r *http.Request) {
	sh := s.shell(r, "New item")
	s.render(w, http.StatusOK, "page_item_new_no_location", struct {
		shell
	}{shell: sh})
}

func (s *Server) renderItemNew(w http.ResponseWriter, r *http.Request, selLoc, errMsg string) {
	ctx := r.Context()
	sh := s.shell(r, "New item")
	opts, err := s.flatLocations(ctx)
	if err != nil || len(opts) == 0 {
		s.renderItemNewNoLocation(w, r)
		return
	}
	if selLoc == "" {
		selLoc = opts[0].ID
	}
	labels, _ := s.Inventory.ListLabels(ctx)
	templates, _ := s.Inventory.ListItemTemplates(ctx)
	defTpl := ""
	if len(templates) > 0 {
		defTpl = templates[0].ID
	}
	s.render(w, http.StatusOK, "page_item_new", struct {
		shell
		Locations         []locOpt
		SelectedLocation  string
		Error             string
		AllLabels         []domain.Label
		Templates         []domain.ItemTemplate
		DefaultTemplateID string
	}{shell: sh, Locations: opts, SelectedLocation: selLoc, Error: errMsg, AllLabels: labels, Templates: templates, DefaultTemplateID: defTpl})
}

func (s *Server) renderItemEdit(w http.ResponseWriter, r *http.Request, id, errMsg string) {
	ctx := r.Context()
	it, err := s.Items.GetByID(ctx, id)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	opts, err := s.flatLocations(ctx)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	itemLabels, err := s.Items.ListLabelsForItem(ctx, id)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	selected := make(map[string]bool)
	for _, lb := range itemLabels {
		selected[lb.ID] = true
	}
	allLabels, _ := s.Inventory.ListLabels(ctx)
	sh := s.shell(r, "Edit item")
	s.render(w, http.StatusOK, "page_item_edit", struct {
		shell
		Item           *domain.Item
		Error          string
		Locations      []locOpt
		AllLabels      []domain.Label
		SelectedLabels map[string]bool
		AdditionalRows []additionalKV
	}{shell: sh, Item: it, Error: errMsg, Locations: opts, AllLabels: allLabels, SelectedLabels: selected, AdditionalRows: additionalFieldPairs(mergedAdditionalForEdit(it))})
}

func mergedAdditionalForEdit(it *domain.Item) json.RawMessage {
	merged, err := service.MergeTemplateIntoAdditional(it.AdditionalData, it.TemplateData)
	if err != nil {
		return it.AdditionalData
	}
	return merged
}

func (s *Server) buildItemDetailPageData(r *http.Request, sh shell, it *domain.Item, editMode bool, formErr string) itemDetailPageData {
	ctx := r.Context()
	labels, _ := s.Items.ListLabelsForItem(ctx, it.ID)
	labelIDs := make([]string, len(labels))
	for i := range labels {
		labelIDs[i] = labels[i].ID
	}
	selectedLabels := make(map[string]bool, len(labelIDs))
	for _, id := range labelIDs {
		selectedLabels[id] = true
	}
	opts, locErr := s.flatLocations(ctx)
	if locErr != nil {
		opts = nil
	}
	allLabels, _ := s.Inventory.ListLabels(ctx)
	tpl, tplErr := s.Inventory.GetItemTemplate(ctx, string(it.TemplateType))
	if tplErr != nil {
		tpl = nil
	}
	merged, err := service.MergeTemplateIntoAdditional(it.AdditionalData, it.TemplateData)
	if err != nil {
		s.Log.Error("item merge", "err", err)
		merged = it.AdditionalData
	}
	attrRows := service.ItemPropertyDetailRows(tpl, merged)
	tplCaption := strings.TrimSpace(string(it.TemplateType))
	if tpl != nil && strings.TrimSpace(tpl.DisplayName) != "" {
		tplCaption = tpl.DisplayName
	}
	const tsLayout = "2 Jan 2006, 15:04 MST"
	systemRows := []itemDetailSystemRow{
		{"Item ID", it.ID},
		{"Created", it.CreatedAt.UTC().Format(tsLayout)},
		{"Last updated", it.UpdatedAt.UTC().Format(tsLayout)},
		{"Initial setup", tplCaption},
	}
	locPath, _ := s.Locs.PathFromRoot(ctx, it.LocationID)
	photo := ""
	if it.PhotoPath != nil && *it.PhotoPath != "" {
		photo = "/items/" + it.ID + "/photo"
	}
	return itemDetailPageData{
		shell: sh, Item: it, Labels: labels, PhotoURL: photo,
		LocationPath: locPath,
		EditMode: editMode, Error: formErr,
		AttrRows: attrRows, SystemRows: systemRows,
		Locations: opts, AllLabels: allLabels, SelectedLabels: selectedLabels,
	}
}

func (s *Server) renderItemUpdateError(w http.ResponseWriter, r *http.Request, id, msg string) {
	ctx := r.Context()
	it, err := s.Items.GetByID(ctx, id)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	if strings.TrimSpace(r.FormValue("return_to")) == "detail" {
		sh := s.shell(r, it.Name)
		d := s.buildItemDetailPageData(r, sh, it, true, msg)
		s.render(w, http.StatusOK, "page_item_detail", d)
		return
	}
	s.renderItemEdit(w, r, id, msg)
}

func (s *Server) ItemNewGet(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	opts, err := s.flatLocations(ctx)
	if err != nil || len(opts) == 0 {
		s.renderItemNewNoLocation(w, r)
		return
	}
	sel := strings.TrimSpace(r.URL.Query().Get("location_id"))
	if sel == "" {
		sel = opts[0].ID
	}
	labels, _ := s.Inventory.ListLabels(ctx)
	templates, _ := s.Inventory.ListItemTemplates(ctx)
	defTpl := ""
	if len(templates) > 0 {
		defTpl = templates[0].ID
	}
	sh := s.shell(r, "New item")
	data := struct {
		shell
		Locations         []locOpt
		SelectedLocation  string
		Error             string
		AllLabels         []domain.Label
		Templates         []domain.ItemTemplate
		DefaultTemplateID string
	}{shell: sh, Locations: opts, SelectedLocation: sel, AllLabels: labels, Templates: templates, DefaultTemplateID: defTpl}
	s.render(w, http.StatusOK, "page_item_new", data)
}

func (s *Server) ItemFieldsGet(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tpl := strings.TrimSpace(r.URL.Query().Get("template_type"))
	if tpl == "" {
		list, _ := s.Inventory.ListItemTemplates(ctx)
		if len(list) > 0 {
			tpl = list[0].ID
		}
	}
	t, err := s.Inventory.GetItemTemplate(ctx, tpl)
	if err != nil {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(`<p class="text-sm text-red-600">Unknown template.</p>`))
		return
	}
	s.render(w, http.StatusOK, "page_item_fields", struct {
		Fields []domain.TemplateField
	}{Fields: t.Fields})
}

func (s *Server) ItemCreatePost(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		http.Error(w, "bad form", http.StatusBadRequest)
		return
	}
	tt, err := s.Inventory.GetItemTemplate(ctx, r.FormValue("template_type"))
	if err != nil {
		s.renderItemNew(w, r, r.FormValue("location_id"), "Invalid template type.")
		return
	}
	add, err := service.ParseAdditionalFromForm(r)
	if err != nil {
		s.renderItemNew(w, r, r.FormValue("location_id"), err.Error())
		return
	}
	form := make(map[string]string)
	for _, f := range tt.Fields {
		form[f.Key] = r.FormValue(f.Key)
	}
	td, err := domain.NormalizeTemplateDataFromFields(tt.Fields, form)
	if err != nil {
		s.renderItemNew(w, r, r.FormValue("location_id"), err.Error())
		return
	}
	merged, err := service.MergeTemplateIntoAdditional(add, td)
	if err != nil {
		s.renderItemNew(w, r, r.FormValue("location_id"), err.Error())
		return
	}
	labelIDs := parseLabelIDs(r)
	emptyTD := json.RawMessage(`{}`)
	it, err := s.Inventory.CreateItem(ctx, r.FormValue("name"), r.FormValue("description"), r.FormValue("location_id"), domain.TemplateType(tt.ID), emptyTD, merged, labelIDs)
	if err != nil {
		s.renderItemNew(w, r, r.FormValue("location_id"), err.Error())
		return
	}
	if f, hdr, err := r.FormFile("photo"); err == nil {
		defer f.Close()
		_ = hdr
		if err := s.saveItemPhoto(ctx, it.ID, f); err != nil {
			s.Log.Error("photo", "err", err)
		} else {
			rel := filepath.ToSlash(filepath.Join("images", it.ID+".webp"))
			it.PhotoPath = &rel
			it.UpdatedAt = time.Now().UTC()
			_ = s.Items.Update(ctx, it)
		}
	}
	http.Redirect(w, r, "/items/"+it.ID, http.StatusSeeOther)
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

func (s *Server) ItemDetail(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := r.PathValue("id")
	it, err := s.Items.GetByID(ctx, id)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	sh := s.shell(r, it.Name)
	editMode := r.URL.Query().Get("edit") == "1"
	if editMode && !sh.IsAdmin {
		http.Redirect(w, r, "/items/"+id, http.StatusSeeOther)
		return
	}
	d := s.buildItemDetailPageData(r, sh, it, editMode, "")
	s.render(w, http.StatusOK, "page_item_detail", d)
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

func (s *Server) ItemEditGet(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := r.PathValue("id")
	it, err := s.Items.GetByID(ctx, id)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	opts, err := s.flatLocations(ctx)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	itemLabels, err := s.Items.ListLabelsForItem(ctx, id)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	selected := make(map[string]bool)
	for _, lb := range itemLabels {
		selected[lb.ID] = true
	}
	allLabels, _ := s.Inventory.ListLabels(ctx)
	sh := s.shell(r, "Edit item")
	data := struct {
		shell
		Item           *domain.Item
		Error          string
		Locations      []locOpt
		AllLabels      []domain.Label
		SelectedLabels map[string]bool
		AdditionalRows []additionalKV
	}{shell: sh, Item: it, Locations: opts, AllLabels: allLabels, SelectedLabels: selected, AdditionalRows: additionalFieldPairs(mergedAdditionalForEdit(it))}
	s.render(w, http.StatusOK, "page_item_edit", data)
}

func (s *Server) ItemUpdatePost(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := r.PathValue("id")
	it, err := s.Items.GetByID(ctx, id)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		http.Error(w, "bad form", http.StatusBadRequest)
		return
	}
	tplID := strings.TrimSpace(r.FormValue("template_type"))
	if _, err := s.Inventory.GetItemTemplate(ctx, tplID); err != nil {
		s.renderItemUpdateError(w, r, id, "Invalid template type.")
		return
	}
	tt := domain.TemplateType(tplID)
	if tt != it.TemplateType {
		s.renderItemUpdateError(w, r, id, "Template type cannot be changed here.")
		return
	}
	add, err := service.ParseAdditionalFromForm(r)
	if err != nil {
		s.renderItemUpdateError(w, r, id, err.Error())
		return
	}
	labelIDs := parseLabelIDs(r)
	emptyTD := json.RawMessage(`{}`)
	it2, err := s.Inventory.UpdateItem(ctx, id, r.FormValue("name"), r.FormValue("description"), r.FormValue("location_id"), tt, emptyTD, add, labelIDs)
	if err != nil {
		s.renderItemUpdateError(w, r, id, err.Error())
		return
	}
	if f, hdr, err := r.FormFile("photo"); err == nil {
		defer f.Close()
		_ = hdr
		if err := s.saveItemPhoto(ctx, id, f); err != nil {
			s.Log.Error("photo", "err", err)
		} else {
			rel := filepath.ToSlash(filepath.Join("images", id+".webp"))
			it2.PhotoPath = &rel
			it2.UpdatedAt = time.Now().UTC()
			_ = s.Items.Update(ctx, it2)
		}
	}
	http.Redirect(w, r, "/items/"+id, http.StatusSeeOther)
}

func (s *Server) ItemDeletePost(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := r.PathValue("id")
	it, err := s.Items.GetByID(ctx, id)
	if err == nil && it.PhotoPath != nil {
		_ = os.Remove(filepath.Join(s.Config.DataDir, filepath.FromSlash(*it.PhotoPath)))
	}
	if err := s.Inventory.DeleteItem(ctx, id); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	http.Redirect(w, r, "/items", http.StatusSeeOther)
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

func (s *Server) Search(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	res, err := s.Items.Search(ctx, q, 50)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	if r.Header.Get("HX-Request") == "true" {
		s.render(w, http.StatusOK, "page_search", struct{ Results []domain.Item }{Results: res})
		return
	}
	sh := s.shell(r, "Search")
	data := struct {
		shell
		Query   string
		Results []domain.Item
	}{shell: sh, Query: q, Results: res}
	s.render(w, http.StatusOK, "page_search_page", data)
}

func (s *Server) AdminHome(w http.ResponseWriter, r *http.Request) {
	sh := s.shell(r, "Admin")
	s.render(w, http.StatusOK, "page_admin", sh)
}

// renderAdminUserManagement loads users, invites, and registration mode and renders the unified user management page.
func (s *Server) renderAdminUserManagement(w http.ResponseWriter, r *http.Request, pageError string) bool {
	ctx := r.Context()
	us, err := s.Admin.ListUsers(ctx)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return false
	}
	list, err := s.Admin.ListInvites(ctx)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return false
	}
	mode, err := s.Admin.GetRegistrationMode(ctx)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return false
	}
	sh := s.shell(r, "User management")
	s.render(w, http.StatusOK, "page_admin_user_management", struct {
		shell
		Users   []domain.User
		Invites []domain.Invite
		Mode    string
		Error   string
	}{shell: sh, Users: us, Invites: list, Mode: string(mode), Error: pageError})
	return true
}

func (s *Server) AdminUsers(w http.ResponseWriter, r *http.Request) {
	s.renderAdminUserManagement(w, r, "")
}

func (s *Server) AdminUsersCreatePost(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var role domain.Role
	switch r.FormValue("role") {
	case "admin":
		role = domain.RoleAdmin
	case "user":
		role = domain.RoleUser
	default:
		http.Error(w, "bad role", http.StatusBadRequest)
		return
	}
	if _, err := s.Admin.CreateUser(ctx, r.FormValue("username"), r.FormValue("email"), r.FormValue("password"), role); err != nil {
		s.renderAdminUserManagement(w, r, err.Error())
		return
	}
	http.Redirect(w, r, "/admin/users", http.StatusSeeOther)
}

func (s *Server) AdminUserRolePost(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := r.PathValue("id")
	role := domain.Role(r.FormValue("role"))
	if role != domain.RoleAdmin && role != domain.RoleUser {
		http.Error(w, "bad role", http.StatusBadRequest)
		return
	}
	_ = s.Admin.SetUserRole(ctx, id, role)
	http.Redirect(w, r, "/admin/users", http.StatusSeeOther)
}

func (s *Server) AdminUserActivePost(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := r.PathValue("id")
	active := r.FormValue("active") == "1"
	_ = s.Admin.SetUserActive(ctx, id, active)
	http.Redirect(w, r, "/admin/users", http.StatusSeeOther)
}

func (s *Server) AdminInvites(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/admin/users", http.StatusSeeOther)
}

func (s *Server) AdminInvitesCreatePost(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	role := domain.Role(r.FormValue("role"))
	if role != domain.RoleAdmin && role != domain.RoleUser {
		http.Error(w, "bad role", http.StatusBadRequest)
		return
	}
	h, _ := strconv.Atoi(strings.TrimSpace(r.FormValue("ttl_hours")))
	ttl := time.Duration(h) * time.Hour
	u, ok := middleware.User(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	if _, err := s.Admin.CreateInvite(ctx, u.ID, role, ttl); err != nil {
		s.renderAdminUserManagement(w, r, err.Error())
		return
	}
	http.Redirect(w, r, "/admin/users", http.StatusSeeOther)
}

func (s *Server) AdminSettings(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/admin/users", http.StatusSeeOther)
}

func (s *Server) AdminSettingsPost(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	m, ok := domain.ParseRegistrationMode(r.FormValue("mode"))
	if !ok {
		http.Error(w, "bad mode", http.StatusBadRequest)
		return
	}
	if err := s.Admin.SetRegistrationMode(ctx, m); err != nil {
		s.renderAdminUserManagement(w, r, err.Error())
		return
	}
	http.Redirect(w, r, "/admin/users", http.StatusSeeOther)
}

func (s *Server) LabelsList(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	list, err := s.Inventory.ListLabels(ctx)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	sh := s.shell(r, "Labels")
	tpls, _ := s.Inventory.ListItemTemplates(ctx)
	titles := make(map[string]string, len(tpls))
	for _, t := range tpls {
		titles[t.ID] = t.DisplayName
	}
	rows := s.labelRows(sh, list, titles)
	s.render(w, http.StatusOK, "page_labels", struct {
		shell
		LabelRows []labelRow
	}{shell: sh, LabelRows: rows})
}

func (s *Server) LabelNewGet(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	templates, _ := s.Inventory.ListItemTemplates(ctx)
	sh := s.shell(r, "New label")
	s.render(w, http.StatusOK, "page_label_new", struct {
		shell
		Error     string
		Templates []domain.ItemTemplate
	}{shell: sh, Templates: templates})
}

func (s *Server) LabelCreatePost(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	rawTpl := strings.TrimSpace(r.FormValue("default_template_type"))
	var def *string
	if rawTpl != "" {
		def = &rawTpl
	}
	if _, err := s.Inventory.CreateLabel(ctx, r.FormValue("name"), r.FormValue("color"), def); err != nil {
		templates, _ := s.Inventory.ListItemTemplates(ctx)
		sh := s.shell(r, "New label")
		s.render(w, http.StatusOK, "page_label_new", struct {
			shell
			Error     string
			Templates []domain.ItemTemplate
		}{shell: sh, Error: err.Error(), Templates: templates})
		return
	}
	http.Redirect(w, r, "/labels", http.StatusSeeOther)
}

func (s *Server) LabelEditGet(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := r.PathValue("id")
	lb, err := s.Labels.GetByID(ctx, id)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	sh := s.shell(r, "Edit label")
	sel := ""
	if lb.DefaultTemplateType != nil {
		sel = *lb.DefaultTemplateType
	}
	templates, _ := s.Inventory.ListItemTemplates(ctx)
	s.render(w, http.StatusOK, "page_label_edit", struct {
		shell
		Label            *domain.Label
		Error            string
		SelectedTemplate string
		Templates        []domain.ItemTemplate
	}{shell: sh, Label: lb, SelectedTemplate: sel, Templates: templates})
}

func (s *Server) LabelUpdatePost(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := r.PathValue("id")
	rawTpl := strings.TrimSpace(r.FormValue("default_template_type"))
	var def *string
	if rawTpl != "" {
		def = &rawTpl
	}
	if _, err := s.Inventory.UpdateLabel(ctx, id, r.FormValue("name"), r.FormValue("color"), def); err != nil {
		lb, _ := s.Labels.GetByID(ctx, id)
		sh := s.shell(r, "Edit label")
		sel := ""
		if lb != nil && lb.DefaultTemplateType != nil {
			sel = *lb.DefaultTemplateType
		}
		templates, _ := s.Inventory.ListItemTemplates(ctx)
		s.render(w, http.StatusOK, "page_label_edit", struct {
			shell
			Label            *domain.Label
			Error            string
			SelectedTemplate string
			Templates        []domain.ItemTemplate
		}{shell: sh, Label: lb, Error: err.Error(), SelectedTemplate: sel, Templates: templates})
		return
	}
	http.Redirect(w, r, "/labels", http.StatusSeeOther)
}

func (s *Server) LabelDeletePost(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := r.PathValue("id")
	if err := s.Inventory.DeleteLabel(ctx, id); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	http.Redirect(w, r, "/labels", http.StatusSeeOther)
}

func (s *Server) AdminLabelsRedirect(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/labels", http.StatusSeeOther)
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

func (s *Server) Handler() (http.Handler, error) {
	mux := http.NewServeMux()
	if err := s.MountStatic(mux); err != nil {
		return nil, err
	}

	mux.HandleFunc("GET /healthz", s.Health)

	mux.HandleFunc("GET /login", s.LoginGet)
	mux.HandleFunc("POST /login", s.LoginPost)
	mux.HandleFunc("GET /register", s.RegisterGet)
	mux.HandleFunc("POST /register", s.RegisterPost)

	mux.Handle("POST /logout", middleware.RequireAuth(http.HandlerFunc(s.LogoutPost)))

	auth := func(h http.HandlerFunc) http.Handler { return middleware.RequireAuth(h) }
	admin := func(h http.HandlerFunc) http.Handler {
		return middleware.RequireAuth(middleware.RequireAdmin(h))
	}

	mux.Handle("GET /{$}", auth(s.Home))

	mux.Handle("GET /profile", auth(http.HandlerFunc(s.ProfileGet)))
	mux.Handle("POST /profile", auth(http.HandlerFunc(s.ProfilePost)))
	mux.Handle("GET /profile/photo", auth(http.HandlerFunc(s.ProfilePhoto)))

	mux.Handle("GET /locations", auth(http.HandlerFunc(s.LocationsList)))
	mux.Handle("GET /locations/new", admin(http.HandlerFunc(s.LocationNewGet)))
	mux.Handle("POST /locations", admin(http.HandlerFunc(s.LocationCreatePost)))
	mux.Handle("GET /locations/{id}", auth(http.HandlerFunc(s.LocationDetail)))
	mux.Handle("GET /locations/{id}/edit", admin(http.HandlerFunc(s.LocationEditGet)))
	mux.Handle("POST /locations/{id}", admin(http.HandlerFunc(s.LocationUpdatePost)))
	mux.Handle("POST /locations/{id}/delete", admin(http.HandlerFunc(s.LocationDeletePost)))
	mux.Handle("GET /locations/{id}/qr.png", auth(http.HandlerFunc(s.LocationQR)))

	mux.Handle("GET /items", auth(http.HandlerFunc(s.ItemsList)))
	mux.Handle("GET /items/new", admin(http.HandlerFunc(s.ItemNewGet)))
	mux.Handle("GET /items/new/fields", admin(http.HandlerFunc(s.ItemFieldsGet)))
	mux.Handle("POST /items", admin(http.HandlerFunc(s.ItemCreatePost)))
	mux.Handle("GET /items/{id}", auth(http.HandlerFunc(s.ItemDetail)))
	mux.Handle("GET /items/{id}/photo", auth(http.HandlerFunc(s.ItemPhoto)))
	mux.Handle("GET /items/{id}/edit", admin(http.HandlerFunc(s.ItemEditGet)))
	mux.Handle("POST /items/{id}", admin(http.HandlerFunc(s.ItemUpdatePost)))
	mux.Handle("POST /items/{id}/delete", admin(http.HandlerFunc(s.ItemDeletePost)))
	mux.Handle("GET /items/{id}/qr.png", auth(http.HandlerFunc(s.ItemQR)))

	mux.HandleFunc("GET /q/{token}", s.QRedirect)

	mux.Handle("GET /search", auth(http.HandlerFunc(s.Search)))

	mux.Handle("GET /labels", auth(http.HandlerFunc(s.LabelsList)))
	mux.Handle("GET /labels/new", admin(http.HandlerFunc(s.LabelNewGet)))
	mux.Handle("POST /labels", admin(http.HandlerFunc(s.LabelCreatePost)))
	mux.Handle("GET /labels/{id}/edit", admin(http.HandlerFunc(s.LabelEditGet)))
	mux.Handle("POST /labels/{id}", admin(http.HandlerFunc(s.LabelUpdatePost)))
	mux.Handle("POST /labels/{id}/delete", admin(http.HandlerFunc(s.LabelDeletePost)))

	mux.Handle("GET /admin", admin(http.HandlerFunc(s.AdminHome)))
	mux.Handle("GET /admin/users", admin(http.HandlerFunc(s.AdminUsers)))
	mux.Handle("POST /admin/users", admin(http.HandlerFunc(s.AdminUsersCreatePost)))
	mux.Handle("POST /admin/users/{id}/role", admin(http.HandlerFunc(s.AdminUserRolePost)))
	mux.Handle("POST /admin/users/{id}/active", admin(http.HandlerFunc(s.AdminUserActivePost)))
	mux.Handle("GET /admin/invites", admin(http.HandlerFunc(s.AdminInvites)))
	mux.Handle("POST /admin/invites", admin(http.HandlerFunc(s.AdminInvitesCreatePost)))
	mux.Handle("GET /admin/settings", admin(http.HandlerFunc(s.AdminSettings)))
	mux.Handle("POST /admin/settings/registration", admin(http.HandlerFunc(s.AdminSettingsPost)))
	mux.Handle("GET /admin/labels", admin(http.HandlerFunc(s.AdminLabelsRedirect)))
	mux.Handle("GET /admin/templates", admin(http.HandlerFunc(s.AdminTemplates)))
	mux.Handle("GET /admin/templates/new", admin(http.HandlerFunc(s.AdminTemplateNewGet)))
	mux.Handle("POST /admin/templates/new", admin(http.HandlerFunc(s.AdminTemplateNewPost)))
	mux.Handle("GET /admin/templates/{id}/edit", admin(http.HandlerFunc(s.AdminTemplateEditGet)))
	mux.Handle("POST /admin/templates/{id}/edit", admin(http.HandlerFunc(s.AdminTemplateEditPost)))
	mux.Handle("POST /admin/templates/{id}/delete", admin(http.HandlerFunc(s.AdminTemplateDeletePost)))
	mux.Handle("GET /admin/backup.zip", admin(http.HandlerFunc(s.AdminBackup)))

	return middleware.Chain(mux,
		middleware.AuthOptional(s.Users, s.JWTSecret, s.Config.CookieSecure),
		middleware.CSRF(s.Config.CookieSecure),
		middleware.RequestLog(s.Log),
		middleware.Recover(s.Log),
	), nil
}
