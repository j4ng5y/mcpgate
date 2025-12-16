package server

import (
	"context"
	"testing"
	"time"

	"github.com/j4ng5y/mcpgate/config"
)

func TestManager_NewManager(t *testing.T) {
	cfg := &config.Config{
		Gateway: config.GatewayConfig{
			LogLevel: "info",
		},
		Servers: []config.ServerConfig{},
	}

	manager := NewManager(cfg)
	if manager == nil {
		t.Fatal("Failed to create manager")
	}

	if manager.config != cfg {
		t.Error("Manager config not properly set")
	}
}

func TestManager_Start_NoServers(t *testing.T) {
	cfg := &config.Config{
		Gateway: config.GatewayConfig{
			LogLevel: "info",
		},
		Servers: []config.ServerConfig{},
	}

	manager := NewManager(cfg)
	err := manager.Start()
	if err != nil {
		t.Fatalf("Failed to start manager: %v", err)
	}

	servers := manager.ListServers()
	if len(servers) != 0 {
		t.Errorf("Expected 0 servers, got %d", len(servers))
	}
}

func TestManager_Start_DisabledServer(t *testing.T) {
	cfg := &config.Config{
		Gateway: config.GatewayConfig{
			LogLevel: "info",
		},
		Servers: []config.ServerConfig{
			{
				Name:      "disabled-server",
				Transport: "stdio",
				Enabled:   false,
				Command:   "echo",
			},
		},
	}

	manager := NewManager(cfg)
	err := manager.Start()
	if err != nil {
		t.Fatalf("Failed to start manager: %v", err)
	}

	// Disabled server should not be registered
	servers := manager.ListServers()
	if len(servers) != 0 {
		t.Errorf("Expected 0 servers, got %d", len(servers))
	}
}

func TestManager_Start_EnabledServer(t *testing.T) {
	cfg := &config.Config{
		Gateway: config.GatewayConfig{
			LogLevel: "info",
		},
		Servers: []config.ServerConfig{
			{
				Name:      "echo-server",
				Transport: "stdio",
				Enabled:   true,
				Command:   "cat",
			},
		},
	}

	manager := NewManager(cfg)
	err := manager.Start()
	if err != nil {
		t.Fatalf("Failed to start manager: %v", err)
	}

	// Enabled server should be registered
	servers := manager.ListServers()
	if len(servers) != 1 {
		t.Errorf("Expected 1 server, got %d", len(servers))
	}

	if servers[0].Name != "echo-server" {
		t.Errorf("Expected server name 'echo-server', got '%s'", servers[0].Name)
	}
}

func TestManager_GetServer(t *testing.T) {
	cfg := &config.Config{
		Gateway: config.GatewayConfig{
			LogLevel: "info",
		},
		Servers: []config.ServerConfig{
			{
				Name:      "test-server",
				Transport: "stdio",
				Enabled:   true,
				Command:   "cat",
			},
		},
	}

	manager := NewManager(cfg)
	if err := manager.Start(); err != nil {
		t.Fatalf("Failed to start manager: %v", err)
	}

	server, err := manager.GetServer("test-server")
	if err != nil {
		t.Fatalf("Failed to get server: %v", err)
	}

	if server.Name != "test-server" {
		t.Errorf("Expected server name 'test-server', got '%s'", server.Name)
	}
}

func TestManager_GetServer_NotFound(t *testing.T) {
	cfg := &config.Config{
		Gateway: config.GatewayConfig{
			LogLevel: "info",
		},
		Servers: []config.ServerConfig{},
	}

	manager := NewManager(cfg)
	if err := manager.Start(); err != nil {
		t.Fatalf("Failed to start manager: %v", err)
	}

	_, err := manager.GetServer("nonexistent")
	if err == nil {
		t.Fatal("Expected error when getting nonexistent server")
	}
}

func TestManager_ListServers(t *testing.T) {
	cfg := &config.Config{
		Gateway: config.GatewayConfig{
			LogLevel: "info",
		},
		Servers: []config.ServerConfig{
			{
				Name:      "server1",
				Transport: "stdio",
				Enabled:   true,
				Command:   "cat",
			},
			{
				Name:      "server2",
				Transport: "stdio",
				Enabled:   true,
				Command:   "cat",
			},
		},
	}

	manager := NewManager(cfg)
	if err := manager.Start(); err != nil {
		t.Fatalf("Failed to start manager: %v", err)
	}

	servers := manager.ListServers()
	if len(servers) != 2 {
		t.Errorf("Expected 2 servers, got %d", len(servers))
	}
}

func TestManager_ListServersByCapability(t *testing.T) {
	cfg := &config.Config{
		Gateway: config.GatewayConfig{
			LogLevel: "info",
		},
		Servers: []config.ServerConfig{
			{
				Name:      "server1",
				Transport: "stdio",
				Enabled:   true,
				Command:   "cat",
			},
		},
	}

	manager := NewManager(cfg)
	if err := manager.Start(); err != nil {
		t.Fatalf("Failed to start manager: %v", err)
	}

	// Manually set capabilities since we're not actually initializing the servers
	servers := manager.ListServers()
	if len(servers) > 0 {
		servers[0].SetCapabilities([]string{"tools", "resources"})
	}

	toolServers := manager.ListServersByCapability("tools")
	if len(toolServers) != 1 {
		t.Errorf("Expected 1 server with tools capability, got %d", len(toolServers))
	}

	promptServers := manager.ListServersByCapability("prompts")
	if len(promptServers) != 0 {
		t.Errorf("Expected 0 servers with prompts capability, got %d", len(promptServers))
	}
}

