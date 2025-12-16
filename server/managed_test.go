package server

import (
	"context"
	"testing"
	"time"

	"github.com/j4ng5y/mcpgate/config"
)

func TestManagedServer_NewManagedServer(t *testing.T) {
	cfg := config.ServerConfig{
		Name:      "test-server",
		Transport: "stdio",
		Command:   "echo",
		Timeout:   30,
	}

	server, err := NewManagedServer(cfg)
	if err != nil {
		t.Fatalf("Failed to create managed server: %v", err)
	}

	if server.Name != "test-server" {
		t.Errorf("Expected server name 'test-server', got '%s'", server.Name)
	}

	if server.IsConnected() {
		t.Error("Server should not be connected initially")
	}

	if server.IsInitialized() {
		t.Error("Server should not be initialized initially")
	}
}

func TestManagedServer_InvalidTransport(t *testing.T) {
	cfg := config.ServerConfig{
		Name:      "test-server",
		Transport: "invalid",
	}

	_, err := NewManagedServer(cfg)
	if err == nil {
		t.Fatal("Expected error for invalid transport type")
	}
}

func TestManagedServer_HasCapability(t *testing.T) {
	server := &ManagedServer{
		Name:         "test-server",
		Capabilities: []string{"tools", "resources"},
	}

	if !server.HasCapability("tools") {
		t.Error("Server should have 'tools' capability")
	}

	if !server.HasCapability("resources") {
		t.Error("Server should have 'resources' capability")
	}

	if server.HasCapability("prompts") {
		t.Error("Server should not have 'prompts' capability")
	}
}

func TestManagedServer_SetCapabilities(t *testing.T) {
	server := &ManagedServer{
		Name:         "test-server",
		Capabilities: []string{},
	}

	caps := []string{"tools", "resources", "prompts"}
	server.SetCapabilities(caps)

	if len(server.Capabilities) != 3 {
		t.Errorf("Expected 3 capabilities, got %d", len(server.Capabilities))
	}

	for _, cap := range caps {
		if !server.HasCapability(cap) {
			t.Errorf("Server should have capability '%s'", cap)
		}
	}
}

func TestManagedServer_IsConnected(t *testing.T) {
	server := &ManagedServer{
		Name: "test-server",
	}

	if server.IsConnected() {
		t.Error("New server should not be connected")
	}
}

func TestManagedServer_IsInitialized(t *testing.T) {
	server := &ManagedServer{
		Name: "test-server",
	}

	if server.IsInitialized() {
		t.Error("New server should not be initialized")
	}
}

func TestManagedServer_GetLastUsed(t *testing.T) {
	server := &ManagedServer{
		Name: "test-server",
	}

	lastUsed := server.GetLastUsed()
	if !lastUsed.IsZero() {
		t.Error("New server should have zero last used time")
	}
}

func TestManagedServer_Disconnect_NotConnected(t *testing.T) {
	server := &ManagedServer{
		Name: "test-server",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	err := server.Disconnect(ctx)
	if err != nil {
		t.Fatalf("Disconnect should not error for non-connected server: %v", err)
	}
}

func TestManagedServer_Connect_InvalidCommand(t *testing.T) {
	cfg := config.ServerConfig{
		Name:      "test-server",
		Transport: "stdio",
		Command:   "/nonexistent/command",
		Timeout:   5,
	}

	server, err := NewManagedServer(cfg)
	if err != nil {
		t.Fatalf("Failed to create managed server: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err = server.Connect(ctx)
	if err == nil {
		t.Fatal("Expected error when connecting to invalid command")
	}

	if server.IsConnected() {
		t.Error("Server should not be connected after failed connect")
	}
}

func TestManagedServer_Capabilities(t *testing.T) {
	server := &ManagedServer{
		Name:         "test-server",
		Capabilities: []string{"tools", "resources"},
	}

	if !server.HasCapability("tools") {
		t.Error("Server should have tools capability")
	}

	if server.HasCapability("nonexistent") {
		t.Error("Server should not have nonexistent capability")
	}
}

func TestManagedServer_Metadata(t *testing.T) {
	metadata := map[string]interface{}{
		"version": "1.0.0",
		"author":  "test",
	}

	server := &ManagedServer{
		Name:     "test-server",
		Metadata: metadata,
	}

	if server.Metadata["version"] != "1.0.0" {
		t.Error("Metadata not properly set")
	}
}

func TestManagedServer_Config(t *testing.T) {
	cfg := config.ServerConfig{
		Name:      "test-server",
		Transport: "stdio",
		Command:   "echo",
		Timeout:   60,
	}

	server, err := NewManagedServer(cfg)
	if err != nil {
		t.Fatalf("Failed to create managed server: %v", err)
	}

	if server.Config.Name != "test-server" {
		t.Error("Config not properly stored")
	}

	if server.Config.Timeout != 60 {
		t.Errorf("Expected timeout 60, got %d", server.Config.Timeout)
	}
}

func TestManagedServer_Concurrency(t *testing.T) {
	server := &ManagedServer{
		Name:         "test-server",
		Capabilities: []string{"tools"},
	}

	// Test concurrent reads
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func() {
			_ = server.IsConnected()
			_ = server.IsInitialized()
			_ = server.GetLastUsed()
			_ = server.HasCapability("tools")
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
}
