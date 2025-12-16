package config

import (
	"log"
	"os"
	"testing"
)

func TestLoadConfig_ValidConfig(t *testing.T) {
	// Create a temporary config file
	configContent := `
[gateway]
log_level = "debug"

[[server]]
name = "test-server"
transport = "stdio"
enabled = true
command = "echo"
args = ["hello"]
timeout = 60

[server.env]
TEST_VAR = "test_value"
`

	tmpFile, err := createTempConfig(configContent)
	if err != nil {
		t.Fatalf("Failed to create temp config: %v", err)
	}
	defer func() {
		_ = os.Remove(tmpFile)
	}()

	cfg, err := LoadConfig(tmpFile)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if cfg.Gateway.LogLevel != "debug" {
		t.Errorf("Expected log_level 'debug', got '%s'", cfg.Gateway.LogLevel)
	}

	if len(cfg.Servers) != 1 {
		t.Errorf("Expected 1 server, got %d", len(cfg.Servers))
	}

	server := cfg.Servers[0]
	if server.Name != "test-server" {
		t.Errorf("Expected server name 'test-server', got '%s'", server.Name)
	}

	if server.Transport != "stdio" {
		t.Errorf("Expected transport 'stdio', got '%s'", server.Transport)
	}

	if server.Timeout != 60 {
		t.Errorf("Expected timeout 60, got %d", server.Timeout)
	}

	if server.Env["TEST_VAR"] != "test_value" {
		t.Errorf("Expected TEST_VAR 'test_value', got '%s'", server.Env["TEST_VAR"])
	}
}

func TestLoadConfig_MissingServerName(t *testing.T) {
	configContent := `
[[server]]
transport = "stdio"
command = "echo"
`

	tmpFile, err := createTempConfig(configContent)
	if err != nil {
		t.Fatalf("Failed to create temp config: %v", err)
	}
	defer func() {
		_ = os.Remove(tmpFile)
	}()

	_, err = LoadConfig(tmpFile)
	if err == nil {
		t.Fatal("Expected error for missing server name")
	}
}

func TestLoadConfig_DefaultValues(t *testing.T) {
	configContent := `
[gateway]

[[server]]
name = "test-server"
command = "test"
enabled = true
`

	tmpFile, err := createTempConfig(configContent)
	if err != nil {
		t.Fatalf("Failed to create temp config: %v", err)
	}
	defer func() {
		_ = os.Remove(tmpFile)
	}()

	cfg, err := LoadConfig(tmpFile)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if cfg.Gateway.LogLevel != "info" {
		t.Errorf("Expected default log_level 'info', got '%s'", cfg.Gateway.LogLevel)
	}

	server := cfg.Servers[0]
	if server.Transport != "stdio" {
		t.Errorf("Expected default transport 'stdio', got '%s'", server.Transport)
	}

	if server.Timeout != 30 {
		t.Errorf("Expected default timeout 30, got %d", server.Timeout)
	}

	if !server.Enabled {
		t.Error("Expected server to be enabled when set to true in config")
	}
}

func TestLoadConfig_MultipleServers(t *testing.T) {
	configContent := `
[[server]]
name = "server1"
transport = "stdio"
command = "cmd1"
enabled = true

[[server]]
name = "server2"
transport = "http"
url = "http://localhost:8000"
enabled = false

[[server]]
name = "server3"
transport = "websocket"
url = "ws://localhost:9000"
enabled = true
`

	tmpFile, err := createTempConfig(configContent)
	if err != nil {
		t.Fatalf("Failed to create temp config: %v", err)
	}
	defer func() {
		_ = os.Remove(tmpFile)
	}()

	cfg, err := LoadConfig(tmpFile)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if len(cfg.Servers) != 3 {
		t.Errorf("Expected 3 servers, got %d", len(cfg.Servers))
	}

	if !cfg.Servers[0].Enabled || cfg.Servers[1].Enabled || !cfg.Servers[2].Enabled {
		t.Error("Server enabled flags not set correctly")
	}
}

