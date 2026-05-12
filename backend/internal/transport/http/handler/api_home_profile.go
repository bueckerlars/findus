package handler

import (
	"context"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"findus/backend/internal/domain"
	"findus/backend/internal/transport/http/middleware"
)

// APIHome returns dashboard payload (same data as legacy Home template).
func (s *Server) APIHome(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	u, ok := middleware.User(ctx)
	if !ok {
		s.writeJSONError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	const recentItemsLimit = 10
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
	itemCount, _ := s.Items.Count(ctx)
	locCount, _ := s.Locs.Count(ctx)
	var labelCount int64
	var allLabels []domain.Label
	if s.Labels != nil {
		labelCount, _ = s.Labels.Count(ctx)
		allLabels, _ = s.Labels.ListAll(ctx)
	}
	locNames := make(map[string]string)
	type recentRow struct {
		Item             domain.Item `json:"item"`
		LocationName     string      `json:"location_name"`
		RecentlyAdded    bool        `json:"recently_added"`
	}
	itemRows := make([]recentRow, 0, len(recentItems))
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
		itemRows = append(itemRows, recentRow{
			Item: it, LocationName: name,
			RecentlyAdded: recentlyCreated(it.CreatedAt, it.UpdatedAt),
		})
	}
	type locRow struct {
		Location         domain.Location `json:"location"`
		SubLocationCount int64           `json:"sub_location_count"`
		RecentlyAdded    bool            `json:"recently_added"`
	}
	locRows := make([]locRow, 0, len(rootLocs))
	for _, loc := range rootLocs {
		var subN int64
		if childCounts != nil {
			subN = childCounts[loc.ID]
		}
		locRows = append(locRows, locRow{
			Location: loc, SubLocationCount: subN,
			RecentlyAdded: recentlyCreated(loc.CreatedAt, loc.UpdatedAt),
		})
	}
	s.writeJSON(w, http.StatusOK, map[string]any{
		"user":             apiUserFrom(u),
		"item_count":       itemCount,
		"location_count":   locCount,
		"label_count":      labelCount,
		"recent_items":     itemRows,
		"home_locations":   locRows,
		"all_labels":       allLabels,
	})
}

// APIProfileGet returns current profile.
func (s *Server) APIProfileGet(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	u, ok := middleware.User(ctx)
	if !ok {
		s.writeJSONError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	if fresh, err := s.Users.GetByID(ctx, u.ID); err == nil {
		u = fresh
	}
	s.writeJSON(w, http.StatusOK, map[string]any{"user": apiUserFrom(u)})
}

type apiProfilePostReq struct {
	Username        string `json:"username"`
	Email           string `json:"email"`
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
	RemoveAvatar    bool   `json:"remove_avatar"`
}

// APIProfilePost updates profile; accepts JSON or multipart (avatar file field "avatar").
func (s *Server) APIProfilePost(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	u, ok := middleware.User(ctx)
	if !ok {
		s.writeJSONError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	ct := r.Header.Get("Content-Type")
	var username, email, curPwd, newPwd string
	var removeAvatar bool
	if strings.HasPrefix(ct, "application/json") {
		var req apiProfilePostReq
		if err := readJSON(r, &req); err != nil {
			s.writeJSONError(w, http.StatusBadRequest, "invalid json")
			return
		}
		username, email, curPwd, newPwd = req.Username, req.Email, req.CurrentPassword, req.NewPassword
		removeAvatar = req.RemoveAvatar
	} else {
		if err := r.ParseMultipartForm(32 << 20); err != nil {
			s.writeJSONError(w, http.StatusBadRequest, "invalid form")
			return
		}
		username = r.FormValue("username")
		email = r.FormValue("email")
		curPwd = r.FormValue("current_password")
		newPwd = r.FormValue("new_password")
		removeAvatar = r.FormValue("remove_avatar") != ""
	}
	u2, err := s.Auth.UpdateProfile(ctx, u.ID, username, email, curPwd, newPwd)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidCurrentPassword):
			s.writeJSONError(w, http.StatusBadRequest, "Current password is incorrect.")
		case errors.Is(err, domain.ErrValidation):
			s.writeJSONError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, domain.ErrConflict):
			s.writeJSONError(w, http.StatusConflict, "That username or email is already in use.")
		default:
			s.Log.Error("profile update", "err", err)
			s.writeJSONError(w, http.StatusInternalServerError, "Could not save profile.")
		}
		return
	}
	rel := filepath.ToSlash(filepath.Join("images", "user-"+u.ID+".webp"))
	if strings.HasPrefix(ct, "multipart/form-data") {
		if f, _, err := r.FormFile("avatar"); err == nil {
			defer f.Close()
			if err := s.saveUserAvatar(ctx, u.ID, f); err != nil {
				s.Log.Error("avatar", "err", err)
				s.writeJSONError(w, http.StatusBadRequest, "Could not process profile image. Other changes were saved.")
				return
			}
			u2.AvatarPath = &rel
			u2.UpdatedAt = time.Now().UTC()
			if err := s.Users.Update(ctx, u2); err != nil {
				s.Log.Error("avatar db", "err", err)
				s.writeJSONError(w, http.StatusInternalServerError, "Could not save profile image path.")
				return
			}
		} else if removeAvatar && u2.AvatarPath != nil && *u2.AvatarPath != "" {
			s.removeUserAvatarFile(ctx, u2)
		}
	} else if removeAvatar && u2.AvatarPath != nil && *u2.AvatarPath != "" {
		s.removeUserAvatarFile(ctx, u2)
	}
	s.writeJSON(w, http.StatusOK, map[string]any{"user": apiUserFrom(u2)})
}

func (s *Server) removeUserAvatarFile(ctx context.Context, u2 *domain.User) {
	p := filepath.Join(s.Config.DataDir, filepath.FromSlash(*u2.AvatarPath))
	_ = os.Remove(p)
	u2.AvatarPath = nil
	u2.UpdatedAt = time.Now().UTC()
	if err := s.Users.Update(ctx, u2); err != nil {
		s.Log.Error("avatar remove", "err", err)
	}
}
