package test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/hungthai1401/occtx/internal/context"
)

// TestHelper provides utilities for testing
type TestHelper struct {
	TempDir     string
	ConfigDir   string
	SettingsDir string
	StateFile   string
	t           *testing.T
}

func NewTestHelper(t *testing.T) *TestHelper {
	tempDir, err := os.MkdirTemp("", "occtx-test-*")
	if err != nil {
		t.Fatal(err)
	}

	configDir := filepath.Join(tempDir, ".config", "opencode")
	settingsDir := filepath.Join(configDir, "settings")
	stateFile := filepath.Join(settingsDir, ".occtx-state.json")

	// Create directories
	if err := os.MkdirAll(settingsDir, 0755); err != nil {
		t.Fatal(err)
	}

	return &TestHelper{
		TempDir:     tempDir,
		ConfigDir:   configDir,
		SettingsDir: settingsDir,
		StateFile:   stateFile,
		t:           t,
	}
}

func (th *TestHelper) Cleanup() {
	os.RemoveAll(th.TempDir)
}

// SetupEnvironment sets up environment variables for cross-platform testing
func (th *TestHelper) SetupEnvironment() (func(), error) {
	var oldVars map[string]string = make(map[string]string)

	if runtime.GOOS == "windows" {
		// On Windows, set USERPROFILE and APPDATA
		if val := os.Getenv("USERPROFILE"); val != "" {
			oldVars["USERPROFILE"] = val
		}
		if val := os.Getenv("APPDATA"); val != "" {
			oldVars["APPDATA"] = val
		}
		if val := os.Getenv("LOCALAPPDATA"); val != "" {
			oldVars["LOCALAPPDATA"] = val
		}

		os.Setenv("USERPROFILE", th.TempDir)
		os.Setenv("APPDATA", filepath.Join(th.TempDir, "AppData", "Roaming"))
		os.Setenv("LOCALAPPDATA", filepath.Join(th.TempDir, "AppData", "Local"))
	} else {
		// On Unix systems, set HOME
		if val := os.Getenv("HOME"); val != "" {
			oldVars["HOME"] = val
		}
		os.Setenv("HOME", th.TempDir)
	}

	// Return cleanup function
	return func() {
		if runtime.GOOS == "windows" {
			if val, ok := oldVars["USERPROFILE"]; ok {
				os.Setenv("USERPROFILE", val)
			} else {
				os.Unsetenv("USERPROFILE")
			}
			if val, ok := oldVars["APPDATA"]; ok {
				os.Setenv("APPDATA", val)
			} else {
				os.Unsetenv("APPDATA")
			}
			if val, ok := oldVars["LOCALAPPDATA"]; ok {
				os.Setenv("LOCALAPPDATA", val)
			} else {
				os.Unsetenv("LOCALAPPDATA")
			}
		} else {
			if val, ok := oldVars["HOME"]; ok {
				os.Setenv("HOME", val)
			} else {
				os.Unsetenv("HOME")
			}
		}
	}, nil
}

func (th *TestHelper) CreateSampleConfig() {
	config := map[string]interface{}{
		"theme": "default",
		"provider": map[string]interface{}{
			"anthropic": map[string]interface{}{
				"api": "https://api.anthropic.com",
				"options": map[string]interface{}{
					"apiKey":  "test-key",
					"timeout": 30000,
				},
			},
		},
		"agent": map[string]interface{}{
			"default": map[string]interface{}{
				"provider": "anthropic",
				"model":    "claude-4-sonnet",
			},
		},
	}

	data, _ := json.MarshalIndent(config, "", "  ")
	activeConfigPath := filepath.Join(th.ConfigDir, "opencode.json")
	os.WriteFile(activeConfigPath, data, 0644)
}

func (th *TestHelper) CreateManagerWithTempDir() *context.Manager {
	// Setup environment for cross-platform testing
	cleanup, err := th.SetupEnvironment()
	if err != nil {
		panic("SetupEnvironment failed: " + err.Error())
	}

	manager, err := context.NewManager(false)
	if err != nil {
		// Restore environment and panic
		cleanup()
		panic("Failed to create manager: " + err.Error())
	}

	// Restore environment
	cleanup()
	return manager
}

func TestContextFormat_String(t *testing.T) {
	tests := []struct {
		format   context.ContextFormat
		expected string
	}{
		{context.FormatJSON, "json"},
		{context.FormatJSONC, "jsonc"},
	}

	for _, tt := range tests {
		if got := tt.format.String(); got != tt.expected {
			t.Errorf("ContextFormat.String() = %v, want %v", got, tt.expected)
		}
	}
}

