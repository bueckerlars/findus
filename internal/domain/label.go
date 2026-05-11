package domain

import "time"

// Label is a reusable tag; optional default_template_type suggests item template when this label is chosen.
type Label struct {
	ID                  string
	Name                string
	Color               string
	DefaultTemplateType *string
	CreatedAt           time.Time
	UpdatedAt           time.Time
}
