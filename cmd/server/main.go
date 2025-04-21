package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/mark3labs/mcp-go/server" // Import directly without alias

	"github.com/tuannvm/haproxy-mcp-server/internal/config"
	"github.com/tuannvm/haproxy-mcp-server/internal/haproxy"
	"github.com/tuannvm/haproxy-mcp-server/internal/mcp"
)

func main() {
	// --- Configuration ---
	cfg, err := config.LoadConfig()
	if err != nil {
		slog.Error("Failed to load configuration", "error", err)
		os.Exit(1)
	}

	// --- Logging ---
	// Configure logging level
	var logLevel slog.Level
	switch strings.ToLower(cfg.LogLevel) {
	case "debug":
		logLevel = slog.LevelDebug
	case "info":
		logLevel = slog.LevelInfo
	case "warn", "warning":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		slog.Warn("Invalid log level, defaulting to 'info'", "configured_level", cfg.LogLevel)
		logLevel = slog.LevelInfo
	}

	// Use text handler for development/stdio mode
	var handler slog.Handler
	if cfg.MCPTransport == "stdio" || os.Getenv("PRETTY_LOG") == "true" {
		handler = slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
			Level: logLevel,
		})
	} else {
		handler = slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
			Level: logLevel,
		})
	}
	slog.SetDefault(slog.New(handler))

	slog.Info("Starting HAProxy MCP Server...")
	slog.Info("Loaded configuration", "config", cfg)

	// --- HAProxy Runtime API Client ---
	var runtimeAPIURL string

	// Use direct URL if provided, otherwise construct from components
	if cfg.HAProxyRuntimeURL != "" {
		runtimeAPIURL = cfg.HAProxyRuntimeURL
	} else {
		// Handle connection based on runtime mode
		switch cfg.HAProxyRuntimeMode {
		case "unix":
			// Unix socket mode
			if cfg.HAProxyRuntimeSocket == "" {
				slog.Error("HAProxy Runtime socket path is empty. Please set HAPROXY_RUNTIME_SOCKET env variable.")
				os.Exit(1)
			}

			// Create a URL with unix socket protocol
			u := &url.URL{
				Scheme: "unix",
				Path:   cfg.HAProxyRuntimeSocket,
			}
			runtimeAPIURL = u.String()

		case "tcp4":
			// TCP4 mode
			if cfg.HAProxyHost == "" {
				slog.Error("HAProxy host is empty. Please set HAPROXY_HOST env variable.")
				os.Exit(1)
			}

			// Create a TCP URL
			u := &url.URL{
				Scheme: "tcp",
				Host:   fmt.Sprintf("%s:%d", cfg.HAProxyHost, cfg.HAProxyPort),
			}
			runtimeAPIURL = u.String()

		default:
			if cfg.HAProxyStatsEnabled && cfg.HAProxyStatsURL != "" {
				slog.Warn("Invalid HAProxy runtime mode, but stats API is enabled. Continuing with stats only.",
					"mode", cfg.HAProxyRuntimeMode)
				runtimeAPIURL = ""
			} else {
				slog.Error("Invalid HAProxy runtime mode and no stats API configured", "mode", cfg.HAProxyRuntimeMode)
				os.Exit(1)
			}
		}
	}

	// --- HAProxy Stats Client ---
	var statsURL string
	if cfg.HAProxyStatsEnabled && cfg.HAProxyStatsURL != "" {
		statsURL = cfg.HAProxyStatsURL
		slog.Info("HAProxy Stats API enabled", "url", statsURL)
	} else {
		slog.Info("HAProxy Stats API disabled")
	}

	// Ensure at least one API is configured
	if runtimeAPIURL == "" && statsURL == "" {
		slog.Error("Neither HAProxy Runtime API nor Stats API is configured")
		os.Exit(1)
	}

	slog.Info("Connecting to HAProxy", "runtimeAPIURL", runtimeAPIURL, "statsURL", statsURL)

	// Create the HAProxy client with the appropriate URLs
	haproxyClient, err := haproxy.NewHAProxyClient(runtimeAPIURL, statsURL)
	if err != nil {
		// Log fatal here as the client is essential for the server's function
		slog.Error("Failed to initialize HAProxy client", "error", err)
		os.Exit(1)
	}

	// --- MCP Server ---
	// Create MCP Server with name and version
	mcpServer := server.NewMCPServer("haproxy-mcp-server", "0.1.0")

	// --- Register Tools ---
	mcp.RegisterTools(mcpServer, haproxyClient) // Use mcp.RegisterTools instead of tools.RegisterTools

	// --- Context and Shutdown Handling ---
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// --- Transport Handling ---
	switch cfg.MCPTransport {
	case "stdio":
		slog.Info("Running MCP server in stdio mode")
		// Use ServeStdio for stdio mode
		if err := server.ServeStdio(mcpServer); err != nil {
			slog.Error("MCP server exited with error (stdio)", "error", err)
			os.Exit(1)
		}
		slog.Info("MCP server (stdio) finished gracefully")

	case "http":
		addr := fmt.Sprintf(":%d", cfg.MCPPort)

		// Create an SSE server
		sseServer := server.NewSSEServer(mcpServer)

		// Create HTTP server with SSE handler
		httpServer := &http.Server{
			Addr:    addr,
			Handler: sseServer,
		}

		go func() {
			slog.Info("Starting HTTP server for MCP SSE transport", "address", addr)
			if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				slog.Error("HTTP server failed", "error", err)
				os.Exit(1)
			}
		}()

		// Wait for shutdown signal
		<-ctx.Done()
		slog.Info("Shutdown signal received, stopping HTTP server...")

		// Graceful shutdown
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			slog.Error("HTTP server graceful shutdown failed", "error", err)
		} else {
			slog.Info("HTTP server stopped gracefully")
		}

	default:
		slog.Error("Invalid MCP_TRANSPORT specified", "transport", cfg.MCPTransport)
		os.Exit(1)
	}

	slog.Info("HAProxy MCP Server stopped.")
}
