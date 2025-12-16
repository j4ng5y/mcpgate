package inject

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

var (
	ErrAgentNotFound     = errors.New("agent not found")
	ErrConfigNotFound    = errors.New("config file not found")
	ErrInvalidConfig     = errors.New("invalid config format")
	ErrAlreadyInjected   = errors.New("mcpgate already injected")
	ErrNotInjected       = errors.New("mcpgate not injected")
)

// Transport represents how mcpgate communicates with an agent
type Transport string

const (
	TransportStdio Transport = "stdio"
	TransportHTTP  Transport = "http"
)

// ServerConfig contains configuration for injecting mcpgate into an agent
type ServerConfig struct {
	Transport    Transport              // stdio or http
	Name         string                 // Server name in agent config
	URL          string                 // For HTTP mode: the URL (e.g., http://localhost:8000)
	Command      string                 // For stdio mode: path to mcpgate binary
	Args         []string               // For stdio mode: arguments to pass
	Options      map[string]interface{} // Additional agent-specific options
}

// Agent represents a supported AI agent
type Agent interface {
	// Name returns the agent name
	Name() string

	// GetConfigPath returns the path to the config file for this agent
	GetConfigPath() (string, error)

	// IsInstalled checks if the agent is installed
	IsInstalled() bool

	// InjectStdio adds mcpgate (stdio mode) to the agent's config
	InjectStdio(command string, args []string, serverName string, options map[string]interface{}) error

	// InjectHTTP adds mcpgate (HTTP mode) to the agent's config
	InjectHTTP(serverURL string, serverName string, options map[string]interface{}) error

	// Eject removes the mcpgate server from the agent's config
	Eject(serverName string) error

	// IsInjected checks if mcpgate is already injected
	IsInjected(serverName string) bool

	// GetBackupPath returns the path to the backup of the original config
	GetBackupPath() string

	// CreateBackup creates a backup of the original config
	CreateBackup() error

	// RestoreBackup restores the original config from backup
	RestoreBackup() error
}

// AgentConfig contains configuration for an agent
type AgentConfig struct {
	Name        string // Agent name
	ConfigPath  string // Full path to config file
	ServerURL   string // URL to mcpgate server
	ServerName  string // Name for the mcpgate entry
	Options     map[string]interface{}
}

// Manager handles injection/ejection across multiple agents
type Manager struct {
	agents map[string]Agent
}

// NewManager creates a new injection manager
func NewManager() *Manager {
	return &Manager{
		agents: make(map[string]Agent),
	}
}

// RegisterAgent registers an agent
func (m *Manager) RegisterAgent(agent Agent) {
	m.agents[agent.Name()] = agent
}

// GetAgent retrieves a registered agent by name
func (m *Manager) GetAgent(name string) (Agent, error) {
	agent, ok := m.agents[name]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrAgentNotFound, name)
	}
	return agent, nil
}

// ListInstalledAgents returns a list of installed agents
func (m *Manager) ListInstalledAgents() []Agent {
	var installed []Agent
	for _, agent := range m.agents {
		if agent.IsInstalled() {
			installed = append(installed, agent)
		}
	}
	return installed
}

// ListInjectedAgents returns a list of agents with mcpgate injected
func (m *Manager) ListInjectedAgents(serverName string) []Agent {
	var injected []Agent
	for _, agent := range m.agents {
		if agent.IsInstalled() && agent.IsInjected(serverName) {
			injected = append(injected, agent)
		}
	}
	return injected
}

// InjectAllStdio injects mcpgate (stdio mode) into all installed agents
func (m *Manager) InjectAllStdio(command string, args []string, serverName string, options map[string]interface{}) error {
	for _, agent := range m.agents {
		if !agent.IsInstalled() {
			continue
		}

		if err := agent.CreateBackup(); err != nil {
			return fmt.Errorf("failed to backup %s config: %w", agent.Name(), err)
		}

		if err := agent.InjectStdio(command, args, serverName, options); err != nil {
			// Try to restore backup on error
			if restoreErr := agent.RestoreBackup(); restoreErr != nil {
				return fmt.Errorf("injection failed and backup restore failed: %w (restore error: %v)", err, restoreErr)
			}
			return fmt.Errorf("failed to inject into %s: %w", agent.Name(), err)
		}
	}
	return nil
}

// InjectAllHTTP injects mcpgate (HTTP mode) into all installed agents
func (m *Manager) InjectAllHTTP(serverURL string, serverName string, options map[string]interface{}) error {
	for _, agent := range m.agents {
		if !agent.IsInstalled() {
			continue
		}

		if err := agent.CreateBackup(); err != nil {
			return fmt.Errorf("failed to backup %s config: %w", agent.Name(), err)
		}

		if err := agent.InjectHTTP(serverURL, serverName, options); err != nil {
			// Try to restore backup on error
			if restoreErr := agent.RestoreBackup(); restoreErr != nil {
				return fmt.Errorf("injection failed and backup restore failed: %w (restore error: %v)", err, restoreErr)
			}
			return fmt.Errorf("failed to inject into %s: %w", agent.Name(), err)
		}
	}
	return nil
}

// EjectAll removes mcpgate from all agents
func (m *Manager) EjectAll(serverName string) error {
	for _, agent := range m.agents {
		if !agent.IsInstalled() {
			continue
		}

		if !agent.IsInjected(serverName) {
			continue
		}

		if err := agent.Eject(serverName); err != nil {
			return fmt.Errorf("failed to eject from %s: %w", agent.Name(), err)
		}
	}
	return nil
}

// ExpandPath expands ~ and environment variables in a path
func ExpandPath(path string) (string, error) {
	// Expand environment variables
	expanded := os.ExpandEnv(path)

	// Expand home directory
	if len(expanded) > 0 && expanded[0] == '~' {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		expanded = filepath.Join(home, expanded[1:])
	}

	return expanded, nil
}

// EnsureDir creates a directory if it doesn't exist
func EnsureDir(path string) error {
	dir := filepath.Dir(path)
	return os.MkdirAll(dir, 0755)
}
