package bitbucket

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// SearchResult represents a code search result (Bitbucket Data Center 8+).
type SearchResult struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

// SearchResponse is the API response for code search.
type SearchResponse struct {
	Values []SearchResult `json:"values"`
}

// SearchContent searches for code in repositories.
// Note: Code search requires Bitbucket Data Center 8+ with search enabled.
// For older versions, returns empty results.
func (c *Client) SearchContent(ctx context.Context, workspaceSlug, query, extension string, opts RequestOpts) (*SearchResponse, error) {
	// Bitbucket Server/DC search API varies by version. Try common paths.
	// /rest/search/1.0/search or /rest/api/1.0/search
	apiPath := "/search?searchQuery=" + url.QueryEscape(query)
	if workspaceSlug != "" {
		apiPath += "&projectKey=" + url.QueryEscape(workspaceSlug)
	}
	if extension != "" {
		apiPath += "&fileExtension=" + url.QueryEscape(extension)
	}
	resp, err := c.doSearch(ctx, http.MethodGet, apiPath, nil, opts)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode == 404 || resp.StatusCode == 501 {
		return &SearchResponse{Values: []SearchResult{}}, nil
	}
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("search failed %d: %s (code search requires Bitbucket Data Center 8+)", resp.StatusCode, string(body))
	}
	var out SearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("search decode: %w", err)
	}
	return &out, nil
}
