package config

import (
	"net/url"
	"os"
	"strconv"
	"strings"
)

// Config holds server configuration from environment.
type Config struct {
	BitbucketURL       string
	MCPHTTPPort        int
	MCPHTTPEndpoint   string
	MCPPublicURL      string // Base URL for MCP server (e.g. https://mcp.example.com). Used for OAuth discovery. Default: http://localhost:{port}
	ProxyHeaders      []string
	ExtraHeaders      map[string]string
	DefaultProjectKey string
	LogLevel          string // "info" (default), "debug", or "off" - BITBUCKET_LOG_LEVEL
}

// Load reads configuration from environment variables.
func Load() (*Config, error) {
	bitbucketURL := os.Getenv("BITBUCKET_URL")
	if bitbucketURL == "" {
		return nil, &ConfigError{Field: "BITBUCKET_URL", Msg: "required"}
	}
	u, err := url.Parse(bitbucketURL)
	if err != nil {
		return nil, &ConfigError{Field: "BITBUCKET_URL", Msg: "invalid URL: " + err.Error()}
	}
	if u.Scheme == "" || u.Host == "" {
		return nil, &ConfigError{Field: "BITBUCKET_URL", Msg: "URL must include scheme and host"}
	}
	// Ensure no trailing slash for consistent path building
	bitbucketURL = strings.TrimSuffix(bitbucketURL, "/")

	port := 3001
	if p := os.Getenv("MCP_HTTP_PORT"); p != "" {
		if port, err = strconv.Atoi(p); err != nil || port < 1 || port > 65535 {
			return nil, &ConfigError{Field: "MCP_HTTP_PORT", Msg: "must be 1-65535"}
		}
	}

	endpoint := "/mcp"
	if e := os.Getenv("MCP_HTTP_ENDPOINT"); e != "" {
		endpoint = "/" + strings.Trim(e, "/")
		if endpoint == "/" {
			endpoint = "/mcp"
		}
	}

	proxyHeaders := parseList(os.Getenv("BITBUCKET_PROXY_HEADERS"))
	extraHeaders := parseExtraHeaders()
	defaultProject := strings.TrimSpace(os.Getenv("BITBUCKET_DEFAULT_PROJECT"))
	if defaultProject != "" && !validProjectKey(defaultProject) {
		return nil, &ConfigError{Field: "BITBUCKET_DEFAULT_PROJECT", Msg: "invalid format (use A-Z0-9_-)"}
	}

	logLevel := strings.ToLower(strings.TrimSpace(os.Getenv("BITBUCKET_LOG_LEVEL")))
	if logLevel == "" && os.Getenv("BITBUCKET_DEBUG") != "" {
		logLevel = "debug" // backward compat
	}
	if logLevel != "" && logLevel != "info" && logLevel != "debug" && logLevel != "off" {
		return nil, &ConfigError{Field: "BITBUCKET_LOG_LEVEL", Msg: "must be info, debug, or off"}
	}
	if logLevel == "" {
		logLevel = "info"
	}

	publicURL := strings.TrimSuffix(strings.TrimSpace(os.Getenv("MCP_PUBLIC_URL")), "/")
	if publicURL == "" {
		publicURL = "http://localhost:" + strconv.Itoa(port)
	} else if u, err := url.Parse(publicURL); err != nil || u.Scheme == "" || u.Host == "" {
		return nil, &ConfigError{Field: "MCP_PUBLIC_URL", Msg: "invalid URL (use e.g. https://mcp.example.com)"}
	}

	return &Config{
		BitbucketURL:       bitbucketURL,
		MCPHTTPPort:        port,
		MCPHTTPEndpoint:    endpoint,
		MCPPublicURL:       publicURL,
		ProxyHeaders:       proxyHeaders,
		ExtraHeaders:       extraHeaders,
		DefaultProjectKey:  defaultProject,
		LogLevel:           logLevel,
	}, nil
}

func validProjectKey(s string) bool {
	if len(s) == 0 || len(s) > 255 {
		return false
	}
	for _, r := range s {
		if (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			continue
		}
		return false
	}
	return true
}

func parseList(s string) []string {
	if s == "" {
		return nil
	}
	var out []string
	for _, v := range strings.Split(s, ",") {
		v = strings.TrimSpace(v)
		if v != "" {
			out = append(out, v)
		}
	}
	return out
}

func parseExtraHeaders() map[string]string {
	out := make(map[string]string)
	prefix := "BITBUCKET_EXTRA_HEADER_"
	for _, e := range os.Environ() {
		if !strings.HasPrefix(e, prefix) {
			continue
		}
		kv := strings.SplitN(e, "=", 2)
		if len(kv) != 2 {
			continue
		}
		key := strings.TrimPrefix(kv[0], prefix)
		key = strings.ReplaceAll(key, "_", "-")
		if key != "" && kv[1] != "" {
			out[key] = kv[1]
		}
	}
	return out
}

// ConfigError represents a configuration error.
type ConfigError struct {
	Field string
	Msg   string
}

func (e *ConfigError) Error() string {
	return "config: " + e.Field + ": " + e.Msg
}
