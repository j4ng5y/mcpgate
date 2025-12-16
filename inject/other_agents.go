package inject

import (
	"fmt"
)

// OpenCode represents the OpenCode agent
type OpenCode struct {
	name string
}

// NewOpenCode creates a new OpenCode agent handler
func NewOpenCode() *OpenCode {
	return &OpenCode{name: "OpenCode"}
}

// Name returns the agent name
func (o *OpenCode) Name() string {
	return o.name
}

// GetConfigPath returns the path to OpenCode's config file
func (o *OpenCode) GetConfigPath() (string, error) {
	return "", fmt.Errorf("opencode configuration not yet implemented")
}

// IsInstalled checks if OpenCode is installed
func (o *OpenCode) IsInstalled() bool {
	return false
}

// GetBackupPath returns the backup file path
func (o *OpenCode) GetBackupPath() string {
	return ""
}

// CreateBackup creates a backup of the config file
func (o *OpenCode) CreateBackup() error {
	return fmt.Errorf("opencode configuration not yet implemented")
}

// RestoreBackup restores the config from backup
func (o *OpenCode) RestoreBackup() error {
	return fmt.Errorf("opencode configuration not yet implemented")
}

// InjectStdio adds mcpgate (stdio mode) to OpenCode's config
func (o *OpenCode) InjectStdio(command string, args []string, serverName string, options map[string]interface{}) error {
	return fmt.Errorf("opencode configuration not yet implemented")
}

// InjectHTTP adds mcpgate (HTTP mode) to OpenCode's config
func (o *OpenCode) InjectHTTP(serverURL string, serverName string, options map[string]interface{}) error {
	return fmt.Errorf("opencode configuration not yet implemented")
}

// Eject removes mcpgate from OpenCode's config
func (o *OpenCode) Eject(serverName string) error {
	return fmt.Errorf("opencode configuration not yet implemented")
}

// IsInjected checks if mcpgate is already injected
func (o *OpenCode) IsInjected(serverName string) bool {
	return false
}
