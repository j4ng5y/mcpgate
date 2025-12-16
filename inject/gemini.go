package inject

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// GeminiCLI represents the Gemini CLI agent
type GeminiCLI struct {
	configPath string
	config     map[string]interface{}
	backupPath string
}

// NewGeminiCLI creates a new Gemini CLI agent handler
func NewGeminiCLI() *GeminiCLI {
	return &GeminiCLI{}
}

// Name returns the agent name
func (g *GeminiCLI) Name() string {
	return "Gemini CLI"
}

// GetConfigPath returns the path to Gemini CLI's config file
func (g *GeminiCLI) GetConfigPath() (string, error) {
	if g.configPath != "" {
		return g.configPath, nil
	}

	configPath, err := ExpandPath("~/.gemini/settings.json")
	if err != nil {
		return "", err
	}

	g.configPath = configPath
	return configPath, nil
}

// IsInstalled checks if Gemini CLI is installed
func (g *GeminiCLI) IsInstalled() bool {
	configPath, err := g.GetConfigPath()
	if err != nil {
		return false
	}

	// Check if parent directory exists
	_, err = os.Stat(filepath.Dir(configPath))
	return err == nil
}

// GetBackupPath returns the backup file path
func (g *GeminiCLI) GetBackupPath() string {
	if g.backupPath == "" {
		g.backupPath = g.configPath + ".backup"
	}
	return g.backupPath
}

// CreateBackup creates a backup of the config file
func (g *GeminiCLI) CreateBackup() error {
	configPath, err := g.GetConfigPath()
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

	dest, err := os.Create(g.GetBackupPath())
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
func (g *GeminiCLI) RestoreBackup() error {
	backupPath := g.GetBackupPath()

	// If backup doesn't exist, nothing to restore
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		return nil
	}

	configPath, err := g.GetConfigPath()
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

// loadConfig loads the Gemini CLI config from disk
func (g *GeminiCLI) loadConfig() error {
	if g.config != nil {
		return nil
	}

	configPath, err := g.GetConfigPath()
	if err != nil {
		return err
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Create empty config structure if file doesn't exist
			g.config = map[string]interface{}{
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

	g.config = config
	return nil
}

// saveConfig saves the Gemini CLI config to disk
func (g *GeminiCLI) saveConfig() error {
	configPath, err := g.GetConfigPath()
	if err != nil {
		return err
	}

	if err := EnsureDir(configPath); err != nil {
		return err
	}

	data, err := json.MarshalIndent(g.config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}

// InjectStdio adds mcpgate (stdio mode) to Gemini CLI's config
func (g *GeminiCLI) InjectStdio(command string, args []string, serverName string, options map[string]interface{}) error {
	if err := g.loadConfig(); err != nil {
		return err
	}

	if g.IsInjected(serverName) {
		return ErrAlreadyInjected
	}

	// Ensure mcpServers key exists
	mcpServers, ok := g.config["mcpServers"].(map[string]interface{})
	if !ok {
		mcpServers = make(map[string]interface{})
		g.config["mcpServers"] = mcpServers
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

	return g.saveConfig()
}

// InjectHTTP adds mcpgate (HTTP mode) to Gemini CLI's config
func (g *GeminiCLI) InjectHTTP(serverURL string, serverName string, options map[string]interface{}) error {
	if err := g.loadConfig(); err != nil {
		return err
	}

	if g.IsInjected(serverName) {
		return ErrAlreadyInjected
	}

	// Ensure mcpServers key exists
	mcpServers, ok := g.config["mcpServers"].(map[string]interface{})
	if !ok {
		mcpServers = make(map[string]interface{})
		g.config["mcpServers"] = mcpServers
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

	return g.saveConfig()
}

// Eject removes mcpgate from Gemini CLI's config
func (g *GeminiCLI) Eject(serverName string) error {
	if err := g.loadConfig(); err != nil {
		return err
	}

	if !g.IsInjected(serverName) {
		return ErrNotInjected
	}

	mcpServers, ok := g.config["mcpServers"].(map[string]interface{})
	if !ok {
		return ErrInvalidConfig
	}

	delete(mcpServers, serverName)

	return g.saveConfig()
}

// IsInjected checks if mcpgate is already injected
func (g *GeminiCLI) IsInjected(serverName string) bool {
	if err := g.loadConfig(); err != nil {
		return false
	}

	mcpServers, ok := g.config["mcpServers"].(map[string]interface{})
	if !ok {
		return false
	}

	_, ok = mcpServers[serverName]
	return ok
}
