package bitbucket

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/n8n/bitbucket-mcp/internal/middleware"
)

// Client performs HTTP requests to Bitbucket Server REST API.
type Client struct {
	api    *resty.Client
	search *resty.Client
}

// NewClient creates a Bitbucket API client with retries and optional debug logging.
func NewClient(baseURL string, extraHeaders map[string]string) *Client {
	base := strings.TrimSuffix(baseURL, "/")
	apiBase := base + "/rest/api/1.0"
	searchBase := base + "/rest/search/1.0"

	api := resty.New().
		SetBaseURL(apiBase).
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept", "application/json").
		SetRetryCount(3).
		SetRetryWaitTime(500 * time.Millisecond).
		SetRetryMaxWaitTime(2 * time.Second)

	search := resty.New().
		SetBaseURL(searchBase).
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept", "application/json").
		SetRetryCount(3).
		SetRetryWaitTime(500 * time.Millisecond).
		SetRetryMaxWaitTime(2 * time.Second)

	for k, v := range extraHeaders {
		api.SetHeader(k, v)
		search.SetHeader(k, v)
	}

	if os.Getenv("BITBUCKET_DEBUG") != "" {
		logger := &debugLogger{log: log.Default()}
		api.SetDebug(true).SetLogger(logger)
		search.SetDebug(true).SetLogger(logger)
	}

	return &Client{api: api, search: search}
}

// RequestOpts holds per-request options (token, proxied headers from context).
type RequestOpts struct {
	Token   string
	Headers map[string]string
}

// do performs an HTTP request to the Bitbucket API (rest/api/1.0).
func (c *Client) do(ctx context.Context, method, apiPath string, body any, opts RequestOpts) (*resty.Response, error) {
	return c.doClient(ctx, c.api, method, apiPath, body, opts)
}

// doSearch performs an HTTP request to the Bitbucket search API (rest/search/1.0).
func (c *Client) doSearch(ctx context.Context, method, apiPath string, body any, opts RequestOpts) (*resty.Response, error) {
	return c.doClient(ctx, c.search, method, apiPath, body, opts)
}

func (c *Client) doClient(ctx context.Context, client *resty.Client, method, apiPath string, body any, opts RequestOpts) (*resty.Response, error) {
	path := strings.TrimPrefix(apiPath, "/")
	req := client.R().SetContext(ctx)
	if opts.Token != "" {
		req.SetAuthToken(opts.Token)
	}
	for k, v := range opts.Headers {
		req.SetHeader(k, v)
	}
	if body != nil {
		req.SetBody(body)
	}

	var resp *resty.Response
	var err error
	switch method {
	case "GET":
		resp, err = req.Get(path)
	case "POST":
		resp, err = req.Post(path)
	case "PUT":
		resp, err = req.Put(path)
	case "DELETE":
		resp, err = req.Delete(path)
	default:
		return nil, fmt.Errorf("unsupported method %s", method)
	}
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// apiError returns a formatted error from a resty response.
func apiError(resp *resty.Response, prefix string) error {
	if prefix != "" {
		return fmt.Errorf("%s %d: %s", prefix, resp.StatusCode(), resp.String())
	}
	return fmt.Errorf("%d: %s", resp.StatusCode(), resp.String())
}

// doJSON performs a request and unmarshals JSON into result. Returns error on 4xx/5xx.
func (c *Client) doJSON(ctx context.Context, client *resty.Client, method, path string, body, result any, opts RequestOpts) error {
	path = strings.TrimPrefix(path, "/")
	req := client.R().SetContext(ctx).SetResult(result)
	if opts.Token != "" {
		req.SetAuthToken(opts.Token)
	}
	for k, v := range opts.Headers {
		req.SetHeader(k, v)
	}
	if body != nil {
		req.SetBody(body)
	}

	var resp *resty.Response
	var err error
	switch method {
	case "GET":
		resp, err = req.Get(path)
	case "POST":
		resp, err = req.Post(path)
	case "PUT":
		resp, err = req.Put(path)
	case "DELETE":
		resp, err = req.Delete(path)
	default:
		return fmt.Errorf("unsupported method %s", method)
	}
	if err != nil {
		return err
	}
	if resp.IsError() {
		return apiError(resp, "bitbucket API error")
	}
	return nil
}

// debugLogger adapts *log.Logger to resty.Logger.
type debugLogger struct{ log *log.Logger }

func (l *debugLogger) Debugf(format string, v ...any) { l.log.Printf("[DEBUG] "+format, v...) }
func (l *debugLogger) Warnf(format string, v ...any) { l.log.Printf("[WARN] "+format, v...) }
func (l *debugLogger) Errorf(format string, v ...any) { l.log.Printf("[ERROR] "+format, v...) }

// RequestOptsFromContext builds RequestOpts from auth TokenInfo and middleware context.
func RequestOptsFromContext(ctx context.Context, token string) RequestOpts {
	opts := RequestOpts{Token: token}
	if proxy := middleware.ProxyHeadersFromContext(ctx); proxy != nil {
		opts.Headers = proxy
	}
	return opts
}