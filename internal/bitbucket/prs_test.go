package bitbucket

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestNewCreatePRRequest(t *testing.T) {
	req := NewCreatePRRequest("PROJ", "repo", "feature", "main", "Title", "Desc")
	if req.FromRef.ID != "refs/heads/feature" {
		t.Errorf("FromRef.ID = %q", req.FromRef.ID)
	}
	if req.ToRef.ID != "refs/heads/main" {
		t.Errorf("ToRef.ID = %q", req.ToRef.ID)
	}
	if req.Title != "Title" {
		t.Errorf("Title = %q", req.Title)
	}
	if req.FromRef.Repository.Slug != "repo" {
		t.Errorf("FromRef.Repository.Slug = %q", req.FromRef.Repository.Slug)
	}
	if req.FromRef.Repository.Project.Key != "PROJ" {
		t.Errorf("FromRef.Repository.Project.Key = %q", req.FromRef.Repository.Project.Key)
	}
}

func TestNewCreatePRRequest_WithRefsPrefix(t *testing.T) {
	req := NewCreatePRRequest("PROJ", "repo", "refs/heads/feature", "refs/heads/main", "T", "D")
	if req.FromRef.ID != "refs/heads/feature" {
		t.Errorf("FromRef.ID = %q (should not double-prefix)", req.FromRef.ID)
	}
	if req.ToRef.ID != "refs/heads/main" {
		t.Errorf("ToRef.ID = %q", req.ToRef.ID)
	}
}

func TestCreatePullRequest(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/rest/api/1.0/projects/PROJ/repos/repo/pull-requests", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":1,"version":0,"title":"PR Title","state":"OPEN","open":true}`))
	})
	client, ts := newTestServer(mux)
	defer ts.Close()

	pr, err := client.CreatePullRequest(context.Background(), "PROJ", "repo",
		NewCreatePRRequest("PROJ", "repo", "feature", "main", "PR Title", ""),
		RequestOpts{})
	if err != nil {
		t.Fatalf("CreatePullRequest: %v", err)
	}
	if pr.ID != 1 {
		t.Errorf("ID = %d", pr.ID)
	}
	if pr.Title != "PR Title" {
		t.Errorf("Title = %q", pr.Title)
	}
}

func TestCreatePullRequest_Error(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/rest/api/1.0/projects/PROJ/repos/repo/pull-requests", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(409)
		_, _ = w.Write([]byte(`{"errors":[{"message":"already exists"}]}`))
	})
	client, ts := newTestServer(mux)
	defer ts.Close()

	_, err := client.CreatePullRequest(context.Background(), "PROJ", "repo",
		NewCreatePRRequest("PROJ", "repo", "feature", "main", "T", ""),
		RequestOpts{})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestGetPullRequest(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/rest/api/1.0/projects/PROJ/repos/repo/pull-requests/42", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":42,"version":3,"title":"My PR","state":"OPEN"}`))
	})
	client, ts := newTestServer(mux)
	defer ts.Close()

	pr, err := client.GetPullRequest(context.Background(), "PROJ", "repo", 42, RequestOpts{})
	if err != nil {
		t.Fatalf("GetPullRequest: %v", err)
	}
	if pr.ID != 42 || pr.Version != 3 {
		t.Errorf("ID=%d Version=%d", pr.ID, pr.Version)
	}
}

