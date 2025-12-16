package inject

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// CodexCLI represents the Codex CLI agent
type CodexCLI struct {
	configPath string
	config     map[string]interface{}
	backupPath string
}

// NewCodexCLI creates a new Codex CLI agent handler
func NewCodexCLI() *CodexCLI {
	return &CodexCLI{}
}

// Name returns the agent name
func (c *CodexCLI) Name() string {
	return "Codex CLI"
}

// GetConfigPath returns the path to Codex CLI's config file
func (c *CodexCLI) GetConfigPath() (string, error) {
	if c.configPath != "" {
		return c.configPath, nil
	}

	configPath, err := ExpandPath("~/.codex/config.toml")
	if err != nil {
		return "", err
	}

	c.configPath = configPath
	return configPath, nil
}

// IsInstalled checks if Codex CLI is installed
func (c *CodexCLI) IsInstalled() bool {
	configPath, err := c.GetConfigPath()
	if err != nil {
		return false
	}

	// Check if parent directory exists
	_, err = os.Stat(filepath.Dir(configPath))
	return err == nil
}

// GetBackupPath returns the backup file path
func (c *CodexCLI) GetBackupPath() string {
	if c.backupPath == "" {
		c.backupPath = c.configPath + ".backup"
	}
	return c.backupPath
}

// CreateBackup creates a backup of the config file
func (c *CodexCLI) CreateBackup() error {
	configPath, err := c.GetConfigPath()
	if err != nil {
		return err
	}

	// If file doesn't exist, no backup needed
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil
	}

	source, err := os.Open(configPath)
	if err != nil {
		return err
	}
	defer func() {
		_ = source.Close()
	}()

	dest, err := os.Create(c.GetBackupPath())
	if err != nil {
		return err
	}
	defer func() {
		_ = dest.Close()
	}()

	_, err = io.Copy(dest, source)
	return err
}

// RestoreBackup restores the config from backup
func (c *CodexCLI) RestoreBackup() error {
	backupPath := c.GetBackupPath()

	// If backup doesn't exist, nothing to restore
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		return nil
	}

	configPath, err := c.GetConfigPath()
	if err != nil {
		return err
	}

	source, err := os.Open(backupPath)
	if err != nil {
		return err
	}
	defer func() {
		_ = source.Close()
	}()

	dest, err := os.Create(configPath)
	if err != nil {
		return err
	}
	defer func() {
		_ = dest.Close()
	}()

	_, err = io.Copy(dest, source)
	return err
}

// loadConfig loads the Codex CLI config from disk
func (c *CodexCLI) loadConfig() error {
	if c.config != nil {
		return nil
	}

	configPath, err := c.GetConfigPath()
	if err != nil {
		return err
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Create empty config structure if file doesn't exist
			c.config = make(map[string]interface{})
			return nil
		}
		return err
	}

	config := make(map[string]interface{})
	if err := toml.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidConfig, err)
	}

	c.config = config
	return nil
}

// saveConfig saves the Codex CLI config to disk
func (c *CodexCLI) saveConfig() error {
	configPath, err := c.GetConfigPath()
	if err != nil {
		return err
	}

	if err := EnsureDir(configPath); err != nil {
		return err
	}

	var buf bytes.Buffer
	encoder := toml.NewEncoder(&buf)
	if err := encoder.Encode(c.config); err != nil {
		return err
	}

	return os.WriteFile(configPath, buf.Bytes(), 0644)
}

// InjectStdio adds mcpgate (stdio mode) to Codex CLI's config
func (c *CodexCLI) InjectStdio(command string, args []string, serverName string, options map[string]interface{}) error {
	if err := c.loadConfig(); err != nil {
		return err
	}

	if c.IsInjected(serverName) {
		return ErrAlreadyInjected
	}

	// Ensure mcp_servers key exists
	var mcpServers map[string]interface{}
	mcpServersRaw, ok := c.config["mcp_servers"]
	if !ok {
		mcpServers = make(map[string]interface{})
		c.config["mcp_servers"] = mcpServers
	} else {
		var okType bool
		mcpServers, okType = mcpServersRaw.(map[string]interface{})
		if !okType {
			mcpServers = make(map[string]interface{})
			c.config["mcp_servers"] = mcpServers
		}
	}

	// Create the mcpgate server config entry for stdio mode
	serverConfig := map[string]interface{}{
		"command": command,
		"args":    args,
	}

	// Add any additional options
	for key, value := range options {
		serverConfig[key] = value
	}

	mcpServers[serverName] = serverConfig

	return c.saveConfig()
}

// InjectHTTP adds mcpgate (HTTP mode) to Codex CLI's config
func (c *CodexCLI) InjectHTTP(serverURL string, serverName string, options map[string]interface{}) error {
	if err := c.loadConfig(); err != nil {
		return err
	}

	if c.IsInjected(serverName) {
		return ErrAlreadyInjected
	}

	// Ensure mcp_servers key exists
	var mcpServers map[string]interface{}
	mcpServersRaw, ok := c.config["mcp_servers"]
	if !ok {
		mcpServers = make(map[string]interface{})
		c.config["mcp_servers"] = mcpServers
	} else {
		var okType bool
		mcpServers, okType = mcpServersRaw.(map[string]interface{})
		if !okType {
			mcpServers = make(map[string]interface{})
			c.config["mcp_servers"] = mcpServers
		}
	}

	// Create the mcpgate server config entry for HTTP mode
	serverConfig := map[string]interface{}{
		"url": serverURL,
	}

	// Add any additional options
	for key, value := range options {
		serverConfig[key] = value
	}

	mcpServers[serverName] = serverConfig

	return c.saveConfig()
}

// Eject removes mcpgate from Codex CLI's config
func (c *CodexCLI) Eject(serverName string) error {
	if err := c.loadConfig(); err != nil {
		return err
	}

	if !c.IsInjected(serverName) {
		return ErrNotInjected
	}

	mcpServersRaw, ok := c.config["mcp_servers"]
	if !ok {
		return ErrInvalidConfig
	}

	mcpServers, ok := mcpServersRaw.(map[string]interface{})
	if !ok {
		return ErrInvalidConfig
	}

	delete(mcpServers, serverName)

	return c.saveConfig()
}

// IsInjected checks if mcpgate is already injected
func (c *CodexCLI) IsInjected(serverName string) bool {
	if err := c.loadConfig(); err != nil {
		return false
	}

	mcpServersRaw, ok := c.config["mcp_servers"]
	if !ok {
		return false
	}

	mcpServers, ok := mcpServersRaw.(map[string]interface{})
	if !ok {
		return false
	}

	_, ok = mcpServers[serverName]
	return ok
}
