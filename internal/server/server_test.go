package server

import (
	"os"
	"testing"
)

// unsetEnv removes the given env vars so tests get clean defaults.
func unsetEnv(t *testing.T, keys ...string) {
	t.Helper()
	for _, k := range keys {
		if err := os.Unsetenv(k); err != nil {
			t.Fatalf("Unsetenv(%q): %v", k, err)
		}
	}
}

func TestLoadConfigDefaults(t *testing.T) {
	unsetEnv(t, "FASTMCP_TRANSPORT", "FASTMCP_HOST", "FASTMCP_PORT",
		"FASTMCP_PATH", "MCP_AUTH_TOKEN")

	cfg := LoadConfig()

	if cfg.Transport != "http" {
		t.Errorf("Transport = %q, want %q", cfg.Transport, "http")
	}
	if cfg.Host != "127.0.0.1" {
		t.Errorf("Host = %q, want %q", cfg.Host, "127.0.0.1")
	}
	if cfg.Port != "8000" {
		t.Errorf("Port = %q, want %q", cfg.Port, "8000")
	}
	if cfg.Path != "" {
		t.Errorf("Path = %q, want empty", cfg.Path)
	}
	if cfg.AuthToken != "" {
		t.Errorf("AuthToken = %q, want empty", cfg.AuthToken)
	}
}

func TestLoadConfigOverrides(t *testing.T) {
	unsetEnv(t, "FASTMCP_TRANSPORT", "FASTMCP_HOST", "FASTMCP_PORT",
		"FASTMCP_PATH", "MCP_AUTH_TOKEN")
	env := map[string]string{
		"FASTMCP_TRANSPORT": "stdio",
		"FASTMCP_HOST":      "0.0.0.0",
		"FASTMCP_PORT":      "9000",
		"FASTMCP_PATH":      "/mcp-custom",
		"MCP_AUTH_TOKEN":    "secret",
	}
	for k, v := range env {
		if err := os.Setenv(k, v); err != nil {
			t.Fatalf("Setenv(%q): %v", k, err)
		}
	}
	t.Cleanup(func() {
		unsetEnv(t, "FASTMCP_TRANSPORT", "FASTMCP_HOST", "FASTMCP_PORT",
			"FASTMCP_PATH", "MCP_AUTH_TOKEN")
	})

	cfg := LoadConfig()

	if cfg.Transport != "stdio" {
		t.Errorf("Transport = %q, want %q", cfg.Transport, "stdio")
	}
	if cfg.Host != "0.0.0.0" {
		t.Errorf("Host = %q, want %q", cfg.Host, "0.0.0.0")
	}
	if cfg.Port != "9000" {
		t.Errorf("Port = %q, want %q", cfg.Port, "9000")
	}
	if cfg.Path != "/mcp-custom" {
		t.Errorf("Path = %q, want %q", cfg.Path, "/mcp-custom")
	}
	if cfg.AuthToken != "secret" {
		t.Errorf("AuthToken = %q, want %q", cfg.AuthToken, "secret")
	}
}
