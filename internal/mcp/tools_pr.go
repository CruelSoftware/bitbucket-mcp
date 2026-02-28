package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/n8n/bitbucket-mcp/internal/bitbucket"
)

func (s *Server) registerPRTools() {
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "bitbucket_create_pull_request",
		Description: "Create a new pull request",
	}, s.createPullRequest)
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "bitbucket_get_pull_request_details",
		Description: "Get pull request details and metadata",
	}, s.getPullRequestDetails)
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "bitbucket_get_pull_request_diff",
		Description: "Retrieve the diff for a pull request",
	}, s.getPullRequestDiff)
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "bitbucket_get_pull_request_reviews",
		Description: "Get PR review status (participants)",
	}, s.getPullRequestReviews)
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "bitbucket_merge_pull_request",
		Description: "Merge a pull request",
	}, s.mergePullRequest)
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "bitbucket_decline_pull_request",
		Description: "Decline a pull request",
	}, s.declinePullRequest)
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "bitbucket_add_pull_request_comment",
		Description: "Add a general comment to a pull request",
	}, s.addPullRequestComment)
}

type createPRArgs struct {
	Repository   string `json:"repository" jsonschema:"required,Repository slug"`
	Title        string `json:"title" jsonschema:"required"`
	SourceBranch string `json:"sourceBranch" jsonschema:"required"`
	TargetBranch string `json:"targetBranch" jsonschema:"required"`
	Description  string `json:"description"`
	WorkspaceSlug string `json:"workspaceSlug" jsonschema:"Project/workspace key"`
}

func (s *Server) createPullRequest(ctx context.Context, req *mcp.CallToolRequest, args createPRArgs) (*mcp.CallToolResult, any, error) {
	opts := s.getOpts(ctx, req)
	projectKey := s.projectKey(args.WorkspaceSlug)
	if projectKey == "" {
		return nil, nil, fmt.Errorf("workspaceSlug required (or set BITBUCKET_DEFAULT_PROJECT)")
	}
	createReq := bitbucket.NewCreatePRRequest(projectKey, args.Repository, args.SourceBranch, args.TargetBranch, args.Title, args.Description)
	pr, err := s.client.CreatePullRequest(ctx, projectKey, args.Repository, createReq, opts)
	if err != nil {
		return nil, nil, err
	}
	data, err := json.Marshal(pr)
	if err != nil {
		return nil, nil, fmt.Errorf("marshal response: %w", err)
	}
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: string(data)}}}, nil, nil
}

type getPRDetailsArgs struct {
	Repository     string `json:"repository" jsonschema:"required"`
	PrID           int    `json:"prId" jsonschema:"required"`
	WorkspaceSlug  string `json:"workspaceSlug"`
}

func (s *Server) getPullRequestDetails(ctx context.Context, req *mcp.CallToolRequest, args getPRDetailsArgs) (*mcp.CallToolResult, any, error) {
	opts := s.getOpts(ctx, req)
	projectKey := s.projectKey(args.WorkspaceSlug)
	if projectKey == "" {
		return nil, nil, fmt.Errorf("workspaceSlug required")
	}
	pr, err := s.client.GetPullRequest(ctx, projectKey, args.Repository, args.PrID, opts)
	if err != nil {
		return nil, nil, err
	}
	data, err := json.Marshal(pr)
	if err != nil {
		return nil, nil, fmt.Errorf("marshal response: %w", err)
	}
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: string(data)}}}, nil, nil
}

type getPRDiffArgs struct {
	Repository    string `json:"repository" jsonschema:"required"`
	PrID          int    `json:"prId" jsonschema:"required"`
	WorkspaceSlug string `json:"workspaceSlug"`
}

