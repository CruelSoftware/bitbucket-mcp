package middleware

import (
	"net/http"
	"strings"
)

// TrailingSlash strips trailing slashes from the request path so /mcp/ and /health/
// are handled the same as /mcp and /health. Root path "/" is left unchanged.
func TrailingSlash(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if len(r.URL.Path) > 1 && strings.HasSuffix(r.URL.Path, "/") {
			r.URL.Path = strings.TrimSuffix(r.URL.Path, "/")
		}
		next.ServeHTTP(w, r)
	})
}
