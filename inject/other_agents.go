package inject

import (
	"fmt"
)

// Codex represents the GitHub Codex agent
type Codex struct {
	name string
}

// NewCodex creates a new Codex agent handler
func NewCodex() *Codex {
	return &Codex{name: "GitHub Codex"}
}

// Name returns the agent name
func (c *Codex) Name() string {
	return c.name
}

// GetConfigPath returns the path to Codex's config file
func (c *Codex) GetConfigPath() (string, error) {
	return "", fmt.Errorf("codex is a cloud service, not a local agent")
}

// IsInstalled checks if Codex is installed
func (c *Codex) IsInstalled() bool {
	return false // Codex is cloud-based
}

// GetBackupPath returns the backup file path
func (c *Codex) GetBackupPath() string {
	return ""
}

// CreateBackup creates a backup of the config file
func (c *Codex) CreateBackup() error {
	return fmt.Errorf("codex is a cloud service, cannot create local backups")
}

// RestoreBackup restores the config from backup
func (c *Codex) RestoreBackup() error {
	return fmt.Errorf("codex is a cloud service, cannot restore local backups")
}

// InjectStdio adds mcpgate (stdio mode) to Codex's config
func (c *Codex) InjectStdio(command string, args []string, serverName string, options map[string]interface{}) error {
	return fmt.Errorf("codex is a cloud service, injection not supported")
}

// InjectHTTP adds mcpgate (HTTP mode) to Codex's config
func (c *Codex) InjectHTTP(serverURL string, serverName string, options map[string]interface{}) error {
	return fmt.Errorf("codex is a cloud service, injection not supported")
}

// Eject removes mcpgate from Codex's config
func (c *Codex) Eject(serverName string) error {
	return fmt.Errorf("codex is a cloud service, ejection not supported")
}

// IsInjected checks if mcpgate is already injected
func (c *Codex) IsInjected(serverName string) bool {
	return false
}

// Gemini represents the Google Gemini agent
type Gemini struct {
	name string
}

// NewGemini creates a new Gemini agent handler
func NewGemini() *Gemini {
	return &Gemini{name: "Google Gemini"}
}

// Name returns the agent name
func (g *Gemini) Name() string {
	return g.name
}

// GetConfigPath returns the path to Gemini's config file
func (g *Gemini) GetConfigPath() (string, error) {
	return "", fmt.Errorf("gemini is a cloud service, not a local agent")
}

// IsInstalled checks if Gemini is installed
func (g *Gemini) IsInstalled() bool {
	return false // Gemini is cloud-based
}

// GetBackupPath returns the backup file path
func (g *Gemini) GetBackupPath() string {
	return ""
}

// CreateBackup creates a backup of the config file
func (g *Gemini) CreateBackup() error {
	return fmt.Errorf("gemini is a cloud service, cannot create local backups")
}

// RestoreBackup restores the config from backup
func (g *Gemini) RestoreBackup() error {
	return fmt.Errorf("gemini is a cloud service, cannot restore local backups")
}

// InjectStdio adds mcpgate (stdio mode) to Gemini's config
func (g *Gemini) InjectStdio(command string, args []string, serverName string, options map[string]interface{}) error {
	return fmt.Errorf("gemini is a cloud service, injection not supported")
}

// InjectHTTP adds mcpgate (HTTP mode) to Gemini's config
func (g *Gemini) InjectHTTP(serverURL string, serverName string, options map[string]interface{}) error {
	return fmt.Errorf("gemini is a cloud service, injection not supported")
}

// Eject removes mcpgate from Gemini's config
func (g *Gemini) Eject(serverName string) error {
	return fmt.Errorf("gemini is a cloud service, ejection not supported")
}

// IsInjected checks if mcpgate is already injected
func (g *Gemini) IsInjected(serverName string) bool {
	return false
}

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
