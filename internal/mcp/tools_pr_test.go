package mcp

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/n8n/bitbucket-mcp/internal/bitbucket"
)

func bbServer(mux *http.ServeMux) (*Server, *httptest.Server) {
	ts := httptest.NewServer(mux)
	client := bitbucket.NewClient(ts.URL, nil, "off")
	return NewServer(client, "PROJ"), ts
}

func TestCreatePullRequest(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/rest/api/1.0/projects/PROJ/repos/repo/pull-requests", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":1,"title":"PR","state":"OPEN"}`))
	})
	srv, ts := bbServer(mux)
	defer ts.Close()

	result, _, err := srv.createPullRequest(context.Background(), &sdkmcp.CallToolRequest{}, createPRArgs{
		Repository: "repo", Title: "PR", SourceBranch: "feat", TargetBranch: "main",
	})
	if err != nil {
		t.Fatalf("createPullRequest: %v", err)
	}
	if len(result.Content) == 0 {
		t.Error("expected content")
	}
}

func TestCreatePullRequest_NoWorkspace(t *testing.T) {
	mux := http.NewServeMux()
	srv, ts := bbServer(mux)
	defer ts.Close()
	srv.defaultProjectKey = ""

	_, _, err := srv.createPullRequest(context.Background(), &sdkmcp.CallToolRequest{}, createPRArgs{
		Repository: "repo", Title: "PR", SourceBranch: "feat", TargetBranch: "main",
	})
	if err == nil {
		t.Fatal("expected error for missing workspace")
	}
}

func TestCreatePullRequest_APIError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/rest/api/1.0/projects/PROJ/repos/repo/pull-requests", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(409)
		_, _ = w.Write([]byte(`conflict`))
	})
	srv, ts := bbServer(mux)
	defer ts.Close()

	_, _, err := srv.createPullRequest(context.Background(), &sdkmcp.CallToolRequest{}, createPRArgs{
		Repository: "repo", Title: "PR", SourceBranch: "feat", TargetBranch: "main",
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestGetPullRequestDetails(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/rest/api/1.0/projects/PROJ/repos/repo/pull-requests/42", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":42,"title":"PR","state":"OPEN"}`))
	})
	srv, ts := bbServer(mux)
	defer ts.Close()

	result, _, err := srv.getPullRequestDetails(context.Background(), &sdkmcp.CallToolRequest{}, getPRDetailsArgs{
		Repository: "repo", PrID: 42,
	})
	if err != nil {
		t.Fatalf("getPullRequestDetails: %v", err)
	}
	if len(result.Content) == 0 {
		t.Error("expected content")
	}
}

