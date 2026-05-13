package service

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"time"

	"findus/backend/internal/domain"
)

func (a *Admin) ListPermissionGroups(ctx context.Context) ([]domain.PermissionGroup, map[string]int64, error) {
	if a.Groups == nil {
		return nil, nil, fmt.Errorf("groups repository not configured")
	}
	gs, err := a.Groups.ListGroups(ctx)
	if err != nil {
		return nil, nil, err
	}
	counts, err := a.Groups.ListMemberCounts(ctx)
	if err != nil {
		return nil, nil, err
	}
	return gs, counts, nil
}

func (a *Admin) GetPermissionGroup(ctx context.Context, id string) (*domain.PermissionGroup, []domain.Permission, error) {
	if a.Groups == nil {
		return nil, nil, fmt.Errorf("groups repository not configured")
	}
	g, err := a.Groups.GetGroup(ctx, id)
	if err != nil {
		return nil, nil, err
	}
	perms, err := a.Groups.GetGroupPermissions(ctx, id)
	if err != nil {
		return nil, nil, err
	}
	return g, perms, nil
}

func (a *Admin) CreatePermissionGroup(ctx context.Context, name string, perms []domain.Permission) (*domain.PermissionGroup, error) {
	if a.Groups == nil {
		return nil, fmt.Errorf("groups repository not configured")
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, fmt.Errorf("%w: group name required", domain.ErrValidation)
	}
	for _, p := range perms {
		if !domain.ValidPermission(p) {
			return nil, fmt.Errorf("%w: unknown permission %q", domain.ErrValidation, p)
		}
	}
	perms = dedupePermissions(perms)
	now := time.Now().UTC()
	g := &domain.PermissionGroup{
		ID:        newID(),
		Name:      name,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := a.Groups.CreateGroup(ctx, g); err != nil {
		return nil, err
	}
	if err := a.Groups.ReplaceGroupPermissions(ctx, g.ID, perms); err != nil {
		_ = a.Groups.DeleteGroup(ctx, g.ID)
		return nil, err
	}
	return g, nil
}

func (a *Admin) UpdatePermissionGroup(ctx context.Context, id, name string, perms []domain.Permission) error {
	if a.Groups == nil {
		return fmt.Errorf("groups repository not configured")
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("%w: group name required", domain.ErrValidation)
	}
	for _, p := range perms {
		if !domain.ValidPermission(p) {
			return fmt.Errorf("%w: unknown permission %q", domain.ErrValidation, p)
		}
	}
	perms = dedupePermissions(perms)
	g, err := a.Groups.GetGroup(ctx, id)
	if err != nil {
		return err
	}
	g.Name = name
	g.UpdatedAt = time.Now().UTC()
	if err := a.Groups.UpdateGroup(ctx, g); err != nil {
		return err
	}
	return a.Groups.ReplaceGroupPermissions(ctx, id, perms)
}

func (a *Admin) DeletePermissionGroup(ctx context.Context, id string) error {
	if a.Groups == nil {
		return fmt.Errorf("groups repository not configured")
	}
	return a.Groups.DeleteGroup(ctx, id)
}

func (a *Admin) SetUserPermissionGroups(ctx context.Context, userID string, groupIDs []string) error {
	if a.Groups == nil {
		return fmt.Errorf("groups repository not configured")
	}
	if _, err := a.Users.GetByID(ctx, userID); err != nil {
		return err
	}
	groupIDs = dedupeStrings(groupIDs)
	for _, gid := range groupIDs {
		if _, err := a.Groups.GetGroup(ctx, gid); err != nil {
			return fmt.Errorf("%w: unknown group %q", domain.ErrValidation, gid)
		}
	}
	return a.Groups.ReplaceUserGroups(ctx, userID, groupIDs)
}

func (a *Admin) MapUserGroupIDs(ctx context.Context) (map[string][]string, error) {
	if a.Groups == nil {
		return map[string][]string{}, nil
	}
	return a.Groups.ListUserGroupIDs(ctx)
}

func dedupePermissions(perms []domain.Permission) []domain.Permission {
	if len(perms) <= 1 {
		return perms
	}
	out := make([]domain.Permission, 0, len(perms))
	for _, p := range perms {
		if !slices.Contains(out, p) {
			out = append(out, p)
		}
	}
	slices.SortFunc(out, func(a, b domain.Permission) int {
		if a < b {
			return -1
		}
		if a > b {
			return 1
		}
		return 0
	})
	return out
}
