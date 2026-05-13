package domain

import "time"

// PermissionGroup is a named collection of permissions (RBAC "role").
type PermissionGroup struct {
	ID        string
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// PermissionGroupMemberCount is returned with group listings.
type PermissionGroupMemberCount struct {
	GroupID     string
	MemberCount int64
}
