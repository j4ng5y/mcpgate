package server

import (
	"testing"

	"github.com/j4ng5y/mcpgate/config"
)

func TestRegistry_Register(t *testing.T) {
	registry := NewRegistry()

	server := &ManagedServer{
		Name: "test-server",
		Config: config.ServerConfig{
			Name:      "test-server",
			Transport: "stdio",
		},
		Capabilities: []string{"tools"},
	}

	err := registry.Register(server)
	if err != nil {
		t.Fatalf("Failed to register server: %v", err)
	}

	// Verify the server was registered
	retrieved, err := registry.Get("test-server")
	if err != nil {
		t.Fatalf("Failed to retrieve server: %v", err)
	}

	if retrieved.Name != "test-server" {
		t.Errorf("Expected server name 'test-server', got '%s'", retrieved.Name)
	}
}

func TestRegistry_RegisterDuplicate(t *testing.T) {
	registry := NewRegistry()

	server := &ManagedServer{
		Name: "test-server",
		Config: config.ServerConfig{
			Name:      "test-server",
			Transport: "stdio",
		},
	}

	if err := registry.Register(server); err != nil {
		t.Fatalf("Failed to register server: %v", err)
	}

	// Try to register again
	err := registry.Register(server)
	if err == nil {
		t.Fatal("Expected error when registering duplicate server")
	}
}

func TestRegistry_Unregister(t *testing.T) {
	registry := NewRegistry()

	server := &ManagedServer{
		Name: "test-server",
		Config: config.ServerConfig{
			Name:      "test-server",
			Transport: "stdio",
		},
	}

	if err := registry.Register(server); err != nil {
		t.Fatalf("Failed to register server: %v", err)
	}

	err := registry.Unregister("test-server")
	if err != nil {
		t.Fatalf("Failed to unregister server: %v", err)
	}

	// Verify the server was removed
	_, err = registry.Get("test-server")
	if err == nil {
		t.Fatal("Expected error when retrieving unregistered server")
	}
}

func TestRegistry_UnregisterNonexistent(t *testing.T) {
	registry := NewRegistry()

	err := registry.Unregister("nonexistent")
	if err == nil {
		t.Fatal("Expected error when unregistering nonexistent server")
	}
}

func TestRegistry_Get(t *testing.T) {
	registry := NewRegistry()

	server := &ManagedServer{
		Name: "test-server",
		Config: config.ServerConfig{
			Name:      "test-server",
			Transport: "stdio",
		},
	}

	if err := registry.Register(server); err != nil {
		t.Fatalf("Failed to register server: %v", err)
	}

	retrieved, err := registry.Get("test-server")
	if err != nil {
		t.Fatalf("Failed to get server: %v", err)
	}

	if retrieved.Name != "test-server" {
		t.Errorf("Expected server name 'test-server', got '%s'", retrieved.Name)
	}
}

func TestRegistry_GetNonexistent(t *testing.T) {
	registry := NewRegistry()

	_, err := registry.Get("nonexistent")
	if err == nil {
		t.Fatal("Expected error when getting nonexistent server")
	}
}

func TestRegistry_List(t *testing.T) {
	registry := NewRegistry()

	servers := []string{"server1", "server2", "server3"}
	for _, name := range servers {
		server := &ManagedServer{
			Name: name,
			Config: config.ServerConfig{
				Name:      name,
				Transport: "stdio",
			},
		}
		if err := registry.Register(server); err != nil {
		t.Fatalf("Failed to register server: %v", err)
	}
	}

	list := registry.List()
	if len(list) != 3 {
		t.Errorf("Expected 3 servers, got %d", len(list))
	}

	for _, srv := range list {
		found := false
		for _, name := range servers {
			if srv.Name == name {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Unexpected server in list: %s", srv.Name)
		}
	}
}

func TestRegistry_ListEmpty(t *testing.T) {
	registry := NewRegistry()

	list := registry.List()
	if len(list) != 0 {
		t.Errorf("Expected 0 servers, got %d", len(list))
	}
}

func TestRegistry_ListByCapability(t *testing.T) {
	registry := NewRegistry()

	server1 := &ManagedServer{
		Name:         "server1",
		Capabilities: []string{"tools", "resources"},
	}

	server2 := &ManagedServer{
		Name:         "server2",
		Capabilities: []string{"tools"},
	}

	server3 := &ManagedServer{
		Name:         "server3",
		Capabilities: []string{"prompts"},
	}

	if err := registry.Register(server1); err != nil {
		t.Fatalf("Failed to register server1: %v", err)
	}

	if err := registry.Register(server2); err != nil {
		t.Fatalf("Failed to register server2: %v", err)
	}

	if err := registry.Register(server3); err != nil {
		t.Fatalf("Failed to register server3: %v", err)
	}

	// Find servers with tools capability
	toolServers := registry.ListByCapability("tools")
	if len(toolServers) != 2 {
		t.Errorf("Expected 2 servers with tools capability, got %d", len(toolServers))
	}

	// Find servers with resources capability
	resourceServers := registry.ListByCapability("resources")
	if len(resourceServers) != 1 {
		t.Errorf("Expected 1 server with resources capability, got %d", len(resourceServers))
	}

	// Find servers with prompts capability
	promptServers := registry.ListByCapability("prompts")
	if len(promptServers) != 1 {
		t.Errorf("Expected 1 server with prompts capability, got %d", len(promptServers))
	}

	// Find servers with nonexistent capability
	noneServers := registry.ListByCapability("nonexistent")
	if len(noneServers) != 0 {
		t.Errorf("Expected 0 servers with nonexistent capability, got %d", len(noneServers))
	}
}

func TestRegistry_Concurrency(t *testing.T) {
	registry := NewRegistry()

	// Test concurrent registration
	done := make(chan bool, 10)
	errChan := make(chan error, 10)

	for i := 0; i < 10; i++ {
		go func(id int) {
			server := &ManagedServer{
				Name: "server" + string(rune(id)),
				Config: config.ServerConfig{
					Name: "server" + string(rune(id)),
				},
			}
			if err := registry.Register(server); err != nil {
				errChan <- err
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Check for any errors from goroutines
	select {
	case err := <-errChan:
		t.Fatalf("Failed to register server: %v", err)
	default:
	}

	// Verify all servers were registered
	list := registry.List()
	if len(list) != 10 {
		t.Errorf("Expected 10 servers after concurrent registration, got %d", len(list))
	}
}
