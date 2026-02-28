package bitbucket

import (
	"bytes"
	"context"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/n8n/bitbucket-mcp/internal/middleware"
)

func newTestServer(handler http.Handler) (*Client, *httptest.Server) {
	ts := httptest.NewServer(handler)
	client := NewClient(ts.URL, nil, "off")
	return client, ts
}

func TestNewClient(t *testing.T) {
	c := NewClient("https://bb.example.com", map[string]string{"X-Custom": "val"}, "off")
	if c.api == nil || c.search == nil {
		t.Fatal("clients should not be nil")
	}
}

func TestNewClient_DebugMode(t *testing.T) {
	c := NewClient("https://bb.example.com", nil, "debug")
	if c.api == nil {
		t.Fatal("api client should not be nil")
	}
}

func TestNewClient_TrailingSlash(t *testing.T) {
	c := NewClient("https://bb.example.com/", nil, "off")
	if c.api == nil {
		t.Fatal("api client should not be nil")
	}
}

func TestDoClient_UnsupportedMethod(t *testing.T) {
	mux := http.NewServeMux()
	client, ts := newTestServer(mux)
	defer ts.Close()

	_, err := client.doClient(context.Background(), client.api, "PATCH", "/test", nil, RequestOpts{})
	if err == nil {
		t.Fatal("expected error for unsupported method")
	}
}

func TestDoClient_AllMethods(t *testing.T) {
	methods := []string{"GET", "POST", "PUT", "DELETE"}
	for _, m := range methods {
		t.Run(m, func(t *testing.T) {
			mux := http.NewServeMux()
			mux.HandleFunc("/rest/api/1.0/test", func(w http.ResponseWriter, r *http.Request) {
				if r.Method != m {
					t.Errorf("method = %s, want %s", r.Method, m)
				}
				w.WriteHeader(200)
			})
			client, ts := newTestServer(mux)
			defer ts.Close()

			resp, err := client.doClient(context.Background(), client.api, m, "/test", nil, RequestOpts{Token: "tok"})
			if err != nil {
				t.Fatalf("doClient: %v", err)
			}
			if resp.StatusCode() != 200 {
				t.Errorf("status = %d", resp.StatusCode())
			}
		})
	}
}

func TestDoClient_WithHeaders(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/rest/api/1.0/test", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Proxy") != "val" {
			t.Errorf("missing proxy header")
		}
		if r.Header.Get("Authorization") == "" {
			t.Errorf("missing auth header")
		}
		w.WriteHeader(200)
	})
	client, ts := newTestServer(mux)
	defer ts.Close()

	opts := RequestOpts{Token: "tok", Headers: map[string]string{"X-Proxy": "val"}}
	_, err := client.doClient(context.Background(), client.api, "GET", "/test", nil, opts)
	if err != nil {
		t.Fatalf("doClient: %v", err)
	}
}

func TestDoClient_WithBody(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/rest/api/1.0/test", func(w http.ResponseWriter, r *http.Request) {
		if r.ContentLength == 0 {
			t.Errorf("expected body")
		}
		w.WriteHeader(200)
	})
	client, ts := newTestServer(mux)
	defer ts.Close()

	_, err := client.doClient(context.Background(), client.api, "POST", "/test", map[string]string{"key": "val"}, RequestOpts{})
	if err != nil {
		t.Fatalf("doClient: %v", err)
	}
}

func TestDoJSON_Success(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/rest/api/1.0/test", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"key":"PROJ","name":"Project"}`))
	})
	client, ts := newTestServer(mux)
	defer ts.Close()

	var out Workspace
	err := client.doJSON(context.Background(), client.api, "GET", "/test", nil, &out, RequestOpts{})
	if err != nil {
		t.Fatalf("doJSON: %v", err)
	}
	if out.Key != "PROJ" {
		t.Errorf("Key = %q", out.Key)
	}
}

func TestDoJSON_ErrorResponse(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/rest/api/1.0/test", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		_, _ = w.Write([]byte(`{"errors":[{"message":"not found"}]}`))
	})
	client, ts := newTestServer(mux)
	defer ts.Close()

	var out Workspace
	err := client.doJSON(context.Background(), client.api, "GET", "/test", nil, &out, RequestOpts{})
	if err == nil {
		t.Fatal("expected error for 404")
	}
}

func TestDoJSON_UnsupportedMethod(t *testing.T) {
	mux := http.NewServeMux()
	client, ts := newTestServer(mux)
	defer ts.Close()

	var out Workspace
	err := client.doJSON(context.Background(), client.api, "PATCH", "/test", nil, &out, RequestOpts{})
	if err == nil {
		t.Fatal("expected error for unsupported method")
	}
}

