package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	sdkauth "github.com/modelcontextprotocol/go-sdk/auth"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/n8n/bitbucket-mcp/internal/auth"
	"github.com/n8n/bitbucket-mcp/internal/bitbucket"
)

// Server wraps the MCP server and Bitbucket client.
type Server struct {
	mcpServer        *mcp.Server
	client           *bitbucket.Client
	defaultProjectKey string
}

// NewServer creates an MCP server with Bitbucket tools.
func NewServer(client *bitbucket.Client, defaultProjectKey string) *Server {
	s := &Server{
		mcpServer:        mcp.NewServer(&mcp.Implementation{Name: "bitbucket-mcp", Version: "1.0.0"}, nil),
		client:           client,
		defaultProjectKey: defaultProjectKey,
	}
	s.registerTools()
	return s
}

func (s *Server) registerTools() {
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "bitbucket_list_workspaces",
		Description: "List all workspaces (projects) the user can access",
	}, s.listWorkspaces)
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "bitbucket_get_user_profile",
		Description: "Get the authenticated user's profile",
	}, s.getUserProfile)
	s.registerPRTools()
	s.registerRepoTools()
	s.registerBranchTools()
}

func (s *Server) projectKey(slug string) string {
	if slug != "" {
		return slug
	}
	return s.defaultProjectKey
}

func (s *Server) getOpts(ctx context.Context, req *mcp.CallToolRequest) bitbucket.RequestOpts {
	token := ""
	if req.Extra != nil && req.Extra.TokenInfo != nil {
		token = req.Extra.TokenInfo.UserID // raw Bearer token from passthrough verifier
	}
	return bitbucket.RequestOptsFromContext(ctx, token)
}

func (s *Server) listWorkspaces(ctx context.Context, req *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
	opts := s.getOpts(ctx, req)
	resp, err := s.client.ListWorkspaces(ctx, opts)
	if err != nil {
		return nil, nil, fmt.Errorf("list workspaces: %w", err)
	}
	data, err := json.Marshal(resp)
	if err != nil {
		return nil, nil, err
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: string(data)}},
	}, nil, nil
}

func (s *Server) getUserProfile(ctx context.Context, req *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
	opts := s.getOpts(ctx, req)
	user, err := s.client.GetCurrentUser(ctx, opts)
	if err != nil {
		return nil, nil, fmt.Errorf("get user profile: %w", err)
	}
	data, err := json.Marshal(user)
	if err != nil {
		return nil, nil, err
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: string(data)}},
	}, nil, nil
}

// Handler returns the HTTP handler for the MCP server.
func (s *Server) Handler() http.Handler {
	return mcp.NewStreamableHTTPHandler(func(r *http.Request) *mcp.Server {
		return s.mcpServer
	}, nil)
}

// AuthMiddleware returns the auth middleware (RequireBearerToken with passthrough verifier).
// resourceMetadataURL is the full URL to Protected Resource Metadata (RFC 9728), e.g. https://mcp.example.com/.well-known/oauth-protected-resource/mcp.
// If empty, WWW-Authenticate on 401 will not include resource_metadata.
func AuthMiddleware(resourceMetadataURL string) func(http.Handler) http.Handler {
	opts := &sdkauth.RequireBearerTokenOptions{
		ResourceMetadataURL: resourceMetadataURL,
		Scopes:              []string{"REPO_READ", "REPO_WRITE"},
	}
	return sdkauth.RequireBearerToken(auth.PassthroughVerifier(), opts)
}
