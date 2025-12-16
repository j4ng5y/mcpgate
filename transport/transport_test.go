package transport

import (
	"context"
	"testing"
	"time"
)

func TestTransportFactory_CreateStdio(t *testing.T) {
	factory := NewFactory()
	config := map[string]interface{}{
		"command": "echo",
		"args":    []interface{}{"test"},
	}

	transport, err := factory.Create("stdio", config)
	if err != nil {
		t.Fatalf("Failed to create stdio transport: %v", err)
	}

	if transport.Name() != "stdio" {
		t.Errorf("Expected transport name 'stdio', got '%s'", transport.Name())
	}

	if transport.IsConnected() {
		t.Error("Transport should not be connected initially")
	}
}

func TestTransportFactory_CreateHTTP(t *testing.T) {
	factory := NewFactory()
	config := map[string]interface{}{
		"url": "http://localhost:8000",
	}

	transport, err := factory.Create("http", config)
	if err != nil {
		t.Fatalf("Failed to create HTTP transport: %v", err)
	}

	if transport.Name() != "http" {
		t.Errorf("Expected transport name 'http', got '%s'", transport.Name())
	}

	if transport.IsConnected() {
		t.Error("Transport should not be connected initially")
	}
}

func TestTransportFactory_CreateWebSocket(t *testing.T) {
	factory := NewFactory()
	config := map[string]interface{}{
		"url": "ws://localhost:9000",
	}

	transport, err := factory.Create("websocket", config)
	if err != nil {
		t.Fatalf("Failed to create WebSocket transport: %v", err)
	}

	if transport.Name() != "websocket" {
		t.Errorf("Expected transport name 'websocket', got '%s'", transport.Name())
	}

	if transport.IsConnected() {
		t.Error("Transport should not be connected initially")
	}
}

func TestTransportFactory_CreateUnixSocket(t *testing.T) {
	factory := NewFactory()
	config := map[string]interface{}{
		"socket_path": "/tmp/test.sock",
	}

	transport, err := factory.Create("unix", config)
	if err != nil {
		t.Fatalf("Failed to create Unix socket transport: %v", err)
	}

	if transport.Name() != "unix" {
		t.Errorf("Expected transport name 'unix', got '%s'", transport.Name())
	}

	if transport.IsConnected() {
		t.Error("Transport should not be connected initially")
	}
}

func TestTransportFactory_InvalidType(t *testing.T) {
	factory := NewFactory()
	config := map[string]interface{}{}

	_, err := factory.Create("invalid", config)
	if err == nil {
		t.Fatal("Expected error for invalid transport type")
	}
}

func TestStdioTransport_MissingCommand(t *testing.T) {
	config := map[string]interface{}{
		// No command specified
	}

	transport, err := NewStdioTransport(config)
	if err != nil {
		t.Fatalf("Failed to create transport: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	err = transport.Connect(ctx)
	if err == nil {
		t.Fatal("Expected error for missing command")
	}
}

func TestStdioTransport_ValidCommand(t *testing.T) {
	config := map[string]interface{}{
		"command": "cat",
	}

	transport, err := NewStdioTransport(config)
	if err != nil {
		t.Fatalf("Failed to create transport: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err = transport.Connect(ctx)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}

	if !transport.IsConnected() {
		t.Error("Transport should be connected")
	}

	// Clean up
	if err := transport.Disconnect(ctx); err != nil {
		t.Fatalf("Failed to disconnect: %v", err)
	}
}

func TestHTTPTransport_MissingURL(t *testing.T) {
	config := map[string]interface{}{
		// No URL specified
	}

	transport, err := NewHTTPTransport(config)
	if err != nil {
		t.Fatalf("Failed to create transport: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	err = transport.Connect(ctx)
	if err == nil {
		t.Fatal("Expected error for missing URL")
	}
}

func TestUnixSocketTransport_MissingSocketPath(t *testing.T) {
	config := map[string]interface{}{
		// No socket_path specified
	}

	transport, err := NewUnixSocketTransport(config)
	if err != nil {
		t.Fatalf("Failed to create transport: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	err = transport.Connect(ctx)
	if err == nil {
		t.Fatal("Expected error for missing socket path")
	}
}

func TestWebSocketTransport_MissingURL(t *testing.T) {
	config := map[string]interface{}{
		// No URL specified
	}

	transport, err := NewWebSocketTransport(config)
	if err != nil {
		t.Fatalf("Failed to create transport: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	err = transport.Connect(ctx)
	if err == nil {
		t.Fatal("Expected error for missing URL")
	}
}

func TestStdioTransport_Disconnect(t *testing.T) {
	config := map[string]interface{}{
		"command": "cat",
	}

	transport, err := NewStdioTransport(config)
	if err != nil {
		t.Fatalf("Failed to create transport: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := transport.Connect(ctx); err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}

	err = transport.Disconnect(ctx)
	if err != nil {
		t.Fatalf("Failed to disconnect: %v", err)
	}

	if transport.IsConnected() {
		t.Error("Transport should be disconnected")
	}
}

func TestHTTPTransport_Disconnect(t *testing.T) {
	config := map[string]interface{}{
		"url": "http://localhost:8000",
	}

	transport, err := NewHTTPTransport(config)
	if err != nil {
		t.Fatalf("Failed to create transport: %v", err)
	}

	// Try to disconnect without connecting
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	if err := transport.Disconnect(ctx); err != nil {
		t.Fatalf("Failed to disconnect: %v", err)
	}

	if transport.IsConnected() {
		t.Error("Transport should not be connected")
	}
}
