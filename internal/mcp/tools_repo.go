package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func (s *Server) registerRepoTools() {
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "bitbucket_list_repositories",
		Description: "List repositories in a project/workspace",
	}, s.listRepositories)
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "bitbucket_get_repository_details",
		Description: "Get repository information",
	}, s.getRepositoryDetails)
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "bitbucket_search_content",
		Description: "Search for code in repositories (Bitbucket Data Center 8+ with search enabled)",
	}, s.searchContent)
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "bitbucket_get_file_content",
		Description: "Read file contents from a repository",
	}, s.getFileContent)
}

type listReposArgs struct {
	WorkspaceSlug string `json:"workspaceSlug" jsonschema:"required,Project key"`
}

func (s *Server) listRepositories(ctx context.Context, req *mcp.CallToolRequest, args listReposArgs) (*mcp.CallToolResult, any, error) {
	opts := s.getOpts(ctx, req)
	resp, err := s.client.ListRepositories(ctx, args.WorkspaceSlug, opts)
	if err != nil {
		return nil, nil, err
	}
	data, err := json.Marshal(resp)
	if err != nil {
		return nil, nil, fmt.Errorf("marshal response: %w", err)
	}
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: string(data)}}}, nil, nil
}

type getRepoDetailsArgs struct {
	WorkspaceSlug string `json:"workspaceSlug" jsonschema:"required"`
	RepoSlug      string `json:"repoSlug" jsonschema:"required"`
}

func (s *Server) getRepositoryDetails(ctx context.Context, req *mcp.CallToolRequest, args getRepoDetailsArgs) (*mcp.CallToolResult, any, error) {
	opts := s.getOpts(ctx, req)
	repo, err := s.client.GetRepository(ctx, args.WorkspaceSlug, args.RepoSlug, opts)
	if err != nil {
		return nil, nil, err
	}
	data, err := json.Marshal(repo)
	if err != nil {
		return nil, nil, fmt.Errorf("marshal response: %w", err)
	}
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: string(data)}}}, nil, nil
}

type searchContentArgs struct {
	WorkspaceSlug string `json:"workspaceSlug"`
	Query         string `json:"query" jsonschema:"required"`
	Extension     string `json:"extension"`
}

func (s *Server) searchContent(ctx context.Context, req *mcp.CallToolRequest, args searchContentArgs) (*mcp.CallToolResult, any, error) {
	opts := s.getOpts(ctx, req)
	resp, err := s.client.SearchContent(ctx, args.WorkspaceSlug, args.Query, args.Extension, opts)
	if err != nil {
		return nil, nil, err
	}
	data, err := json.Marshal(resp)
	if err != nil {
		return nil, nil, fmt.Errorf("marshal response: %w", err)
	}
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: string(data)}}}, nil, nil
}

type getFileContentArgs struct {
	WorkspaceSlug string `json:"workspaceSlug" jsonschema:"Project key (default: BITBUCKET_DEFAULT_PROJECT)"`
	RepoSlug      string `json:"repoSlug" jsonschema:"required"`
	FilePath      string `json:"filePath" jsonschema:"required"`
	Ref           string `json:"ref" jsonschema:"Ref or branch (e.g. refs/heads/master)"`
}

func (s *Server) getFileContent(ctx context.Context, req *mcp.CallToolRequest, args getFileContentArgs) (*mcp.CallToolResult, any, error) {
	opts := s.getOpts(ctx, req)
	content, err := s.client.GetFileContent(ctx, args.WorkspaceSlug, args.RepoSlug, args.FilePath, args.Ref, opts)
	if err != nil {
		return nil, nil, err
	}
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: content}}}, nil, nil
}