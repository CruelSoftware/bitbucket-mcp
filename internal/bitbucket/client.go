package bitbucket

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/n8n/bitbucket-mcp/internal/middleware"
)

// Client performs HTTP requests to Bitbucket Server REST API.
type Client struct {
	baseURL     string
	httpClient  *http.Client
	extraHeaders map[string]string
}

// NewClient creates a Bitbucket API client.
func NewClient(baseURL string, extraHeaders map[string]string) *Client {
	return &Client{
		baseURL:      strings.TrimSuffix(baseURL, "/"),
		httpClient:   &http.Client{},
		extraHeaders: extraHeaders,
	}
}

// RequestOpts holds per-request options (token, proxied headers from context).
type RequestOpts struct {
	Token   string
	Headers map[string]string
}

// do performs an HTTP request to the Bitbucket API (rest/api/1.0).
func (c *Client) do(ctx context.Context, method, apiPath string, body io.Reader, opts RequestOpts) (*http.Response, error) {
	return c.doBase(ctx, method, "rest/api/1.0", apiPath, body, opts)
}

// doSearch performs an HTTP request to the Bitbucket search API (rest/search/1.0).
func (c *Client) doSearch(ctx context.Context, method, apiPath string, body io.Reader, opts RequestOpts) (*http.Response, error) {
	return c.doBase(ctx, method, "rest/search/1.0", apiPath, body, opts)
}

func (c *Client) doBase(ctx context.Context, method, apiBase, apiPath string, body io.Reader, opts RequestOpts) (*http.Response, error) {
	u, err := url.Parse(c.baseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid base URL: %w", err)
	}
	u.Path = path.Join(u.Path, apiBase, strings.TrimPrefix(apiPath, "/"))
	req, err := http.NewRequestWithContext(ctx, method, u.String(), body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	if opts.Token != "" {
		req.Header.Set("Authorization", "Bearer "+opts.Token)
	}
	for k, v := range c.extraHeaders {
		req.Header.Set(k, v)
	}
	for k, v := range opts.Headers {
		req.Header.Set(k, v)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// RequestOptsFromContext builds RequestOpts from auth TokenInfo and middleware context.
func RequestOptsFromContext(ctx context.Context, token string) RequestOpts {
	opts := RequestOpts{Token: token}
	if proxy := middleware.ProxyHeadersFromContext(ctx); proxy != nil {
		opts.Headers = proxy
	}
	return opts
}

// decodeJSON decodes response body into v, closing the body.
func decodeJSON(resp *http.Response, v any) error {
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("bitbucket API error %d: %s", resp.StatusCode, string(body))
	}
	return json.NewDecoder(resp.Body).Decode(v)
}