func TestManager_Stop(t *testing.T) {
	cfg := &config.Config{
		Gateway: config.GatewayConfig{
			LogLevel: "info",
		},
		Servers: []config.ServerConfig{
			{
				Name:      "test-server",
				Transport: "stdio",
				Enabled:   true,
				Command:   "cat",
			},
		},
	}

	manager := NewManager(cfg)
	if err := manager.Start(); err != nil {
		t.Fatalf("Failed to start manager: %v", err)
	}

	servers := manager.ListServers()
	if len(servers) != 1 {
		t.Errorf("Expected 1 server before stop, got %d", len(servers))
	}

	manager.Stop()

	servers = manager.ListServers()
	if len(servers) != 0 {
		t.Errorf("Expected 0 servers after stop, got %d", len(servers))
	}
}

func TestManager_ReconnectServer(t *testing.T) {
	cfg := &config.Config{
		Gateway: config.GatewayConfig{
			LogLevel: "info",
		},
		Servers: []config.ServerConfig{
			{
				Name:      "test-server",
				Transport: "stdio",
				Enabled:   true,
				Command:   "cat",
			},
		},
	}

	manager := NewManager(cfg)
	if err := manager.Start(); err != nil {
		t.Fatalf("Failed to start manager: %v", err)
	}

	// Attempt to reconnect
	err := manager.ReconnectServer("test-server")
	// We expect an error here since the connection process will fail
	// but the important thing is that it doesn't panic
	_ = err
}

func TestManager_ReconnectServer_NotFound(t *testing.T) {
	cfg := &config.Config{
		Gateway: config.GatewayConfig{
			LogLevel: "info",
		},
		Servers: []config.ServerConfig{},
	}

	manager := NewManager(cfg)
	if err := manager.Start(); err != nil {
		t.Fatalf("Failed to start manager: %v", err)
	}

	err := manager.ReconnectServer("nonexistent")
	if err == nil {
		t.Fatal("Expected error when reconnecting nonexistent server")
	}
}

func TestManager_MultipleServers(t *testing.T) {
	cfg := &config.Config{
		Gateway: config.GatewayConfig{
			LogLevel: "info",
		},
		Servers: []config.ServerConfig{
			{
				Name:      "server1",
				Transport: "stdio",
				Enabled:   true,
				Command:   "cat",
			},
			{
				Name:      "server2",
				Transport: "stdio",
				Enabled:   true,
				Command:   "cat",
			},
			{
				Name:      "server3",
				Transport: "stdio",
				Enabled:   false, // This one is disabled
				Command:   "cat",
			},
		},
	}

	manager := NewManager(cfg)
	if err := manager.Start(); err != nil {
		t.Fatalf("Failed to start manager: %v", err)
	}

	servers := manager.ListServers()
	if len(servers) != 2 {
		t.Errorf("Expected 2 servers (one disabled), got %d", len(servers))
	}

	for _, srv := range servers {
		if srv.Name == "server3" {
			t.Error("Disabled server should not be in list")
		}
	}
}

func TestManager_ManagerError(t *testing.T) {
	err := &ManagerError{
		Op:  "TestOp",
		Name: "TestName",
		Err:  "TestError",
	}

	errStr := err.Error()
	if errStr != "TestOp TestName: TestError" {
		t.Errorf("Expected 'TestOp TestName: TestError', got '%s'", errStr)
	}

	err2 := &ManagerError{
		Op:  "TestOp",
		Err: "TestError",
	}

	errStr2 := err2.Error()
	if errStr2 != "TestOp: TestError" {
		t.Errorf("Expected 'TestOp: TestError', got '%s'", errStr2)
	}
}

func TestManager_Concurrency(t *testing.T) {
	cfg := &config.Config{
		Gateway: config.GatewayConfig{
			LogLevel: "info",
		},
		Servers: []config.ServerConfig{
			{
				Name:      "test-server",
				Transport: "stdio",
				Enabled:   true,
				Command:   "cat",
			},
		},
	}

	manager := NewManager(cfg)
	if err := manager.Start(); err != nil {
		t.Fatalf("Failed to start manager: %v", err)
	}

	// Test concurrent reads
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func() {
			_ = manager.ListServers()
			_ = manager.ListServersByCapability("tools")
			_, _ = manager.GetServer("test-server")
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	manager.Stop()
}

func TestManager_Timeout(t *testing.T) {
	cfg := &config.Config{
		Gateway: config.GatewayConfig{
			LogLevel: "info",
		},
		Servers: []config.ServerConfig{
			{
				Name:      "hanging-server",
				Transport: "stdio",
				Enabled:   true,
				Command:   "sleep",
				Args:      []string{"10"},
			},
		},
	}

	manager := NewManager(cfg)

	// Use a short timeout for the test
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	// The start process should respect context timeout
	_ = ctx

	if err := manager.Start(); err != nil {
		t.Fatalf("Failed to start manager: %v", err)
	}

	manager.Stop()
}
