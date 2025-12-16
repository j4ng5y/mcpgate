package config

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
)

// Config represents the gateway configuration
type Config struct {
	Gateway GatewayConfig `toml:"gateway"`
	Servers []ServerConfig `toml:"server"`
}

// GatewayConfig represents gateway-level configuration
type GatewayConfig struct {
	LogLevel string `toml:"log_level"`
	LogFile  string `toml:"log_file"`
}

// ServerConfig represents a single upstream MCP server configuration
type ServerConfig struct {
	Name       string                 `toml:"name"`
	Transport  string                 `toml:"transport"`
	Enabled    bool                   `toml:"enabled"`
	Command    string                 `toml:"command"`
	Args       []string               `toml:"args"`
	Env        map[string]string      `toml:"env"`
	URL        string                 `toml:"url"`
	SocketPath string                 `toml:"socket_path"`
	Timeout    int                    `toml:"timeout"`
	Metadata   map[string]interface{} `toml:"metadata"`
}

// LoadConfig loads the configuration from a TOML file
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Set defaults
	if cfg.Gateway.LogLevel == "" {
		cfg.Gateway.LogLevel = "info"
	}

	// Validate servers
	for i, srv := range cfg.Servers {
		if srv.Name == "" {
			return nil, fmt.Errorf("server %d missing required field: name", i)
		}
		if srv.Transport == "" {
			cfg.Servers[i].Transport = "stdio"
		}
		if srv.Timeout == 0 {
			cfg.Servers[i].Timeout = 30
		}
	}

	return &cfg, nil
}
