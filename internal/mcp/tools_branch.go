package mcp

import (
	"context"
	"encoding/json"

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
	WorkspaceSlug string `json:"workspaceSlug" jsonschema:"required"`
	Repository   string `json:"repository" jsonschema:"required"`
	Name         string `json:"name" jsonschema:"required"`
	StartPoint   string `json:"startPoint" jsonschema:"Base branch (default: master)"`
}

func (s *Server) createBranch(ctx context.Context, req *mcp.CallToolRequest, args createBranchArgs) (*mcp.CallToolResult, any, error) {
	opts := s.getOpts(ctx, req)
	branch, err := s.client.CreateBranch(ctx, args.WorkspaceSlug, args.Repository, args.Name, args.StartPoint, opts)
	if err != nil {
		return nil, nil, err
	}
	data, _ := json.Marshal(branch)
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: string(data)}}}, nil, nil
}

type listBranchesArgs struct {
	WorkspaceSlug string `json:"workspaceSlug" jsonschema:"required"`
	Repository    string `json:"repository" jsonschema:"required"`
}

func (s *Server) listRepositoryBranches(ctx context.Context, req *mcp.CallToolRequest, args listBranchesArgs) (*mcp.CallToolResult, any, error) {
	opts := s.getOpts(ctx, req)
	resp, err := s.client.ListBranches(ctx, args.WorkspaceSlug, args.Repository, opts)
	if err != nil {
		return nil, nil, err
	}
	data, _ := json.Marshal(resp)
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: string(data)}}}, nil, nil
}
