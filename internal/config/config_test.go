package config

import (
	"os"
	"testing"
)

func clearEnv() {
	for _, key := range []string{
		"BITBUCKET_URL", "MCP_HTTP_PORT", "MCP_HTTP_ENDPOINT",
		"BITBUCKET_PROXY_HEADERS", "BITBUCKET_DEFAULT_PROJECT",
		"BITBUCKET_LOG_LEVEL", "BITBUCKET_DEBUG",
	} {
		_ = os.Unsetenv(key)
	}
}

func TestLoad(t *testing.T) {
	clearEnv()
	_ = os.Setenv("BITBUCKET_URL", "https://bitbucket.example.com")
	defer clearEnv()

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
	clearEnv()
	_, err := Load()
	if err == nil {
		t.Fatal("expected error for missing BITBUCKET_URL")
	}
}

func TestLoad_InvalidURL(t *testing.T) {
	clearEnv()
	_ = os.Setenv("BITBUCKET_URL", "://invalid")
	defer clearEnv()

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for invalid URL")
	}
}

func TestLoad_URLWithoutScheme(t *testing.T) {
	clearEnv()
	_ = os.Setenv("BITBUCKET_URL", "bitbucket.example.com")
	defer clearEnv()

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for URL without scheme")
	}
}

func TestLoad_TrailingSlash(t *testing.T) {
	clearEnv()
	_ = os.Setenv("BITBUCKET_URL", "https://bitbucket.example.com/")
	defer clearEnv()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.BitbucketURL != "https://bitbucket.example.com" {
		t.Errorf("BitbucketURL = %q (trailing slash not removed)", cfg.BitbucketURL)
	}
}

func TestLoad_CustomPort(t *testing.T) {
	clearEnv()
	_ = os.Setenv("BITBUCKET_URL", "https://bitbucket.example.com")
	_ = os.Setenv("MCP_HTTP_PORT", "8080")
	defer clearEnv()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.MCPHTTPPort != 8080 {
		t.Errorf("MCPHTTPPort = %d", cfg.MCPHTTPPort)
	}
}

func TestLoad_InvalidPort(t *testing.T) {
	clearEnv()
	_ = os.Setenv("BITBUCKET_URL", "https://bitbucket.example.com")
	_ = os.Setenv("MCP_HTTP_PORT", "99999")
	defer clearEnv()

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for invalid port")
	}
}

func TestLoad_NonNumericPort(t *testing.T) {
	clearEnv()
	_ = os.Setenv("BITBUCKET_URL", "https://bitbucket.example.com")
	_ = os.Setenv("MCP_HTTP_PORT", "abc")
	defer clearEnv()

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for non-numeric port")
	}
}

func TestLoad_ZeroPort(t *testing.T) {
	clearEnv()
	_ = os.Setenv("BITBUCKET_URL", "https://bitbucket.example.com")
	_ = os.Setenv("MCP_HTTP_PORT", "0")
	defer clearEnv()

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for port 0")
	}
}

func TestLoad_CustomEndpoint(t *testing.T) {
	clearEnv()
	_ = os.Setenv("BITBUCKET_URL", "https://bitbucket.example.com")
	_ = os.Setenv("MCP_HTTP_ENDPOINT", "/api/mcp")
	defer clearEnv()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.MCPHTTPEndpoint != "/api/mcp" {
		t.Errorf("MCPHTTPEndpoint = %q", cfg.MCPHTTPEndpoint)
	}
}

func TestLoad_EmptyEndpointFallback(t *testing.T) {
	clearEnv()
	_ = os.Setenv("BITBUCKET_URL", "https://bitbucket.example.com")
	_ = os.Setenv("MCP_HTTP_ENDPOINT", "/")
	defer clearEnv()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.MCPHTTPEndpoint != "/mcp" {
		t.Errorf("MCPHTTPEndpoint = %q, want /mcp", cfg.MCPHTTPEndpoint)
	}
}

func TestLoad_ProxyHeaders(t *testing.T) {
	clearEnv()
	_ = os.Setenv("BITBUCKET_URL", "https://bitbucket.example.com")
	_ = os.Setenv("BITBUCKET_PROXY_HEADERS", "X-Request-Id, X-Custom")
	defer clearEnv()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(cfg.ProxyHeaders) != 2 {
		t.Fatalf("ProxyHeaders len = %d", len(cfg.ProxyHeaders))
	}
	if cfg.ProxyHeaders[0] != "X-Request-Id" || cfg.ProxyHeaders[1] != "X-Custom" {
		t.Errorf("ProxyHeaders = %v", cfg.ProxyHeaders)
	}
}

