package transport

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"sync"
)

// UnixSocketTransport communicates via Unix domain socket
type UnixSocketTransport struct {
	config    map[string]interface{}
	conn      net.Conn
	reader    *bufio.Reader
	mutex     sync.RWMutex
	connected bool
	respChan  chan json.RawMessage
	done      chan struct{}
}

// Connect establishes a Unix socket connection
func (t *UnixSocketTransport) Connect(ctx context.Context) error {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	if t.connected {
		return nil
	}

	socketPath, ok := t.config["socket_path"].(string)
	if !ok {
		return fmt.Errorf("unix socket transport requires 'socket_path' configuration")
	}

	dialer := net.Dialer{}
	conn, err := dialer.DialContext(ctx, "unix", socketPath)
	if err != nil {
		return fmt.Errorf("failed to connect to unix socket: %w", err)
	}

	t.conn = conn
	t.reader = bufio.NewReader(conn)
	t.connected = true
	t.respChan = make(chan json.RawMessage, 100)
	t.done = make(chan struct{})

	// Start reading responses in background
	go t.readResponses()

	return nil
}

// readResponses reads JSON responses from Unix socket
func (t *UnixSocketTransport) readResponses() {
	defer close(t.respChan)
	for {
		select {
		case <-t.done:
			return
		default:
		}

		line, err := t.reader.ReadBytes('\n')
		if err != nil {
			t.mutex.Lock()
			t.connected = false
			t.mutex.Unlock()
			return
		}

		t.respChan <- json.RawMessage(line)
	}
}

// Disconnect closes the Unix socket connection
func (t *UnixSocketTransport) Disconnect(ctx context.Context) error {
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

// SendRequest sends a request via Unix socket
func (t *UnixSocketTransport) SendRequest(ctx context.Context, request interface{}) (json.RawMessage, error) {
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

	if _, err := conn.Write(append(data, '\n')); err != nil {
		return nil, fmt.Errorf("failed to write to socket: %w", err)
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
func (t *UnixSocketTransport) IsConnected() bool {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	return t.connected
}

// Name returns transport type name
func (t *UnixSocketTransport) Name() string {
	return "unix"
}
