package server

import (
	"fmt"
	"sync"
)

// Registry manages registered MCP servers
type Registry struct {
	servers map[string]*ManagedServer
	mutex   sync.RWMutex
}

// NewRegistry creates a new server registry
func NewRegistry() *Registry {
	return &Registry{
		servers: make(map[string]*ManagedServer),
	}
}

// Register registers a server
func (r *Registry) Register(server *ManagedServer) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, exists := r.servers[server.Name]; exists {
		return fmt.Errorf("server %s already registered", server.Name)
	}

	r.servers[server.Name] = server
	return nil
}

// Unregister removes a server from the registry
func (r *Registry) Unregister(name string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, exists := r.servers[name]; !exists {
		return fmt.Errorf("server %s not registered", name)
	}

	delete(r.servers, name)
	return nil
}

// Get retrieves a server by name
func (r *Registry) Get(name string) (*ManagedServer, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	server, exists := r.servers[name]
	if !exists {
		return nil, fmt.Errorf("server %s not found", name)
	}

	return server, nil
}

// List returns all registered servers
func (r *Registry) List() []*ManagedServer {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	servers := make([]*ManagedServer, 0, len(r.servers))
	for _, server := range r.servers {
		servers = append(servers, server)
	}

	return servers
}

// ListByCapability returns servers that support a specific capability
func (r *Registry) ListByCapability(capability string) []*ManagedServer {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	var result []*ManagedServer
	for _, server := range r.servers {
		if server.HasCapability(capability) {
			result = append(result, server)
		}
	}

	return result
}
