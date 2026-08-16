# ============================================================
# Chile BCN MCP Server — multi-stage Go build
#
# Build stage pins the builder to the native platform (BUILDPLATFORM)
# so Go cross-compiles for the target arch without QEMU emulation
# (official Docker pattern for Go multi-platform builds).
# ============================================================

# Stage 1: Build Go binary
FROM --platform=$BUILDPLATFORM golang:1.26.6 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download
# Injected automatically by buildx for each platform in the build.
ARG TARGETOS TARGETARCH
ARG VERSION=dev

COPY . .

RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -trimpath -ldflags="-s -w -X github.com/alvarosdev/chile-bcn-mcp/internal/version.Version=${VERSION}" -o /out/chile-bcn-mcp ./cmd/chile-bcn-mcp

# Stage 2: Runtime
FROM gcr.io/distroless/static-debian13:nonroot
COPY --from=builder /out/chile-bcn-mcp /usr/local/bin/

# API resources and prompts are baked into the binary via go:embed
# (internal/config/api.resources.yaml and internal/prompts/prompts.yaml)
# — no external config required at runtime.

ENV GOMEMLIMIT=256MiB

ENV FASTMCP_HOST=0.0.0.0
ENV FASTMCP_TRANSPORT=http

EXPOSE 8000

ENTRYPOINT ["chile-bcn-mcp"]
