package inject

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
)

// Zed represents the Zed editor agent
type Zed struct {
	configPath string
	config     map[string]interface{}
	backupPath string
}

// NewZed creates a new Zed agent handler
func NewZed() *Zed {
	return &Zed{}
}

// Name returns the agent name
func (z *Zed) Name() string {
	return "Zed"
}

// GetConfigPath returns the path to Zed's config file
func (z *Zed) GetConfigPath() (string, error) {
	if z.configPath != "" {
		return z.configPath, nil
	}

	var configPath string
	switch runtime.GOOS {
	case "darwin":
		configPath = "~/Library/Application Support/Zed/settings.json"
	case "linux":
		configPath = "~/.config/zed/settings.json"
	case "windows":
		configPath = "~/AppData/Roaming/Zed/settings.json"
	default:
		return "", fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}

	expanded, err := ExpandPath(configPath)
	if err != nil {
		return "", err
	}

	z.configPath = expanded
	return expanded, nil
}

// IsInstalled checks if Zed is installed
func (z *Zed) IsInstalled() bool {
	configPath, err := z.GetConfigPath()
	if err != nil {
		return false
	}

	_, err = os.Stat(filepath.Dir(configPath))
	return err == nil
}

// GetBackupPath returns the backup file path
func (z *Zed) GetBackupPath() string {
	if z.backupPath == "" {
		z.backupPath = z.configPath + ".backup"
	}
	return z.backupPath
}

// CreateBackup creates a backup of the config file
func (z *Zed) CreateBackup() error {
	configPath, err := z.GetConfigPath()
	if err != nil {
		return err
	}

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

	dest, err := os.Create(z.GetBackupPath())
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
func (z *Zed) RestoreBackup() error {
	backupPath := z.GetBackupPath()

	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		return nil
	}

	configPath, err := z.GetConfigPath()
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

// loadConfig loads the Zed config from disk
func (z *Zed) loadConfig() error {
	if z.config != nil {
		return nil
	}

	configPath, err := z.GetConfigPath()
	if err != nil {
		return err
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			z.config = make(map[string]interface{})
			return nil
		}
		return err
	}

	config := make(map[string]interface{})
	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidConfig, err)
	}

	z.config = config
	return nil
}

// saveConfig saves the Zed config to disk
func (z *Zed) saveConfig() error {
	configPath, err := z.GetConfigPath()
	if err != nil {
		return err
	}

	if err := EnsureDir(configPath); err != nil {
		return err
	}

	data, err := json.MarshalIndent(z.config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}

// InjectStdio adds mcpgate (stdio mode) to Zed's config
func (z *Zed) InjectStdio(command string, args []string, serverName string, options map[string]interface{}) error {
	if err := z.loadConfig(); err != nil {
		return err
	}

	if z.IsInjected(serverName) {
		return ErrAlreadyInjected
	}

	// Ensure mcpServers key exists
	mcpServers, ok := z.config["mcpServers"].(map[string]interface{})
	if !ok {
		mcpServers = make(map[string]interface{})
		z.config["mcpServers"] = mcpServers
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

	return z.saveConfig()
}

// InjectHTTP adds mcpgate (HTTP mode) to Zed's config
func (z *Zed) InjectHTTP(serverURL string, serverName string, options map[string]interface{}) error {
	if err := z.loadConfig(); err != nil {
		return err
	}

	if z.IsInjected(serverName) {
		return ErrAlreadyInjected
	}

	// Ensure mcpServers key exists
	mcpServers, ok := z.config["mcpServers"].(map[string]interface{})
	if !ok {
		mcpServers = make(map[string]interface{})
		z.config["mcpServers"] = mcpServers
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

	return z.saveConfig()
}

// Eject removes mcpgate from Zed's config
func (z *Zed) Eject(serverName string) error {
	if err := z.loadConfig(); err != nil {
		return err
	}

	if !z.IsInjected(serverName) {
		return ErrNotInjected
	}

	mcpServers, ok := z.config["mcpServers"].(map[string]interface{})
	if !ok {
		return ErrInvalidConfig
	}

	delete(mcpServers, serverName)

	return z.saveConfig()
}

// IsInjected checks if mcpgate is already injected
func (z *Zed) IsInjected(serverName string) bool {
	if err := z.loadConfig(); err != nil {
		return false
	}

	mcpServers, ok := z.config["mcpServers"].(map[string]interface{})
	if !ok {
		return false
	}

	_, ok = mcpServers[serverName]
	return ok
}
