package search

import (
	"encoding/json"
	"fmt"
	"strings"
)

// FlattenJSON extracts primitive string, number, and bool values from JSON (object values and array elements).
func FlattenJSON(raw []byte) string {
	raw = trimSpaceBytes(raw)
	if len(raw) == 0 {
		return ""
	}
	var v any
	if err := json.Unmarshal(raw, &v); err != nil {
		return string(raw)
	}
	var parts []string
	walkJSONValue(v, &parts)
	return strings.Join(parts, " ")
}

func trimSpaceBytes(b []byte) []byte {
	return []byte(strings.TrimSpace(string(b)))
}

func walkJSONValue(v any, parts *[]string) {
	switch x := v.(type) {
	case string:
		s := strings.TrimSpace(x)
		if s != "" {
			*parts = append(*parts, s)
		}
	case float64:
		*parts = append(*parts, trimFloatString(fmt.Sprint(x)))
	case bool:
		*parts = append(*parts, fmt.Sprint(x))
	case json.Number:
		*parts = append(*parts, x.String())
	case []any:
		for _, e := range x {
			walkJSONValue(e, parts)
		}
	case map[string]any:
		for _, e := range x {
			walkJSONValue(e, parts)
		}
	case nil:
	default:
		*parts = append(*parts, fmt.Sprint(x))
	}
}

func trimFloatString(s string) string {
	if strings.Contains(s, ".") {
		s = strings.TrimRight(s, "0")
		s = strings.TrimSuffix(s, ".")
	}
	return s
}
