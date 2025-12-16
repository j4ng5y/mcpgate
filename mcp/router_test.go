package mcp

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/j4ng5y/mcpgate/config"
	"github.com/j4ng5y/mcpgate/server"
)

func TestRouter_NewRouter(t *testing.T) {
	cfg := &config.Config{
		Servers: []config.ServerConfig{},
	}
	manager := server.NewManager(cfg)
	if err := manager.Start(); err != nil {
		t.Fatalf("Failed to start manager: %v", err)
	}

	router := NewRouter(manager)
	if router == nil {
		t.Fatal("Failed to create router")
	}
}

func TestRouter_Route_InvalidJSONRPCVersion(t *testing.T) {
	cfg := &config.Config{
		Servers: []config.ServerConfig{},
	}
	manager := server.NewManager(cfg)
	if err := manager.Start(); err != nil {
		t.Fatalf("Failed to start manager: %v", err)
	}
	router := NewRouter(manager)

	ctx := context.Background()
	req := &Request{
		JSONRPC: "1.0",
		ID:      1,
		Method:  "test",
	}

	resp := router.Route(ctx, req)
	if resp.Error == nil {
		t.Fatal("Expected error for invalid JSON-RPC version")
	}

	if resp.Error.Code != InvalidRequest {
		t.Errorf("Expected error code %d, got %d", InvalidRequest, resp.Error.Code)
	}
}

func TestRouter_Route_ListServers(t *testing.T) {
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
	manager := server.NewManager(cfg)
	if err := manager.Start(); err != nil {
		t.Fatalf("Failed to start manager: %v", err)
	}
	router := NewRouter(manager)

	ctx := context.Background()
	req := &Request{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "gateway/list_servers",
	}

	resp := router.Route(ctx, req)
	if resp.Error != nil {
		t.Fatalf("Unexpected error: %v", resp.Error)
	}

	if resp.Result == nil {
		t.Fatal("Expected result in response")
	}

	// Result should be a list
	resultList, ok := resp.Result.([]map[string]interface{})
	if !ok {
		t.Errorf("Expected result to be a list, got %T", resp.Result)
	}

	if len(resultList) != 1 {
		t.Errorf("Expected 1 server in result, got %d", len(resultList))
	}

	manager.Stop()
}

func TestRouter_Route_GetServer(t *testing.T) {
	cfg := &config.Config{
		Servers: []config.ServerConfig{
			{
				Name:      "test-server",
				Transport: "stdio",
				Enabled:   true,
				Command:   "cat",
			},
		},
	}
	manager := server.NewManager(cfg)
	if err := manager.Start(); err != nil {
		t.Fatalf("Failed to start manager: %v", err)
	}
	router := NewRouter(manager)

	ctx := context.Background()
	params := map[string]interface{}{
		"name": "test-server",
	}
	paramsJSON, _ := json.Marshal(params)

	req := &Request{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "gateway/get_server",
		Params:  paramsJSON,
	}

	resp := router.Route(ctx, req)
	if resp.Error != nil {
		t.Fatalf("Unexpected error: %v", resp.Error)
	}

	if resp.Result == nil {
		t.Fatal("Expected result in response")
	}

	manager.Stop()
}

func TestRouter_Route_GetServer_NotFound(t *testing.T) {
	cfg := &config.Config{
		Servers: []config.ServerConfig{},
	}
	manager := server.NewManager(cfg)
	if err := manager.Start(); err != nil {
		t.Fatalf("Failed to start manager: %v", err)
	}
	router := NewRouter(manager)

	ctx := context.Background()
	params := map[string]interface{}{
		"name": "nonexistent",
	}
	paramsJSON, _ := json.Marshal(params)

	req := &Request{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "gateway/get_server",
		Params:  paramsJSON,
	}

	resp := router.Route(ctx, req)
	if resp.Error == nil {
		t.Fatal("Expected error for nonexistent server")
	}

	manager.Stop()
}

func TestRouter_Route_InvalidParams(t *testing.T) {
	cfg := &config.Config{
		Servers: []config.ServerConfig{},
	}
	manager := server.NewManager(cfg)
	if err := manager.Start(); err != nil {
		t.Fatalf("Failed to start manager: %v", err)
	}
	router := NewRouter(manager)

	ctx := context.Background()
	req := &Request{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "gateway/get_server",
		Params:  json.RawMessage(`{invalid json}`),
	}

	resp := router.Route(ctx, req)
	if resp.Error == nil {
		t.Fatal("Expected error for invalid params")
	}

	if resp.Error.Code != InvalidParams {
		t.Errorf("Expected error code %d, got %d", InvalidParams, resp.Error.Code)
	}

	manager.Stop()
}

