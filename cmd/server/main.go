package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/n8n/bitbucket-mcp/internal/config"
	"github.com/n8n/bitbucket-mcp/internal/middleware"
	"github.com/n8n/bitbucket-mcp/internal/mcp"
	"github.com/n8n/bitbucket-mcp/internal/bitbucket"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	client := bitbucket.NewClient(cfg.BitbucketURL, cfg.ExtraHeaders)
	srv := mcp.NewServer(client, cfg.DefaultProjectKey)

	// Middleware chain: header proxy -> auth -> MCP handler
	handler := middleware.ProxyHeaders(cfg.ProxyHeaders)(mcp.AuthMiddleware()(srv.Handler()))

	http.Handle(cfg.MCPHTTPEndpoint, handler)
	http.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"status":"healthy"}`)
	})

	addr := fmt.Sprintf(":%d", cfg.MCPHTTPPort)
	log.Printf("Bitbucket MCP server listening on %s%s", addr, cfg.MCPHTTPEndpoint)
	log.Fatal(http.ListenAndServe(addr, nil))
}
