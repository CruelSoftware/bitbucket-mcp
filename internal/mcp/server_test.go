package mcp

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	sdkauth "github.com/modelcontextprotocol/go-sdk/auth"
	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/n8n/bitbucket-mcp/internal/bitbucket"
)

func newTestBitbucket(handler http.Handler) (*bitbucket.Client, *httptest.Server) {
	ts := httptest.NewServer(handler)
	client := bitbucket.NewClient(ts.URL, nil, "off")
	return client, ts
}

func TestNewServer(t *testing.T) {
	client := bitbucket.NewClient("https://bb.example.com", nil, "off")
	srv := NewServer(client, "PROJ")
	if srv.mcpServer == nil {
		t.Error("mcpServer should not be nil")
	}
	if srv.client == nil {
		t.Error("client should not be nil")
	}
	if srv.defaultProjectKey != "PROJ" {
		t.Errorf("defaultProjectKey = %q", srv.defaultProjectKey)
	}
}

func TestProjectKey_WithSlug(t *testing.T) {
	srv := &Server{defaultProjectKey: "DEFAULT"}
	if got := srv.projectKey("CUSTOM"); got != "CUSTOM" {
		t.Errorf("projectKey = %q, want CUSTOM", got)
	}
}

func TestProjectKey_FallbackToDefault(t *testing.T) {
	srv := &Server{defaultProjectKey: "DEFAULT"}
	if got := srv.projectKey(""); got != "DEFAULT" {
		t.Errorf("projectKey = %q, want DEFAULT", got)
	}
}

func TestProjectKey_Empty(t *testing.T) {
	srv := &Server{defaultProjectKey: ""}
	if got := srv.projectKey(""); got != "" {
		t.Errorf("projectKey = %q, want empty", got)
	}
}

func TestGetOpts_NilExtra(t *testing.T) {
	srv := &Server{}
	req := &sdkmcp.CallToolRequest{}
	opts := srv.getOpts(context.Background(), req)
	if opts.Token != "" {
		t.Errorf("Token = %q", opts.Token)
	}
}

func TestGetOpts_WithToken(t *testing.T) {
	srv := &Server{}
	req := &sdkmcp.CallToolRequest{
		Extra: &sdkmcp.RequestExtra{
			TokenInfo: &sdkauth.TokenInfo{
				UserID:     "my-bearer-token",
				Expiration: time.Now().Add(time.Hour),
			},
		},
	}
	opts := srv.getOpts(context.Background(), req)
	if opts.Token != "my-bearer-token" {
		t.Errorf("Token = %q", opts.Token)
	}
}

func TestGetOpts_NilTokenInfo(t *testing.T) {
	srv := &Server{}
	req := &sdkmcp.CallToolRequest{
		Extra: &sdkmcp.RequestExtra{},
	}
	opts := srv.getOpts(context.Background(), req)
	if opts.Token != "" {
		t.Errorf("Token = %q", opts.Token)
	}
}

func TestHandler(t *testing.T) {
	client := bitbucket.NewClient("https://bb.example.com", nil, "off")
	srv := NewServer(client, "PROJ")
	h := srv.Handler()
	if h == nil {
		t.Error("Handler should not be nil")
	}
}

func TestHandler_HTTPRequest(t *testing.T) {
	client := bitbucket.NewClient("https://bb.example.com", nil, "off")
	srv := NewServer(client, "PROJ")
	h := srv.Handler()

	req := httptest.NewRequest("GET", "/mcp", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	// MCP handler should respond (even if not a valid MCP request, it should not panic)
	if rec.Code == 0 {
		t.Error("expected non-zero status code")
	}
}

func TestAuthMiddleware(t *testing.T) {
	mw := AuthMiddleware()
	if mw == nil {
		t.Error("AuthMiddleware should not be nil")
	}

	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	})
	wrapped := mw(inner)
	if wrapped == nil {
		t.Error("wrapped handler should not be nil")
	}
}

func TestListWorkspaces(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/rest/api/1.0/projects", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"values":[{"key":"PROJ","name":"Project"}],"size":1,"isLastPage":true}`))
	})
	client, ts := newTestBitbucket(mux)
	defer ts.Close()

	srv := NewServer(client, "PROJ")
	result, _, err := srv.listWorkspaces(context.Background(), &sdkmcp.CallToolRequest{}, struct{}{})
	if err != nil {
		t.Fatalf("listWorkspaces: %v", err)
	}
	if len(result.Content) == 0 {
		t.Error("expected content")
	}
}

func TestListWorkspaces_Error(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/rest/api/1.0/projects", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		_, _ = w.Write([]byte(`error`))
	})
	client, ts := newTestBitbucket(mux)
	defer ts.Close()

	srv := NewServer(client, "PROJ")
	_, _, err := srv.listWorkspaces(context.Background(), &sdkmcp.CallToolRequest{}, struct{}{})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestGetUserProfile(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/rest/api/1.0/users/current", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"name":"jdoe","displayName":"John","id":1,"active":true}`))
	})
	client, ts := newTestBitbucket(mux)
	defer ts.Close()

	srv := NewServer(client, "PROJ")
	result, _, err := srv.getUserProfile(context.Background(), &sdkmcp.CallToolRequest{}, struct{}{})
	if err != nil {
		t.Fatalf("getUserProfile: %v", err)
	}
	if len(result.Content) == 0 {
		t.Error("expected content")
	}
}

func TestGetUserProfile_Error(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/rest/api/1.0/users/current", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(401)
		_, _ = w.Write([]byte(`unauthorized`))
	})
	client, ts := newTestBitbucket(mux)
	defer ts.Close()

	srv := NewServer(client, "PROJ")
	_, _, err := srv.getUserProfile(context.Background(), &sdkmcp.CallToolRequest{}, struct{}{})
	if err == nil {
		t.Fatal("expected error")
	}
}
