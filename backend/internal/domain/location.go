package domain

import "time"

type Location struct {
	ID          string
	Name        string
	ParentID    *string
	Description string
	QRToken     string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// LocationPathElement is one segment from root to leaf (root first).
type LocationPathElement struct {
	ID   string
	Name string
}
