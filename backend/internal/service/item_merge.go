package service

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"findus/backend/internal/domain"
)

func rawJSONToStringMap(raw json.RawMessage) map[string]string {
	if len(raw) == 0 || string(raw) == "null" {
		return make(map[string]string)
	}
	var m map[string]interface{}
	if err := json.Unmarshal(raw, &m); err != nil || m == nil {
		return make(map[string]string)
	}
	out := make(map[string]string, len(m))
	for k, v := range m {
		out[k] = valueToString(v)
	}
	return out
}

func valueToString(v interface{}) string {
	switch x := v.(type) {
	case nil:
		return ""
	case string:
		return x
	case bool:
		if x {
			return "true"
		}
		return "false"
	case float64:
		if x == float64(int64(x)) {
			return fmt.Sprintf("%.0f", x)
		}
		return fmt.Sprint(x)
	case json.Number:
		return x.String()
	default:
		b, err := json.Marshal(x)
		if err != nil {
			return fmt.Sprint(x)
		}
		return string(b)
	}
}

// MergeTemplateIntoAdditional merges template_data into additional_data. Keys already present in additional win.
func MergeTemplateIntoAdditional(additional, templateData json.RawMessage) (json.RawMessage, error) {
	add := rawJSONToStringMap(additional)
	tpl := rawJSONToStringMap(templateData)
	for k, v := range tpl {
		if _, ok := add[k]; !ok {
			add[k] = v
		}
	}
	if len(add) == 0 {
		return json.RawMessage(`{}`), nil
	}
	b, err := json.Marshal(add)
	if err != nil {
		return nil, err
	}
	if err := ValidateAdditionalJSON(b); err != nil {
		return nil, err
	}
	return b, nil
}

// ItemPropertyDetailRow is one user-editable attribute row for item detail (Key/RawValue are stored; DisplayValue is for read-only view).
type ItemPropertyDetailRow struct {
	Key          string
	Label        string
	DisplayValue string
	RawValue     string
}

// ItemPropertyDetailRows builds rows from merged JSON using template field labels when available.
func ItemPropertyDetailRows(tpl *domain.ItemTemplate, merged json.RawMessage) []ItemPropertyDetailRow {
	m := rawJSONToStringMap(merged)
	if len(m) == 0 {
		return nil
	}
	seen := make(map[string]struct{})
	var rows []ItemPropertyDetailRow
	if tpl != nil {
		for i := range tpl.Fields {
			f := &tpl.Fields[i]
			v, ok := m[f.Key]
			if !ok || strings.TrimSpace(v) == "" {
				continue
			}
			rows = append(rows, ItemPropertyDetailRow{
				Key: f.Key, Label: f.Label,
				DisplayValue: displaySelectLabel(f, v),
				RawValue:     v,
			})
			seen[f.Key] = struct{}{}
		}
	}
	var rest []string
	for k := range m {
		if _, ok := seen[k]; ok {
			continue
		}
		if strings.TrimSpace(m[k]) == "" {
			continue
		}
		rest = append(rest, k)
	}
	sort.Strings(rest)
	for _, k := range rest {
		v := m[k]
		rows = append(rows, ItemPropertyDetailRow{
			Key: k, Label: humanizeKey(k),
			DisplayValue: v, RawValue: v,
		})
	}
	return rows
}

func displaySelectLabel(f *domain.TemplateField, stored string) string {
	if f == nil || f.Widget != "select" {
		return stored
	}
	for _, o := range f.Options {
		if o.Value == stored {
			return o.Label
		}
	}
	return stored
}

func humanizeKey(k string) string {
	k = strings.TrimSpace(k)
	if k == "" {
		return k
	}
	k = strings.ReplaceAll(k, "_", " ")
	return k
}