func (s *Server) getPullRequestDiff(ctx context.Context, req *mcp.CallToolRequest, args getPRDiffArgs) (*mcp.CallToolResult, any, error) {
	opts := s.getOpts(ctx, req)
	projectKey := s.projectKey(args.WorkspaceSlug)
	if projectKey == "" {
		return nil, nil, fmt.Errorf("workspaceSlug required")
	}
	diff, err := s.client.GetPullRequestDiff(ctx, projectKey, args.Repository, args.PrID, opts)
	if err != nil {
		return nil, nil, err
	}
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: diff}}}, nil, nil
}

type getPRReviewsArgs struct {
	Repository    string `json:"repository" jsonschema:"required"`
	PrID          int    `json:"prId" jsonschema:"required"`
	WorkspaceSlug string `json:"workspaceSlug"`
}

func (s *Server) getPullRequestReviews(ctx context.Context, req *mcp.CallToolRequest, args getPRReviewsArgs) (*mcp.CallToolResult, any, error) {
	opts := s.getOpts(ctx, req)
	projectKey := s.projectKey(args.WorkspaceSlug)
	if projectKey == "" {
		return nil, nil, fmt.Errorf("workspaceSlug required")
	}
	resp, err := s.client.GetPullRequestParticipants(ctx, projectKey, args.Repository, args.PrID, opts)
	if err != nil {
		return nil, nil, err
	}
	data, err := json.Marshal(resp)
	if err != nil {
		return nil, nil, fmt.Errorf("marshal response: %w", err)
	}
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: string(data)}}}, nil, nil
}

type mergePRArgs struct {
	Repository    string `json:"repository" jsonschema:"required"`
	PrID          int    `json:"prId" jsonschema:"required"`
	Version       int    `json:"version" jsonschema:"required,PR version from get_pull_request_details"`
	WorkspaceSlug string `json:"workspaceSlug"`
}

func (s *Server) mergePullRequest(ctx context.Context, req *mcp.CallToolRequest, args mergePRArgs) (*mcp.CallToolResult, any, error) {
	opts := s.getOpts(ctx, req)
	projectKey := s.projectKey(args.WorkspaceSlug)
	if projectKey == "" {
		return nil, nil, fmt.Errorf("workspaceSlug required")
	}
	err := s.client.MergePullRequest(ctx, projectKey, args.Repository, args.PrID, args.Version, opts)
	if err != nil {
		return nil, nil, err
	}
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: "merged"}}}, nil, nil
}

type declinePRArgs struct {
	Repository    string `json:"repository" jsonschema:"required"`
	PrID          int    `json:"prId" jsonschema:"required"`
	Version       int    `json:"version" jsonschema:"required"`
	WorkspaceSlug string `json:"workspaceSlug"`
}

func (s *Server) declinePullRequest(ctx context.Context, req *mcp.CallToolRequest, args declinePRArgs) (*mcp.CallToolResult, any, error) {
	opts := s.getOpts(ctx, req)
	projectKey := s.projectKey(args.WorkspaceSlug)
	if projectKey == "" {
		return nil, nil, fmt.Errorf("workspaceSlug required")
	}
	err := s.client.DeclinePullRequest(ctx, projectKey, args.Repository, args.PrID, args.Version, opts)
	if err != nil {
		return nil, nil, err
	}
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: "declined"}}}, nil, nil
}

type addPRCommentArgs struct {
	Repository    string `json:"repository" jsonschema:"required"`
	PrID          int    `json:"prId" jsonschema:"required"`
	Text          string `json:"text" jsonschema:"required"`
	WorkspaceSlug string `json:"workspaceSlug"`
}

func (s *Server) addPullRequestComment(ctx context.Context, req *mcp.CallToolRequest, args addPRCommentArgs) (*mcp.CallToolResult, any, error) {
	opts := s.getOpts(ctx, req)
	projectKey := s.projectKey(args.WorkspaceSlug)
	if projectKey == "" {
		return nil, nil, fmt.Errorf("workspaceSlug required")
	}
	err := s.client.AddPullRequestComment(ctx, projectKey, args.Repository, args.PrID, args.Text, opts)
	if err != nil {
		return nil, nil, err
	}
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: "comment added"}}}, nil, nil
}
