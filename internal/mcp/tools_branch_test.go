package mcp

import (
	"context"
	"net/http"
	"testing"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestCreateBranch(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/rest/api/1.0/projects/PROJ/repos/repo/branches", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"refs/heads/feat","displayId":"feat"}`))
	})
	srv, ts := bbServer(mux)
	defer ts.Close()

	result, _, err := srv.createBranch(context.Background(), &sdkmcp.CallToolRequest{}, createBranchArgs{
		Repository: "repo", Name: "feat", StartPoint: "main",
	})
	if err != nil {
		t.Fatalf("createBranch: %v", err)
	}
	if len(result.Content) == 0 {
		t.Error("expected content")
	}
}

func TestCreateBranch_NoWorkspace(t *testing.T) {
	mux := http.NewServeMux()
	srv, ts := bbServer(mux)
	defer ts.Close()
	srv.defaultProjectKey = ""

	_, _, err := srv.createBranch(context.Background(), &sdkmcp.CallToolRequest{}, createBranchArgs{
		Repository: "repo", Name: "feat",
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestCreateBranch_APIError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/rest/api/1.0/projects/PROJ/repos/repo/branches", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(409)
		_, _ = w.Write([]byte(`conflict`))
	})
	srv, ts := bbServer(mux)
	defer ts.Close()

	_, _, err := srv.createBranch(context.Background(), &sdkmcp.CallToolRequest{}, createBranchArgs{
		Repository: "repo", Name: "feat", StartPoint: "main",
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestListRepositoryBranches(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/rest/api/1.0/projects/PROJ/repos/repo/branches", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"values":[{"id":"refs/heads/main","displayId":"main"}],"size":1,"isLastPage":true}`))
	})
	srv, ts := bbServer(mux)
	defer ts.Close()

	result, _, err := srv.listRepositoryBranches(context.Background(), &sdkmcp.CallToolRequest{}, listBranchesArgs{
		Repository: "repo",
	})
	if err != nil {
		t.Fatalf("listRepositoryBranches: %v", err)
	}
	if len(result.Content) == 0 {
		t.Error("expected content")
	}
}

func TestListRepositoryBranches_NoWorkspace(t *testing.T) {
	mux := http.NewServeMux()
	srv, ts := bbServer(mux)
	defer ts.Close()
	srv.defaultProjectKey = ""

	_, _, err := srv.listRepositoryBranches(context.Background(), &sdkmcp.CallToolRequest{}, listBranchesArgs{
		Repository: "repo",
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestListRepositoryBranches_APIError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/rest/api/1.0/projects/PROJ/repos/repo/branches", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		_, _ = w.Write([]byte(`error`))
	})
	srv, ts := bbServer(mux)
	defer ts.Close()

	_, _, err := srv.listRepositoryBranches(context.Background(), &sdkmcp.CallToolRequest{}, listBranchesArgs{
		Repository: "repo",
	})
	if err == nil {
		t.Fatal("expected error")
	}
}
