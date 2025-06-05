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

	"github.com/mark3labs/mcp-go/server"
	"github.com/tuannvm/haproxy-mcp-server/internal/config"
	"github.com/tuannvm/haproxy-mcp-server/internal/haproxy"
	"github.com/tuannvm/haproxy-mcp-server/internal/mcp"
)

func main() {
	// Configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		slog.Error("Failed to load configuration", "error", err)
		os.Exit(1)
	}

	// Logging
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

	var handler slog.Handler
	if cfg.MCPTransport == "stdio" || os.Getenv("PRETTY_LOG") == "true" {
		handler = slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: logLevel})
	} else {
		handler = slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: logLevel})
	}
	slog.SetDefault(slog.New(handler))

	slog.Info("Starting HAProxy MCP Server...")
	slog.Info("Loaded configuration", "config", cfg)

	var runtimeAPIURL string
	if cfg.HAProxyRuntimeURL != "" {
		runtimeAPIURL = cfg.HAProxyRuntimeURL
	} else {
		switch cfg.HAProxyRuntimeMode {
		case "unix":
			if cfg.HAProxyRuntimeSocket == "" {
				slog.Error("HAProxy Runtime socket path is empty. Please set HAPROXY_RUNTIME_SOCKET env variable.")
				os.Exit(1)
			}
			u := &url.URL{
				Scheme: "unix",
				Path:   cfg.HAProxyRuntimeSocket,
			}
			runtimeAPIURL = u.String()
		case "tcp4":
			if cfg.HAProxyHost == "" {
				slog.Error("HAProxy host is empty. Please set HAPROXY_HOST env variable.")
				os.Exit(1)
			}
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

	var statsURL string
	if cfg.HAProxyStatsEnabled && cfg.HAProxyStatsURL != "" {
		statsURL = cfg.HAProxyStatsURL
		slog.Info("HAProxy Stats API enabled", "url", statsURL)
	} else {
		slog.Info("HAProxy Stats API disabled")
	}

	if runtimeAPIURL == "" && statsURL == "" {
		slog.Error("Neither HAProxy Runtime API nor Stats API is configured")
		os.Exit(1)
	}

	slog.Info("Connecting to HAProxy", "runtimeAPIURL", runtimeAPIURL, "statsURL", statsURL)

	haproxyClient, err := haproxy.NewHAProxyClient(runtimeAPIURL, statsURL)
	if err != nil {
		slog.Error("Failed to initialize HAProxy client", "error", err)
		os.Exit(1)
	}

	mcpServer := server.NewMCPServer("haproxy-mcp-server", "0.1.0")
	mcp.RegisterTools(mcpServer, haproxyClient)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	switch cfg.MCPTransport {
	case "stdio":
		slog.Info("Running MCP server in stdio mode")
		if err := server.ServeStdio(mcpServer); err != nil {
			slog.Error("MCP server exited with error (stdio)", "error", err)
			os.Exit(1)
		}
		slog.Info("MCP server (stdio) finished gracefully")
	case "http":
		addr := fmt.Sprintf(":%d", cfg.MCPPort)
		sseServer := server.NewSSEServer(mcpServer)
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
		<-ctx.Done()
		slog.Info("Shutdown signal received, stopping HTTP server...")
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
