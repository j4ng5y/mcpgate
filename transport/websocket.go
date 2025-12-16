package transport

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// WebSocketTransport communicates with a remote MCP server via WebSocket
type WebSocketTransport struct {
	config    map[string]interface{}
	conn      *websocket.Conn
	url       string
	mutex     sync.RWMutex
	connected bool
	respChan  chan json.RawMessage
	done      chan struct{}
	timeout   time.Duration
}

// Connect establishes a WebSocket connection
func (t *WebSocketTransport) Connect(ctx context.Context) error {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	if t.connected {
		return nil
	}

	url, ok := t.config["url"].(string)
	if !ok {
		return fmt.Errorf("websocket transport requires 'url' configuration")
	}

	timeoutSec := 30
	if timeout, ok := t.config["timeout"].(int); ok {
		timeoutSec = timeout
	}

	t.url = url
	t.timeout = time.Duration(timeoutSec) * time.Second

	dialer := websocket.Dialer{
		HandshakeTimeout: t.timeout,
	}

	conn, _, err := dialer.DialContext(ctx, t.url, nil)
	if err != nil {
		return fmt.Errorf("failed to connect to websocket: %w", err)
	}

	t.conn = conn
	t.connected = true
	t.respChan = make(chan json.RawMessage, 100)
	t.done = make(chan struct{})

	// Start reading responses in background
	go t.readResponses()

	return nil
}

// readResponses reads JSON responses from WebSocket
func (t *WebSocketTransport) readResponses() {
	defer close(t.respChan)
	for {
		select {
		case <-t.done:
			return
		default:
		}

		if err := t.conn.SetReadDeadline(time.Now().Add(t.timeout)); err != nil {
			t.mutex.Lock()
			t.connected = false
			t.mutex.Unlock()
			log.Printf("Error setting read deadline: %v", err)
			return
		}

		messageType, data, err := t.conn.ReadMessage()
		if err != nil {
			t.mutex.Lock()
			t.connected = false
			t.mutex.Unlock()
			return
		}

		if messageType == websocket.TextMessage {
			t.respChan <- json.RawMessage(data)
		}
	}
}

// Disconnect closes the WebSocket connection
func (t *WebSocketTransport) Disconnect(ctx context.Context) error {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	if !t.connected {
		return nil
	}

	close(t.done)
	t.connected = false

	if t.conn != nil {
		if err := t.conn.Close(); err != nil {
			log.Printf("Error closing connection: %v", err)
		}
	}

	return nil
}

// SendRequest sends a request via WebSocket
func (t *WebSocketTransport) SendRequest(ctx context.Context, request interface{}) (json.RawMessage, error) {
	t.mutex.RLock()
	if !t.connected {
		t.mutex.RUnlock()
		return nil, fmt.Errorf("not connected")
	}
	conn := t.conn
	t.mutex.RUnlock()

	data, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	if err := conn.SetWriteDeadline(time.Now().Add(t.timeout)); err != nil {
		return nil, fmt.Errorf("failed to set write deadline: %w", err)
	}

	if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
		return nil, fmt.Errorf("failed to write to websocket: %w", err)
	}

	// Wait for response with timeout
	select {
	case resp := <-t.respChan:
		return resp, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// IsConnected returns connection status
func (t *WebSocketTransport) IsConnected() bool {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	return t.connected
}

// Name returns transport type name
func (t *WebSocketTransport) Name() string {
	return "websocket"
}