func TestGetPullRequestDetails_NoWorkspace(t *testing.T) {
	mux := http.NewServeMux()
	srv, ts := bbServer(mux)
	defer ts.Close()
	srv.defaultProjectKey = ""

	_, _, err := srv.getPullRequestDetails(context.Background(), &sdkmcp.CallToolRequest{}, getPRDetailsArgs{
		Repository: "repo", PrID: 1,
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestGetPullRequestDiff(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/rest/api/1.0/projects/PROJ/repos/repo/pull-requests/1.diff", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("diff content"))
	})
	srv, ts := bbServer(mux)
	defer ts.Close()

	result, _, err := srv.getPullRequestDiff(context.Background(), &sdkmcp.CallToolRequest{}, getPRDiffArgs{
		Repository: "repo", PrID: 1,
	})
	if err != nil {
		t.Fatalf("getPullRequestDiff: %v", err)
	}
	if len(result.Content) == 0 {
		t.Error("expected content")
	}
}

func TestGetPullRequestDiff_NoWorkspace(t *testing.T) {
	mux := http.NewServeMux()
	srv, ts := bbServer(mux)
	defer ts.Close()
	srv.defaultProjectKey = ""

	_, _, err := srv.getPullRequestDiff(context.Background(), &sdkmcp.CallToolRequest{}, getPRDiffArgs{
		Repository: "repo", PrID: 1,
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestGetPullRequestReviews(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/rest/api/1.0/projects/PROJ/repos/repo/pull-requests/1/participants", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"values":[{"user":{"name":"alice"},"approved":true}]}`))
	})
	srv, ts := bbServer(mux)
	defer ts.Close()

	result, _, err := srv.getPullRequestReviews(context.Background(), &sdkmcp.CallToolRequest{}, getPRReviewsArgs{
		Repository: "repo", PrID: 1,
	})
	if err != nil {
		t.Fatalf("getPullRequestReviews: %v", err)
	}
	if len(result.Content) == 0 {
		t.Error("expected content")
	}
}

func TestGetPullRequestReviews_NoWorkspace(t *testing.T) {
	mux := http.NewServeMux()
	srv, ts := bbServer(mux)
	defer ts.Close()
	srv.defaultProjectKey = ""

	_, _, err := srv.getPullRequestReviews(context.Background(), &sdkmcp.CallToolRequest{}, getPRReviewsArgs{
		Repository: "repo", PrID: 1,
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestMergePullRequest(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/rest/api/1.0/projects/PROJ/repos/repo/pull-requests/1/merge", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"id":1,"state":"MERGED"}`))
	})
	srv, ts := bbServer(mux)
	defer ts.Close()

	result, _, err := srv.mergePullRequest(context.Background(), &sdkmcp.CallToolRequest{}, mergePRArgs{
		Repository: "repo", PrID: 1, Version: 3,
	})
	if err != nil {
		t.Fatalf("mergePullRequest: %v", err)
	}
	if len(result.Content) == 0 {
		t.Error("expected content")
	}
}

func TestMergePullRequest_NoWorkspace(t *testing.T) {
	mux := http.NewServeMux()
	srv, ts := bbServer(mux)
	defer ts.Close()
	srv.defaultProjectKey = ""

	_, _, err := srv.mergePullRequest(context.Background(), &sdkmcp.CallToolRequest{}, mergePRArgs{
		Repository: "repo", PrID: 1, Version: 3,
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestDeclinePullRequest(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/rest/api/1.0/projects/PROJ/repos/repo/pull-requests/1/decline", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	})
	srv, ts := bbServer(mux)
	defer ts.Close()

	result, _, err := srv.declinePullRequest(context.Background(), &sdkmcp.CallToolRequest{}, declinePRArgs{
		Repository: "repo", PrID: 1, Version: 3,
	})
	if err != nil {
		t.Fatalf("declinePullRequest: %v", err)
	}
	if len(result.Content) == 0 {
		t.Error("expected content")
	}
}

func TestDeclinePullRequest_NoWorkspace(t *testing.T) {
	mux := http.NewServeMux()
	srv, ts := bbServer(mux)
	defer ts.Close()
	srv.defaultProjectKey = ""

	_, _, err := srv.declinePullRequest(context.Background(), &sdkmcp.CallToolRequest{}, declinePRArgs{
		Repository: "repo", PrID: 1, Version: 3,
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestAddPullRequestComment(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/rest/api/1.0/projects/PROJ/repos/repo/pull-requests/1/comments", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(201)
		_, _ = w.Write([]byte(`{"id":10}`))
	})
	srv, ts := bbServer(mux)
	defer ts.Close()

	result, _, err := srv.addPullRequestComment(context.Background(), &sdkmcp.CallToolRequest{}, addPRCommentArgs{
		Repository: "repo", PrID: 1, Text: "LGTM",
	})
	if err != nil {
		t.Fatalf("addPullRequestComment: %v", err)
	}
	if len(result.Content) == 0 {
		t.Error("expected content")
	}
}

func TestAddPullRequestComment_NoWorkspace(t *testing.T) {
	mux := http.NewServeMux()
	srv, ts := bbServer(mux)
	defer ts.Close()
	srv.defaultProjectKey = ""

	_, _, err := srv.addPullRequestComment(context.Background(), &sdkmcp.CallToolRequest{}, addPRCommentArgs{
		Repository: "repo", PrID: 1, Text: "text",
	})
	if err == nil {
		t.Fatal("expected error")
	}
}
