package bitbucket

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// BrowseLine represents a line of file content.
type BrowseLine struct {
	Text string `json:"text"`
}

// BrowseResponse is the API response for file content.
type BrowseResponse struct {
	Lines []BrowseLine `json:"lines"`
}

// GetFileContent returns the content of a file at the given path.
func (c *Client) GetFileContent(ctx context.Context, projectKey, repoSlug, filePath, ref string, opts RequestOpts) (string, error) {
	path := "/projects/" + url.PathEscape(projectKey) + "/repos/" + url.PathEscape(repoSlug) + "/browse/" + url.PathEscape(filePath)
	if ref != "" {
		path += "?at=" + url.QueryEscape(ref)
	}
	var out BrowseResponse
	if err := c.doJSON(ctx, c.api, http.MethodGet, path, nil, &out, opts); err != nil {
		return "", fmt.Errorf("get file content: %w", err)
	}
	var b strings.Builder
	for i, line := range out.Lines {
		if i > 0 {
			b.WriteByte('\n')
		}
		b.WriteString(line.Text)
	}
	return b.String(), nil
}
