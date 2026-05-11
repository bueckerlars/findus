package service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"unicode/utf8"

	"findus/internal/domain"
)

const (
	maxAdditionalJSONBytes = 65536
	maxAdditionalKeys      = 64
	maxAdditionalKeyRunes  = 128
	maxAdditionalValRunes  = 4000
)

// ParseAdditionalFromForm builds a JSON object from repeated add_k / add_v fields.
func ParseAdditionalFromForm(r *http.Request) (json.RawMessage, error) {
	keys := r.Form["add_k"]
	vals := r.Form["add_v"]
	if len(keys) != len(vals) {
		return nil, fmt.Errorf("%w: additional key/value count mismatch", domain.ErrValidation)
	}
	m := make(map[string]string)
	for i := range keys {
		k := strings.TrimSpace(keys[i])
		if k == "" {
			continue
		}
		if utf8.RuneCountInString(k) > maxAdditionalKeyRunes {
			return nil, fmt.Errorf("%w: additional key too long", domain.ErrValidation)
		}
		v := strings.TrimSpace(vals[i])
		if utf8.RuneCountInString(v) > maxAdditionalValRunes {
			return nil, fmt.Errorf("%w: additional value too long", domain.ErrValidation)
		}
		m[k] = v
	}
	if len(m) == 0 {
		return json.RawMessage(`{}`), nil
	}
	if len(m) > maxAdditionalKeys {
		return nil, fmt.Errorf("%w: too many additional fields", domain.ErrValidation)
	}
	b, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	if err := ValidateAdditionalJSON(b); err != nil {
		return nil, err
	}
	return b, nil
}

// ValidateAdditionalJSON ensures payload is a JSON object within size limits.
func ValidateAdditionalJSON(b json.RawMessage) error {
	if len(b) == 0 {
		return nil
	}
	if len(b) > maxAdditionalJSONBytes {
		return fmt.Errorf("%w: additional_data too large", domain.ErrValidation)
	}
	var v map[string]json.RawMessage
	if err := json.Unmarshal(b, &v); err != nil {
		return fmt.Errorf("%w: additional_data must be a JSON object", domain.ErrValidation)
	}
	if len(v) > maxAdditionalKeys {
		return fmt.Errorf("%w: too many keys in additional_data", domain.ErrValidation)
	}
	return nil
}
