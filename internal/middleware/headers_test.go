package middleware

import (
	"context"
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
		if _, exists := headers["X-Ignored"]; exists {
			t.Error("X-Ignored should not be proxied")
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

func TestProxyHeaders_NoMatchingHeaders(t *testing.T) {
	allowlist := []string{"X-Special"}
	handler := ProxyHeaders(allowlist)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		headers := ProxyHeadersFromContext(r.Context())
		if headers == nil {
			t.Error("headers map should not be nil")
			return
		}
		if len(headers) != 0 {
			t.Errorf("expected empty headers map, got %v", headers)
		}
	}))

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Other", "value")

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
}

func TestProxyHeaders_WhitespaceInAllowlist(t *testing.T) {
	allowlist := []string{"  X-Trimmed  "}
	handler := ProxyHeaders(allowlist)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		headers := ProxyHeadersFromContext(r.Context())
		if headers == nil {
			t.Error("headers should not be nil")
			return
		}
		if headers["X-Trimmed"] != "val" {
			t.Errorf("X-Trimmed = %q", headers["X-Trimmed"])
		}
	}))

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Trimmed", "val")

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
}

func TestProxyHeadersFromContext_NilContext(t *testing.T) {
	ctx := context.Background()
	headers := ProxyHeadersFromContext(ctx)
	if headers != nil {
		t.Errorf("expected nil, got %v", headers)
	}
}

func TestProxyHeadersFromContext_WithHeaders(t *testing.T) {
	hdrs := map[string]string{"X-Test": "val"}
	ctx := context.WithValue(context.Background(), proxyHeadersKey, hdrs)
	got := ProxyHeadersFromContext(ctx)
	if got == nil {
		t.Fatal("expected headers")
	}
	if got["X-Test"] != "val" {
		t.Errorf("X-Test = %q", got["X-Test"])
	}
}

func TestWithProxyHeaders(t *testing.T) {
	ctx := WithProxyHeaders(context.Background(), map[string]string{"X-Test": "val"})
	got := ProxyHeadersFromContext(ctx)
	if got == nil {
		t.Fatal("expected headers")
	}
	if got["X-Test"] != "val" {
		t.Errorf("X-Test = %q", got["X-Test"])
	}
}

func TestProxyHeaders_EmptySliceAllowlist(t *testing.T) {
	handler := ProxyHeaders([]string{})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		headers := ProxyHeadersFromContext(r.Context())
		if headers != nil {
			t.Errorf("headers should be nil for empty slice, got %v", headers)
		}
	}))

	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
}