func TestContextFormat_FileExtension(t *testing.T) {
	tests := []struct {
		format   context.ContextFormat
		expected string
	}{
		{context.FormatJSON, ".json"},
		{context.FormatJSONC, ".jsonc"},
	}

	for _, tt := range tests {
		if got := tt.format.FileExtension(); got != tt.expected {
			t.Errorf("ContextFormat.FileExtension() = %v, want %v", got, tt.expected)
		}
	}
}

func TestContextFormat_DisplayName(t *testing.T) {
	tests := []struct {
		format   context.ContextFormat
		expected string
	}{
		{context.FormatJSON, "JSON"},
		{context.FormatJSONC, "JSONC"},
	}

	for _, tt := range tests {
		if got := tt.format.DisplayName(); got != tt.expected {
			t.Errorf("ContextFormat.DisplayName() = %v, want %v", got, tt.expected)
		}
	}
}

func TestParseFormat(t *testing.T) {
	tests := []struct {
		input       string
		expected    context.ContextFormat
		expectError bool
	}{
		{"json", context.FormatJSON, false},
		{"jsonc", context.FormatJSONC, false},
		{"yaml", context.FormatJSON, true},
		{"", context.FormatJSON, true},
		{"XML", context.FormatJSON, true},
	}

	for _, tt := range tests {
		got, err := context.ParseFormat(tt.input)
		if tt.expectError {
			if err == nil {
				t.Errorf("ParseFormat(%q) expected error, got nil", tt.input)
			}
		} else {
			if err != nil {
				t.Errorf("ParseFormat(%q) unexpected error: %v", tt.input, err)
			}
			if got != tt.expected {
				t.Errorf("ParseFormat(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		}
	}
}

func TestManager_CreateContext_WithMockedPaths(t *testing.T) {
	th := NewTestHelper(t)
	defer th.Cleanup()

	th.CreateSampleConfig()

	// Setup environment for cross-platform testing
	cleanup, err := th.SetupEnvironment()
	if err != nil {
		t.Fatalf("SetupEnvironment failed: %v", err)
	}
	defer cleanup()

	manager, err := context.NewManager(false)
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}

	// Test creating JSON context
	err = manager.CreateContext("test-json")
	if err != nil {
		t.Fatalf("CreateContext failed: %v", err)
	}

	// Verify file exists
	contextPath := filepath.Join(th.SettingsDir, "test-json.json")
	if _, err := os.Stat(contextPath); os.IsNotExist(err) {
		t.Error("Context file was not created")
	}

	// Verify content
	data, err := os.ReadFile(contextPath)
	if err != nil {
		t.Fatalf("Failed to read context file: %v", err)
	}

	var config map[string]interface{}
	if err := json.Unmarshal(data, &config); err != nil {
		t.Fatalf("Context file contains invalid JSON: %v", err)
	}

	if config["theme"] != "default" {
		t.Error("Context file missing expected content")
	}
}

func TestManager_CreateContextWithFormat_WithMockedPaths(t *testing.T) {
	th := NewTestHelper(t)
	defer th.Cleanup()

	th.CreateSampleConfig()

	// Setup environment for cross-platform testing
	cleanup, err := th.SetupEnvironment()
	if err != nil {
		t.Fatalf("SetupEnvironment failed: %v", err)
	}
	defer cleanup()

	manager, err := context.NewManager(false)
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}

	// Test creating JSONC context
	err = manager.CreateContextWithFormat("test-jsonc", context.FormatJSONC)
	if err != nil {
		t.Fatalf("CreateContextWithFormat failed: %v", err)
	}

	// Verify file exists with correct extension
	contextPath := filepath.Join(th.SettingsDir, "test-jsonc.jsonc")
	if _, err := os.Stat(contextPath); os.IsNotExist(err) {
		t.Error("JSONC context file was not created")
	}

	// Verify content has comments
	data, err := os.ReadFile(contextPath)
	if err != nil {
		t.Fatalf("Failed to read context file: %v", err)
	}

	content := string(data)
	if !containsString(content, "// opencode context: test-jsonc") {
		t.Error("JSONC context file missing expected comment header")
	}
	if !containsString(content, "// Format: JSONC") {
		t.Error("JSONC context file missing format comment")
	}
}

func TestManager_ListContexts_WithMockedPaths(t *testing.T) {
	th := NewTestHelper(t)
	defer th.Cleanup()

	th.CreateSampleConfig()

	// Setup environment for cross-platform testing
	cleanup, err := th.SetupEnvironment()
	if err != nil {
		t.Fatalf("SetupEnvironment failed: %v", err)
	}
	defer cleanup()

	manager, err := context.NewManager(false)
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}

	// Create test contexts
	manager.CreateContext("context1")
	manager.CreateContextWithFormat("context2", context.FormatJSONC)

	contexts, err := manager.ListContexts()
	if err != nil {
		t.Fatalf("ListContexts failed: %v", err)
	}

	if len(contexts) != 2 {
		t.Errorf("Expected 2 contexts, got %d", len(contexts))
	}

	// Check context names
	names := make(map[string]bool)
	for _, ctx := range contexts {
		names[ctx.Name] = true
	}

	if !names["context1"] || !names["context2"] {
		t.Error("Missing expected context names")
	}
}

