package mcp

import (
	"encoding/json"
	"net/http"
)

// ProtectedResourceMetadata is the OAuth 2.0 Protected Resource Metadata (RFC 9728).
type ProtectedResourceMetadata struct {
	Resource               string   `json:"resource"`
	AuthorizationServers   []string `json:"authorization_servers"`
	ScopesSupported        []string `json:"scopes_supported,omitempty"`
	BearerMethodsSupported []string `json:"bearer_methods_supported,omitempty"`
}

// ProtectedResourceMetadataHandler returns an http.Handler that serves Protected Resource Metadata at
// /.well-known/oauth-protected-resource and path-specific variants.
func ProtectedResourceMetadataHandler(publicURL, mcpEndpoint, bitbucketURL string) http.Handler {
	resource := publicURL + mcpEndpoint
	meta := ProtectedResourceMetadata{
		Resource:               resource,
		AuthorizationServers:   []string{bitbucketURL},
		ScopesSupported:        []string{"REPO_READ", "REPO_WRITE", "PROJECT_READ"},
		BearerMethodsSupported: []string{"header"},
	}
	body, _ := json.Marshal(meta)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.Header().Set("Allow", "GET")
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "public, max-age=3600")
		_, _ = w.Write(body)
	})
}