func TestLoadConfig_FileNotFound(t *testing.T) {
	_, err := LoadConfig("/nonexistent/config.toml")
	if err == nil {
		t.Fatal("Expected error for missing file")
	}
}

func TestLoadConfig_InvalidTOML(t *testing.T) {
	configContent := `
this is not valid toml [[[
`

	tmpFile, err := createTempConfig(configContent)
	if err != nil {
		t.Fatalf("Failed to create temp config: %v", err)
	}
	defer func() {
		_ = os.Remove(tmpFile)
	}()

	_, err = LoadConfig(tmpFile)
	if err == nil {
		t.Fatal("Expected error for invalid TOML")
	}
}

func TestLoadConfig_AllTransportTypes(t *testing.T) {
	configContent := `
[[server]]
name = "stdio-server"
transport = "stdio"
command = "test"

[[server]]
name = "http-server"
transport = "http"
url = "http://localhost:8000"

[[server]]
name = "ws-server"
transport = "websocket"
url = "ws://localhost:9000"

[[server]]
name = "unix-server"
transport = "unix"
socket_path = "/tmp/test.sock"
`

	tmpFile, err := createTempConfig(configContent)
	if err != nil {
		t.Fatalf("Failed to create temp config: %v", err)
	}
	if tmpFile == "" {
		t.Fatal("Empty temp file path")
	}
	defer func(name string) {
		_ = os.Remove(name)
	}(tmpFile)

	cfg, err := LoadConfig(tmpFile)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	transports := []string{"stdio", "http", "websocket", "unix"}
	for i, expectedTransport := range transports {
		if cfg.Servers[i].Transport != expectedTransport {
			t.Errorf("Server %d: expected transport '%s', got '%s'", i, expectedTransport, cfg.Servers[i].Transport)
		}
	}
}

func TestLoadConfig_Metadata(t *testing.T) {
	configContent := `
[[server]]
name = "test-server"
command = "test"

[server.metadata]
version = "1.0.0"
description = "Test server"
custom_field = "custom_value"
`

	tmpFile, err := createTempConfig(configContent)
	if err != nil {
		t.Fatalf("Failed to create temp config: %v", err)
	}
	defer func() {
		_ = os.Remove(tmpFile)
	}()

	cfg, err := LoadConfig(tmpFile)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	metadata := cfg.Servers[0].Metadata
	if metadata["version"] != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got '%v'", metadata["version"])
	}

	if metadata["description"] != "Test server" {
		t.Errorf("Expected description 'Test server', got '%v'", metadata["description"])
	}
}

func TestLoadConfig_ServerArgs(t *testing.T) {
	configContent := `
[[server]]
name = "test-server"
command = "python"
args = ["-m", "module", "--flag", "value"]
`

	tmpFile, err := createTempConfig(configContent)
	if err != nil {
		t.Fatalf("Failed to create temp config: %v", err)
	}
	defer func() {
		_ = os.Remove(tmpFile)
	}()

	cfg, err := LoadConfig(tmpFile)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	args := cfg.Servers[0].Args
	expectedArgs := []string{"-m", "module", "--flag", "value"}
	if len(args) != len(expectedArgs) {
		t.Errorf("Expected %d args, got %d", len(expectedArgs), len(args))
	}

	for i, arg := range args {
		if arg != expectedArgs[i] {
			t.Errorf("Arg %d: expected '%s', got '%s'", i, expectedArgs[i], arg)
		}
	}
}

// Helper function to create temporary config files
func createTempConfig(content string) (string, error) {
	tmpDir := os.TempDir()

	f, err := os.CreateTemp(tmpDir, "test-config-*.toml")
	if err != nil {
		return "", err
	}
	defer func() {
		_ = f.Close()
	}()

	if _, err := f.WriteString(content); err != nil {
		if err := os.Remove(f.Name()); err != nil {
			log.Printf("Error removing file: %v", err)
		}
		return "", err
	}

	return f.Name(), nil
}
