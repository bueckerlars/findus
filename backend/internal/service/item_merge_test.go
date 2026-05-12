package service

import (
	"encoding/json"
	"testing"

	"findus/backend/internal/domain"
)

func TestMergeTemplateIntoAdditional_AdditionalWins(t *testing.T) {
	add := json.RawMessage(`{"a":"from_add","b":"only_add"}`)
	td := json.RawMessage(`{"a":"from_tpl","c":"only_tpl"}`)
	out, err := MergeTemplateIntoAdditional(add, td)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]string
	if err := json.Unmarshal(out, &m); err != nil {
		t.Fatal(err)
	}
	if m["a"] != "from_add" {
		t.Fatalf("a: got %q", m["a"])
	}
	if m["b"] != "only_add" || m["c"] != "only_tpl" {
		t.Fatalf("merge: %#v", m)
	}
}

func TestItemPropertyDetailRows_SelectLabel(t *testing.T) {
	tpl := &domain.ItemTemplate{
		Fields: []domain.TemplateField{
			{Key: "kind", Label: "Kind", Widget: "select", Options: []domain.FieldOption{{Value: "x", Label: "Extra"}}},
			{Key: "note", Label: "Note", Widget: "text"},
		},
	}
	merged := json.RawMessage(`{"kind":"x","note":"hello","orphan":"z"}`)
	rows := ItemPropertyDetailRows(tpl, merged)
	if len(rows) != 3 {
		t.Fatalf("len=%d %#v", len(rows), rows)
	}
	if rows[0].Key != "kind" || rows[0].Label != "Kind" || rows[0].DisplayValue != "Extra" || rows[0].RawValue != "x" {
		t.Fatalf("row0: %+v", rows[0])
	}
}
