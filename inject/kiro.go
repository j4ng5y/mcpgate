package inject

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// Kiro represents the Kiro IDE agent
type Kiro struct {
	configPath string
	config     map[string]interface{}
	backupPath string
}

// NewKiro creates a new Kiro agent handler
func NewKiro() *Kiro {
	return &Kiro{}
}

// Name returns the agent name
func (k *Kiro) Name() string {
	return "Kiro"
}

// GetConfigPath returns the path to Kiro's config file (user-level)
func (k *Kiro) GetConfigPath() (string, error) {
	if k.configPath != "" {
		return k.configPath, nil
	}

	configPath, err := ExpandPath("~/.kiro/settings/mcp.json")
	if err != nil {
		return "", err
	}

	k.configPath = configPath
	return configPath, nil
}

// IsInstalled checks if Kiro is installed
func (k *Kiro) IsInstalled() bool {
	configPath, err := k.GetConfigPath()
	if err != nil {
		return false
	}

	// Check if parent directory exists
	_, err = os.Stat(filepath.Dir(configPath))
	return err == nil
}

// GetBackupPath returns the backup file path
func (k *Kiro) GetBackupPath() string {
	if k.backupPath == "" {
		k.backupPath = k.configPath + ".backup"
	}
	return k.backupPath
}

// CreateBackup creates a backup of the config file
func (k *Kiro) CreateBackup() error {
	configPath, err := k.GetConfigPath()
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

	dest, err := os.Create(k.GetBackupPath())
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
func (k *Kiro) RestoreBackup() error {
	backupPath := k.GetBackupPath()

	// If backup doesn't exist, nothing to restore
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		return nil
	}

	configPath, err := k.GetConfigPath()
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

// loadConfig loads the Kiro config from disk
func (k *Kiro) loadConfig() error {
	if k.config != nil {
		return nil
	}

	configPath, err := k.GetConfigPath()
	if err != nil {
		return err
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Create empty config structure if file doesn't exist
			k.config = map[string]interface{}{
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

	k.config = config
	return nil
}

// saveConfig saves the Kiro config to disk
func (k *Kiro) saveConfig() error {
	configPath, err := k.GetConfigPath()
	if err != nil {
		return err
	}

	if err := EnsureDir(configPath); err != nil {
		return err
	}

	data, err := json.MarshalIndent(k.config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}

// InjectStdio adds mcpgate (stdio mode) to Kiro's config
func (k *Kiro) InjectStdio(command string, args []string, serverName string, options map[string]interface{}) error {
	if err := k.loadConfig(); err != nil {
		return err
	}

	if k.IsInjected(serverName) {
		return ErrAlreadyInjected
	}

	// Ensure mcpServers key exists
	mcpServers, ok := k.config["mcpServers"].(map[string]interface{})
	if !ok {
		mcpServers = make(map[string]interface{})
		k.config["mcpServers"] = mcpServers
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

	return k.saveConfig()
}

// InjectHTTP adds mcpgate (HTTP mode) to Kiro's config
func (k *Kiro) InjectHTTP(serverURL string, serverName string, options map[string]interface{}) error {
	if err := k.loadConfig(); err != nil {
		return err
	}

	if k.IsInjected(serverName) {
		return ErrAlreadyInjected
	}

	// Ensure mcpServers key exists
	mcpServers, ok := k.config["mcpServers"].(map[string]interface{})
	if !ok {
		mcpServers = make(map[string]interface{})
		k.config["mcpServers"] = mcpServers
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

	return k.saveConfig()
}

// Eject removes mcpgate from Kiro's config
func (k *Kiro) Eject(serverName string) error {
	if err := k.loadConfig(); err != nil {
		return err
	}

	if !k.IsInjected(serverName) {
		return ErrNotInjected
	}

	mcpServers, ok := k.config["mcpServers"].(map[string]interface{})
	if !ok {
		return ErrInvalidConfig
	}

	delete(mcpServers, serverName)

	return k.saveConfig()
}

// IsInjected checks if mcpgate is already injected
func (k *Kiro) IsInjected(serverName string) bool {
	if err := k.loadConfig(); err != nil {
		return false
	}

	mcpServers, ok := k.config["mcpServers"].(map[string]interface{})
	if !ok {
		return false
	}

	_, ok = mcpServers[serverName]
	return ok
}
