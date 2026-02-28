package mcp

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestProtectedResourceMetadataHandler(t *testing.T) {
	h := ProtectedResourceMetadataHandler("https://mcp.example.com", "/mcp", "https://bitbucket.example.com")
	req := httptest.NewRequest("GET", "/.well-known/oauth-protected-resource", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("Code = %d", rec.Code)
	}
	if rec.Header().Get("Content-Type") != "application/json" {
		t.Errorf("Content-Type = %q", rec.Header().Get("Content-Type"))
	}
	var meta ProtectedResourceMetadata
	if err := json.NewDecoder(rec.Body).Decode(&meta); err != nil {
		t.Fatalf("Decode: %v", err)
	}
	if meta.Resource != "https://mcp.example.com/mcp" {
		t.Errorf("resource = %q", meta.Resource)
	}
	if len(meta.AuthorizationServers) != 1 || meta.AuthorizationServers[0] != "https://bitbucket.example.com" {
		t.Errorf("authorization_servers = %v", meta.AuthorizationServers)
	}
}

func TestProtectedResourceMetadataHandler_MethodNotAllowed(t *testing.T) {
	h := ProtectedResourceMetadataHandler("https://mcp.example.com", "/mcp", "https://bitbucket.example.com")
	req := httptest.NewRequest("POST", "/.well-known/oauth-protected-resource", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("Code = %d", rec.Code)
	}
}
