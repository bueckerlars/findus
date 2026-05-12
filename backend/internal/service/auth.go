package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"golang.org/x/crypto/bcrypt"

	"findus/backend/internal/domain"
	"findus/backend/internal/repository"
)

type Auth struct {
	Users    repository.UserRepository
	Settings repository.SettingsRepository
	Invites  repository.InviteRepository
}

func (a *Auth) Login(ctx context.Context, username, password string) (*domain.User, error) {
	u, err := a.Users.GetByUsername(ctx, strings.TrimSpace(username))
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, domain.ErrUnauthorized
		}
		return nil, err
	}
	if !u.IsActive {
		return nil, domain.ErrUnauthorized
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return nil, domain.ErrUnauthorized
	}
	return u, nil
}

func (a *Auth) Register(ctx context.Context, username, email, password, inviteToken string) (*domain.User, error) {
	username = strings.TrimSpace(username)
	email = strings.TrimSpace(strings.ToLower(email))
	if err := validateUserCreds(username, email, password); err != nil {
		return nil, err
	}
	n, err := a.Users.Count(ctx)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	var role domain.Role
	var consumedInvite *domain.Invite

	if n == 0 {
		role = domain.RoleAdmin
	} else {
		mode, ok, err := a.registrationMode(ctx)
		if err != nil {
			return nil, err
		}
		if !ok {
			mode = domain.RegistrationAdminOnly
		}
		switch mode {
		case domain.RegistrationAdminOnly:
			return nil, domain.ErrRegistrationClosed
		case domain.RegistrationOpen:
			role = domain.RoleUser
		case domain.RegistrationInvite:
			inv, err := a.Invites.GetByToken(ctx, strings.TrimSpace(inviteToken))
			if err != nil {
				return nil, domain.ErrInvalidInvite
			}
			if inv.UsedAt != nil || time.Now().After(inv.ExpiresAt) {
				return nil, domain.ErrInvalidInvite
			}
			consumedInvite = inv
			role = inv.Role
		default:
			return nil, domain.ErrRegistrationClosed
		}
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	u := &domain.User{
		ID:           newID(),
		Username:     username,
		Email:        email,
		PasswordHash: string(hash),
		Role:         role,
		IsActive:     true,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := a.Users.Create(ctx, u); err != nil {
		return nil, err
	}
	if consumedInvite != nil {
		if err := a.Invites.MarkUsed(ctx, consumedInvite.ID, now); err != nil {
			return nil, err
		}
	}
	if n == 0 {
		_ = a.Settings.Set(ctx, string(domain.SettingRegistrationMode), string(domain.RegistrationAdminOnly))
	}
	return u, nil
}

// UpdateProfile verifies currentPassword, then updates username, email, and optionally password.
// newPassword empty means leave password unchanged.
func (a *Auth) UpdateProfile(ctx context.Context, userID, username, email, currentPassword, newPassword string) (*domain.User, error) {
	u, err := a.Users.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(currentPassword)); err != nil {
		return nil, domain.ErrInvalidCurrentPassword
	}
	username = strings.TrimSpace(username)
	email = strings.TrimSpace(strings.ToLower(email))
	if err := validateProfileUsernameEmail(username, email); err != nil {
		return nil, err
	}
	newPassword = strings.TrimSpace(newPassword)
	if newPassword != "" && len(newPassword) < 10 {
		return nil, fmt.Errorf("%w: password min 10 chars", domain.ErrValidation)
	}
	if other, err := a.Users.GetByUsername(ctx, username); err == nil {
		if other.ID != userID {
			return nil, fmt.Errorf("username taken: %w", domain.ErrConflict)
		}
	} else if !errors.Is(err, domain.ErrNotFound) {
		return nil, err
	}
	if other, err := a.Users.GetByEmail(ctx, email); err == nil {
		if other.ID != userID {
			return nil, fmt.Errorf("email taken: %w", domain.ErrConflict)
		}
	} else if !errors.Is(err, domain.ErrNotFound) {
		return nil, err
	}
	u.Username = username
	u.Email = email
	u.UpdatedAt = time.Now().UTC()
	if newPassword != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
		if err != nil {
			return nil, err
		}
		u.PasswordHash = string(hash)
	}
	if err := a.Users.Update(ctx, u); err != nil {
		return nil, err
	}
	return u, nil
}

func validateProfileUsernameEmail(username, email string) error {
	if utf8.RuneCountInString(username) < 2 || utf8.RuneCountInString(username) > 64 {
		return fmt.Errorf("%w: username length", domain.ErrValidation)
	}
	if !strings.Contains(email, "@") || utf8.RuneCountInString(email) > 254 {
		return fmt.Errorf("%w: email", domain.ErrValidation)
	}
	return nil
}

func (a *Auth) registrationMode(ctx context.Context) (domain.RegistrationMode, bool, error) {
	v, ok, err := a.Settings.Get(ctx, domain.SettingRegistrationMode)
	if err != nil || !ok {
		return "", ok, err
	}
	m, valid := domain.ParseRegistrationMode(v)
	if !valid {
		return "", false, nil
	}
	return m, true, nil
}

func validateUserCreds(username, email, password string) error {
	if utf8.RuneCountInString(username) < 2 || utf8.RuneCountInString(username) > 64 {
		return fmt.Errorf("%w: username length", domain.ErrValidation)
	}
	if !strings.Contains(email, "@") || utf8.RuneCountInString(email) > 254 {
		return fmt.Errorf("%w: email", domain.ErrValidation)
	}
	if len(password) < 10 {
		return fmt.Errorf("%w: password min 10 chars", domain.ErrValidation)
	}
	return nil
}
