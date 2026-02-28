package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTrailingSlash_StripsSlash(t *testing.T) {
	handler := TrailingSlash(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/mcp" {
			t.Errorf("path = %q, want /mcp", r.URL.Path)
		}
		w.WriteHeader(200)
	}))
	req := httptest.NewRequest("POST", "/mcp/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != 200 {
		t.Errorf("Code = %d", rec.Code)
	}
}

func TestTrailingSlash_LeavesPathWithoutSlash(t *testing.T) {
	handler := TrailingSlash(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/health" {
			t.Errorf("path = %q, want /health", r.URL.Path)
		}
		w.WriteHeader(200)
	}))
	req := httptest.NewRequest("GET", "/health", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != 200 {
		t.Errorf("Code = %d", rec.Code)
	}
}

func TestTrailingSlash_LeavesRoot(t *testing.T) {
	handler := TrailingSlash(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			t.Errorf("path = %q, want /", r.URL.Path)
		}
		w.WriteHeader(200)
	}))
	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != 200 {
		t.Errorf("Code = %d", rec.Code)
	}
}
