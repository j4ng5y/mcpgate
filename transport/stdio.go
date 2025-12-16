package transport

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"sync"
)

// StdioTransport communicates with a subprocess via stdio
type StdioTransport struct {
	config    map[string]interface{}
	cmd       *exec.Cmd
	stdin     io.WriteCloser
	stdout    *bufio.Reader
	mutex     sync.RWMutex
	connected bool
	respChan  chan json.RawMessage
	done      chan struct{}
}

// Connect starts the subprocess and establishes communication
func (t *StdioTransport) Connect(ctx context.Context) error {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	if t.connected {
		return nil
	}

	command, ok := t.config["command"].(string)
	if !ok {
		return fmt.Errorf("stdio transport requires 'command' configuration")
	}

	args := []string{}
	if argsList, ok := t.config["args"].([]interface{}); ok {
		for _, arg := range argsList {
			if s, ok := arg.(string); ok {
				args = append(args, s)
			}
		}
	}

	t.cmd = exec.CommandContext(ctx, command, args...)

	// Set up environment variables
	t.cmd.Env = os.Environ()
	if envMap, ok := t.config["env"].(map[string]interface{}); ok {
		for key, val := range envMap {
			if str, ok := val.(string); ok {
				t.cmd.Env = append(t.cmd.Env, key+"="+str)
			}
		}
	}

	var err error
	t.stdin, err = t.cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	stdout, err := t.cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	if err := t.cmd.Start(); err != nil {
		return fmt.Errorf("failed to start subprocess: %w", err)
	}

	t.stdout = bufio.NewReader(stdout)
	t.connected = true
	t.respChan = make(chan json.RawMessage, 100)
	t.done = make(chan struct{})

	// Start reading responses in background
	go t.readResponses()

	return nil
}

// readResponses reads JSON responses from subprocess
func (t *StdioTransport) readResponses() {
	defer close(t.respChan)
	for {
		select {
		case <-t.done:
			return
		default:
		}

		line, err := t.stdout.ReadBytes('\n')
		if err != nil {
			t.mutex.Lock()
			t.connected = false
			t.mutex.Unlock()
			return
		}

		t.respChan <- json.RawMessage(line)
	}
}

// Disconnect stops the subprocess
func (t *StdioTransport) Disconnect(ctx context.Context) error {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	if !t.connected {
		return nil
	}

	close(t.done)
	t.connected = false

	if t.cmd != nil && t.cmd.Process != nil {
		if err := t.cmd.Process.Kill(); err != nil {
			log.Printf("Error killing process: %v", err)
		}
		if err := t.cmd.Wait(); err != nil {
			log.Printf("Error waiting for process: %v", err)
		}
	}

	return nil
}

// SendRequest sends a request to the subprocess
func (t *StdioTransport) SendRequest(ctx context.Context, request interface{}) (json.RawMessage, error) {
	t.mutex.RLock()
	if !t.connected {
		t.mutex.RUnlock()
		return nil, fmt.Errorf("not connected")
	}
	t.mutex.RUnlock()

	// Send request
	data, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	if _, err := t.stdin.Write(append(data, '\n')); err != nil {
		return nil, fmt.Errorf("failed to write to subprocess: %w", err)
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
func (t *StdioTransport) IsConnected() bool {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	return t.connected
}

// Name returns transport type name
func (t *StdioTransport) Name() string {
	return "stdio"
}