func TestManager_GetContext_WithMockedPaths(t *testing.T) {
	th := NewTestHelper(t)
	defer th.Cleanup()

	th.CreateSampleConfig()

	// Setup environment for cross-platform testing
	cleanup, err := th.SetupEnvironment()
	if err != nil {
		t.Fatalf("SetupEnvironment failed: %v", err)
	}
	defer cleanup()

	manager, err := context.NewManager(false)
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}

	// Create test context
	manager.CreateContext("test-context")

	// Get context
	ctx, err := manager.GetContext("test-context")
	if err != nil {
		t.Fatalf("GetContext failed: %v", err)
	}

	if ctx.Name != "test-context" {
		t.Errorf("Expected context name 'test-context', got '%s'", ctx.Name)
	}

	if ctx.Data["theme"] != "default" {
		t.Error("Context data missing expected content")
	}
}

func TestManager_GetContext_JSONC_WithMockedPaths(t *testing.T) {
	th := NewTestHelper(t)
	defer th.Cleanup()

	th.CreateSampleConfig()

	// Setup environment for cross-platform testing
	cleanup, err := th.SetupEnvironment()
	if err != nil {
		t.Fatalf("SetupEnvironment failed: %v", err)
	}
	defer cleanup()

	manager, err := context.NewManager(false)
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}

	// Create JSONC context
	manager.CreateContextWithFormat("test-jsonc", context.FormatJSONC)

	// Get context (should strip comments)
	ctx, err := manager.GetContext("test-jsonc")
	if err != nil {
		t.Fatalf("GetContext failed: %v", err)
	}

	if ctx.Name != "test-jsonc" {
		t.Errorf("Expected context name 'test-jsonc', got '%s'", ctx.Name)
	}

	if ctx.Data["theme"] != "default" {
		t.Error("JSONC context data missing expected content after comment stripping")
	}
}

func TestManager_DeleteContext_WithMockedPaths(t *testing.T) {
	th := NewTestHelper(t)
	defer th.Cleanup()

	th.CreateSampleConfig()

	// Setup environment for cross-platform testing
	cleanup, err := th.SetupEnvironment()
	if err != nil {
		t.Fatalf("SetupEnvironment failed: %v", err)
	}
	defer cleanup()

	manager, err := context.NewManager(false)
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}

	// Create and delete context
	manager.CreateContext("to-delete")

	err = manager.DeleteContext("to-delete")
	if err != nil {
		t.Fatalf("DeleteContext failed: %v", err)
	}

	// Verify file is gone
	contextPath := filepath.Join(th.SettingsDir, "to-delete.json")
	if _, err := os.Stat(contextPath); !os.IsNotExist(err) {
		t.Error("Context file was not deleted")
	}
}

func TestManager_RenameContext_WithMockedPaths(t *testing.T) {
	th := NewTestHelper(t)
	defer th.Cleanup()

	th.CreateSampleConfig()

	// Setup environment for cross-platform testing
	cleanup, err := th.SetupEnvironment()
	if err != nil {
		t.Fatalf("SetupEnvironment failed: %v", err)
	}
	defer cleanup()

	manager, err := context.NewManager(false)
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}

	// Create and rename context
	manager.CreateContext("old-name")

	err = manager.RenameContext("old-name", "new-name")
	if err != nil {
		t.Fatalf("RenameContext failed: %v", err)
	}

	// Verify old file is gone
	oldPath := filepath.Join(th.SettingsDir, "old-name.json")
	if _, err := os.Stat(oldPath); !os.IsNotExist(err) {
		t.Error("Old context file was not removed")
	}

	// Verify new file exists
	newPath := filepath.Join(th.SettingsDir, "new-name.json")
	if _, err := os.Stat(newPath); os.IsNotExist(err) {
		t.Error("New context file was not created")
	}
}

// Helper function
func containsString(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			containsString(s[1:], substr) ||
			(len(s) > 0 && s[:len(substr)] == substr))
}
