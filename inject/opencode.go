package inject

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// OpenCode represents the OpenCode agent
type OpenCode struct {
	configPath string
	config     map[string]interface{}
	backupPath string
}

// NewOpenCode creates a new OpenCode agent handler
func NewOpenCode() *OpenCode {
	return &OpenCode{}
}

// Name returns the agent name
func (o *OpenCode) Name() string {
	return "OpenCode"
}

// GetConfigPath returns the path to OpenCode's config file
func (o *OpenCode) GetConfigPath() (string, error) {
	if o.configPath != "" {
		return o.configPath, nil
	}

	configPath, err := ExpandPath("~/.config/opencode/opencode.json")
	if err != nil {
		return "", err
	}

	o.configPath = configPath
	return configPath, nil
}

// IsInstalled checks if OpenCode is installed
func (o *OpenCode) IsInstalled() bool {
	configPath, err := o.GetConfigPath()
	if err != nil {
		return false
	}

	// Check if parent directory exists
	_, err = os.Stat(filepath.Dir(configPath))
	return err == nil
}

// GetBackupPath returns the backup file path
func (o *OpenCode) GetBackupPath() string {
	if o.backupPath == "" {
		o.backupPath = o.configPath + ".backup"
	}
	return o.backupPath
}

// CreateBackup creates a backup of the config file
func (o *OpenCode) CreateBackup() error {
	configPath, err := o.GetConfigPath()
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

	dest, err := os.Create(o.GetBackupPath())
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
func (o *OpenCode) RestoreBackup() error {
	backupPath := o.GetBackupPath()

	// If backup doesn't exist, nothing to restore
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		return nil
	}

	configPath, err := o.GetConfigPath()
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

// loadConfig loads the OpenCode config from disk
func (o *OpenCode) loadConfig() error {
	if o.config != nil {
		return nil
	}

	configPath, err := o.GetConfigPath()
	if err != nil {
		return err
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Create empty config structure if file doesn't exist
			o.config = map[string]interface{}{
				"mcp": map[string]interface{}{},
			}
			return nil
		}
		return err
	}

	config := make(map[string]interface{})
	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidConfig, err)
	}

	o.config = config
	return nil
}

// saveConfig saves the OpenCode config to disk
func (o *OpenCode) saveConfig() error {
	configPath, err := o.GetConfigPath()
	if err != nil {
		return err
	}

	if err := EnsureDir(configPath); err != nil {
		return err
	}

	data, err := json.MarshalIndent(o.config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}

// InjectStdio adds mcpgate (stdio mode) to OpenCode's config
func (o *OpenCode) InjectStdio(command string, args []string, serverName string, options map[string]interface{}) error {
	if err := o.loadConfig(); err != nil {
		return err
	}

	if o.IsInjected(serverName) {
		return ErrAlreadyInjected
	}

	// Ensure mcp key exists
	var mcp map[string]interface{}
	mcpRaw, ok := o.config["mcp"]
	if !ok {
		mcp = make(map[string]interface{})
		o.config["mcp"] = mcp
	} else {
		var okType bool
		mcp, okType = mcpRaw.(map[string]interface{})
		if !okType {
			mcp = make(map[string]interface{})
			o.config["mcp"] = mcp
		}
	}

	// Create the mcpgate server config entry for stdio mode
	serverConfig := map[string]interface{}{
		"type":    "local",
		"command": args,
		"enabled": true,
	}

	// For local mode, prepend the command path to args
	if len(args) > 0 {
		fullCommand := append([]string{command}, args...)
		serverConfig["command"] = fullCommand
	} else {
		serverConfig["command"] = []string{command}
	}

	// Add any additional options
	for key, value := range options {
		serverConfig[key] = value
	}

	mcp[serverName] = serverConfig

	return o.saveConfig()
}

// InjectHTTP adds mcpgate (HTTP mode) to OpenCode's config
func (o *OpenCode) InjectHTTP(serverURL string, serverName string, options map[string]interface{}) error {
	if err := o.loadConfig(); err != nil {
		return err
	}

	if o.IsInjected(serverName) {
		return ErrAlreadyInjected
	}

	// Ensure mcp key exists
	var mcp map[string]interface{}
	mcpRaw, ok := o.config["mcp"]
	if !ok {
		mcp = make(map[string]interface{})
		o.config["mcp"] = mcp
	} else {
		var okType bool
		mcp, okType = mcpRaw.(map[string]interface{})
		if !okType {
			mcp = make(map[string]interface{})
			o.config["mcp"] = mcp
		}
	}

	// Create the mcpgate server config entry for HTTP mode
	serverConfig := map[string]interface{}{
		"type":    "remote",
		"url":     serverURL,
		"enabled": true,
	}

	// Add any additional options
	for key, value := range options {
		serverConfig[key] = value
	}

	mcp[serverName] = serverConfig

	return o.saveConfig()
}

// Eject removes mcpgate from OpenCode's config
func (o *OpenCode) Eject(serverName string) error {
	if err := o.loadConfig(); err != nil {
		return err
	}

	if !o.IsInjected(serverName) {
		return ErrNotInjected
	}

	mcpRaw, ok := o.config["mcp"]
	if !ok {
		return ErrInvalidConfig
	}

	mcp, ok := mcpRaw.(map[string]interface{})
	if !ok {
		return ErrInvalidConfig
	}

	delete(mcp, serverName)

	return o.saveConfig()
}

// IsInjected checks if mcpgate is already injected
func (o *OpenCode) IsInjected(serverName string) bool {
	if err := o.loadConfig(); err != nil {
		return false
	}

	mcpRaw, ok := o.config["mcp"]
	if !ok {
		return false
	}

	mcp, ok := mcpRaw.(map[string]interface{})
	if !ok {
		return false
	}

	_, ok = mcp[serverName]
	return ok
}
