// +build integration

package main

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/j4ng5y/mcpgate/config"
	"github.com/j4ng5y/mcpgate/mcp"
	"github.com/j4ng5y/mcpgate/server"
)

func TestIntegration_FullWorkflow(t *testing.T) {
	// Create a test configuration
	cfg := &config.Config{
		Gateway: config.GatewayConfig{
			LogLevel: "debug",
		},
		Servers: []config.ServerConfig{
			{
				Name:      "echo-server",
				Transport: "stdio",
				Enabled:   true,
				Command:   "cat",
				Timeout:   10,
			},
		},
	}

	// Create and start the manager
	mgr := server.NewManager(cfg)
	err := mgr.Start()
	if err != nil {
		t.Fatalf("Failed to start manager: %v", err)
	}
	defer mgr.Stop()

	// Create the router
	router := mcp.NewRouter(mgr)

	// Test listing servers
	ctx := context.Background()
	req := &mcp.Request{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "gateway/list_servers",
	}

	resp := router.Route(ctx, req)
	if resp.Error != nil {
		t.Fatalf("gateway/list_servers failed: %v", resp.Error)
	}

	if resp.Result == nil {
		t.Fatal("Expected result from gateway/list_servers")
	}

	t.Logf("Servers: %v", resp.Result)
}

func TestIntegration_MultipleServers(t *testing.T) {
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
				Timeout:   5,
			},
			{
				Name:      "server2",
				Transport: "stdio",
				Enabled:   true,
				Command:   "cat",
				Timeout:   5,
			},
			{
				Name:      "server3",
				Transport: "stdio",
				Enabled:   false, // Disabled
				Command:   "cat",
			},
		},
	}

	mgr := server.NewManager(cfg)
	mgr.Start()
	defer mgr.Stop()

	servers := mgr.ListServers()
	if len(servers) != 2 {
		t.Errorf("Expected 2 enabled servers, got %d", len(servers))
	}

	// Check that server3 is not in the list
	for _, srv := range servers {
		if srv.Name == "server3" {
			t.Fatal("Disabled server should not be in list")
		}
	}
}

func TestIntegration_ConfigLoading(t *testing.T) {
	// Create a temporary config file
	configContent := `
[gateway]
log_level = "debug"

[[server]]
name = "test-server"
transport = "stdio"
enabled = true
command = "cat"
timeout = 30

[server.env]
TEST_ENV = "test_value"

[server.metadata]
version = "1.0.0"
`

	tmpFile, err := createTempConfig(configContent)
	if err != nil {
		t.Fatalf("Failed to create temp config: %v", err)
	}
	defer os.Remove(tmpFile)

	// Load configuration
	cfg, err := config.LoadConfig(tmpFile)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if cfg.Gateway.LogLevel != "debug" {
		t.Errorf("Expected log level 'debug', got '%s'", cfg.Gateway.LogLevel)
	}

	if len(cfg.Servers) != 1 {
		t.Errorf("Expected 1 server, got %d", len(cfg.Servers))
	}

	srv := cfg.Servers[0]
	if srv.Name != "test-server" {
		t.Errorf("Expected server name 'test-server', got '%s'", srv.Name)
	}

	if srv.Env["TEST_ENV"] != "test_value" {
		t.Error("Environment variables not loaded correctly")
	}

	if srv.Metadata["version"] != "1.0.0" {
		t.Error("Metadata not loaded correctly")
	}
}

func TestIntegration_RouterWithMultipleServers(t *testing.T) {
	cfg := &config.Config{
		Servers: []config.ServerConfig{
			{
				Name:      "tools-server",
				Transport: "stdio",
				Enabled:   true,
				Command:   "cat",
			},
			{
				Name:      "resources-server",
				Transport: "stdio",
				Enabled:   true,
				Command:   "cat",
			},
		},
	}

	mgr := server.NewManager(cfg)
	mgr.Start()
	defer mgr.Stop()

	// Set capabilities for the servers
	servers := mgr.ListServers()
	if len(servers) >= 1 {
		servers[0].SetCapabilities([]string{"tools"})
	}
	if len(servers) >= 2 {
		servers[1].SetCapabilities([]string{"resources"})
	}

	router := mcp.NewRouter(mgr)

	// Test getting server details
	ctx := context.Background()
	req := &mcp.Request{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "gateway/get_server",
		Params:  json.RawMessage(`{"name": "tools-server"}`),
	}

	resp := router.Route(ctx, req)
	if resp.Error != nil {
		t.Fatalf("Failed to get server: %v", resp.Error)
	}

	// Test listing capabilities
	req2 := &mcp.Request{
		JSONRPC: "2.0",
		ID:      2,
		Method:  "gateway/capabilities",
	}

	resp2 := router.Route(ctx, req2)
	if resp2.Error != nil {
		t.Fatalf("Failed to get capabilities: %v", resp2.Error)
	}
}

