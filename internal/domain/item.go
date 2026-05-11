package domain

import (
	"encoding/json"
	"time"
)

type TemplateType string

type Item struct {
	ID             string
	Name           string
	Description    string
	LocationID     string
	TemplateType   TemplateType
	TemplateData   json.RawMessage
	AdditionalData json.RawMessage
	PhotoPath      *string
	QRToken        string
	CreatedAt      time.Time
	UpdatedAt      time.Time
	// Labels populated when loaded with label join (optional).
	Labels []Label
}
