package service

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseAdditionalFromForm_Empty(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(""))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	require.NoError(t, req.ParseForm())
	b, err := ParseAdditionalFromForm(req)
	require.NoError(t, err)
	require.Equal(t, "{}", string(b))
}

func TestParseAdditionalFromForm_Pairs(t *testing.T) {
	v := url.Values{}
	v.Add("add_k", "serial")
	v.Add("add_v", "ABC-1")
	v.Add("add_k", "  ")
	v.Add("add_v", "ignored")
	v.Add("add_k", "note")
	v.Add("add_v", "hello")
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(v.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	require.NoError(t, req.ParseForm())
	b, err := ParseAdditionalFromForm(req)
	require.NoError(t, err)
	require.JSONEq(t, `{"note":"hello","serial":"ABC-1"}`, string(b))
}
