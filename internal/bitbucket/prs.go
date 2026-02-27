package bitbucket

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// PullRequest represents a Bitbucket pull request.
type PullRequest struct {
	ID          int    `json:"id"`
	Version     int    `json:"version"`
	Title       string `json:"title"`
	Description string `json:"description"`
	State       string `json:"state"`
	Open        bool   `json:"open"`
	Closed      bool   `json:"closed"`
	FromRef     *Ref   `json:"fromRef"`
	ToRef       *Ref   `json:"toRef"`
	Author      *User  `json:"author"`
}

// Ref represents a branch reference.
type Ref struct {
	ID         string      `json:"id"`
	DisplayID  string      `json:"displayId"`
	Repository *Repository `json:"repository"`
}

// RefInput is a minimal ref for create PR (repository with project key).
type RefInput struct {
	ID         string               `json:"id"`
	Repository *RepositoryRefInput  `json:"repository"`
}

type RepositoryRefInput struct {
	Slug    string         `json:"slug"`
	Project *ProjectRefInput `json:"project"`
}

type ProjectRefInput struct {
	Key string `json:"key"`
}

// CreatePRRequest is the request body for creating a PR.
type CreatePRRequest struct {
	Title       string    `json:"title"`
	Description string    `json:"description"`
	FromRef     *RefInput `json:"fromRef"`
	ToRef       *RefInput `json:"toRef"`
}

// NewCreatePRRequest builds a CreatePRRequest for the same repository.
func NewCreatePRRequest(projectKey, repoSlug, sourceBranch, targetBranch, title, description string) CreatePRRequest {
	repo := &RepositoryRefInput{Slug: repoSlug, Project: &ProjectRefInput{Key: projectKey}}
	fromID := sourceBranch
	if !strings.HasPrefix(fromID, "refs/heads/") {
		fromID = "refs/heads/" + sourceBranch
	}
	toID := targetBranch
	if !strings.HasPrefix(toID, "refs/heads/") {
		toID = "refs/heads/" + targetBranch
	}
	return CreatePRRequest{
		Title:       title,
		Description: description,
		FromRef:     &RefInput{ID: fromID, Repository: repo},
		ToRef:       &RefInput{ID: toID, Repository: repo},
	}
}

// CreatePullRequest creates a new pull request.
func (c *Client) CreatePullRequest(ctx context.Context, projectKey, repoSlug string, req CreatePRRequest, opts RequestOpts) (*PullRequest, error) {
	path := "/projects/" + url.PathEscape(projectKey) + "/repos/" + url.PathEscape(repoSlug) + "/pull-requests"
	body, _ := json.Marshal(req)
	resp, err := c.do(ctx, http.MethodPost, path, bytes.NewReader(body), opts)
	if err != nil {
		return nil, err
	}
	var out PullRequest
	if err := decodeJSON(resp, &out); err != nil {
		return nil, fmt.Errorf("create pull request: %w", err)
	}
	return &out, nil
}

// GetPullRequest returns pull request details.
func (c *Client) GetPullRequest(ctx context.Context, projectKey, repoSlug string, prID int, opts RequestOpts) (*PullRequest, error) {
	path := fmt.Sprintf("/projects/%s/repos/%s/pull-requests/%d",
		url.PathEscape(projectKey), url.PathEscape(repoSlug), prID)
	resp, err := c.do(ctx, http.MethodGet, path, nil, opts)
	if err != nil {
		return nil, err
	}
	var out PullRequest
	if err := decodeJSON(resp, &out); err != nil {
		return nil, fmt.Errorf("get pull request: %w", err)
	}
	return &out, nil
}

// MergePullRequest merges a pull request.
func (c *Client) MergePullRequest(ctx context.Context, projectKey, repoSlug string, prID, version int, opts RequestOpts) error {
	path := fmt.Sprintf("/projects/%s/repos/%s/pull-requests/%d/merge?version=%d",
		url.PathEscape(projectKey), url.PathEscape(repoSlug), prID, version)
	resp, err := c.do(ctx, http.MethodPost, path, nil, opts)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("merge failed %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

// DeclinePullRequest declines a pull request.
func (c *Client) DeclinePullRequest(ctx context.Context, projectKey, repoSlug string, prID, version int, opts RequestOpts) error {
	path := fmt.Sprintf("/projects/%s/repos/%s/pull-requests/%d/decline?version=%d",
		url.PathEscape(projectKey), url.PathEscape(repoSlug), prID, version)
	resp, err := c.do(ctx, http.MethodPost, path, nil, opts)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("decline failed %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

// GetPullRequestDiff returns the raw diff for a pull request.
func (c *Client) GetPullRequestDiff(ctx context.Context, projectKey, repoSlug string, prID int, opts RequestOpts) (string, error) {
	path := fmt.Sprintf("/projects/%s/repos/%s/pull-requests/%d.diff",
		url.PathEscape(projectKey), url.PathEscape(repoSlug), prID)
	resp, err := c.do(ctx, http.MethodGet, path, nil, opts)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("get diff failed %d: %s", resp.StatusCode, string(body))
	}
	body, err := io.ReadAll(resp.Body)
	return string(body), err
}

// AddPRCommentRequest is the request body for adding a PR comment.
type AddPRCommentRequest struct {
	Text string `json:"text"`
}

// AddPullRequestComment adds a general comment to a pull request.
func (c *Client) AddPullRequestComment(ctx context.Context, projectKey, repoSlug string, prID int, text string, opts RequestOpts) error {
	path := fmt.Sprintf("/projects/%s/repos/%s/pull-requests/%d/comments",
		url.PathEscape(projectKey), url.PathEscape(repoSlug), prID)
	body, _ := json.Marshal(AddPRCommentRequest{Text: text})
	resp, err := c.do(ctx, http.MethodPost, path, bytes.NewReader(body), opts)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("add comment failed %d: %s", resp.StatusCode, string(b))
	}
	return nil
}

// Participant represents a PR participant (reviewer).
type Participant struct {
	User       *User `json:"user"`
	Approved   bool  `json:"approved"`
	Status     string `json:"status"`
}

// ParticipantsResponse is the API response for PR participants.
type ParticipantsResponse struct {
	Values []Participant `json:"values"`
}

// GetPullRequestParticipants returns PR participants (reviewers).
func (c *Client) GetPullRequestParticipants(ctx context.Context, projectKey, repoSlug string, prID int, opts RequestOpts) (*ParticipantsResponse, error) {
	path := fmt.Sprintf("/projects/%s/repos/%s/pull-requests/%d/participants",
		url.PathEscape(projectKey), url.PathEscape(repoSlug), prID)
	resp, err := c.do(ctx, http.MethodGet, path, nil, opts)
	if err != nil {
		return nil, err
	}
	var out ParticipantsResponse
	if err := decodeJSON(resp, &out); err != nil {
		return nil, fmt.Errorf("get participants: %w", err)
	}
	return &out, nil
}
