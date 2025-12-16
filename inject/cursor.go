package inject

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
)

// Cursor represents the Cursor editor agent
type Cursor struct {
	configPath string
	config     map[string]interface{}
	backupPath string
}

// NewCursor creates a new Cursor agent handler
func NewCursor() *Cursor {
	return &Cursor{}
}

// Name returns the agent name
func (c *Cursor) Name() string {
	return "Cursor"
}

// GetConfigPath returns the path to Cursor's config file
func (c *Cursor) GetConfigPath() (string, error) {
	if c.configPath != "" {
		return c.configPath, nil
	}

	var configPath string
	switch runtime.GOOS {
	case "darwin":
		configPath = "~/Library/Application Support/Cursor/User/settings.json"
	case "linux":
		configPath = "~/.config/Cursor/User/settings.json"
	case "windows":
		configPath = "~/AppData/Roaming/Cursor/User/settings.json"
	default:
		return "", fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}

	expanded, err := ExpandPath(configPath)
	if err != nil {
		return "", err
	}

	c.configPath = expanded
	return expanded, nil
}

// IsInstalled checks if Cursor is installed
func (c *Cursor) IsInstalled() bool {
	configPath, err := c.GetConfigPath()
	if err != nil {
		return false
	}

	_, err = os.Stat(filepath.Dir(configPath))
	return err == nil
}

// GetBackupPath returns the backup file path
func (c *Cursor) GetBackupPath() string {
	if c.backupPath == "" {
		c.backupPath = c.configPath + ".backup"
	}
	return c.backupPath
}

// CreateBackup creates a backup of the config file
func (c *Cursor) CreateBackup() error {
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
func (c *Cursor) RestoreBackup() error {
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

// loadConfig loads the Cursor config from disk
func (c *Cursor) loadConfig() error {
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
			c.config = make(map[string]interface{})
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

// saveConfig saves the Cursor config to disk
func (c *Cursor) saveConfig() error {
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

// InjectStdio adds mcpgate (stdio mode) to Cursor's config
func (c *Cursor) InjectStdio(command string, args []string, serverName string, options map[string]interface{}) error {
	if err := c.loadConfig(); err != nil {
		return err
	}

	if c.IsInjected(serverName) {
		return ErrAlreadyInjected
	}

	// Cursor uses "modelContextProtocol.servers" instead of "mcpServers"
	mcpServers, ok := c.config["modelContextProtocol"].(map[string]interface{})
	if !ok {
		mcpServers = make(map[string]interface{})
		c.config["modelContextProtocol"] = mcpServers
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

	// Ensure servers key exists
	servers, ok := mcpServers["servers"].(map[string]interface{})
	if !ok {
		servers = make(map[string]interface{})
		mcpServers["servers"] = servers
	}

	servers[serverName] = serverConfig

	return c.saveConfig()
}

// InjectHTTP adds mcpgate (HTTP mode) to Cursor's config
func (c *Cursor) InjectHTTP(serverURL string, serverName string, options map[string]interface{}) error {
	if err := c.loadConfig(); err != nil {
		return err
	}

	if c.IsInjected(serverName) {
		return ErrAlreadyInjected
	}

	// Cursor uses "modelContextProtocol.servers" instead of "mcpServers"
	mcpServers, ok := c.config["modelContextProtocol"].(map[string]interface{})
	if !ok {
		mcpServers = make(map[string]interface{})
		c.config["modelContextProtocol"] = mcpServers
	}

	// Create the mcpgate server config entry for HTTP mode
	serverConfig := map[string]interface{}{
		"url": serverURL,
	}

	// Add any additional options
	for key, value := range options {
		serverConfig[key] = value
	}

	// Ensure servers key exists
	servers, ok := mcpServers["servers"].(map[string]interface{})
	if !ok {
		servers = make(map[string]interface{})
		mcpServers["servers"] = servers
	}

	servers[serverName] = serverConfig

	return c.saveConfig()
}

// Eject removes mcpgate from Cursor's config
func (c *Cursor) Eject(serverName string) error {
	if err := c.loadConfig(); err != nil {
		return err
	}

	if !c.IsInjected(serverName) {
		return ErrNotInjected
	}

	mcpServers, ok := c.config["modelContextProtocol"].(map[string]interface{})
	if !ok {
		return ErrInvalidConfig
	}

	servers, ok := mcpServers["servers"].(map[string]interface{})
	if !ok {
		return ErrInvalidConfig
	}

	delete(servers, serverName)

	return c.saveConfig()
}

// IsInjected checks if mcpgate is already injected
func (c *Cursor) IsInjected(serverName string) bool {
	if err := c.loadConfig(); err != nil {
		return false
	}

	mcpServers, ok := c.config["modelContextProtocol"].(map[string]interface{})
	if !ok {
		return false
	}

	servers, ok := mcpServers["servers"].(map[string]interface{})
	if !ok {
		return false
	}

	_, ok = servers[serverName]
	return ok
}
