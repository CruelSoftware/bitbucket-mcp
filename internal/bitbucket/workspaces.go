package bitbucket

import (
	"context"
	"fmt"
	"net/http"
)

// Workspace represents a Bitbucket workspace (project).
type Workspace struct {
	Key  string `json:"key"`
	Name string `json:"name"`
}

// WorkspacesResponse is the API response for listing workspaces.
type WorkspacesResponse struct {
	Values        []Workspace `json:"values"`
	Size          int         `json:"size"`
	Limit         int         `json:"limit"`
	IsLastPage    bool        `json:"isLastPage"`
	Start         int         `json:"start"`
	NextPageStart int         `json:"nextPageStart"`
}

// ListWorkspaces returns all workspaces (projects) the user can access.
func (c *Client) ListWorkspaces(ctx context.Context, opts RequestOpts) (*WorkspacesResponse, error) {
	resp, err := c.do(ctx, http.MethodGet, "/projects", nil, opts)
	if err != nil {
		return nil, err
	}
	var out WorkspacesResponse
	if err := decodeJSON(resp, &out); err != nil {
		return nil, fmt.Errorf("list workspaces: %w", err)
	}
	return &out, nil
}
