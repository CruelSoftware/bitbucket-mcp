package bitbucket

import (
	"context"
	"net/http"
	"testing"
)

func TestListRepositories(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/rest/api/1.0/projects/PROJ/repos", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"values":[{"slug":"my-repo","name":"My Repo","id":1}],"size":1,"isLastPage":true}`))
	})
	client, ts := newTestServer(mux)
	defer ts.Close()

	resp, err := client.ListRepositories(context.Background(), "PROJ", RequestOpts{})
	if err != nil {
		t.Fatalf("ListRepositories: %v", err)
	}
	if len(resp.Values) != 1 {
		t.Fatalf("got %d repos", len(resp.Values))
	}
	if resp.Values[0].Slug != "my-repo" {
		t.Errorf("Slug = %q", resp.Values[0].Slug)
	}
}

func TestListRepositories_Error(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/rest/api/1.0/projects/PROJ/repos", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(403)
		_, _ = w.Write([]byte(`{"errors":[{"message":"forbidden"}]}`))
	})
	client, ts := newTestServer(mux)
	defer ts.Close()

	_, err := client.ListRepositories(context.Background(), "PROJ", RequestOpts{})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestGetRepository(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/rest/api/1.0/projects/PROJ/repos/my-repo", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"slug":"my-repo","name":"My Repo","id":1,"project":{"key":"PROJ","name":"Project"}}`))
	})
	client, ts := newTestServer(mux)
	defer ts.Close()

	repo, err := client.GetRepository(context.Background(), "PROJ", "my-repo", RequestOpts{})
	if err != nil {
		t.Fatalf("GetRepository: %v", err)
	}
	if repo.Slug != "my-repo" {
		t.Errorf("Slug = %q", repo.Slug)
	}
	if repo.Project == nil || repo.Project.Key != "PROJ" {
		t.Errorf("Project = %+v", repo.Project)
	}
}

func TestGetRepository_Error(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/rest/api/1.0/projects/PROJ/repos/bad", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		_, _ = w.Write([]byte(`not found`))
	})
	client, ts := newTestServer(mux)
	defer ts.Close()

	_, err := client.GetRepository(context.Background(), "PROJ", "bad", RequestOpts{})
	if err == nil {
		t.Fatal("expected error")
	}
}
