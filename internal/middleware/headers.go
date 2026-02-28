package middleware

import (
	"context"
	"net/http"
	"strings"
)

type contextKey string

const proxyHeadersKey contextKey = "bitbucket_proxy_headers"

// ProxyHeaders extracts allowlisted headers from the request and stores them in context.
// Headers are stored as map[string]string with canonical names (e.g. X-Request-Id).
func ProxyHeaders(allowlist []string) func(http.Handler) http.Handler {
	if len(allowlist) == 0 {
		return func(next http.Handler) http.Handler { return next }
	}
	set := make(map[string]bool)
	for _, h := range allowlist {
		set[strings.TrimSpace(h)] = true
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			headers := make(map[string]string)
			for k, v := range r.Header {
				canon := http.CanonicalHeaderKey(k)
				if set[canon] && len(v) > 0 {
					headers[canon] = v[0]
				}
			}
			ctx := context.WithValue(r.Context(), proxyHeadersKey, headers)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// ProxyHeadersFromContext returns the proxied headers from context, or nil.
func ProxyHeadersFromContext(ctx context.Context) map[string]string {
	v := ctx.Value(proxyHeadersKey)
	if v == nil {
		return nil
	}
	return v.(map[string]string)
}
