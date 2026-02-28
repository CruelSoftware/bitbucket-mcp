package config

import (
	"os"
	"testing"
)

func TestLoad(t *testing.T) {
	_ = os.Setenv("BITBUCKET_URL", "https://bitbucket.example.com")
	_ = os.Unsetenv("MCP_HTTP_PORT")
	_ = os.Unsetenv("MCP_HTTP_ENDPOINT")
	_ = os.Unsetenv("BITBUCKET_PROXY_HEADERS")
	_ = os.Unsetenv("BITBUCKET_DEFAULT_PROJECT")
	_ = os.Unsetenv("BITBUCKET_LOG_LEVEL")
	_ = os.Unsetenv("BITBUCKET_DEBUG")
	defer func() {
		_ = os.Unsetenv("BITBUCKET_URL")
	}()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.BitbucketURL != "https://bitbucket.example.com" {
		t.Errorf("BitbucketURL = %q", cfg.BitbucketURL)
	}
	if cfg.MCPHTTPPort != 3001 {
		t.Errorf("MCPHTTPPort = %d", cfg.MCPHTTPPort)
	}
	if cfg.MCPHTTPEndpoint != "/mcp" {
		t.Errorf("MCPHTTPEndpoint = %q", cfg.MCPHTTPEndpoint)
	}
	if cfg.LogLevel != "info" {
		t.Errorf("LogLevel = %q, want info", cfg.LogLevel)
	}
}

func TestLoad_MissingURL(t *testing.T) {
	_ = os.Unsetenv("BITBUCKET_URL")
	_, err := Load()
	if err == nil {
		t.Fatal("expected error for missing BITBUCKET_URL")
	}
}

func TestValidProjectKey(t *testing.T) {
	tests := []struct {
		s    string
		want bool
	}{
		{"PROJ", true},
		{"my-project", true},
		{"TEAM_1", true},
		{"", false},
		{"proj/test", false},
		{"../../../etc", false},
	}
	for _, tt := range tests {
		if got := validProjectKey(tt.s); got != tt.want {
			t.Errorf("validProjectKey(%q) = %v, want %v", tt.s, got, tt.want)
		}
	}
}
