package middleware

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCSRF_SkipsBodyParseWhenTokenInHeader(t *testing.T) {
	const tok = "deadbeefcafebabe"
	h := CSRF(false)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(b)
	}))

	body := `{"kind":"sequential","prefix":"item","width":4}`
	req := httptest.NewRequest(http.MethodPost, "/api/admin/settings/item-ids", strings.NewReader(body))
	req.AddCookie(&http.Cookie{Name: csrfCookieName, Value: tok})
	req.Header.Set(csrfHeaderName, tok)
	// Wrong Content-Type (would previously trigger ParseForm and drain the body).
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d body=%q", rr.Code, rr.Body.String())
	}
	if rr.Body.String() != body {
		t.Fatalf("handler body: got %q want %q", rr.Body.String(), body)
	}
}

func TestParseRequestForCSRF_ParsesDeclaredURLEncoded(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/api/x", strings.NewReader("csrf_token=abc&a=1"))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	if err := parseRequestForCSRF(r); err != nil {
		t.Fatal(err)
	}
	if r.FormValue("csrf_token") != "abc" {
		t.Fatalf("form not parsed: %#v", r.Form)
	}
}

func TestParseRequestForCSRF_DoesNotReadJSONContentType(t *testing.T) {
	body := `{"a":1}`
	r := httptest.NewRequest(http.MethodPost, "/api/x", strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json; charset=utf-8")

	if err := parseRequestForCSRF(r); err != nil {
		t.Fatal(err)
	}
	got, err := io.ReadAll(r.Body)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != body {
		t.Fatalf("got %q want %q", got, body)
	}
}
