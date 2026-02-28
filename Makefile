# Build for current platform (default)
build:
	go build -o build/bitbucket-mcp ./cmd/server
	go build -o build/fake-bitbucket ./cmd/fake-bitbucket

# Build for macOS (Apple Silicon)
build-darwin-arm64:
	GOOS=darwin GOARCH=arm64 go build -o build/darwin_arm64/bitbucket-mcp ./cmd/server
	GOOS=darwin GOARCH=arm64 go build -o build/darwin_arm64/fake-bitbucket ./cmd/fake-bitbucket

# Build for macOS (Intel)
build-darwin-amd64:
	GOOS=darwin GOARCH=amd64 go build -o build/darwin_amd64/bitbucket-mcp ./cmd/server
	GOOS=darwin GOARCH=amd64 go build -o build/darwin_amd64/fake-bitbucket ./cmd/fake-bitbucket

# Build for Linux
build-linux:
	GOOS=linux GOARCH=amd64 go build -o build/linux_amd64/bitbucket-mcp ./cmd/server
	GOOS=linux GOARCH=amd64 go build -o build/linux_amd64/fake-bitbucket ./cmd/fake-bitbucket

.PHONY: build build-darwin-arm64 build-darwin-amd64 build-linux
