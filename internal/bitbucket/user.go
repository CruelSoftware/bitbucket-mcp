package bitbucket

import (
	"context"
	"fmt"
	"net/http"
)

// User represents a Bitbucket user.
type User struct {
	Name         string `json:"name"`
	EmailAddress string `json:"emailAddress"`
	DisplayName  string `json:"displayName"`
	ID           int    `json:"id"`
	Active       bool   `json:"active"`
}

// GetCurrentUser returns the authenticated user's profile.
func (c *Client) GetCurrentUser(ctx context.Context, opts RequestOpts) (*User, error) {
	var out User
	if err := c.doJSON(ctx, c.api, http.MethodGet, "/users/current", nil, &out, opts); err != nil {
		return nil, fmt.Errorf("get current user: %w", err)
	}
	return &out, nil
}