func TestDoJSON_AllMethods(t *testing.T) {
	methods := []string{"GET", "POST", "PUT", "DELETE"}
	for _, m := range methods {
		t.Run(m, func(t *testing.T) {
			mux := http.NewServeMux()
			mux.HandleFunc("/rest/api/1.0/test", func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"key":"OK"}`))
			})
			client, ts := newTestServer(mux)
			defer ts.Close()

			var out Workspace
			err := client.doJSON(context.Background(), client.api, m, "/test", nil, &out, RequestOpts{})
			if err != nil {
				t.Fatalf("doJSON %s: %v", m, err)
			}
		})
	}
}

func TestDoJSON_WithBodyAndAuth(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/rest/api/1.0/test", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") == "" {
			t.Errorf("missing auth")
		}
		if r.Header.Get("X-Custom") != "hdr" {
			t.Errorf("missing custom header")
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{}`))
	})
	client, ts := newTestServer(mux)
	defer ts.Close()

	var out struct{}
	opts := RequestOpts{Token: "tok", Headers: map[string]string{"X-Custom": "hdr"}}
	err := client.doJSON(context.Background(), client.api, "POST", "/test", map[string]string{"a": "b"}, &out, opts)
	if err != nil {
		t.Fatalf("doJSON: %v", err)
	}
}

func TestApiError_WithPrefix(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/rest/api/1.0/err", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		_, _ = w.Write([]byte("server error"))
	})
	client, ts := newTestServer(mux)
	defer ts.Close()

	resp, _ := client.doClient(context.Background(), client.api, "GET", "/err", nil, RequestOpts{})
	err := apiError(resp, "test prefix")
	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != "test prefix 500: server error" {
		t.Errorf("error = %q", err.Error())
	}
}

func TestApiError_WithoutPrefix(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/rest/api/1.0/err", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(400)
		_, _ = w.Write([]byte("bad request"))
	})
	client, ts := newTestServer(mux)
	defer ts.Close()

	resp, _ := client.doClient(context.Background(), client.api, "GET", "/err", nil, RequestOpts{})
	err := apiError(resp, "")
	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != "400: bad request" {
		t.Errorf("error = %q", err.Error())
	}
}

func TestRequestOptsFromContext_WithToken(t *testing.T) {
	ctx := context.Background()
	opts := RequestOptsFromContext(ctx, "my-token")
	if opts.Token != "my-token" {
		t.Errorf("Token = %q", opts.Token)
	}
}

func TestRequestOptsFromContext_NoToken(t *testing.T) {
	ctx := context.Background()
	opts := RequestOptsFromContext(ctx, "")
	if opts.Token != "" {
		t.Errorf("Token = %q", opts.Token)
	}
}

func TestRequestOptsFromContext_WithProxyHeaders(t *testing.T) {
	ctx := middleware.WithProxyHeaders(context.Background(), map[string]string{"X-Request-Id": "123"})
	opts := RequestOptsFromContext(ctx, "tok")
	if opts.Token != "tok" {
		t.Errorf("Token = %q", opts.Token)
	}
	if opts.Headers["X-Request-Id"] != "123" {
		t.Errorf("Headers = %v", opts.Headers)
	}
}

func TestDebugLogger(t *testing.T) {
	var buf bytes.Buffer
	l := &debugLogger{log: log.New(&buf, "", 0)}
	l.Debugf("debug %s", "msg")
	l.Warnf("warn %s", "msg")
	l.Errorf("error %s", "msg")

	out := buf.String()
	if !bytes.Contains([]byte(out), []byte("[DEBUG] debug msg")) {
		t.Errorf("missing debug output in: %s", out)
	}
	if !bytes.Contains([]byte(out), []byte("[WARN] warn msg")) {
		t.Errorf("missing warn output in: %s", out)
	}
	if !bytes.Contains([]byte(out), []byte("[ERROR] error msg")) {
		t.Errorf("missing error output in: %s", out)
	}
}

func TestDo(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/rest/api/1.0/test", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	})
	client, ts := newTestServer(mux)
	defer ts.Close()

	resp, err := client.do(context.Background(), "GET", "/test", nil, RequestOpts{})
	if err != nil {
		t.Fatalf("do: %v", err)
	}
	if resp.StatusCode() != 200 {
		t.Errorf("status = %d", resp.StatusCode())
	}
}

func TestDoSearch(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/rest/search/1.0/test", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	})
	client, ts := newTestServer(mux)
	defer ts.Close()

	resp, err := client.doSearch(context.Background(), "GET", "/test", nil, RequestOpts{})
	if err != nil {
		t.Fatalf("doSearch: %v", err)
	}
	if resp.StatusCode() != 200 {
		t.Errorf("status = %d", resp.StatusCode())
	}
}
