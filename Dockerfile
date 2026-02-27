FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /bitbucket-mcp ./cmd/server

FROM alpine:3.19
RUN apk --no-cache add ca-certificates
COPY --from=builder /bitbucket-mcp /bitbucket-mcp
EXPOSE 3001
ENTRYPOINT ["/bitbucket-mcp"]