func TestRouter_Route_ServerStatus(t *testing.T) {
	cfg := &config.Config{
		Servers: []config.ServerConfig{
			{
				Name:      "test-server",
				Transport: "stdio",
				Enabled:   true,
				Command:   "cat",
			},
		},
	}
	manager := server.NewManager(cfg)
	if err := manager.Start(); err != nil {
		t.Fatalf("Failed to start manager: %v", err)
	}
	router := NewRouter(manager)

	ctx := context.Background()
	params := map[string]interface{}{
		"name": "test-server",
	}
	paramsJSON, _ := json.Marshal(params)

	req := &Request{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "gateway/server_status",
		Params:  paramsJSON,
	}

	resp := router.Route(ctx, req)
	if resp.Error != nil {
		t.Fatalf("Unexpected error: %v", resp.Error)
	}

	manager.Stop()
}

func TestRouter_Route_Capabilities_All(t *testing.T) {
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
	manager := server.NewManager(cfg)
	if err := manager.Start(); err != nil {
		t.Fatalf("Failed to start manager: %v", err)
	}
	router := NewRouter(manager)

	ctx := context.Background()
	req := &Request{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "gateway/capabilities",
	}

	resp := router.Route(ctx, req)
	if resp.Error != nil {
		t.Fatalf("Unexpected error: %v", resp.Error)
	}

	if resp.Result == nil {
		t.Fatal("Expected result in response")
	}

	manager.Stop()
}

func TestRouter_Route_Capabilities_Specific(t *testing.T) {
	cfg := &config.Config{
		Servers: []config.ServerConfig{
			{
				Name:      "test-server",
				Transport: "stdio",
				Enabled:   true,
				Command:   "cat",
			},
		},
	}
	manager := server.NewManager(cfg)
	if err := manager.Start(); err != nil {
		t.Fatalf("Failed to start manager: %v", err)
	}
	router := NewRouter(manager)

	ctx := context.Background()
	params := map[string]interface{}{
		"name": "test-server",
	}
	paramsJSON, _ := json.Marshal(params)

	req := &Request{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "gateway/capabilities",
		Params:  paramsJSON,
	}

	resp := router.Route(ctx, req)
	if resp.Error != nil {
		t.Fatalf("Unexpected error: %v", resp.Error)
	}

	manager.Stop()
}

func TestRouter_ExtractCapability(t *testing.T) {
	router := &Router{}

	tests := []struct {
		method       string
		expectedCap  string
	}{
		{"tools/list", "tools"},
		{"tools/call", "tools"},
		{"resources/list", "resources"},
		{"resources/read", "resources"},
		{"prompts/list", "prompts"},
		{"prompts/get", "prompts"},
		{"unknown/method", ""},
		{"something", ""},
	}

	for _, test := range tests {
		cap := router.extractCapability(test.method)
		if cap != test.expectedCap {
			t.Errorf("Method %s: expected capability '%s', got '%s'", test.method, test.expectedCap, cap)
		}
	}
}

func TestRouter_JSONRPCErrorCodes(t *testing.T) {
	if ParseError != -32700 {
		t.Errorf("ParseError should be -32700, got %d", ParseError)
	}

	if InvalidRequest != -32600 {
		t.Errorf("InvalidRequest should be -32600, got %d", InvalidRequest)
	}

	if MethodNotFound != -32601 {
		t.Errorf("MethodNotFound should be -32601, got %d", MethodNotFound)
	}

	if InvalidParams != -32602 {
		t.Errorf("InvalidParams should be -32602, got %d", InvalidParams)
	}

	if InternalError != -32603 {
		t.Errorf("InternalError should be -32603, got %d", InternalError)
	}
}

func TestRouter_RequestID(t *testing.T) {
	cfg := &config.Config{
		Servers: []config.ServerConfig{},
	}
	manager := server.NewManager(cfg)
	if err := manager.Start(); err != nil {
		t.Fatalf("Failed to start manager: %v", err)
	}
	router := NewRouter(manager)

	ctx := context.Background()

	tests := []interface{}{
		1,
		"string-id",
		1.5,
		nil,
	}

	for _, id := range tests {
		req := &Request{
			JSONRPC: "2.0",
			ID:      id,
			Method:  "gateway/list_servers",
		}

		resp := router.Route(ctx, req)
		if resp.ID != id {
			t.Errorf("Response ID mismatch: expected %v, got %v", id, resp.ID)
		}
	}

	manager.Stop()
}
