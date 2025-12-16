package inject

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestManager_RegisterAgent(t *testing.T) {
	manager := NewManager()
	claude := NewClaude()

	manager.RegisterAgent(claude)

	agent, err := manager.GetAgent("Claude Desktop")
	if err != nil {
		t.Fatalf("Failed to get agent: %v", err)
	}

	if agent.Name() != "Claude Desktop" {
		t.Errorf("Expected agent name 'Claude Desktop', got '%s'", agent.Name())
	}
}

func TestManager_GetAgent_NotFound(t *testing.T) {
	manager := NewManager()

	_, err := manager.GetAgent("NonExistent")
	if err == nil {
		t.Fatal("Expected error for nonexistent agent")
	}

	if !errors.Is(err, ErrAgentNotFound) {
		t.Errorf("Expected ErrAgentNotFound, got %v", err)
	}
}

func TestManager_ListInstalledAgents(t *testing.T) {
	manager := NewManager()
	manager.RegisterAgent(NewClaude())
	manager.RegisterAgent(NewCursor())
	manager.RegisterAgent(NewZed())

	installed := manager.ListInstalledAgents()

	// Note: This might be 0, 1, 2, or 3 depending on what's installed locally
	// We're just checking that the function runs without error
	if installed == nil {
		t.Error("Expected non-nil installed agents list")
	}
}

func TestClaude_ExpandPath(t *testing.T) {
	tests := []struct {
		input  string
		prefix string // What the expanded path should start with after expansion
	}{
		{"~/test", "test"},
		{"~/.claude/config.json", ".claude/config.json"},
	}

	for _, test := range tests {
		expanded, err := ExpandPath(test.input)
		if err != nil {
			t.Errorf("Failed to expand path %s: %v", test.input, err)
		}

		if filepath.Base(expanded) != filepath.Base(test.input) {
			t.Errorf("Expanded path %s doesn't end correctly", expanded)
		}
	}
}

func TestEnsureDir(t *testing.T) {
	tmpDir := t.TempDir()
	newDir := filepath.Join(tmpDir, "test", "nested", "dir", "file.txt")

	err := EnsureDir(newDir)
	if err != nil {
		t.Fatalf("Failed to ensure directory: %v", err)
	}

	dirPath := filepath.Dir(newDir)
	_, err = os.Stat(dirPath)
	if err != nil {
		t.Errorf("Directory was not created: %v", err)
	}
}

func TestClaude_Name(t *testing.T) {
	claude := NewClaude()
	if claude.Name() != "Claude Desktop" {
		t.Errorf("Expected name 'Claude Desktop', got '%s'", claude.Name())
	}
}

func TestCursor_Name(t *testing.T) {
	cursor := NewCursor()
	if cursor.Name() != "Cursor" {
		t.Errorf("Expected name 'Cursor', got '%s'", cursor.Name())
	}
}

func TestZed_Name(t *testing.T) {
	zed := NewZed()
	if zed.Name() != "Zed" {
		t.Errorf("Expected name 'Zed', got '%s'", zed.Name())
	}
}

func TestCodex_Name(t *testing.T) {
	codex := NewCodex()
	if codex.Name() != "GitHub Codex" {
		t.Errorf("Expected name 'GitHub Codex', got '%s'", codex.Name())
	}
}

func TestGemini_Name(t *testing.T) {
	gemini := NewGemini()
	if gemini.Name() != "Google Gemini" {
		t.Errorf("Expected name 'Google Gemini', got '%s'", gemini.Name())
	}
}

func TestOpenCode_Name(t *testing.T) {
	opencode := NewOpenCode()
	if opencode.Name() != "OpenCode" {
		t.Errorf("Expected name 'OpenCode', got '%s'", opencode.Name())
	}
}

func TestClaude_InjectHTTP_Eject_MemoryConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "claude_config.json")

	claude := NewClaude()
	// Override config path for testing
	claude.configPath = configPath

	// Test that we can inject via HTTP
	err := claude.InjectHTTP("http://localhost:8000", "mcpgate", nil)
	if err != nil {
		t.Fatalf("Failed to inject HTTP: %v", err)
	}

	// Test that IsInjected returns true after injection
	isInjected := claude.IsInjected("mcpgate")
	if !isInjected {
		t.Error("Expected IsInjected to return true after HTTP injection")
	}

	// Test that we can eject
	err = claude.Eject("mcpgate")
	if err != nil {
		t.Fatalf("Failed to eject: %v", err)
	}

	// Test that IsInjected returns false after ejection
	isInjected = claude.IsInjected("mcpgate")
	if isInjected {
		t.Error("Expected IsInjected to return false after ejection")
	}
}

