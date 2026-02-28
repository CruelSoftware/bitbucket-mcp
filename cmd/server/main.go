package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/n8n/bitbucket-mcp/internal/bitbucket"
	"github.com/n8n/bitbucket-mcp/internal/config"
	"github.com/n8n/bitbucket-mcp/internal/mcp"
	"github.com/n8n/bitbucket-mcp/internal/middleware"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	client := bitbucket.NewClient(cfg.BitbucketURL, cfg.ExtraHeaders, cfg.LogLevel)
	srv := mcp.NewServer(client, cfg.DefaultProjectKey)

	handler := middleware.ProxyHeaders(cfg.ProxyHeaders)(mcp.AuthMiddleware()(srv.Handler()))

	mux := http.NewServeMux()
	mux.Handle(cfg.MCPHTTPEndpoint, handler)
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, `{"status":"healthy"}`)
	})

	addr := fmt.Sprintf(":%d", cfg.MCPHTTPPort)
	server := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		log.Println("shutting down...")
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		_ = server.Shutdown(ctx)
	}()

	log.Printf("Bitbucket MCP server listening on %s%s", addr, cfg.MCPHTTPEndpoint)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server: %v", err)
	}
}
