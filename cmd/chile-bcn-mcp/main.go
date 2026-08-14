// Chile BCN MCP Server.
//
// Supports stdio and streamable HTTP transports with optional
// bearer-token authentication.
package main

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/modelcontextprotocol/go-sdk/auth"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"dev.alvaros.chile-bcn-mcp/internal/bcn"
	"dev.alvaros.chile-bcn-mcp/internal/config"
	"dev.alvaros.chile-bcn-mcp/internal/server"
	"dev.alvaros.chile-bcn-mcp/internal/tools"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	// Read all configuration once at startup — never in hot paths.
	cfg := server.LoadConfig()

	logger.Info("Starting Chile BCN MCP Server",
		"transport", cfg.Transport,
	)

	// Load the API resources contract once at startup — fail fast on an
	// invalid contract. The path is FIXED: config changes are deployed by
	// rebuilding the image (config/ is baked in), never by hot-reload.
	const resourcesPath = "config/api.resources.yaml"
	resources, err := config.Load(resourcesPath)
	if err != nil {
		logger.Error("Invalid API resources contract", "path", resourcesPath, "error", err)
		os.Exit(1)
	}
	logger.Info("API resources loaded", "resources", len(resources.Resources))

	// Build the process-wide law client (the injected singleton) and
	// register all tools with it.
	lawClient := bcn.NewClient(resources, logger)

	// Create MCP server and register tools.
	srv := server.New(logger)
	tools.RegisterTools(srv, lawClient)

	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	switch cfg.Transport {
	case "stdio":
		logger.Info("Starting stdio transport")
		if err := srv.Run(ctx, &mcp.StdioTransport{}); err != nil {
			logger.Error("Server error", "error", err)
			os.Exit(1)
		}

	default: // "http", "streamable-http", "sse"
		if err := runHTTP(ctx, logger, srv, cfg); err != nil {
			logger.Error("HTTP server error", "error", err)
			os.Exit(1)
		}
	}
}

// runHTTP starts the server with streamable HTTP transport.
func runHTTP(ctx context.Context, logger *slog.Logger, srv *mcp.Server, cfg server.Config) error {
	handler := mcp.NewStreamableHTTPHandler(
		func(r *http.Request) *mcp.Server { return srv },
		&mcp.StreamableHTTPOptions{},
	)

	// Wrap with auth middleware if a token is configured.
	var mux http.Handler = handler
	if cfg.AuthToken != "" {
		mux = staticTokenAuthMiddleware(cfg.AuthToken)(handler)
		logger.Info("Authentication enabled — Bearer token required")
	}

	// Route: MCP at configured path, health at /health.
	mcpPath := cfg.Path
	if mcpPath == "" {
		mcpPath = "/mcp"
	}
	mux2 := http.NewServeMux()
	mux2.Handle(mcpPath, mux)
	mux2.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
	})

	addr := fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)
	httpSrv := &http.Server{
		Addr:         addr,
		Handler:      mux2,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	logger.Info("Listening", "addr", addr, "path", mcpPath)

	// Graceful shutdown: wait for signal, then drain.
	go func() {
		<-ctx.Done()
		logger.Info("Shutting down HTTP server...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := httpSrv.Shutdown(shutdownCtx); err != nil {
			logger.Error("HTTP shutdown error", "error", err)
		}
	}()

	if err := httpSrv.ListenAndServe(); err != http.ErrServerClosed {
		return err
	}
	return nil
}

// staticTokenAuthMiddleware wraps a handler with bearer-token auth against a
// static env-var token. The TokenInfo carries no Expiration (a static token
// is valid for the process lifetime), so AllowMissingExpiration MUST be set —
// otherwise the SDK rejects every request with "token missing expiration".
func staticTokenAuthMiddleware(token string) func(http.Handler) http.Handler {
	verifier := func(_ context.Context, tokenStr string, _ *http.Request) (*auth.TokenInfo, error) {
		if subtle.ConstantTimeCompare([]byte(tokenStr), []byte(token)) != 1 {
			return nil, auth.ErrInvalidToken
		}
		return &auth.TokenInfo{}, nil
	}
	return auth.RequireBearerToken(verifier, &auth.RequireBearerTokenOptions{
		AllowMissingExpiration: true,
	})
}
