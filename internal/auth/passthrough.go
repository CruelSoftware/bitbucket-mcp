package auth

import (
	"context"
	"fmt"
	"net/http"
	"time"

	sdkauth "github.com/modelcontextprotocol/go-sdk/auth"
)

// PassthroughVerifier returns a TokenVerifier that accepts any non-empty token
// and returns TokenInfo with UserID set to the raw token (for proxying to Bitbucket).
// Expiration is set to 24h from now to satisfy auth.RequireBearerToken checks.
func PassthroughVerifier() sdkauth.TokenVerifier {
	return func(_ context.Context, token string, _ *http.Request) (*sdkauth.TokenInfo, error) {
		if token == "" {
			return nil, fmt.Errorf("%w: empty token", sdkauth.ErrInvalidToken)
		}
		return &sdkauth.TokenInfo{
			UserID:     token,
			Expiration: time.Now().Add(24 * time.Hour),
		}, nil
	}
}