func TestIntegration_RegistryOperations(t *testing.T) {
	cfg := &config.Config{
		Servers: []config.ServerConfig{
			{
				Name:      "server1",
				Transport: "stdio",
				Enabled:   true,
				Command:   "cat",
			},
		},
	}

	mgr := server.NewManager(cfg)
	mgr.Start()
	defer mgr.Stop()

	// Get a specific server
	srv, err := mgr.GetServer("server1")
	if err != nil {
		t.Fatalf("Failed to get server: %v", err)
	}

	if srv.Name != "server1" {
		t.Errorf("Expected server name 'server1', got '%s'", srv.Name)
	}

	// List all servers
	servers := mgr.ListServers()
	if len(servers) != 1 {
		t.Errorf("Expected 1 server, got %d", len(servers))
	}

	// Set capabilities
	srv.SetCapabilities([]string{"tools", "resources"})
	if !srv.HasCapability("tools") {
		t.Error("Server should have tools capability")
	}

	// List by capability
	toolServers := mgr.ListServersByCapability("tools")
	if len(toolServers) != 1 {
		t.Errorf("Expected 1 server with tools capability, got %d", len(toolServers))
	}
}

func TestIntegration_ResponseStructure(t *testing.T) {
	cfg := &config.Config{
		Servers: []config.ServerConfig{},
	}

	mgr := server.NewManager(cfg)
	mgr.Start()
	defer mgr.Stop()

	router := mcp.NewRouter(mgr)

	ctx := context.Background()
	req := &mcp.Request{
		JSONRPC: "2.0",
		ID:      123,
		Method:  "gateway/list_servers",
	}

	resp := router.Route(ctx, req)

	// Verify response structure
	if resp.JSONRPC != "2.0" {
		t.Errorf("Expected JSONRPC 2.0, got %s", resp.JSONRPC)
	}

	if resp.ID != 123 {
		t.Errorf("Expected ID 123, got %v", resp.ID)
	}

	if resp.Error != nil {
		t.Errorf("Expected no error, got: %v", resp.Error)
	}
}

func TestIntegration_ErrorHandling(t *testing.T) {
	cfg := &config.Config{
		Servers: []config.ServerConfig{},
	}

	mgr := server.NewManager(cfg)
	mgr.Start()
	defer mgr.Stop()

	router := mcp.NewRouter(mgr)

	ctx := context.Background()

	// Test invalid JSON-RPC version
	req := &mcp.Request{
		JSONRPC: "1.0",
		ID:      1,
		Method:  "test",
	}

	resp := router.Route(ctx, req)
	if resp.Error == nil {
		t.Fatal("Expected error for invalid JSONRPC version")
	}

	if resp.Error.Code != mcp.InvalidRequest {
		t.Errorf("Expected error code %d, got %d", mcp.InvalidRequest, resp.Error.Code)
	}
}

func TestIntegration_ConcurrentRequests(t *testing.T) {
	cfg := &config.Config{
		Servers: []config.ServerConfig{
			{
				Name:      "server1",
				Transport: "stdio",
				Enabled:   true,
				Command:   "cat",
			},
		},
	}

	mgr := server.NewManager(cfg)
	mgr.Start()
	defer mgr.Stop()

	router := mcp.NewRouter(mgr)
	ctx := context.Background()

	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(id int) {
			req := &mcp.Request{
				JSONRPC: "2.0",
				ID:      id,
				Method:  "gateway/list_servers",
			}

			resp := router.Route(ctx, req)
			if resp.Error != nil {
				t.Errorf("Request %d failed: %v", id, resp.Error)
			}

			done <- true
		}(i)
	}

	// Wait for all requests to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestIntegration_ContextTimeout(t *testing.T) {
	cfg := &config.Config{
		Servers: []config.ServerConfig{},
	}

	mgr := server.NewManager(cfg)
	mgr.Start()
	defer mgr.Stop()

	router := mcp.NewRouter(mgr)

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	req := &mcp.Request{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "gateway/list_servers",
	}

	resp := router.Route(ctx, req)
	if resp.Error != nil && resp.Error.Code != mcp.InternalError {
		t.Errorf("Unexpected error: %v", resp.Error)
	}
}

func createTempConfig(content string) (string, error) {
	tmpDir := os.TempDir()
	f, err := os.CreateTemp(tmpDir, "test-config-*.toml")
	if err != nil {
		return "", err
	}
	defer f.Close()

	if _, err := f.WriteString(content); err != nil {
		os.Remove(f.Name())
		return "", err
	}

	return f.Name(), nil
}
