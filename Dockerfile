# ============================================================
# Chile BCN MCP Server — multi-stage Go build
#
# Build stage pins the builder to the native platform (BUILDPLATFORM)
# so Go cross-compiles for the target arch without QEMU emulation
# (official Docker pattern for Go multi-platform builds).
# ============================================================

# Stage 1: Build Go binary
FROM --platform=$BUILDPLATFORM golang:1.26 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

# Injected automatically by buildx for each platform in the build.
ARG TARGETOS TARGETARCH

COPY . .

RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -ldflags="-s -w" -o /out/chile-bcn-mcp ./cmd/chile-bcn-mcp

# Stage 2: Runtime
FROM alpine:3.23

RUN apk add --no-cache ca-certificates curl

COPY --from=builder /out/chile-bcn-mcp /usr/local/bin/

# API resources contract (the server's default path is relative to the workdir).
# Overridable at runtime with API_RESOURCES.
COPY config/ /app/config/
WORKDIR /app

# Security: non-root user
RUN adduser -D appuser
USER appuser

ENV GOMEMLIMIT=256MiB

# Runtime defaults (can be overridden)
ENV FASTMCP_HOST=0.0.0.0
ENV FASTMCP_TRANSPORT=http

EXPOSE 8000

ENTRYPOINT ["chile-bcn-mcp"]
