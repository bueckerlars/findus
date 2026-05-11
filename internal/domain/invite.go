package domain

import "time"

type Invite struct {
	ID        string
	Token     string
	CreatedBy string
	Role      Role
	ExpiresAt time.Time
	UsedAt    *time.Time
	CreatedAt time.Time
}
