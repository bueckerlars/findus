package domain

// UI theme IDs persisted in users.ui_theme and sent as JSON field "theme".

const DefaultUITheme = "default"

var allowedUIThemes = map[string]struct{}{
	DefaultUITheme: {},
	"ocean":        {},
	"forest":       {},
	"amber":        {},
	"rose":         {},
	"night":        {},
	"ocean-night":  {},
	"forest-night": {},
	"amber-night":  {},
	"rose-night":   {},
}

// ValidUITheme reports whether id is an allowed persisted theme slug.
func ValidUITheme(id string) bool {
	_, ok := allowedUIThemes[id]
	return ok
}

// NormalizeUITheme returns id if valid, otherwise DefaultUITheme.
func NormalizeUITheme(id string) string {
	if ValidUITheme(id) {
		return id
	}
	return DefaultUITheme
}
