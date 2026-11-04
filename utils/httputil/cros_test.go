package httputil_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/magic-lib/go-plat-utils/utils/httputil"
)

func corsHeaders(origins []string, reqOrigin string) http.Header {
	handler := httputil.CROSMiddleware(nil, origins...)(func(w http.ResponseWriter, r *http.Request) {})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	if reqOrigin != "" {
		req.Header.Set("Origin", reqOrigin)
	}
	rec := httptest.NewRecorder()
	handler(rec, req)
	return rec.Header()
}

func TestCORSWildcardNoCredentials(t *testing.T) {
	for _, origins := range [][]string{nil, {"*"}, {"a.com", "*"}} {
		h := corsHeaders(origins, "https://evil.example")
		if got := h.Get("Access-Control-Allow-Origin"); got != "*" {
			t.Errorf("origins=%v: Allow-Origin = %q, want *", origins, got)
		}
		if got := h.Get("Access-Control-Allow-Credentials"); got != "" {
			t.Errorf("origins=%v: 通配时不应允许携带凭证, got %q", origins, got)
		}
	}
}

func TestCORSExplicitOrigin(t *testing.T) {
	h := corsHeaders([]string{"example.com"}, "https://api.example.com")
	if got := h.Get("Access-Control-Allow-Origin"); got != "https://api.example.com" {
		t.Errorf("Allow-Origin = %q", got)
	}
	if got := h.Get("Access-Control-Allow-Credentials"); got != "true" {
		t.Errorf("显式配置的域名应允许携带凭证, got %q", got)
	}

	h = corsHeaders([]string{"example.com"}, "https://evil.example")
	if got := h.Get("Access-Control-Allow-Origin"); got != "" {
		t.Errorf("未在白名单的来源不应设置 Allow-Origin, got %q", got)
	}
}
