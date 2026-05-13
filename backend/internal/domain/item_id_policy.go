package domain

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"
)

const SettingItemIDPolicy = "item_id_policy"

// ItemIDKind selects how new item primary keys are generated.
type ItemIDKind string

const (
	ItemIDKindSequential ItemIDKind = "sequential"
)

// ItemIDPolicy is stored as JSON under SettingItemIDPolicy.
type ItemIDPolicy struct {
	Kind    ItemIDKind `json:"kind"`
	Prefix  string     `json:"prefix,omitempty"`
	Width   int        `json:"width,omitempty"`
	NextSeq int64      `json:"next_seq,omitempty"`
}

// DefaultItemIDPolicy is used when no setting is stored (new installs).
func DefaultItemIDPolicy() ItemIDPolicy {
	return ItemIDPolicy{
		Kind:    ItemIDKindSequential,
		Prefix:  "item",
		Width:   4,
		NextSeq: 1,
	}
}

// EffectiveKey returns a comparable tuple for "did display/id scheme change".
func (p ItemIDPolicy) EffectiveKey() string {
	p = p.Normalize()
	return string(p.Kind) + "\x00" + p.Prefix + "\x00" + fmt.Sprintf("%d", p.Width)
}

// Normalize fills defaults after JSON parse.
func (p ItemIDPolicy) Normalize() ItemIDPolicy {
	switch p.Kind {
	case ItemIDKindSequential:
		if p.Width < 1 {
			p.Width = 1
		}
		if p.Width > 12 {
			p.Width = 12
		}
		if p.NextSeq < 1 {
			p.NextSeq = 1
		}
		return p
	default:
		return DefaultItemIDPolicy()
	}
}

const (
	maxItemIDPolicyPrefixRunes = 48
	maxItemIDTotalLen          = 200
)

// Validate checks policy fields; use on admin input and after JSON parse.
func (p ItemIDPolicy) Validate() error {
	p = p.Normalize()
	switch p.Kind {
	case ItemIDKindSequential:
		prefix := strings.TrimSpace(p.Prefix)
		if utf8.RuneCountInString(prefix) > maxItemIDPolicyPrefixRunes {
			return fmt.Errorf("%w: prefix too long", ErrValidation)
		}
		for _, r := range prefix {
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
				continue
			}
			return fmt.Errorf("%w: prefix contains invalid character", ErrValidation)
		}
		if p.Width < 1 || p.Width > 12 {
			return fmt.Errorf("%w: width", ErrValidation)
		}
		if p.NextSeq < 1 {
			return fmt.Errorf("%w: next_seq", ErrValidation)
		}
		// Sample one id length
		sample := FormatSequentialID(prefix, p.Width, 1)
		if len(sample) > maxItemIDTotalLen {
			return fmt.Errorf("%w: id would be too long", ErrValidation)
		}
		return nil
	default:
		return fmt.Errorf("%w: kind", ErrValidation)
	}
}

// FormatSequentialID builds prefix + "_" + zero-padded number (no validation).
func FormatSequentialID(prefix string, width int, seq int64) string {
	return fmt.Sprintf("%s_%0*d", prefix, width, seq)
}

// ItemIDMatchesPolicy reports whether id conforms to the given policy (after Normalize).
func ItemIDMatchesPolicy(id string, p ItemIDPolicy) bool {
	id = strings.TrimSpace(id)
	if id == "" {
		return false
	}
	p = p.Normalize()
	switch p.Kind {
	case ItemIDKindSequential:
		prefix := strings.TrimSpace(p.Prefix)
		reNew, err := regexp.Compile(fmt.Sprintf(`^%s_\d{%d}$`, regexp.QuoteMeta(prefix), p.Width))
		if err != nil {
			return false
		}
		if reNew.MatchString(id) {
			return true
		}
		// Legacy: prefix directly followed by digits (no separator).
		reLegacy, err := regexp.Compile(fmt.Sprintf(`^%s\d{%d}$`, regexp.QuoteMeta(prefix), p.Width))
		if err != nil {
			return false
		}
		return reLegacy.MatchString(id)
	default:
		return false
	}
}

// ParseItemIDPolicyJSON unmarshals and normalizes; invalid kind falls back to DefaultItemIDPolicy.
func ParseItemIDPolicyJSON(raw string) (ItemIDPolicy, error) {
	if strings.TrimSpace(raw) == "" {
		return DefaultItemIDPolicy(), nil
	}
	var p ItemIDPolicy
	if err := json.Unmarshal([]byte(raw), &p); err != nil {
		return ItemIDPolicy{}, err
	}
	switch p.Kind {
	case ItemIDKindSequential:
	default:
		return DefaultItemIDPolicy(), nil
	}
	return p.Normalize(), nil
}

// ItemIDPolicyJSON returns canonical JSON for storage.
func ItemIDPolicyJSON(p ItemIDPolicy) ([]byte, error) {
	if err := p.Validate(); err != nil {
		return nil, err
	}
	p = p.Normalize()
	return json.Marshal(p)
}

// ItemIDMigration describes one primary-key change for items (+ labels + photo_path).
type ItemIDMigration struct {
	OldID        string
	NewID        string
	NewPhotoPath *string // nil means SQL NULL; non-nil is the new photo_path value
}
