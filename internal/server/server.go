// Package server provides the chile-bcn-mcp server setup and configuration.
package server

import (
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/alvarosdev/chile-bcn-mcp/internal/version"
)

// Config holds all runtime configuration read from environment variables.
type Config struct {
	Transport string
	Host      string
	Port      string
	Path      string
	AuthToken string
}

// LoadConfig reads configuration from environment variables.
// Call once at startup — never in hot paths.
func LoadConfig() Config {
	cfg := Config{
		Transport: envOrDefault("MCP_TRANSPORT", "http"),
		Host:      envOrDefault("MCP_HOST", "127.0.0.1"),
		Port:      envOrDefault("MCP_PORT", "8000"),
		Path:      os.Getenv("MCP_PATH"),
		AuthToken: os.Getenv("MCP_AUTH_TOKEN"),
	}
	return cfg
}

func envOrDefault(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

// New creates a new MCP server. Tools must be registered by the caller
// before starting the server.
func New(logger *slog.Logger) *mcp.Server {
	opts := &mcp.ServerOptions{
		Instructions: "Use the chile-bcn-mcp tools to help with your tasks.",
		Logger:       logger,
		// KeepAlive detects abandoned sessions and cleans up goroutines
		// (mitigates SDK goroutine leak in streamable HTTP, issue #499).
		KeepAlive:                 5 * time.Minute,
		KeepAliveFailureThreshold: 3,
	}

	return mcp.NewServer(
		&mcp.Implementation{
			Name:    "chile-bcn-mcp-server",
			Title:   "Chile BCN MCP Server",
			Version: strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(version.Version), "v")),
		},
		opts,
	)
}
