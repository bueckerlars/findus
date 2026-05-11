package domain

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

// ItemTemplate defines how structured fields are collected for items of a given template_type.
type ItemTemplate struct {
	ID          string
	DisplayName string
	Fields      []TemplateField
	FieldsJSON  json.RawMessage
	SortOrder   int
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// TemplateField is one dynamic field in an item template (stored as JSON in item_templates.fields_json).
type TemplateField struct {
	Key         string        `json:"key"`
	Label       string        `json:"label"`
	Widget      string        `json:"widget"` // "text" or "select"
	Required    bool          `json:"required"`
	Placeholder string        `json:"placeholder,omitempty"`
	Pattern     string        `json:"pattern,omitempty"`
	MinInt      *int          `json:"min_int,omitempty"`
	MaxInt      *int          `json:"max_int,omitempty"`
	MaxLen      int           `json:"max_len,omitempty"` // UTF-8 runes; 0 uses a safe default for text
	Options     []FieldOption `json:"options,omitempty"`
}

// FieldOption is one choice for a select widget.
type FieldOption struct {
	Value string `json:"value"`
	Label string `json:"label"`
}

// ParseTemplateFieldsJSON unmarshals and validates fields_json content.
func ParseTemplateFieldsJSON(raw []byte) ([]TemplateField, error) {
	if len(strings.TrimSpace(string(raw))) == 0 {
		raw = []byte("[]")
	}
	var fields []TemplateField
	if err := json.Unmarshal(raw, &fields); err != nil {
		return nil, fmt.Errorf("fields_json: %w", err)
	}
	seen := make(map[string]struct{})
	for i := range fields {
		f := &fields[i]
		f.Key = strings.TrimSpace(f.Key)
		if f.Key == "" {
			return nil, fmt.Errorf("%w: field key at index %d", ErrValidation, i)
		}
		if _, dup := seen[f.Key]; dup {
			return nil, fmt.Errorf("%w: duplicate field key %q", ErrValidation, f.Key)
		}
		seen[f.Key] = struct{}{}
		f.Label = strings.TrimSpace(f.Label)
		if f.Label == "" {
			return nil, fmt.Errorf("%w: field label for key %q", ErrValidation, f.Key)
		}
		switch f.Widget {
		case "text", "select":
		default:
			return nil, fmt.Errorf("%w: widget for key %q", ErrValidation, f.Key)
		}
		if f.Widget == "select" {
			if len(f.Options) == 0 {
				return nil, fmt.Errorf("%w: select options for key %q", ErrValidation, f.Key)
			}
			oval := make(map[string]struct{})
			for j, o := range f.Options {
				o.Value = strings.TrimSpace(o.Value)
				o.Label = strings.TrimSpace(o.Label)
				if o.Value == "" || o.Label == "" {
					return nil, fmt.Errorf("%w: option %d for key %q", ErrValidation, j, f.Key)
				}
				if _, dup := oval[o.Value]; dup {
					return nil, fmt.Errorf("%w: duplicate option value %q for key %q", ErrValidation, o.Value, f.Key)
				}
				oval[o.Value] = struct{}{}
				f.Options[j] = o
			}
		}
		if f.Pattern != "" {
			if _, err := regexp.Compile(f.Pattern); err != nil {
				return nil, fmt.Errorf("%w: pattern for key %q: %v", ErrValidation, f.Key, err)
			}
		}
		if f.Widget == "text" && f.MinInt != nil && f.MaxInt != nil && *f.MinInt > *f.MaxInt {
			return nil, fmt.Errorf("%w: min_int/max_int for key %q", ErrValidation, f.Key)
		}
	}
	return fields, nil
}

// MarshalTemplateFields returns canonical JSON for storing fields_json.
func MarshalTemplateFields(fields []TemplateField) (json.RawMessage, error) {
	if fields == nil {
		return []byte("[]"), nil
	}
	b, err := json.Marshal(fields)
	if err != nil {
		return nil, err
	}
	return json.RawMessage(b), nil
}

const defaultTextMaxLen = 512

// NormalizeTemplateDataFromFields builds template_data JSON from form values using the field schema.
func NormalizeTemplateDataFromFields(fields []TemplateField, form map[string]string) (json.RawMessage, error) {
	out := make(map[string]string)
	for _, f := range fields {
		raw := ""
		if form != nil {
			raw = form[f.Key]
		}
		v := strings.TrimSpace(raw)
		switch f.Widget {
		case "select":
			if v == "" {
				if f.Required {
					return nil, fmt.Errorf("%w: %s", ErrValidation, f.Key)
				}
				continue
			}
			ok := false
			for _, o := range f.Options {
				if o.Value == v {
					ok = true
					break
				}
			}
			if !ok {
				return nil, fmt.Errorf("%w: %s", ErrValidation, f.Key)
			}
			out[f.Key] = v
		case "text":
			if v == "" {
				if f.Required {
					return nil, fmt.Errorf("%w: %s", ErrValidation, f.Key)
				}
				continue
			}
			maxLen := f.MaxLen
			if maxLen <= 0 {
				maxLen = defaultTextMaxLen
			}
			if utf8.RuneCountInString(v) > maxLen {
				return nil, fmt.Errorf("%w: %s length", ErrValidation, f.Key)
			}
			if f.Pattern != "" {
				re, err := regexp.Compile(f.Pattern)
				if err != nil {
					return nil, fmt.Errorf("%w: %s pattern", ErrValidation, f.Key)
				}
				if !re.MatchString(v) {
					return nil, fmt.Errorf("%w: %s", ErrValidation, f.Key)
				}
			}
			if f.MinInt != nil || f.MaxInt != nil {
				n, err := strconv.Atoi(v)
				if err != nil {
					return nil, fmt.Errorf("%w: %s", ErrValidation, f.Key)
				}
				if f.MinInt != nil && n < *f.MinInt {
					return nil, fmt.Errorf("%w: %s range", ErrValidation, f.Key)
				}
				if f.MaxInt != nil && n > *f.MaxInt {
					return nil, fmt.Errorf("%w: %s range", ErrValidation, f.Key)
				}
			}
			out[f.Key] = v
		default:
			return nil, fmt.Errorf("%w: widget %s", ErrValidation, f.Widget)
		}
	}
	if len(out) == 0 {
		return json.RawMessage(`{}`), nil
	}
	b, err := json.Marshal(out)
	if err != nil {
		return nil, err
	}
	return json.RawMessage(b), nil
}
