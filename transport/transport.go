package transport

import (
	"context"
	"encoding/json"
	"fmt"
)

// Transport defines the interface for communication with upstream MCP servers
type Transport interface {
	// Connect establishes a connection to the upstream server
	Connect(ctx context.Context) error

	// Disconnect closes the connection
	Disconnect(ctx context.Context) error

	// SendRequest sends a JSON-RPC request and waits for response
	SendRequest(ctx context.Context, request interface{}) (json.RawMessage, error)

	// IsConnected returns whether the transport is currently connected
	IsConnected() bool

	// Name returns the transport type name
	Name() string
}

// Factory creates transports based on type
type Factory struct{}

// NewFactory creates a new transport factory
func NewFactory() *Factory {
	return &Factory{}
}

// Create creates a new transport instance
func (f *Factory) Create(transportType string, config map[string]interface{}) (Transport, error) {
	switch transportType {
	case "stdio":
		return NewStdioTransport(config)
	case "http":
		return NewHTTPTransport(config)
	case "websocket":
		return NewWebSocketTransport(config)
	case "unix":
		return NewUnixSocketTransport(config)
	default:
		return nil, fmt.Errorf("unknown transport type: %s", transportType)
	}
}

// NewStdioTransport creates a new stdio transport
func NewStdioTransport(config map[string]interface{}) (Transport, error) {
	return &StdioTransport{
		config: config,
	}, nil
}

// NewHTTPTransport creates a new HTTP transport
func NewHTTPTransport(config map[string]interface{}) (Transport, error) {
	return &HTTPTransport{
		config: config,
	}, nil
}

// NewWebSocketTransport creates a new WebSocket transport
func NewWebSocketTransport(config map[string]interface{}) (Transport, error) {
	return &WebSocketTransport{
		config: config,
	}, nil
}

// NewUnixSocketTransport creates a new Unix socket transport
func NewUnixSocketTransport(config map[string]interface{}) (Transport, error) {
	return &UnixSocketTransport{
		config: config,
	}, nil
}
