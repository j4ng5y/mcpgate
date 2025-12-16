package inject

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
)

// Windsurf represents the Windsurf IDE agent
type Windsurf struct {
	configPath string
	config     map[string]interface{}
	backupPath string
}

// NewWindsurf creates a new Windsurf agent handler
func NewWindsurf() *Windsurf {
	return &Windsurf{}
}

// Name returns the agent name
func (w *Windsurf) Name() string {
	return "Windsurf"
}

// GetConfigPath returns the path to Windsurf's config file
func (w *Windsurf) GetConfigPath() (string, error) {
	if w.configPath != "" {
		return w.configPath, nil
	}

	var configPath string
	switch runtime.GOOS {
	case "darwin", "linux":
		configPath = "~/.codeium/windsurf/mcp_config.json"
	case "windows":
		configPath = "~\\.codeium\\windsurf\\mcp_config.json"
	default:
		return "", fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}

	expanded, err := ExpandPath(configPath)
	if err != nil {
		return "", err
	}

	w.configPath = expanded
	return expanded, nil
}

// IsInstalled checks if Windsurf is installed
func (w *Windsurf) IsInstalled() bool {
	configPath, err := w.GetConfigPath()
	if err != nil {
		return false
	}

	// Check if parent directory exists
	_, err = os.Stat(filepath.Dir(configPath))
	return err == nil
}

// GetBackupPath returns the backup file path
func (w *Windsurf) GetBackupPath() string {
	if w.backupPath == "" {
		w.backupPath = w.configPath + ".backup"
	}
	return w.backupPath
}

// CreateBackup creates a backup of the config file
func (w *Windsurf) CreateBackup() error {
	configPath, err := w.GetConfigPath()
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

	dest, err := os.Create(w.GetBackupPath())
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
func (w *Windsurf) RestoreBackup() error {
	backupPath := w.GetBackupPath()

	// If backup doesn't exist, nothing to restore
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		return nil
	}

	configPath, err := w.GetConfigPath()
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

// loadConfig loads the Windsurf config from disk
func (w *Windsurf) loadConfig() error {
	if w.config != nil {
		return nil
	}

	configPath, err := w.GetConfigPath()
	if err != nil {
		return err
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Create empty config structure if file doesn't exist
			w.config = map[string]interface{}{
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

	w.config = config
	return nil
}

// saveConfig saves the Windsurf config to disk
func (w *Windsurf) saveConfig() error {
	configPath, err := w.GetConfigPath()
	if err != nil {
		return err
	}

	if err := EnsureDir(configPath); err != nil {
		return err
	}

	data, err := json.MarshalIndent(w.config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}

// InjectStdio adds mcpgate (stdio mode) to Windsurf's config
func (w *Windsurf) InjectStdio(command string, args []string, serverName string, options map[string]interface{}) error {
	if err := w.loadConfig(); err != nil {
		return err
	}

	if w.IsInjected(serverName) {
		return ErrAlreadyInjected
	}

	// Ensure mcpServers key exists
	mcpServers, ok := w.config["mcpServers"].(map[string]interface{})
	if !ok {
		mcpServers = make(map[string]interface{})
		w.config["mcpServers"] = mcpServers
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

	return w.saveConfig()
}

// InjectHTTP adds mcpgate (HTTP mode) to Windsurf's config
func (w *Windsurf) InjectHTTP(serverURL string, serverName string, options map[string]interface{}) error {
	if err := w.loadConfig(); err != nil {
		return err
	}

	if w.IsInjected(serverName) {
		return ErrAlreadyInjected
	}

	// Ensure mcpServers key exists
	mcpServers, ok := w.config["mcpServers"].(map[string]interface{})
	if !ok {
		mcpServers = make(map[string]interface{})
		w.config["mcpServers"] = mcpServers
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

	return w.saveConfig()
}

// Eject removes mcpgate from Windsurf's config
func (w *Windsurf) Eject(serverName string) error {
	if err := w.loadConfig(); err != nil {
		return err
	}

	if !w.IsInjected(serverName) {
		return ErrNotInjected
	}

	mcpServers, ok := w.config["mcpServers"].(map[string]interface{})
	if !ok {
		return ErrInvalidConfig
	}

	delete(mcpServers, serverName)

	return w.saveConfig()
}

// IsInjected checks if mcpgate is already injected
func (w *Windsurf) IsInjected(serverName string) bool {
	if err := w.loadConfig(); err != nil {
		return false
	}

	mcpServers, ok := w.config["mcpServers"].(map[string]interface{})
	if !ok {
		return false
	}

	_, ok = mcpServers[serverName]
	return ok
}
