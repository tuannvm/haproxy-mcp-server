package config

import (
	"log/slog"
	"strings"

	"github.com/spf13/viper"
)

// Config holds the application configuration.
type Config struct {
	// HAProxy Settings
	HAProxyHost          string `mapstructure:"HAPROXY_HOST"`
	HAProxyPort          int    `mapstructure:"HAPROXY_PORT"`
	HAProxyRuntimeMode   string `mapstructure:"HAPROXY_RUNTIME_MODE"`   // "tcp4" or "unix"
	HAProxyRuntimeSocket string `mapstructure:"HAPROXY_RUNTIME_SOCKET"` // Used only when HAProxyRuntimeMode is "unix"

	// MCP Server Settings
	MCPTransport string `mapstructure:"MCP_TRANSPORT"`
	MCPPort      int    `mapstructure:"MCP_PORT"`

	// Logging Settings
	LogLevel string `mapstructure:"LOG_LEVEL"`
}

// LoadConfig reads configuration from environment variables and sets defaults.
func LoadConfig() (*Config, error) {
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Set Defaults
	viper.SetDefault("HAPROXY_HOST", "127.0.0.1")                             // Default to localhost
	viper.SetDefault("HAPROXY_PORT", 9999)                                    // Default HAProxy stats port with TCP socket enabled
	viper.SetDefault("HAPROXY_RUNTIME_MODE", "tcp4")                          // Default to TCP4 connections
	viper.SetDefault("HAPROXY_RUNTIME_SOCKET", "/var/run/haproxy/admin.sock") // Only used in unix mode
	viper.SetDefault("MCP_TRANSPORT", "stdio")                                // Default to stdio
	viper.SetDefault("MCP_PORT", 8080)                                        // Default port for http transport
	viper.SetDefault("LOG_LEVEL", "info")

	var config Config
	err := viper.Unmarshal(&config)
	if err != nil {
		slog.Error("Failed to unmarshal configuration", "error", err)
		return nil, err
	}

	slog.Info("Configuration loaded", "config", config) // Be careful logging sensitive defaults
	return &config, nil
}
