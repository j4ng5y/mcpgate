package transport

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"
)

// HTTPTransport communicates with a remote MCP server via HTTP
type HTTPTransport struct {
	config    map[string]interface{}
	client    *http.Client
	baseURL   string
	mutex     sync.RWMutex
	connected bool
	timeout   time.Duration
}

// Connect establishes an HTTP connection (validates connectivity)
func (t *HTTPTransport) Connect(ctx context.Context) error {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	if t.connected {
		return nil
	}

	url, ok := t.config["url"].(string)
	if !ok {
		return fmt.Errorf("http transport requires 'url' configuration")
	}

	timeoutSec := 30
	if timeout, ok := t.config["timeout"].(int); ok {
		timeoutSec = timeout
	}

	t.baseURL = url
	t.timeout = time.Duration(timeoutSec) * time.Second
	t.client = &http.Client{
		Timeout: t.timeout,
	}

	// Test connectivity
	req, err := http.NewRequestWithContext(ctx, "GET", t.baseURL+"/health", nil)
	if err == nil {
		resp, err := t.client.Do(req)
		if err == nil {
			if err := resp.Body.Close(); err != nil {
				log.Printf("Error closing response body: %v", err)
			}
		}
	}

	t.connected = true
	return nil
}

// Disconnect closes the HTTP connection
func (t *HTTPTransport) Disconnect(ctx context.Context) error {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	if t.client != nil {
		t.client.CloseIdleConnections()
	}
	t.connected = false
	return nil
}

// SendRequest sends a JSON-RPC request via HTTP POST
func (t *HTTPTransport) SendRequest(ctx context.Context, request interface{}) (json.RawMessage, error) {
	t.mutex.RLock()
	if !t.connected {
		t.mutex.RUnlock()
		return nil, fmt.Errorf("not connected")
	}
	baseURL := t.baseURL
	client := t.client
	t.mutex.RUnlock()

	data, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", baseURL+"/rpc", bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request failed: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("Error closing response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("http error %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	return json.RawMessage(body), nil
}

// IsConnected returns connection status
func (t *HTTPTransport) IsConnected() bool {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	return t.connected
}

// Name returns transport type name
func (t *HTTPTransport) Name() string {
	return "http"
}
