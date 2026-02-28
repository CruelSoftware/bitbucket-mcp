package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/n8n/bitbucket-mcp/internal/middleware"
)

func main() {
	port := flag.Int("port", 7990, "port to listen on")
	flag.Parse()

	mux := http.NewServeMux()

	// Projects (workspaces)
	mux.HandleFunc("/rest/api/1.0/projects", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"values":[{"key":"FAKE","name":"Fake Project"}],"size":1,"limit":25,"isLastPage":true,"start":0}`))
	})

	// Current user
	mux.HandleFunc("/rest/api/1.0/users/current", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"name":"fakeuser","displayName":"Fake User","id":1,"active":true}`))
	})

	// Repos (for bitbucket_list_repos)
	mux.HandleFunc("/rest/api/1.0/projects/FAKE/repos", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"values":[{"slug":"my-repo","name":"my-repo"}],"size":1,"isLastPage":true}`))
	})

	// Branches (for bitbucket_list_branches)
	mux.HandleFunc("/rest/api/1.0/projects/FAKE/repos/my-repo/branches", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"values":[{"id":"refs/heads/main","displayId":"main"}],"size":1,"isLastPage":true}`))
	})

	// Search (for bitbucket_search_code)
	mux.HandleFunc("/rest/search/1.0/search", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"query":"test","entities":{"code":{"count":0,"limit":25}}}`))
	})

	addr := fmt.Sprintf(":%d", *port)
	log.Printf("Fake Bitbucket server listening on %s", addr)
	log.Printf("Set BITBUCKET_URL=http://localhost%s", addr)
	if err := http.ListenAndServe(addr, middleware.TrailingSlash(mux)); err != nil {
		log.Fatal(err)
	}
}
