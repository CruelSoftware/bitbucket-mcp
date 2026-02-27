package bitbucket

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
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
	resp, err := c.do(ctx, http.MethodGet, path, nil, opts)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("get file failed %d: %s", resp.StatusCode, string(body))
	}
	var out BrowseResponse
	if err := decodeJSON(resp, &out); err != nil {
		return "", fmt.Errorf("get file content: %w", err)
	}
	var content string
	for _, line := range out.Lines {
		content += line.Text + "\n"
	}
	if len(content) > 0 && content[len(content)-1] == '\n' {
		content = content[:len(content)-1]
	}
	return content, nil
}
