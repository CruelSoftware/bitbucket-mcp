package middleware

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestLogRequests_Off(t *testing.T) {
	handler := LogRequests("off", nil)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))
	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != 200 {
		t.Errorf("Code = %d", rec.Code)
	}
}

func TestLogRequests_ExtractsMethod(t *testing.T) {
	var buf bytes.Buffer
	logger := log.New(&buf, "", 0)
	handler := LogRequests("info", logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		if !strings.Contains(string(body), "initialize") {
			t.Errorf("body truncated: %s", body)
		}
		w.WriteHeader(200)
	}))
	req := httptest.NewRequest("POST", "/mcp", strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != 200 {
		t.Errorf("Code = %d", rec.Code)
	}
	if !strings.Contains(buf.String(), "initialize") {
		t.Errorf("log should contain 'initialize', got %q", buf.String())
	}
}

func TestLogRequests_FallbackToPath(t *testing.T) {
	var buf bytes.Buffer
	logger := log.New(&buf, "", 0)
	handler := LogRequests("info", logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))
	req := httptest.NewRequest("GET", "/health", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if !strings.Contains(buf.String(), "GET /health") {
		t.Errorf("log should contain 'GET /health', got %q", buf.String())
	}
}

func TestLogRequests_NoContentType(t *testing.T) {
	var buf bytes.Buffer
	logger := log.New(&buf, "", 0)
	handler := LogRequests("info", logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))
	req := httptest.NewRequest("POST", "/mcp", strings.NewReader(`{"method":"tools/list"}`))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if !strings.Contains(buf.String(), "POST /mcp") {
		t.Errorf("log should fallback to 'POST /mcp', got %q", buf.String())
	}
}

func TestLogRequests_ToolsCall(t *testing.T) {
	var buf bytes.Buffer
	logger := log.New(&buf, "", 0)
	handler := LogRequests("info", logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))
	req := httptest.NewRequest("POST", "/mcp", strings.NewReader(`{"method":"tools/call","params":{"name":"bitbucket_list_workspaces"}}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if !strings.Contains(buf.String(), "tools/call") {
		t.Errorf("log should contain 'tools/call', got %q", buf.String())
	}
}

func TestLogRequests_NilLogger(t *testing.T) {
	handler := LogRequests("info", nil)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))
	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != 200 {
		t.Errorf("Code = %d", rec.Code)
	}
}
