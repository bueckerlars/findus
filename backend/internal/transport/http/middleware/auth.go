package middleware

import (
	"net/http"
	"net/url"
	"time"

	"findus/backend/internal/authjwt"
	"findus/backend/internal/domain"
	"findus/backend/internal/repository"
)

const sessionCookie = "findus_session"

func AuthOptional(users repository.UserRepository, secret []byte, secure bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c, err := r.Cookie(sessionCookie)
			if err != nil || c.Value == "" {
				next.ServeHTTP(w, r)
				return
			}
			claims, err := authjwt.Parse(secret, c.Value)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}
			u, err := users.GetByID(r.Context(), claims.UserID)
			if err != nil || !u.IsActive {
				next.ServeHTTP(w, r)
				return
			}
			if string(u.Role) != claims.Role {
				next.ServeHTTP(w, r)
				return
			}
			ctx := WithUser(r.Context(), u)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := User(r.Context()); !ok {
			if r.Header.Get("HX-Request") == "true" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			http.Redirect(w, r, "/login?next="+url.QueryEscape(r.URL.RequestURI()), http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u, ok := User(r.Context())
		if !ok {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		if !u.Role.IsAdmin() {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// IssueSessionCookie sets JWT cookie (call after successful login).
func IssueSessionCookie(w http.ResponseWriter, secret []byte, u *domain.User, secure bool) error {
	token, err := authjwt.Sign(secret, u.ID, u.Role, 7*24*time.Hour)
	if err != nil {
		return err
	}
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookie,
		Value:    token,
		Path:     "/",
		MaxAge:   86400 * 7,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	})
	return nil
}

func ClearSessionCookie(w http.ResponseWriter, secure bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookie,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	})
}
