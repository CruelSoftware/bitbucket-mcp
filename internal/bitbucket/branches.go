package bitbucket

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// Branch represents a Bitbucket branch.
type Branch struct {
	ID              string `json:"id"`
	DisplayID       string `json:"displayId"`
	LatestCommit    string `json:"latestCommit"`
	LatestChangeset string `json:"latestChangeset"`
	IsDefault       bool   `json:"isDefault"`
}

// BranchesResponse is the paginated API response for listing branches.
type BranchesResponse struct {
	Values        []Branch `json:"values"`
	Size          int      `json:"size"`
	Limit         int      `json:"limit"`
	IsLastPage    bool     `json:"isLastPage"`
	Start         int      `json:"start"`
	NextPageStart int      `json:"nextPageStart"`
}

// ListBranches returns branches for a repository.
func (c *Client) ListBranches(ctx context.Context, projectKey, repoSlug string, opts RequestOpts) (*BranchesResponse, error) {
	path := "/projects/" + url.PathEscape(projectKey) + "/repos/" + url.PathEscape(repoSlug) + "/branches"
	var out BranchesResponse
	if err := c.doJSON(ctx, c.api, http.MethodGet, path, nil, &out, opts); err != nil {
		return nil, fmt.Errorf("list branches: %w", err)
	}
	return &out, nil
}

// CreateBranchRequest is the request body for creating a branch.
type CreateBranchRequest struct {
	Name       string `json:"name"`
	StartPoint string `json:"startPoint"`
}

// CreateBranch creates a new branch.
func (c *Client) CreateBranch(ctx context.Context, projectKey, repoSlug string, name, startPoint string, opts RequestOpts) (*Branch, error) {
	path := "/projects/" + url.PathEscape(projectKey) + "/repos/" + url.PathEscape(repoSlug) + "/branches"
	req := CreateBranchRequest{Name: name, StartPoint: startPoint}
	if req.StartPoint == "" {
		req.StartPoint = "refs/heads/master"
	} else if !strings.HasPrefix(req.StartPoint, "refs/") {
		req.StartPoint = "refs/heads/" + req.StartPoint
	}
	var out Branch
	if err := c.doJSON(ctx, c.api, http.MethodPost, path, req, &out, opts); err != nil {
		return nil, fmt.Errorf("create branch: %w", err)
	}
	return &out, nil
}
