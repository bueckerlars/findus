package domain

import "errors"

var (
	ErrNotFound               = errors.New("not found")
	ErrConflict               = errors.New("conflict")
	ErrValidation             = errors.New("validation failed")
	ErrUnauthorized           = errors.New("unauthorized")
	ErrForbidden              = errors.New("forbidden")
	ErrInvalidInvite          = errors.New("invalid or expired invite")
	ErrRegistrationClosed     = errors.New("registration is not allowed")
	ErrInvalidCurrentPassword = errors.New("invalid current password")
)
