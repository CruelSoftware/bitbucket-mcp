# Bitbucket MCP

Go-based MCP server for Bitbucket Server/Data Center. Enables AI systems to interact with Bitbucket repositories, pull requests, and code via the Model Context Protocol.

## Features

- **Per-request auth**: Send `Authorization: Bearer <your_bitbucket_token>` with each request
- **Custom headers**: Proxy or add headers to Bitbucket requests
- **15 tools**: PRs, repos, branches, user profile, file content, search
- **HTTP transport**: Streamable HTTP + SSE

## Quick Start

```bash
# Docker
docker run -p 3001:3001 \
  -e BITBUCKET_URL="https://your-bitbucket.example.com" \
  ghcr.io/<org>/bitbucket-mcp:latest

# Or build locally
go build -o bitbucket-mcp ./cmd/server
BITBUCKET_URL=https://your-bitbucket.example.com ./bitbucket-mcp
```

## Configuration

| Variable | Required | Description |
|----------|----------|-------------|
| `BITBUCKET_URL` | Yes | Bitbucket Server base URL |
| `MCP_HTTP_PORT` | No | Port (default 3001) |
| `MCP_HTTP_ENDPOINT` | No | Path (default /mcp) |
| `BITBUCKET_PROXY_HEADERS` | No | Comma-separated headers to forward to Bitbucket |
| `BITBUCKET_EXTRA_HEADER_<NAME>` | No | Static header (e.g. `BITBUCKET_EXTRA_HEADER_X_CUSTOM=value`) |
| `BITBUCKET_DEFAULT_PROJECT` | No | Default project key when tools omit workspaceSlug |

## Authentication

Each request must include:

```
Authorization: Bearer <your_bitbucket_personal_access_token>
```

The token is proxied to Bitbucket; no server-level token is used.

## Endpoints

- `POST /mcp` - MCP requests
- `GET /mcp` - SSE stream
- `GET /health` - Health check (no auth)

## Available Tools

- `bitbucket_list_workspaces` - List projects
- `bitbucket_list_repositories` - List repos in a project
- `bitbucket_get_repository_details` - Get repo info
- `bitbucket_search_content` - Search code (Bitbucket DC 8+)
- `bitbucket_get_file_content` - Read file contents
- `bitbucket_create_pull_request` - Create PR
- `bitbucket_get_pull_request_details` - Get PR details
- `bitbucket_get_pull_request_diff` - Get PR diff
- `bitbucket_get_pull_request_reviews` - Get PR participants
- `bitbucket_merge_pull_request` - Merge PR
- `bitbucket_decline_pull_request` - Decline PR
- `bitbucket_add_pull_request_comment` - Add PR comment
- `bitbucket_create_branch` - Create branch
- `bitbucket_list_repository_branches` - List branches
- `bitbucket_get_user_profile` - Get current user

## Docker Publishing (GHCR)

Images are built and pushed to GitHub Container Registry on push to main/master or tags `v*`.

**Image**: `ghcr.io/<owner>/<repo>` (e.g. `ghcr.io/n8n/bitbucket-mcp`)

**Setup**: No extra config. `GITHUB_TOKEN` has `packages: write`. For public packages, go to repo Settings → Packages → select the package → Change visibility to Public.

**Tags**: `latest` on main, semver on tags (e.g. `v1.0.0` → `1.0.0`, `1.0`, `1`).

## License

Apache License 2.0 — see [LICENSE](LICENSE).
