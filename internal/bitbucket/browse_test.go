package bitbucket

import (
	"context"
	"net/http"
	"testing"
)

func TestGetFileContent(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/rest/api/1.0/projects/PROJ/repos/repo/browse/src%2Fmain.go", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"lines":[{"text":"package main"},{"text":""},{"text":"func main() {}"}]}`))
	})
	client, ts := newTestServer(mux)
	defer ts.Close()

	content, err := client.GetFileContent(context.Background(), "PROJ", "repo", "src/main.go", "", RequestOpts{})
	if err != nil {
		t.Fatalf("GetFileContent: %v", err)
	}
	expected := "package main\n\nfunc main() {}"
	if content != expected {
		t.Errorf("content = %q, want %q", content, expected)
	}
}

func TestGetFileContent_WithRef(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/rest/api/1.0/projects/PROJ/repos/repo/browse/file.txt", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("at") != "refs/heads/develop" {
			t.Errorf("at = %q", r.URL.Query().Get("at"))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"lines":[{"text":"hello"}]}`))
	})
	client, ts := newTestServer(mux)
	defer ts.Close()

	content, err := client.GetFileContent(context.Background(), "PROJ", "repo", "file.txt", "refs/heads/develop", RequestOpts{})
	if err != nil {
		t.Fatalf("GetFileContent: %v", err)
	}
	if content != "hello" {
		t.Errorf("content = %q", content)
	}
}

func TestGetFileContent_EmptyFile(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/rest/api/1.0/projects/PROJ/repos/repo/browse/empty.txt", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"lines":[]}`))
	})
	client, ts := newTestServer(mux)
	defer ts.Close()

	content, err := client.GetFileContent(context.Background(), "PROJ", "repo", "empty.txt", "", RequestOpts{})
	if err != nil {
		t.Fatalf("GetFileContent: %v", err)
	}
	if content != "" {
		t.Errorf("content = %q, want empty", content)
	}
}

func TestGetFileContent_Error(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/rest/api/1.0/projects/PROJ/repos/repo/browse/missing.txt", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		_, _ = w.Write([]byte(`not found`))
	})
	client, ts := newTestServer(mux)
	defer ts.Close()

	_, err := client.GetFileContent(context.Background(), "PROJ", "repo", "missing.txt", "", RequestOpts{})
	if err == nil {
		t.Fatal("expected error")
	}
}
