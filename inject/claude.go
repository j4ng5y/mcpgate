package inject

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// Claude represents the Claude Desktop agent
type Claude struct {
	configPath string
	config     map[string]interface{}
	backupPath string
}

// NewClaude creates a new Claude agent handler
func NewClaude() *Claude {
	return &Claude{}
}

// Name returns the agent name
func (c *Claude) Name() string {
	return "Claude Desktop"
}

// GetConfigPath returns the path to Claude's config file
func (c *Claude) GetConfigPath() (string, error) {
	if c.configPath != "" {
		return c.configPath, nil
	}

	configPath, err := ExpandPath("~/.claude/claude_desktop_config.json")
	if err != nil {
		return "", err
	}

	c.configPath = configPath
	return configPath, nil
}

// IsInstalled checks if Claude Desktop is installed
func (c *Claude) IsInstalled() bool {
	configPath, err := c.GetConfigPath()
	if err != nil {
		return false
	}

	_, err = os.Stat(configPath)
	return err == nil
}

// GetBackupPath returns the backup file path
func (c *Claude) GetBackupPath() string {
	if c.backupPath == "" {
		c.backupPath = c.configPath + ".backup"
	}
	return c.backupPath
}

// CreateBackup creates a backup of the config file
func (c *Claude) CreateBackup() error {
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
func (c *Claude) RestoreBackup() error {
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

// loadConfig loads the Claude config from disk
func (c *Claude) loadConfig() error {
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
			c.config = map[string]interface{}{
				"mcpServers": map[string]interface{}{},
			}
			return nil
		}
		return err
	}

	config := make(map[string]interface{})
	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidConfig, err)
	}

	c.config = config
	return nil
}

// saveConfig saves the Claude config to disk
func (c *Claude) saveConfig() error {
	configPath, err := c.GetConfigPath()
	if err != nil {
		return err
	}

	if err := EnsureDir(configPath); err != nil {
		return err
	}

	data, err := json.MarshalIndent(c.config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}

// InjectStdio adds mcpgate (stdio mode) to Claude's config
func (c *Claude) InjectStdio(command string, args []string, serverName string, options map[string]interface{}) error {
	if err := c.loadConfig(); err != nil {
		return err
	}

	if c.IsInjected(serverName) {
		return ErrAlreadyInjected
	}

	// Ensure mcpServers key exists
	mcpServers, ok := c.config["mcpServers"].(map[string]interface{})
	if !ok {
		mcpServers = make(map[string]interface{})
		c.config["mcpServers"] = mcpServers
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

// InjectHTTP adds mcpgate (HTTP mode) to Claude's config
func (c *Claude) InjectHTTP(serverURL string, serverName string, options map[string]interface{}) error {
	if err := c.loadConfig(); err != nil {
		return err
	}

	if c.IsInjected(serverName) {
		return ErrAlreadyInjected
	}

	// Ensure mcpServers key exists
	mcpServers, ok := c.config["mcpServers"].(map[string]interface{})
	if !ok {
		mcpServers = make(map[string]interface{})
		c.config["mcpServers"] = mcpServers
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

// Eject removes mcpgate from Claude's config
func (c *Claude) Eject(serverName string) error {
	if err := c.loadConfig(); err != nil {
		return err
	}

	if !c.IsInjected(serverName) {
		return ErrNotInjected
	}

	mcpServers, ok := c.config["mcpServers"].(map[string]interface{})
	if !ok {
		return ErrInvalidConfig
	}

	delete(mcpServers, serverName)

	return c.saveConfig()
}

// IsInjected checks if mcpgate is already injected
func (c *Claude) IsInjected(serverName string) bool {
	if err := c.loadConfig(); err != nil {
		return false
	}

	mcpServers, ok := c.config["mcpServers"].(map[string]interface{})
	if !ok {
		return false
	}

	_, ok = mcpServers[serverName]
	return ok
}
