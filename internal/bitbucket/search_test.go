package bitbucket

import (
	"context"
	"net/http"
	"testing"
)

func TestSearchContent(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/rest/search/1.0/search", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("searchQuery") != "hello" {
			t.Errorf("searchQuery = %q", q.Get("searchQuery"))
		}
		if q.Get("projectKey") != "PROJ" {
			t.Errorf("projectKey = %q", q.Get("projectKey"))
		}
		if q.Get("fileExtension") != "go" {
			t.Errorf("fileExtension = %q", q.Get("fileExtension"))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"values":[{"path":"main.go","content":"hello world"}]}`))
	})
	client, ts := newTestServer(mux)
	defer ts.Close()

	resp, err := client.SearchContent(context.Background(), "PROJ", "hello", "go", RequestOpts{})
	if err != nil {
		t.Fatalf("SearchContent: %v", err)
	}
	if len(resp.Values) != 1 {
		t.Fatalf("got %d results", len(resp.Values))
	}
	if resp.Values[0].Path != "main.go" {
		t.Errorf("Path = %q", resp.Values[0].Path)
	}
}

func TestSearchContent_NoWorkspace(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/rest/search/1.0/search", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("projectKey") != "" {
			t.Errorf("projectKey should be empty, got %q", r.URL.Query().Get("projectKey"))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"values":[]}`))
	})
	client, ts := newTestServer(mux)
	defer ts.Close()

	resp, err := client.SearchContent(context.Background(), "", "test", "", RequestOpts{})
	if err != nil {
		t.Fatalf("SearchContent: %v", err)
	}
	if len(resp.Values) != 0 {
		t.Errorf("expected 0 results, got %d", len(resp.Values))
	}
}

func TestSearchContent_NotSupported(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/rest/search/1.0/search", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
	})
	client, ts := newTestServer(mux)
	defer ts.Close()

	resp, err := client.SearchContent(context.Background(), "", "test", "", RequestOpts{})
	if err != nil {
		t.Fatalf("SearchContent 404 should not error: %v", err)
	}
	if len(resp.Values) != 0 {
		t.Errorf("expected empty results for 404")
	}
}

func TestSearchContent_501NotImplemented(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/rest/search/1.0/search", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(501)
	})
	client, ts := newTestServer(mux)
	defer ts.Close()

	resp, err := client.SearchContent(context.Background(), "", "test", "", RequestOpts{})
	if err != nil {
		t.Fatalf("SearchContent 501 should not error: %v", err)
	}
	if len(resp.Values) != 0 {
		t.Errorf("expected empty results for 501")
	}
}

func TestSearchContent_ServerError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/rest/search/1.0/search", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		_, _ = w.Write([]byte(`internal error`))
	})
	client, ts := newTestServer(mux)
	defer ts.Close()

	_, err := client.SearchContent(context.Background(), "", "test", "", RequestOpts{})
	if err == nil {
		t.Fatal("expected error for 500")
	}
}

func TestSearchContent_TransportError(t *testing.T) {
	mux := http.NewServeMux()
	client, ts := newTestServer(mux)
	ts.Close()

	_, err := client.SearchContent(context.Background(), "", "test", "", RequestOpts{})
	if err == nil {
		t.Fatal("expected error for closed server")
	}
}

func TestSearchContent_InvalidJSON(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/rest/search/1.0/search", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`not json`))
	})
	client, ts := newTestServer(mux)
	defer ts.Close()

	_, err := client.SearchContent(context.Background(), "", "test", "", RequestOpts{})
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}