func TestLoad_ExtraHeaders(t *testing.T) {
	clearEnv()
	_ = os.Setenv("BITBUCKET_URL", "https://bitbucket.example.com")
	_ = os.Setenv("BITBUCKET_EXTRA_HEADER_X_MY_HEADER", "myval")
	defer func() {
		clearEnv()
		_ = os.Unsetenv("BITBUCKET_EXTRA_HEADER_X_MY_HEADER")
	}()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.ExtraHeaders["X-MY-HEADER"] != "myval" {
		t.Errorf("ExtraHeaders = %v", cfg.ExtraHeaders)
	}
}

func TestLoad_DefaultProject(t *testing.T) {
	clearEnv()
	_ = os.Setenv("BITBUCKET_URL", "https://bitbucket.example.com")
	_ = os.Setenv("BITBUCKET_DEFAULT_PROJECT", "MYPROJ")
	defer clearEnv()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.DefaultProjectKey != "MYPROJ" {
		t.Errorf("DefaultProjectKey = %q", cfg.DefaultProjectKey)
	}
}

func TestLoad_InvalidDefaultProject(t *testing.T) {
	clearEnv()
	_ = os.Setenv("BITBUCKET_URL", "https://bitbucket.example.com")
	_ = os.Setenv("BITBUCKET_DEFAULT_PROJECT", "../../etc")
	defer clearEnv()

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for invalid project key")
	}
}

func TestLoad_LogLevelDebug(t *testing.T) {
	clearEnv()
	_ = os.Setenv("BITBUCKET_URL", "https://bitbucket.example.com")
	_ = os.Setenv("BITBUCKET_LOG_LEVEL", "debug")
	defer clearEnv()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.LogLevel != "debug" {
		t.Errorf("LogLevel = %q", cfg.LogLevel)
	}
}

func TestLoad_LogLevelOff(t *testing.T) {
	clearEnv()
	_ = os.Setenv("BITBUCKET_URL", "https://bitbucket.example.com")
	_ = os.Setenv("BITBUCKET_LOG_LEVEL", "off")
	defer clearEnv()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.LogLevel != "off" {
		t.Errorf("LogLevel = %q", cfg.LogLevel)
	}
}

func TestLoad_InvalidLogLevel(t *testing.T) {
	clearEnv()
	_ = os.Setenv("BITBUCKET_URL", "https://bitbucket.example.com")
	_ = os.Setenv("BITBUCKET_LOG_LEVEL", "verbose")
	defer clearEnv()

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for invalid log level")
	}
}

func TestLoad_BackwardCompatDebug(t *testing.T) {
	clearEnv()
	_ = os.Setenv("BITBUCKET_URL", "https://bitbucket.example.com")
	_ = os.Setenv("BITBUCKET_DEBUG", "1")
	defer clearEnv()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.LogLevel != "debug" {
		t.Errorf("LogLevel = %q, want debug (backward compat)", cfg.LogLevel)
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
		{"a", true},
		{"ABC123", true},
		{"", false},
		{"proj/test", false},
		{"../../../etc", false},
		{"proj space", false},
		{"proj@key", false},
	}
	for _, tt := range tests {
		if got := validProjectKey(tt.s); got != tt.want {
			t.Errorf("validProjectKey(%q) = %v, want %v", tt.s, got, tt.want)
		}
	}
}

func TestParseList(t *testing.T) {
	tests := []struct {
		input string
		want  int
	}{
		{"", 0},
		{"a", 1},
		{"a,b,c", 3},
		{" a , b , c ", 3},
		{",,,", 0},
		{"a,,b", 2},
	}
	for _, tt := range tests {
		got := parseList(tt.input)
		if len(got) != tt.want {
			t.Errorf("parseList(%q) = %d items, want %d", tt.input, len(got), tt.want)
		}
	}
}

func TestConfigError(t *testing.T) {
	e := &ConfigError{Field: "TEST", Msg: "bad value"}
	if e.Error() != "config: TEST: bad value" {
		t.Errorf("Error = %q", e.Error())
	}
}
