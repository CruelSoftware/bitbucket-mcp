package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func (s *Server) registerBranchTools() {
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "bitbucket_create_branch",
		Description: "Create a new branch",
	}, s.createBranch)
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "bitbucket_list_repository_branches",
		Description: "List repository branches",
	}, s.listRepositoryBranches)
}

type createBranchArgs struct {
	WorkspaceSlug string `json:"workspaceSlug" jsonschema:"description=Project key (default: BITBUCKET_DEFAULT_PROJECT)"`
	Repository   string `json:"repository" jsonschema:"required"`
	Name         string `json:"name" jsonschema:"required"`
	StartPoint   string `json:"startPoint" jsonschema:"description=Base branch (default: master)"`
}

func (s *Server) createBranch(ctx context.Context, req *mcp.CallToolRequest, args createBranchArgs) (*mcp.CallToolResult, any, error) {
	projectKey := s.projectKey(args.WorkspaceSlug)
	if projectKey == "" {
		return nil, nil, fmt.Errorf("workspaceSlug required (or set BITBUCKET_DEFAULT_PROJECT)")
	}
	opts := s.getOpts(ctx, req)
	branch, err := s.client.CreateBranch(ctx, projectKey, args.Repository, args.Name, args.StartPoint, opts)
	if err != nil {
		return nil, nil, err
	}
	data, _ := json.Marshal(branch)
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: string(data)}}}, nil, nil
}

type listBranchesArgs struct {
	WorkspaceSlug string `json:"workspaceSlug" jsonschema:"description=Project key (default: BITBUCKET_DEFAULT_PROJECT)"`
	Repository    string `json:"repository" jsonschema:"required"`
}

func (s *Server) listRepositoryBranches(ctx context.Context, req *mcp.CallToolRequest, args listBranchesArgs) (*mcp.CallToolResult, any, error) {
	projectKey := s.projectKey(args.WorkspaceSlug)
	if projectKey == "" {
		return nil, nil, fmt.Errorf("workspaceSlug required (or set BITBUCKET_DEFAULT_PROJECT)")
	}
	opts := s.getOpts(ctx, req)
	resp, err := s.client.ListBranches(ctx, projectKey, args.Repository, opts)
	if err != nil {
		return nil, nil, err
	}
	data, _ := json.Marshal(resp)
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: string(data)}}}, nil, nil
}
