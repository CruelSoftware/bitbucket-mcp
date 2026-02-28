#!/bin/bash
# Smoke test: fake Bitbucket + MCP server + basic MCP requests
set -e

cd "$(dirname "$0")/.."
ROOT=$(pwd)

# Build for current platform (fixes "no suitable architecture" on macOS)
BUILD_DIR="${ROOT}/build/$(go env GOOS)_$(go env GOARCH)"
mkdir -p "$BUILD_DIR"
FAKE_BB="${BUILD_DIR}/fake-bitbucket"
MCP_BIN="${BUILD_DIR}/bitbucket-mcp"

echo "Building for $(go env GOOS)/$(go env GOARCH)..."
go build -o "$FAKE_BB" ./cmd/fake-bitbucket
go build -o "$MCP_BIN" ./cmd/server

# Start fake Bitbucket
echo "Starting fake Bitbucket on :7990..."
"$FAKE_BB" -port 7990 &
FAKE_PID=$!
trap "kill $FAKE_PID 2>/dev/null || true" EXIT

sleep 1
curl -s -o /dev/null -w "%{http_code}" http://localhost:7990/rest/api/1.0/projects | grep -q 200 || { echo "Fake Bitbucket not ready"; exit 1; }
echo "Fake Bitbucket ready"

# Start MCP server
echo "Starting MCP server on :3001..."
BITBUCKET_URL=http://localhost:7990 MCP_HTTP_PORT=3001 BITBUCKET_LOG_LEVEL=info "$MCP_BIN" &
MCP_PID=$!
trap "kill $FAKE_PID $MCP_PID 2>/dev/null || true" EXIT

sleep 1
curl -s http://localhost:3001/health | grep -q healthy || { echo "MCP server not ready"; exit 1; }
curl -s http://localhost:3001/health/ | grep -q healthy || { echo "MCP server /health/ not ready"; exit 1; }
echo "MCP server ready"

# MCP requests (Bearer token required; use fake token for fake Bitbucket)
TOKEN="fake-token-for-testing"
MCP_URL="http://localhost:3001/mcp"
ACCEPT="Accept: application/json, text/event-stream"
AUTH="Authorization: Bearer $TOKEN"
CT="Content-Type: application/json"
MCP_VER="MCP-Protocol-Version: 2025-03-26"

echo ""
echo "1. Health check"
curl -s http://localhost:3001/health
echo ""

echo ""
echo "2. MCP initialize (get session ID)"
INIT_RESP=$(curl -s -D - -X POST "$MCP_URL" -H "$ACCEPT" -H "$CT" -H "$AUTH" -H "$MCP_VER" \
  -d '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-03-26","capabilities":{},"clientInfo":{"name":"smoke-test","version":"1.0"}}}')
SESSION_ID=$(echo "$INIT_RESP" | grep -i "Mcp-Session-Id:" | head -1 | sed 's/.*: *//' | tr -d '\r')
echo "Session ID: $SESSION_ID"
echo "$INIT_RESP" | grep "data:" | head -1 | sed 's/^data: //' | head -c 400
echo ""

if [ -z "$SESSION_ID" ]; then
  echo "ERROR: No session ID from initialize"
  exit 1
fi

echo ""
echo "3. MCP initialized (complete handshake)"
curl -s -X POST "$MCP_URL" -H "$ACCEPT" -H "$CT" -H "$AUTH" -H "$MCP_VER" -H "Mcp-Session-Id: $SESSION_ID" \
  -d '{"jsonrpc":"2.0","method":"notifications/initialized"}' | head -c 200
echo ""

echo ""
echo "4. MCP tools/list"
curl -s -X POST "$MCP_URL" -H "$ACCEPT" -H "$CT" -H "$AUTH" -H "$MCP_VER" -H "Mcp-Session-Id: $SESSION_ID" \
  -d '{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}' | head -c 800
echo ""

echo ""
echo "5. MCP tools/call bitbucket_list_workspaces"
WS_RESP=$(curl -s -X POST "$MCP_URL" -H "$ACCEPT" -H "$CT" -H "$AUTH" -H "$MCP_VER" -H "Mcp-Session-Id: $SESSION_ID" \
  -d '{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"bitbucket_list_workspaces","arguments":{}}}')
echo "$WS_RESP" | grep -q 'FAKE' || { echo "ERROR: expected FAKE workspace in: $WS_RESP"; exit 1; }
echo "  -> workspaces include FAKE"

echo ""
echo "6. MCP tools/call bitbucket_get_user_profile"
USER_RESP=$(curl -s -X POST "$MCP_URL" -H "$ACCEPT" -H "$CT" -H "$AUTH" -H "$MCP_VER" -H "Mcp-Session-Id: $SESSION_ID" \
  -d '{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"bitbucket_get_user_profile","arguments":{}}}')
echo "$USER_RESP" | grep -q 'Fake User' || { echo "ERROR: expected Fake User"; exit 1; }
echo "  -> user: Fake User"

echo ""
echo "Smoke test passed."
