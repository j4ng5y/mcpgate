package server

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/j4ng5y/mcpgate/config"
)

// Manager manages the lifecycle of upstream MCP servers
type Manager struct {
	config   *config.Config
	registry *Registry
	servers  map[string]*ManagedServer
	mutex    sync.RWMutex
	done     chan struct{}
}

// NewManager creates a new server manager
func NewManager(cfg *config.Config) *Manager {
	return &Manager{
		config:   cfg,
		registry: NewRegistry(),
		servers:  make(map[string]*ManagedServer),
		done:     make(chan struct{}),
	}
}

// Start initializes and starts all configured servers
func (m *Manager) Start() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	for _, serverCfg := range m.config.Servers {
		if !serverCfg.Enabled {
			log.Printf("Skipping disabled server: %s", serverCfg.Name)
			continue
		}

		managed, err := NewManagedServer(serverCfg)
		if err != nil {
			log.Printf("Failed to create managed server %s: %v", serverCfg.Name, err)
			continue
		}

		m.servers[serverCfg.Name] = managed

		if err := m.registry.Register(managed); err != nil {
			log.Printf("Failed to register server %s: %v", serverCfg.Name, err)
			continue
		}

		log.Printf("Registered server: %s", serverCfg.Name)
	}

	// Connect all servers with retries
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	for name, server := range m.servers {
		if err := m.connectWithRetry(ctx, server, 3); err != nil {
			log.Printf("Failed to connect server %s after retries: %v", name, err)
		}
	}

	return nil
}

// connectWithRetry attempts to connect with exponential backoff
func (m *Manager) connectWithRetry(ctx context.Context, server *ManagedServer, maxRetries int) error {
	var lastErr error
	for attempt := 1; attempt <= maxRetries; attempt++ {
		if err := server.Connect(ctx); err == nil {
			log.Printf("Connected to server %s", server.Name)
			return nil
		} else {
			lastErr = err
			if attempt < maxRetries {
				backoff := time.Duration(attempt) * time.Second
				log.Printf("Retrying connection to %s in %v (attempt %d/%d)", server.Name, backoff, attempt, maxRetries)
				select {
				case <-time.After(backoff):
				case <-ctx.Done():
					return ctx.Err()
				}
			}
		}
	}
	return lastErr
}

// Stop disconnects all servers
func (m *Manager) Stop() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	for name, server := range m.servers {
		if err := server.Disconnect(ctx); err != nil {
			log.Printf("Error disconnecting server %s: %v", name, err)
		}
		// Also unregister from registry
		if err := m.registry.Unregister(name); err != nil {
			log.Printf("Error unregistering server %s: %v", name, err)
		}
	}

	m.servers = make(map[string]*ManagedServer)
}

// GetServer retrieves a managed server by name
func (m *Manager) GetServer(name string) (*ManagedServer, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	server, err := m.registry.Get(name)
	if err != nil {
		return nil, err
	}

	return server, nil
}

// ListServers returns all managed servers
func (m *Manager) ListServers() []*ManagedServer {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	return m.registry.List()
}

// ListServersByCapability returns servers with a specific capability
func (m *Manager) ListServersByCapability(capability string) []*ManagedServer {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	return m.registry.ListByCapability(capability)
}

// ReconnectServer reconnects a specific server
func (m *Manager) ReconnectServer(name string) error {
	m.mutex.Lock()
	server, exists := m.servers[name]
	m.mutex.Unlock()

	if !exists {
		return &ManagerError{Op: "ReconnectServer", Name: name, Err: "not found"}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Disconnect(ctx); err != nil {
		log.Printf("Error disconnecting server %s: %v", name, err)
	}
	return m.connectWithRetry(ctx, server, 3)
}

// ManagerError represents a manager operation error
type ManagerError struct {
	Op   string
	Name string
	Err  string
}

func (e *ManagerError) Error() string {
	if e.Name != "" {
		return e.Op + " " + e.Name + ": " + e.Err
	}
	return e.Op + ": " + e.Err
}