func TestClaude_InjectStdio_Eject_MemoryConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "claude_config.json")

	claude := NewClaude()
	// Override config path for testing
	claude.configPath = configPath

	// Test that we can inject via stdio
	err := claude.InjectStdio("/path/to/mcpgate", []string{"server", "-c", "config.toml"}, "mcpgate", nil)
	if err != nil {
		t.Fatalf("Failed to inject stdio: %v", err)
	}

	// Test that IsInjected returns true after injection
	isInjected := claude.IsInjected("mcpgate")
	if !isInjected {
		t.Error("Expected IsInjected to return true after stdio injection")
	}

	// Test that we can eject
	err = claude.Eject("mcpgate")
	if err != nil {
		t.Fatalf("Failed to eject: %v", err)
	}

	// Test that IsInjected returns false after ejection
	isInjected = claude.IsInjected("mcpgate")
	if isInjected {
		t.Error("Expected IsInjected to return false after ejection")
	}
}

func TestCursor_InjectHTTP_MemoryConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "cursor_config.json")

	cursor := NewCursor()
	// Override config path for testing
	cursor.configPath = configPath

	// Test that we can inject via HTTP
	err := cursor.InjectHTTP("http://localhost:8000", "mcpgate", nil)
	if err != nil {
		t.Fatalf("Failed to inject HTTP: %v", err)
	}

	// Test that IsInjected returns true after injection
	isInjected := cursor.IsInjected("mcpgate")
	if !isInjected {
		t.Error("Expected IsInjected to return true after HTTP injection")
	}
}

func TestCursor_InjectStdio_MemoryConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "cursor_config.json")

	cursor := NewCursor()
	// Override config path for testing
	cursor.configPath = configPath

	// Test that we can inject via stdio
	err := cursor.InjectStdio("/path/to/mcpgate", []string{"server"}, "mcpgate", nil)
	if err != nil {
		t.Fatalf("Failed to inject stdio: %v", err)
	}

	// Test that IsInjected returns true after injection
	isInjected := cursor.IsInjected("mcpgate")
	if !isInjected {
		t.Error("Expected IsInjected to return true after stdio injection")
	}
}

func TestZed_InjectHTTP_MemoryConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "zed_config.json")

	zed := NewZed()
	// Override config path for testing
	zed.configPath = configPath

	// Test that we can inject via HTTP
	err := zed.InjectHTTP("http://localhost:8000", "mcpgate", nil)
	if err != nil {
		t.Fatalf("Failed to inject HTTP: %v", err)
	}

	// Test that IsInjected returns true after injection
	isInjected := zed.IsInjected("mcpgate")
	if !isInjected {
		t.Error("Expected IsInjected to return true after HTTP injection")
	}
}

func TestZed_InjectStdio_MemoryConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "zed_config.json")

	zed := NewZed()
	// Override config path for testing
	zed.configPath = configPath

	// Test that we can inject via stdio
	err := zed.InjectStdio("/path/to/mcpgate", []string{"server"}, "mcpgate", nil)
	if err != nil {
		t.Fatalf("Failed to inject stdio: %v", err)
	}

	// Test that IsInjected returns true after injection
	isInjected := zed.IsInjected("mcpgate")
	if !isInjected {
		t.Error("Expected IsInjected to return true after stdio injection")
	}
}

func TestCloudServices_NotInstalled(t *testing.T) {
	codex := NewCodex()
	if codex.IsInstalled() {
		t.Error("Codex should not be installed (cloud service)")
	}

	gemini := NewGemini()
	if gemini.IsInstalled() {
		t.Error("Gemini should not be installed (cloud service)")
	}
}

func TestCloudServices_OperationsFail(t *testing.T) {
	codex := NewCodex()

	_, err := codex.GetConfigPath()
	if err == nil {
		t.Error("Expected error from Codex.GetConfigPath()")
	}

	err = codex.CreateBackup()
	if err == nil {
		t.Error("Expected error from Codex.CreateBackup()")
	}

	err = codex.InjectHTTP("http://localhost:8000", "mcpgate", nil)
	if err == nil {
		t.Error("Expected error from Codex.InjectHTTP()")
	}

	err = codex.InjectStdio("/path/to/mcpgate", []string{"server"}, "mcpgate", nil)
	if err == nil {
		t.Error("Expected error from Codex.InjectStdio()")
	}
}

func TestManager_InjectAllHTTP_NoAgents(t *testing.T) {
	manager := NewManager()

	// Should not error if no agents
	err := manager.InjectAllHTTP("http://localhost:8000", "mcpgate", nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

func TestManager_InjectAllStdio_NoAgents(t *testing.T) {
	manager := NewManager()

	// Should not error if no agents
	err := manager.InjectAllStdio("/path/to/mcpgate", []string{"server"}, "mcpgate", nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

func TestManager_EjectAll_NoAgents(t *testing.T) {
	manager := NewManager()

	// Should not error if no agents
	err := manager.EjectAll("mcpgate")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}
