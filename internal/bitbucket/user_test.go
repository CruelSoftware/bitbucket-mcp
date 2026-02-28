package bitbucket

import (
	"context"
	"net/http"
	"testing"
)

func TestGetCurrentUser(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/rest/api/1.0/users/current", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"name":"jdoe","emailAddress":"jdoe@example.com","displayName":"John Doe","id":1,"active":true}`))
	})
	client, ts := newTestServer(mux)
	defer ts.Close()

	user, err := client.GetCurrentUser(context.Background(), RequestOpts{})
	if err != nil {
		t.Fatalf("GetCurrentUser: %v", err)
	}
	if user.Name != "jdoe" {
		t.Errorf("Name = %q", user.Name)
	}
	if user.DisplayName != "John Doe" {
		t.Errorf("DisplayName = %q", user.DisplayName)
	}
	if user.EmailAddress != "jdoe@example.com" {
		t.Errorf("EmailAddress = %q", user.EmailAddress)
	}
	if !user.Active {
		t.Error("expected active user")
	}
}

func TestGetCurrentUser_Error(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/rest/api/1.0/users/current", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(401)
		_, _ = w.Write([]byte(`unauthorized`))
	})
	client, ts := newTestServer(mux)
	defer ts.Close()

	_, err := client.GetCurrentUser(context.Background(), RequestOpts{})
	if err == nil {
		t.Fatal("expected error")
	}
}
