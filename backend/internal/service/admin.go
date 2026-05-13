package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"findus/backend/internal/domain"
	"findus/backend/internal/repository"
)

type Admin struct {
	Users    repository.UserRepository
	Settings repository.SettingsRepository
	Invites  repository.InviteRepository
	Groups   repository.PermissionGroupRepository
}

func (a *Admin) SetRegistrationMode(ctx context.Context, mode domain.RegistrationMode) error {
	if _, ok := domain.ParseRegistrationMode(string(mode)); !ok {
		return fmt.Errorf("%w: registration_mode", domain.ErrValidation)
	}
	return a.Settings.Set(ctx, domain.SettingRegistrationMode, string(mode))
}

func (a *Admin) GetRegistrationMode(ctx context.Context) (domain.RegistrationMode, error) {
	v, ok, err := a.Settings.Get(ctx, domain.SettingRegistrationMode)
	if err != nil {
		return "", err
	}
	if !ok {
		return domain.RegistrationAdminOnly, nil
	}
	m, ok := domain.ParseRegistrationMode(v)
	if !ok {
		return domain.RegistrationAdminOnly, nil
	}
	return m, nil
}

func (a *Admin) CreateInvite(ctx context.Context, adminID string, role domain.Role, ttl time.Duration) (*domain.Invite, error) {
	if role != domain.RoleAdmin && role != domain.RoleUser {
		return nil, fmt.Errorf("%w: role", domain.ErrValidation)
	}
	if ttl < time.Hour {
		ttl = 24 * time.Hour
	}
	if ttl > 30*24*time.Hour {
		ttl = 30 * 24 * time.Hour
	}
	now := time.Now().UTC()
	inv := &domain.Invite{
		ID:        newID(),
		Token:     newID() + newID(), // long unguessable token
		CreatedBy: adminID,
		Role:      role,
		ExpiresAt: now.Add(ttl),
		CreatedAt: now,
	}
	if err := a.Invites.Create(ctx, inv); err != nil {
		return nil, err
	}
	return inv, nil
}

func (a *Admin) ListInvites(ctx context.Context) ([]domain.Invite, error) {
	return a.Invites.ListRecent(ctx, 200)
}

func (a *Admin) CreateUser(ctx context.Context, username, email, password string, role domain.Role) (*domain.User, error) {
	username = strings.TrimSpace(username)
	emailNorm, err := normalizeRegistrationEmail(email)
	if err != nil {
		return nil, err
	}
	if err := validateUserCreds(username, emailNorm, password); err != nil {
		return nil, err
	}
	if _, err := a.Users.GetByUsername(ctx, username); err == nil {
		return nil, fmt.Errorf("username already taken: %w", domain.ErrConflict)
	} else if !errors.Is(err, domain.ErrNotFound) {
		return nil, err
	}
	if _, err := a.Users.GetByEmail(ctx, emailNorm); err == nil {
		return nil, fmt.Errorf("email already registered: %w", domain.ErrConflict)
	} else if !errors.Is(err, domain.ErrNotFound) {
		return nil, err
	}
	if role != domain.RoleAdmin && role != domain.RoleUser {
		return nil, fmt.Errorf("%w: role", domain.ErrValidation)
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	u := &domain.User{
		ID:           newID(),
		Username:     username,
		Email:        emailNorm,
		PasswordHash: string(hash),
		Role:         role,
		IsActive:     true,
		CreatedAt:    now,
		UpdatedAt:    now,
		UITheme:      domain.DefaultUITheme,
	}
	if err := a.Users.Create(ctx, u); err != nil {
		return nil, err
	}
	return u, nil
}

func (a *Admin) ListUsers(ctx context.Context) ([]domain.User, error) {
	return a.Users.List(ctx)
}

func (a *Admin) SetUserActive(ctx context.Context, id string, active bool) error {
	u, err := a.Users.GetByID(ctx, id)
	if err != nil {
		return err
	}
	u.IsActive = active
	u.UpdatedAt = time.Now().UTC()
	return a.Users.Update(ctx, u)
}

func (a *Admin) SetUserRole(ctx context.Context, id string, role domain.Role) error {
	if role != domain.RoleAdmin && role != domain.RoleUser {
		return fmt.Errorf("%w: role", domain.ErrValidation)
	}
	u, err := a.Users.GetByID(ctx, id)
	if err != nil {
		return err
	}
	u.Role = role
	u.UpdatedAt = time.Now().UTC()
	return a.Users.Update(ctx, u)
}
