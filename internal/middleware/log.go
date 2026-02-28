package middleware

import (
	"bytes"
	"io"
	"log"
	"net/http"
)

// LogRequests logs MCP requests when logLevel is not "off".
// If logger is nil, log.Default() is used.
func LogRequests(logLevel string, logger *log.Logger) func(http.Handler) http.Handler {
	if logLevel == "off" {
		return func(next http.Handler) http.Handler { return next }
	}
	if logger == nil {
		logger = log.Default()
	}
	l := logger
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			method := ""
			if r.Method == http.MethodPost && r.Header.Get("Content-Type") == "application/json" {
				body, _ := io.ReadAll(r.Body)
				_ = r.Body.Close()
				r.Body = io.NopCloser(bytes.NewReader(body))
				if m := extractJSONMethod(body); m != "" {
					method = m
				}
			}
			if method == "" {
				method = r.Method + " " + r.URL.Path
			}
			l.Printf("MCP request: %s", method)
			next.ServeHTTP(w, r)
		})
	}
}

func extractJSONMethod(body []byte) string {
	const prefix = `"method"`
	i := bytes.Index(body, []byte(prefix))
	if i < 0 {
		return ""
	}
	i += len(prefix)
	// Skip to value: ":" and optional space, then "
	for i < len(body) && body[i] != '"' {
		i++
	}
	if i >= len(body) || body[i] != '"' {
		return ""
	}
	i++ // skip opening "
	start := i
	for i < len(body) && body[i] != '"' {
		i++
	}
	if i > start {
		return string(body[start:i])
	}
	return ""
}