func TestGetPullRequest_Error(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/rest/api/1.0/projects/PROJ/repos/repo/pull-requests/999", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		_, _ = w.Write([]byte(`not found`))
	})
	client, ts := newTestServer(mux)
	defer ts.Close()

	_, err := client.GetPullRequest(context.Background(), "PROJ", "repo", 999, RequestOpts{})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestMergePullRequest(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/rest/api/1.0/projects/PROJ/repos/repo/pull-requests/1/merge", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s", r.Method)
		}
		if r.URL.Query().Get("version") != "3" {
			t.Errorf("version = %q", r.URL.Query().Get("version"))
		}
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"id":1,"state":"MERGED"}`))
	})
	client, ts := newTestServer(mux)
	defer ts.Close()

	err := client.MergePullRequest(context.Background(), "PROJ", "repo", 1, 3, RequestOpts{})
	if err != nil {
		t.Fatalf("MergePullRequest: %v", err)
	}
}

func TestMergePullRequest_Error(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/rest/api/1.0/projects/PROJ/repos/repo/pull-requests/1/merge", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(409)
		_, _ = w.Write([]byte(`conflict`))
	})
	client, ts := newTestServer(mux)
	defer ts.Close()

	err := client.MergePullRequest(context.Background(), "PROJ", "repo", 1, 3, RequestOpts{})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestDeclinePullRequest(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/rest/api/1.0/projects/PROJ/repos/repo/pull-requests/1/decline", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s", r.Method)
		}
		w.WriteHeader(200)
	})
	client, ts := newTestServer(mux)
	defer ts.Close()

	err := client.DeclinePullRequest(context.Background(), "PROJ", "repo", 1, 3, RequestOpts{})
	if err != nil {
		t.Fatalf("DeclinePullRequest: %v", err)
	}
}

func TestDeclinePullRequest_Error(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/rest/api/1.0/projects/PROJ/repos/repo/pull-requests/1/decline", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		_, _ = w.Write([]byte(`error`))
	})
	client, ts := newTestServer(mux)
	defer ts.Close()

	err := client.DeclinePullRequest(context.Background(), "PROJ", "repo", 1, 3, RequestOpts{})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestGetPullRequestDiff(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/rest/api/1.0/projects/PROJ/repos/repo/pull-requests/1.diff", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("diff --git a/file.go b/file.go\n+new line"))
	})
	client, ts := newTestServer(mux)
	defer ts.Close()

	diff, err := client.GetPullRequestDiff(context.Background(), "PROJ", "repo", 1, RequestOpts{})
	if err != nil {
		t.Fatalf("GetPullRequestDiff: %v", err)
	}
	if diff == "" {
		t.Error("diff should not be empty")
	}
}

func TestGetPullRequestDiff_Error(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/rest/api/1.0/projects/PROJ/repos/repo/pull-requests/1.diff", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		_, _ = w.Write([]byte(`not found`))
	})
	client, ts := newTestServer(mux)
	defer ts.Close()

	_, err := client.GetPullRequestDiff(context.Background(), "PROJ", "repo", 1, RequestOpts{})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestAddPullRequestComment(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/rest/api/1.0/projects/PROJ/repos/repo/pull-requests/1/comments", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s", r.Method)
		}
		var body AddPRCommentRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decode body: %v", err)
		}
		if body.Text != "LGTM" {
			t.Errorf("Text = %q", body.Text)
		}
		w.WriteHeader(201)
		_, _ = w.Write([]byte(`{"id":10}`))
	})
	client, ts := newTestServer(mux)
	defer ts.Close()

	err := client.AddPullRequestComment(context.Background(), "PROJ", "repo", 1, "LGTM", RequestOpts{})
	if err != nil {
		t.Fatalf("AddPullRequestComment: %v", err)
	}
}

func TestAddPullRequestComment_Error(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/rest/api/1.0/projects/PROJ/repos/repo/pull-requests/1/comments", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(403)
		_, _ = w.Write([]byte(`forbidden`))
	})
	client, ts := newTestServer(mux)
	defer ts.Close()

	err := client.AddPullRequestComment(context.Background(), "PROJ", "repo", 1, "text", RequestOpts{})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestGetPullRequestParticipants(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/rest/api/1.0/projects/PROJ/repos/repo/pull-requests/1/participants", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"values":[{"user":{"name":"alice","displayName":"Alice"},"approved":true,"status":"APPROVED"}]}`))
	})
	client, ts := newTestServer(mux)
	defer ts.Close()

	resp, err := client.GetPullRequestParticipants(context.Background(), "PROJ", "repo", 1, RequestOpts{})
	if err != nil {
		t.Fatalf("GetPullRequestParticipants: %v", err)
	}
	if len(resp.Values) != 1 {
		t.Fatalf("got %d participants", len(resp.Values))
	}
	if resp.Values[0].User.Name != "alice" {
		t.Errorf("Name = %q", resp.Values[0].User.Name)
	}
	if !resp.Values[0].Approved {
		t.Error("expected approved")
	}
}

func TestMergePullRequest_TransportError(t *testing.T) {
	mux := http.NewServeMux()
	client, ts := newTestServer(mux)
	ts.Close()

	err := client.MergePullRequest(context.Background(), "PROJ", "repo", 1, 3, RequestOpts{})
	if err == nil {
		t.Fatal("expected error for closed server")
	}
}

func TestDeclinePullRequest_TransportError(t *testing.T) {
	mux := http.NewServeMux()
	client, ts := newTestServer(mux)
	ts.Close()

	err := client.DeclinePullRequest(context.Background(), "PROJ", "repo", 1, 3, RequestOpts{})
	if err == nil {
		t.Fatal("expected error for closed server")
	}
}

func TestGetPullRequestDiff_TransportError(t *testing.T) {
	mux := http.NewServeMux()
	client, ts := newTestServer(mux)
	ts.Close()

	_, err := client.GetPullRequestDiff(context.Background(), "PROJ", "repo", 1, RequestOpts{})
	if err == nil {
		t.Fatal("expected error for closed server")
	}
}

func TestAddPullRequestComment_TransportError(t *testing.T) {
	mux := http.NewServeMux()
	client, ts := newTestServer(mux)
	ts.Close()

	err := client.AddPullRequestComment(context.Background(), "PROJ", "repo", 1, "text", RequestOpts{})
	if err == nil {
		t.Fatal("expected error for closed server")
	}
}

func TestGetPullRequestParticipants_Error(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/rest/api/1.0/projects/PROJ/repos/repo/pull-requests/1/participants", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		_, _ = w.Write([]byte(`not found`))
	})
	client, ts := newTestServer(mux)
	defer ts.Close()

	_, err := client.GetPullRequestParticipants(context.Background(), "PROJ", "repo", 1, RequestOpts{})
	if err == nil {
		t.Fatal("expected error")
	}
}
