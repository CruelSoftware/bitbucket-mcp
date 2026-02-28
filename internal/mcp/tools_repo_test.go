package mcp

import (
	"context"
	"net/http"
	"testing"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestListRepositories(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/rest/api/1.0/projects/PROJ/repos", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"values":[{"slug":"repo","name":"Repo"}],"size":1,"isLastPage":true}`))
	})
	srv, ts := bbServer(mux)
	defer ts.Close()

	result, _, err := srv.listRepositories(context.Background(), &sdkmcp.CallToolRequest{}, listReposArgs{
		WorkspaceSlug: "PROJ",
	})
	if err != nil {
		t.Fatalf("listRepositories: %v", err)
	}
	if len(result.Content) == 0 {
		t.Error("expected content")
	}
}

func TestListRepositories_Error(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/rest/api/1.0/projects/PROJ/repos", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		_, _ = w.Write([]byte(`error`))
	})
	srv, ts := bbServer(mux)
	defer ts.Close()

	_, _, err := srv.listRepositories(context.Background(), &sdkmcp.CallToolRequest{}, listReposArgs{
		WorkspaceSlug: "PROJ",
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestGetRepositoryDetails(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/rest/api/1.0/projects/PROJ/repos/repo", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"slug":"repo","name":"Repo","id":1}`))
	})
	srv, ts := bbServer(mux)
	defer ts.Close()

	result, _, err := srv.getRepositoryDetails(context.Background(), &sdkmcp.CallToolRequest{}, getRepoDetailsArgs{
		WorkspaceSlug: "PROJ", RepoSlug: "repo",
	})
	if err != nil {
		t.Fatalf("getRepositoryDetails: %v", err)
	}
	if len(result.Content) == 0 {
		t.Error("expected content")
	}
}

func TestGetRepositoryDetails_Error(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/rest/api/1.0/projects/PROJ/repos/bad", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		_, _ = w.Write([]byte(`not found`))
	})
	srv, ts := bbServer(mux)
	defer ts.Close()

	_, _, err := srv.getRepositoryDetails(context.Background(), &sdkmcp.CallToolRequest{}, getRepoDetailsArgs{
		WorkspaceSlug: "PROJ", RepoSlug: "bad",
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestSearchContent(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/rest/search/1.0/search", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"values":[{"path":"main.go","content":"hello"}]}`))
	})
	srv, ts := bbServer(mux)
	defer ts.Close()

	result, _, err := srv.searchContent(context.Background(), &sdkmcp.CallToolRequest{}, searchContentArgs{
		Query: "hello",
	})
	if err != nil {
		t.Fatalf("searchContent: %v", err)
	}
	if len(result.Content) == 0 {
		t.Error("expected content")
	}
}

func TestSearchContent_Error(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/rest/search/1.0/search", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		_, _ = w.Write([]byte(`error`))
	})
	srv, ts := bbServer(mux)
	defer ts.Close()

	_, _, err := srv.searchContent(context.Background(), &sdkmcp.CallToolRequest{}, searchContentArgs{
		Query: "hello",
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestGetFileContent(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/rest/api/1.0/projects/PROJ/repos/repo/browse/file.go", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"lines":[{"text":"package main"}]}`))
	})
	srv, ts := bbServer(mux)
	defer ts.Close()

	result, _, err := srv.getFileContent(context.Background(), &sdkmcp.CallToolRequest{}, getFileContentArgs{
		WorkspaceSlug: "PROJ", RepoSlug: "repo", FilePath: "file.go",
	})
	if err != nil {
		t.Fatalf("getFileContent: %v", err)
	}
	if len(result.Content) == 0 {
		t.Error("expected content")
	}
}

func TestGetFileContent_Error(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/rest/api/1.0/projects/PROJ/repos/repo/browse/missing.go", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		_, _ = w.Write([]byte(`not found`))
	})
	srv, ts := bbServer(mux)
	defer ts.Close()

	_, _, err := srv.getFileContent(context.Background(), &sdkmcp.CallToolRequest{}, getFileContentArgs{
		WorkspaceSlug: "PROJ", RepoSlug: "repo", FilePath: "missing.go",
	})
	if err == nil {
		t.Fatal("expected error")
	}
}
