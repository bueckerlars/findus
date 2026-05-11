package render

import (
	"html/template"
	"io/fs"
	"net/url"
	"strings"
	"unicode/utf8"

	"findus/internal/domain"
)

func Parse(fsys fs.FS) (*template.Template, error) {
	funcs := template.FuncMap{
		"isAdmin": func(u *domain.User) bool {
			return u != nil && u.Role.IsAdmin()
		},
		"initial": func(username string) string {
			username = strings.TrimSpace(username)
			if username == "" {
				return "?"
			}
			r, _ := utf8.DecodeRuneInString(username)
			if r == utf8.RuneError {
				return "?"
			}
			return strings.ToUpper(string(r))
		},
		// navActive reports whether currentPath belongs to a nav section (exact or nested route).
		"navActive": func(currentPath, section string) bool {
			currentPath = strings.TrimSpace(currentPath)
			if section == "/" {
				return currentPath == "" || currentPath == "/"
			}
			if currentPath == section {
				return true
			}
			return strings.HasPrefix(currentPath, section+"/")
		},
		"urlQuery": url.QueryEscape,
	}
	root := template.New("").Funcs(funcs)
	err := fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(strings.ToLower(path), ".html") {
			return nil
		}
		b, err := fs.ReadFile(fsys, path)
		if err != nil {
			return err
		}
		_, err = root.Parse(string(b))
		return err
	})
	return root, err
}
