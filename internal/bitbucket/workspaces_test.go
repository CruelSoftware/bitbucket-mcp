package bitbucket

import (
	"context"
	"net/http"
	"testing"
)

func TestListWorkspaces(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/rest/api/1.0/projects", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"values":[{"key":"PROJ","name":"Project"}],"size":1,"limit":25,"isLastPage":true,"start":0}`))
	})
	client, ts := newTestServer(mux)
	defer ts.Close()

	resp, err := client.ListWorkspaces(context.Background(), RequestOpts{})
	if err != nil {
		t.Fatalf("ListWorkspaces: %v", err)
	}
	if len(resp.Values) != 1 {
		t.Fatalf("got %d workspaces", len(resp.Values))
	}
	if resp.Values[0].Key != "PROJ" {
		t.Errorf("Key = %q", resp.Values[0].Key)
	}
	if resp.Values[0].Name != "Project" {
		t.Errorf("Name = %q", resp.Values[0].Name)
	}
}

func TestListWorkspaces_Error(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/rest/api/1.0/projects", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		_, _ = w.Write([]byte(`{"errors":[{"message":"internal"}]}`))
	})
	client, ts := newTestServer(mux)
	defer ts.Close()

	_, err := client.ListWorkspaces(context.Background(), RequestOpts{})
	if err == nil {
		t.Fatal("expected error")
	}
}
