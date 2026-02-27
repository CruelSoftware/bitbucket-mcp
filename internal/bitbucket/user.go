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
	resp, err := c.do(ctx, http.MethodGet, "/users/current", nil, opts)
	if err != nil {
		return nil, err
	}
	var out User
	if err := decodeJSON(resp, &out); err != nil {
		return nil, fmt.Errorf("get current user: %w", err)
	}
	return &out, nil
}
