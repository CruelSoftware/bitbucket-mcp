# Bitbucket MCP Server

Go-based [Model Context Protocol](https://modelcontextprotocol.io/) (MCP) server for **Bitbucket Server / Data Center**. Enables AI assistants (Cursor, Claude Desktop, etc.) to interact with Bitbucket repositories, pull requests, branches, and code search.

## Features

- **Per-request auth** — each request passes a Bitbucket personal access token via `Authorization: Bearer`
- **15 tools** — PRs, repos, branches, user profile, file content, code search
- **HTTP transport** — Streamable HTTP + SSE (no stdio required)
- **Header proxying** — forward or inject custom headers to Bitbucket
- **Graceful shutdown** — handles SIGINT/SIGTERM cleanly

---

## Setup Guide

### Prerequisites

- **Bitbucket Server / Data Center** instance (v7+; code search requires v8+)
- **Personal Access Token** from Bitbucket with repository read/write permissions
- **Docker** (recommended) or **Go 1.22+** for building from source

### 1. Run with Docker (recommended)

```bash
docker run -d \
  --name bitbucket-mcp \
  -p 3001:3001 \
  -e BITBUCKET_URL="https://your-bitbucket.example.com" \
  -e BITBUCKET_DEFAULT_PROJECT="MYPROJ" \
  ghcr.io/cruelsoftware/bitbucket-mcp:latest
```

Verify it's running:

```bash
curl http://localhost:3001/health
# {"status":"healthy"}
```

### 2. Build from Source

```bash
git clone https://github.com/cruelsoftware/bitbucket-mcp.git
cd bitbucket-mcp
go build -o bitbucket-mcp ./cmd/server

BITBUCKET_URL=https://your-bitbucket.example.com ./bitbucket-mcp
```

### 3. Configure Your MCP Client

#### Cursor IDE

Add to your Cursor MCP settings (`.cursor/mcp.json` or global settings):

```json
{
  "mcpServers": {
    "bitbucket": {
      "url": "http://localhost:3001/mcp",
      "headers": {
        "Authorization": "Bearer <YOUR_BITBUCKET_PERSONAL_ACCESS_TOKEN>"
      }
    }
  }
}
```

#### Claude Desktop

Add to `claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "bitbucket": {
      "url": "http://localhost:3001/mcp",
      "headers": {
        "Authorization": "Bearer <YOUR_BITBUCKET_PERSONAL_ACCESS_TOKEN>"
      }
    }
  }
}
```

#### n8n

Use the **MCP Client** node with:
- **SSE URL**: `http://bitbucket-mcp:3001/mcp` (if running in same Docker network)
- **Headers**: `Authorization: Bearer <token>`

#### Generic MCP Client

Any MCP client that supports Streamable HTTP transport can connect:

```
POST http://localhost:3001/mcp   → MCP JSON-RPC requests
GET  http://localhost:3001/mcp   → SSE event stream
GET  http://localhost:3001/health → Health check (no auth)
```

All MCP requests must include the `Authorization: Bearer <token>` header. The token is forwarded to Bitbucket — no server-level credential is stored.

---

## Configuration

All configuration is via environment variables:

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `BITBUCKET_URL` | **Yes** | — | Bitbucket Server base URL (e.g. `https://bitbucket.example.com`) |
| `MCP_HTTP_PORT` | No | `3001` | Server listen port |
| `MCP_HTTP_ENDPOINT` | No | `/mcp` | MCP endpoint path |
| `BITBUCKET_DEFAULT_PROJECT` | No | — | Default project key when tools omit `workspaceSlug` |
| `BITBUCKET_PROXY_HEADERS` | No | — | Comma-separated headers to forward from MCP client to Bitbucket (e.g. `X-Request-Id,X-Trace-Id`) |
| `BITBUCKET_EXTRA_HEADER_<NAME>` | No | — | Static header injected into every Bitbucket request. Underscores become hyphens (e.g. `BITBUCKET_EXTRA_HEADER_X_CUSTOM=value` → `X-CUSTOM: value`) |
| `BITBUCKET_LOG_LEVEL` | No | `info` | Log level: `info`, `debug`, or `off` |

### Docker Compose Example

```yaml
services:
  bitbucket-mcp:
    image: ghcr.io/cruelsoftware/bitbucket-mcp:latest
    ports:
      - "3001:3001"
    environment:
      BITBUCKET_URL: "https://bitbucket.example.com"
      BITBUCKET_DEFAULT_PROJECT: "MYPROJ"
      BITBUCKET_LOG_LEVEL: "info"
    restart: unless-stopped
```

---

## Available Tools

### Workspaces & User
| Tool | Description |
|------|-------------|
| `bitbucket_list_workspaces` | List all projects the user can access |
| `bitbucket_get_user_profile` | Get the authenticated user's profile |

### Repositories
| Tool | Description |
|------|-------------|
| `bitbucket_list_repositories` | List repositories in a project |
| `bitbucket_get_repository_details` | Get repository metadata |
| `bitbucket_search_content` | Search code across repos (Bitbucket DC 8+) |
| `bitbucket_get_file_content` | Read file contents at a given ref |

### Pull Requests
| Tool | Description |
|------|-------------|
| `bitbucket_create_pull_request` | Create a new pull request |
| `bitbucket_get_pull_request_details` | Get PR details and metadata |
| `bitbucket_get_pull_request_diff` | Get the raw diff for a PR |
| `bitbucket_get_pull_request_reviews` | Get PR participants/reviewers |
| `bitbucket_merge_pull_request` | Merge a pull request |
| `bitbucket_decline_pull_request` | Decline a pull request |
| `bitbucket_add_pull_request_comment` | Add a comment to a PR |

### Branches
| Tool | Description |
|------|-------------|
| `bitbucket_create_branch` | Create a new branch |
| `bitbucket_list_repository_branches` | List branches in a repository |

---

## Authentication

This server uses **passthrough authentication**. Each MCP request must include:

```
Authorization: Bearer <your_bitbucket_personal_access_token>
```

The token is forwarded directly to Bitbucket's REST API. No server-side credentials are stored or configured.

### Creating a Bitbucket Personal Access Token

1. Go to your Bitbucket instance → **Profile** → **Manage account** → **Personal access tokens**
2. Click **Create a token**
3. Grant permissions: **Repository Read** (minimum), **Repository Write** (for creating PRs, branches, merges)
4. Copy the token and use it in your MCP client config

---

## Development

```bash
# Run tests
go test -race ./...

# Run tests with coverage
go test -race -coverprofile=coverage.out -covermode=atomic ./...
go tool cover -func=coverage.out

# Run linter
golangci-lint run

# Build
go build -o bitbucket-mcp ./cmd/server
```

CI enforces a **90% minimum test coverage** threshold.

---

## Docker Publishing (GHCR)

Images are built and pushed to GitHub Container Registry on push to `main`/`master` or tags `v*`.

- **Image**: `ghcr.io/cruelsoftware/bitbucket-mcp:latest`
- **Tags**: `latest` on main, semver on tags (e.g. `v1.0.0` → `1.0.0`, `1.0`)
- **Setup**: No extra config needed. `GITHUB_TOKEN` has `packages: write` by default.

---

## License

Apache License 2.0 — see [LICENSE](LICENSE).
