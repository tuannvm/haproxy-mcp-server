package config

import (
	"strings"

	log "github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

// Config holds the application configuration.
type Config struct {
	// HAProxy Settings
	HAProxyConfigFile     string `mapstructure:"HAPROXY_CONFIG_FILE"`
	HAProxyBinaryPath     string `mapstructure:"HAPROXY_BINARY_PATH"`
	HAProxyTransactionDir string `mapstructure:"HAPROXY_TRANSACTION_DIR"`
	HAProxyRuntimeSocket  string `mapstructure:"HAPROXY_RUNTIME_SOCKET"`
	// TODO: Add Master Socket path if needed

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
	viper.SetDefault("HAPROXY_CONFIG_FILE", "/etc/haproxy/haproxy.cfg")
	viper.SetDefault("HAPROXY_BINARY_PATH", "/usr/sbin/haproxy") // Common path, adjust if needed
	viper.SetDefault("HAPROXY_TRANSACTION_DIR", "/tmp/haproxy_transactions") // Needs write access
	viper.SetDefault("HAPROXY_RUNTIME_SOCKET", "/var/run/haproxy/admin.sock")
	viper.SetDefault("MCP_TRANSPORT", "stdio") // Default to stdio
	viper.SetDefault("MCP_PORT", 8080)         // Default port for http transport
	viper.SetDefault("LOG_LEVEL", "info")

	var config Config
	err := viper.Unmarshal(&config)
	if err != nil {
		log.Error().Err(err).Msg("Failed to unmarshal configuration")
		return nil, err
	}

	log.Info().Interface("configLoaded", config).Msg("Configuration loaded") // Be careful logging sensitive defaults
	return &config, nil
}
