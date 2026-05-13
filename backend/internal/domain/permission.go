package domain

// Permission is a fixed capability string stored in perm_group_permissions.
type Permission string

const (
	PermItemsWrite     Permission = "items.write"
	PermLabelsWrite    Permission = "labels.write"
	PermLocationsWrite Permission = "locations.write"
)

var allPermissions = []Permission{
	PermItemsWrite,
	PermLabelsWrite,
	PermLocationsWrite,
}

// AllPermissions returns every defined permission (e.g. for admin /api/me).
func AllPermissions() []Permission {
	out := make([]Permission, len(allPermissions))
	copy(out, allPermissions)
	return out
}

// ValidPermission reports whether p is a known permission constant.
func ValidPermission(p Permission) bool {
	switch p {
	case PermItemsWrite, PermLabelsWrite, PermLocationsWrite:
		return true
	default:
		return false
	}
}

// ParsePermission parses a stored permission string.
func ParsePermission(s string) (Permission, bool) {
	p := Permission(s)
	if ValidPermission(p) {
		return p, true
	}
	return "", false
}
