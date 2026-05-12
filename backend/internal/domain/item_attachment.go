package domain

import (
	"encoding/json"
	"time"
)

// ItemAttachment is a user-uploaded file linked to an item (e.g. warranty PDF).
// MetadataJSON is reserved for future pipelines (OCR status, extracted text, etc.).
type ItemAttachment struct {
	ID               string
	ItemID           string
	OriginalFilename string
	StoragePath      string
	MimeType         string
	SizeBytes        int64
	Title            string
	MetadataJSON     json.RawMessage
	CreatedAt        time.Time
}
