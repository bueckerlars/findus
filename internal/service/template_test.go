package service

import (
	"testing"

	"github.com/stretchr/testify/require"

	"findus/internal/domain"
)

func TestNormalizeTemplateDataFromFields_Standard(t *testing.T) {
	b, err := domain.NormalizeTemplateDataFromFields(nil, nil)
	require.NoError(t, err)
	require.Equal(t, `{}`, string(b))
}

func TestNormalizeTemplateDataFromFields_Electronics(t *testing.T) {
	raw := []byte(`[{"key":"condition","label":"Condition","widget":"select","required":true,"options":[{"value":"working","label":"Working"},{"value":"broken","label":"Broken"}]},{"key":"power_cable","label":"Power cable","widget":"select","required":true,"options":[{"value":"yes","label":"Yes"},{"value":"no","label":"No"}]}]`)
	fields, err := domain.ParseTemplateFieldsJSON(raw)
	require.NoError(t, err)
	b, err := domain.NormalizeTemplateDataFromFields(fields, map[string]string{
		"condition":   "working",
		"power_cable": "no",
	})
	require.NoError(t, err)
	require.Contains(t, string(b), "working")
}

func TestNormalizeTemplateDataFromFields_InvalidYear(t *testing.T) {
	raw := []byte(`[{"key":"year","label":"Year","widget":"text","required":true,"pattern":"^\\d{4}$","min_int":1900,"max_int":2100},{"key":"category","label":"Category","widget":"select","required":true,"options":[{"value":"tax","label":"Tax"}]}]`)
	fields, err := domain.ParseTemplateFieldsJSON(raw)
	require.NoError(t, err)
	_, err = domain.NormalizeTemplateDataFromFields(fields, map[string]string{
		"year":     "abc",
		"category": "tax",
	})
	require.Error(t, err)
}
