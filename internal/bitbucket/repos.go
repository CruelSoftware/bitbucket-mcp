package bitbucket

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// Repository represents a Bitbucket repository.
type Repository struct {
	Slug    string `json:"slug"`
	Name    string `json:"name"`
	ID     int    `json:"id"`
	Project *struct {
		Key  string `json:"key"`
		Name string `json:"name"`
	} `json:"project"`
}

// ReposResponse is the paginated API response for listing repos.
type ReposResponse struct {
	Values        []Repository `json:"values"`
	Size          int          `json:"size"`
	Limit         int          `json:"limit"`
	IsLastPage    bool         `json:"isLastPage"`
	Start         int          `json:"start"`
	NextPageStart int          `json:"nextPageStart"`
}

// ListRepositories returns repositories for a project (workspace).
func (c *Client) ListRepositories(ctx context.Context, projectKey string, opts RequestOpts) (*ReposResponse, error) {
	path := "/projects/" + url.PathEscape(projectKey) + "/repos"
	resp, err := c.do(ctx, http.MethodGet, path, nil, opts)
	if err != nil {
		return nil, err
	}
	var out ReposResponse
	if err := decodeJSON(resp, &out); err != nil {
		return nil, fmt.Errorf("list repositories: %w", err)
	}
	return &out, nil
}

// GetRepository returns repository details.
func (c *Client) GetRepository(ctx context.Context, projectKey, repoSlug string, opts RequestOpts) (*Repository, error) {
	path := "/projects/" + url.PathEscape(projectKey) + "/repos/" + url.PathEscape(repoSlug)
	resp, err := c.do(ctx, http.MethodGet, path, nil, opts)
	if err != nil {
		return nil, err
	}
	var out Repository
	if err := decodeJSON(resp, &out); err != nil {
		return nil, fmt.Errorf("get repository: %w", err)
	}
	return &out, nil
}
