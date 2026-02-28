package bitbucket

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestListBranches(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/rest/api/1.0/projects/PROJ/repos/repo/branches", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"values":[{"id":"refs/heads/main","displayId":"main","isDefault":true}],"size":1,"isLastPage":true}`))
	})
	client, ts := newTestServer(mux)
	defer ts.Close()

	resp, err := client.ListBranches(context.Background(), "PROJ", "repo", RequestOpts{})
	if err != nil {
		t.Fatalf("ListBranches: %v", err)
	}
	if len(resp.Values) != 1 {
		t.Fatalf("got %d branches", len(resp.Values))
	}
	if resp.Values[0].DisplayID != "main" {
		t.Errorf("DisplayID = %q", resp.Values[0].DisplayID)
	}
	if !resp.Values[0].IsDefault {
		t.Error("expected default branch")
	}
}

func TestListBranches_Error(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/rest/api/1.0/projects/PROJ/repos/repo/branches", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		_, _ = w.Write([]byte(`error`))
	})
	client, ts := newTestServer(mux)
	defer ts.Close()

	_, err := client.ListBranches(context.Background(), "PROJ", "repo", RequestOpts{})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestCreateBranch(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/rest/api/1.0/projects/PROJ/repos/repo/branches", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			return
		}
		var body CreateBranchRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decode: %v", err)
		}
		if body.Name != "feature/new" {
			t.Errorf("Name = %q", body.Name)
		}
		if body.StartPoint != "refs/heads/main" {
			t.Errorf("StartPoint = %q", body.StartPoint)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"refs/heads/feature/new","displayId":"feature/new"}`))
	})
	client, ts := newTestServer(mux)
	defer ts.Close()

	branch, err := client.CreateBranch(context.Background(), "PROJ", "repo", "feature/new", "main", RequestOpts{})
	if err != nil {
		t.Fatalf("CreateBranch: %v", err)
	}
	if branch.DisplayID != "feature/new" {
		t.Errorf("DisplayID = %q", branch.DisplayID)
	}
}

func TestCreateBranch_DefaultStartPoint(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/rest/api/1.0/projects/PROJ/repos/repo/branches", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			return
		}
		var body CreateBranchRequest
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body.StartPoint != "refs/heads/master" {
			t.Errorf("StartPoint = %q, want refs/heads/master", body.StartPoint)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"refs/heads/feat","displayId":"feat"}`))
	})
	client, ts := newTestServer(mux)
	defer ts.Close()

	_, err := client.CreateBranch(context.Background(), "PROJ", "repo", "feat", "", RequestOpts{})
	if err != nil {
		t.Fatalf("CreateBranch: %v", err)
	}
}

func TestCreateBranch_WithRefsPrefix(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/rest/api/1.0/projects/PROJ/repos/repo/branches", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			return
		}
		var body CreateBranchRequest
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body.StartPoint != "refs/heads/develop" {
			t.Errorf("StartPoint = %q (should not double-prefix)", body.StartPoint)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"refs/heads/feat","displayId":"feat"}`))
	})
	client, ts := newTestServer(mux)
	defer ts.Close()

	_, err := client.CreateBranch(context.Background(), "PROJ", "repo", "feat", "refs/heads/develop", RequestOpts{})
	if err != nil {
		t.Fatalf("CreateBranch: %v", err)
	}
}

func TestCreateBranch_Error(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/rest/api/1.0/projects/PROJ/repos/repo/branches", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(409)
		_, _ = w.Write([]byte(`already exists`))
	})
	client, ts := newTestServer(mux)
	defer ts.Close()

	_, err := client.CreateBranch(context.Background(), "PROJ", "repo", "feat", "main", RequestOpts{})
	if err == nil {
		t.Fatal("expected error")
	}
}
