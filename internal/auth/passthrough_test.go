package auth

import (
	"context"
	"net/http"
	"testing"
)

func TestPassthroughVerifier(t *testing.T) {
	verifier := PassthroughVerifier()
	ctx := context.Background()
	req, _ := http.NewRequest("GET", "/", nil)

	// Valid token
	info, err := verifier(ctx, "my-token-123", req)
	if err != nil {
		t.Fatalf("verifier: %v", err)
	}
	if info.UserID != "my-token-123" {
		t.Errorf("UserID = %q", info.UserID)
	}
	if info.Expiration.IsZero() {
		t.Error("Expiration should be set")
	}

	// Empty token
	_, err = verifier(ctx, "", req)
	if err == nil {
		t.Error("expected error for empty token")
	}
}
