package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestProxyHeaders(t *testing.T) {
	allowlist := []string{"X-Request-Id", "X-Custom-Header"}
	handler := ProxyHeaders(allowlist)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		headers := ProxyHeadersFromContext(r.Context())
		if headers == nil {
			t.Error("headers should not be nil")
			return
		}
		if headers["X-Request-Id"] != "req-123" {
			t.Errorf("X-Request-Id = %q", headers["X-Request-Id"])
		}
		if headers["X-Custom-Header"] != "custom-value" {
			t.Errorf("X-Custom-Header = %q", headers["X-Custom-Header"])
		}
	}))

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Request-Id", "req-123")
	req.Header.Set("X-Custom-Header", "custom-value")
	req.Header.Set("X-Ignored", "ignored")

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
}

func TestProxyHeaders_EmptyAllowlist(t *testing.T) {
	handler := ProxyHeaders(nil)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		headers := ProxyHeadersFromContext(r.Context())
		if headers != nil {
			t.Errorf("headers should be nil, got %v", headers)
		}
	}))

	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
}
