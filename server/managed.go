package server

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/j4ng5y/mcpgate/config"
	"github.com/j4ng5y/mcpgate/transport"
)

// ManagedServer wraps an upstream MCP server with connection management
type ManagedServer struct {
	Name        string
	Config      config.ServerConfig
	Transport   transport.Transport
	Capabilities []string
	Metadata    map[string]interface{}

	mutex       sync.RWMutex
	initialized bool
	connected   bool
	lastError   error
	lastUsed    time.Time
}

// NewManagedServer creates a new managed server
func NewManagedServer(cfg config.ServerConfig) (*ManagedServer, error) {
	factory := transport.NewFactory()

	// Convert config to map for transport
	configMap := map[string]interface{}{
		"command":     cfg.Command,
		"args":        cfg.Args,
		"env":         cfg.Env,
		"url":         cfg.URL,
		"socket_path": cfg.SocketPath,
		"timeout":     cfg.Timeout,
	}

	t, err := factory.Create(cfg.Transport, configMap)
	if err != nil {
		return nil, err
	}

	return &ManagedServer{
		Name:        cfg.Name,
		Config:      cfg,
		Transport:   t,
		Capabilities: []string{},
		Metadata:    cfg.Metadata,
	}, nil
}

// Connect establishes a connection to the upstream server
func (s *ManagedServer) Connect(ctx context.Context) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.connected {
		return nil
	}

	if err := s.Transport.Connect(ctx); err != nil {
		s.lastError = err
		log.Printf("Failed to connect to server %s: %v", s.Name, err)
		return err
	}

	s.connected = true
	s.lastUsed = time.Now()

	// Initialize the server
	if err := s.initialize(ctx); err != nil {
		s.connected = false
		s.lastError = err
		log.Printf("Failed to initialize server %s: %v", s.Name, err)
		return err
	}

	return nil
}

// initialize sends the initialize request to the server
func (s *ManagedServer) initialize(ctx context.Context) error {
	req := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "initialize",
		"params":  map[string]interface{}{},
	}

	resp, err := s.Transport.SendRequest(ctx, req)
	if err != nil {
		return err
	}

	var response map[string]interface{}
	if err := json.Unmarshal(resp, &response); err != nil {
		return err
	}

	// Check for error in response
	if errObj, exists := response["error"]; exists && errObj != nil {
		errMap, ok := errObj.(map[string]interface{})
		if ok {
			code, _ := errMap["code"].(float64)
			message, _ := errMap["message"].(string)
			return &JSONRPCError{
				Code:    int(code),
				Message: message,
			}
		}
	}

	s.initialized = true
	return nil
}

// Disconnect closes the connection to the upstream server
func (s *ManagedServer) Disconnect(ctx context.Context) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if !s.connected {
		return nil
	}

	s.connected = false
	return s.Transport.Disconnect(ctx)
}

// SendRequest forwards a request to the upstream server
// Returns raw JSON response that can be parsed by the router
func (s *ManagedServer) SendRequest(ctx context.Context, request interface{}) (json.RawMessage, error) {
	s.mutex.Lock()
	s.lastUsed = time.Now()
	connected := s.connected
	initialized := s.initialized
	s.mutex.Unlock()

	if !connected || !initialized {
		errResp := map[string]interface{}{
			"jsonrpc": "2.0",
			"error": map[string]interface{}{
				"code":    -32603,
				"message": "Server not connected or initialized",
			},
		}
		data, _ := json.Marshal(errResp)
		return json.RawMessage(data), nil
	}

	resp, err := s.Transport.SendRequest(ctx, request)
	if err != nil {
		errResp := map[string]interface{}{
			"jsonrpc": "2.0",
			"error": map[string]interface{}{
				"code":    -32603,
				"message": err.Error(),
			},
		}
		data, _ := json.Marshal(errResp)
		return json.RawMessage(data), nil
	}

	return resp, nil
}

// IsConnected returns connection status
func (s *ManagedServer) IsConnected() bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.connected
}

// IsInitialized returns initialization status
func (s *ManagedServer) IsInitialized() bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.initialized
}

// HasCapability checks if server has a specific capability
func (s *ManagedServer) HasCapability(capability string) bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	for _, cap := range s.Capabilities {
		if cap == capability {
			return true
		}
	}
	return false
}

// GetLastUsed returns the last time this server was used
func (s *ManagedServer) GetLastUsed() time.Time {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.lastUsed
}

// SetCapabilities updates the server's capabilities
func (s *ManagedServer) SetCapabilities(caps []string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.Capabilities = caps
}

// JSONRPCError represents a JSON-RPC error
type JSONRPCError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func (e *JSONRPCError) Error() string {
	return e.Message
}
