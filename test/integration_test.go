package test

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// IntegrationTestHelper provides utilities for integration testing
type IntegrationTestHelper struct {
	TempDir     string
	ConfigDir   string
	SettingsDir string
	BinaryPath  string
	t           *testing.T
}

func NewIntegrationTestHelper(t *testing.T) *IntegrationTestHelper {
	tempDir, err := os.MkdirTemp("", "occtx-integration-*")
	if err != nil {
		t.Fatal(err)
	}

	configDir := filepath.Join(tempDir, ".config", "opencode")
	settingsDir := filepath.Join(configDir, "settings")

	// Create directories
	if err := os.MkdirAll(settingsDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Build binary for testing
	binaryName := "occtx"
	if runtime.GOOS == "windows" {
		binaryName = "occtx.exe"
	}
	binaryPath := filepath.Join(tempDir, binaryName)
	buildCmd := exec.Command("go", "build", "-o", binaryPath, "../.")
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build binary: %v", err)
	}

	return &IntegrationTestHelper{
		TempDir:     tempDir,
		ConfigDir:   configDir,
		SettingsDir: settingsDir,
		BinaryPath:  binaryPath,
		t:           t,
	}
}

func (ith *IntegrationTestHelper) Cleanup() {
	os.RemoveAll(ith.TempDir)
}

func (ith *IntegrationTestHelper) CreateSampleConfig() {
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
	activeConfigPath := filepath.Join(ith.ConfigDir, "opencode.json")
	if err := os.WriteFile(activeConfigPath, data, 0644); err != nil {
		ith.t.Fatal(err)
	}
}

func (ith *IntegrationTestHelper) RunCommand(args ...string) (string, string, error) {
	cmd := exec.Command(ith.BinaryPath, args...)

	// Set HOME to temp directory so it uses our test config
	cmd.Env = append(os.Environ(), "HOME="+ith.TempDir)

	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	return stdout.String(), stderr.String(), err
}

func TestIntegration_BasicWorkflow(t *testing.T) {
	// Skip integration tests on Windows due to path and binary execution complexities
	if runtime.GOOS == "windows" {
		t.Skip("Integration tests skipped on Windows")
	}

	ith := NewIntegrationTestHelper(t)
	defer ith.Cleanup()

	ith.CreateSampleConfig()

	// Test listing empty contexts
	stdout, _, err := ith.RunCommand()
	if err != nil {
		t.Fatalf("List command failed: %v", err)
	}
	if !strings.Contains(stdout, "No global contexts found") {
		t.Error("Expected 'No global contexts found' message")
	}

	// Test creating context
	stdout, _, err = ith.RunCommand("-n", "development")
	if err != nil {
		t.Fatalf("Create command failed: %v", err)
	}
	if !strings.Contains(stdout, "Context 'development' created successfully") {
		t.Error("Expected success message for context creation")
	}

	// Test listing contexts with one context
	stdout, _, err = ith.RunCommand()
	if err != nil {
		t.Fatalf("List command failed: %v", err)
	}
	if !strings.Contains(stdout, "development") {
		t.Error("Expected 'development' context in list")
	}

	// Test switching to context
	stdout, _, err = ith.RunCommand("development")
	if err != nil {
		t.Fatalf("Switch command failed: %v", err)
	}
	if !strings.Contains(stdout, "Switched to context: development") {
		t.Error("Expected switch success message")
	}

	// Test showing current context
	stdout, _, err = ith.RunCommand("-c")
	if err != nil {
		t.Fatalf("Current command failed: %v", err)
	}
	if strings.TrimSpace(stdout) != "development" {
		t.Errorf("Expected current context 'development', got '%s'", strings.TrimSpace(stdout))
	}
}

func TestIntegration_FormatSupport(t *testing.T) {
	// Skip integration tests on Windows due to path and binary execution complexities
	if runtime.GOOS == "windows" {
		t.Skip("Integration tests skipped on Windows")
	}

	ith := NewIntegrationTestHelper(t)
	defer ith.Cleanup()

	ith.CreateSampleConfig()

	// Test creating JSON context
	stdout, _, err := ith.RunCommand("-n", "json-context", "-f", "json")
	if err != nil {
		t.Fatalf("Create JSON context failed: %v", err)
	}
	if !strings.Contains(stdout, "JSON format") {
		t.Error("Expected JSON format message")
	}

	// Test creating JSONC context
	stdout, _, err = ith.RunCommand("-n", "jsonc-context", "-f", "jsonc")
	if err != nil {
		t.Fatalf("Create JSONC context failed: %v", err)
	}
	if !strings.Contains(stdout, "JSONC format") {
		t.Error("Expected JSONC format message")
	}

	// Test invalid format
	_, stderr, err := ith.RunCommand("-n", "invalid-context", "-f", "yaml")
	if err == nil {
		t.Error("Expected error for invalid format")
	}
	if !strings.Contains(stderr, "invalid format 'yaml'") {
		t.Error("Expected invalid format error message")
	}

	// Verify JSONC file has comments
	jsoncPath := filepath.Join(ith.SettingsDir, "jsonc-context.jsonc")
	content, err := os.ReadFile(jsoncPath)
	if err != nil {
		t.Fatalf("Failed to read JSONC file: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "// opencode context: jsonc-context") {
		t.Error("JSONC file should contain context comment")
	}
	if !strings.Contains(contentStr, "// Format: JSONC") {
		t.Error("JSONC file should contain format comment")
	}
}

func TestIntegration_ContextManagement(t *testing.T) {
	// Skip integration tests on Windows due to path and binary execution complexities
	if runtime.GOOS == "windows" {
		t.Skip("Integration tests skipped on Windows")
	}

	ith := NewIntegrationTestHelper(t)
	defer ith.Cleanup()

	ith.CreateSampleConfig()

	// Create multiple contexts
	contexts := []string{"dev", "staging", "prod"}
	for _, ctx := range contexts {
		_, _, err := ith.RunCommand("-n", ctx)
		if err != nil {
			t.Fatalf("Failed to create context %s: %v", ctx, err)
		}
	}

	// Test listing multiple contexts
	stdout, _, err := ith.RunCommand()
	if err != nil {
		t.Fatalf("List command failed: %v", err)
	}
	for _, ctx := range contexts {
		if !strings.Contains(stdout, ctx) {
			t.Errorf("Expected context '%s' in list", ctx)
		}
	}

	// Test switching between contexts
	_, _, err = ith.RunCommand("staging")
	if err != nil {
		t.Fatalf("Failed to switch to staging: %v", err)
	}

	_, _, err = ith.RunCommand("prod")
	if err != nil {
		t.Fatalf("Failed to switch to prod: %v", err)
	}

	// Test switching to previous
	stdout, _, err = ith.RunCommand("-")
	if err != nil {
		t.Fatalf("Failed to switch to previous: %v", err)
	}
	if !strings.Contains(stdout, "staging") {
		t.Error("Expected to switch back to staging")
	}

	// Test renaming context
	_, _, err = ith.RunCommand("-r", "dev", "development")
	if err != nil {
		t.Fatalf("Failed to rename context: %v", err)
	}

	// Verify rename worked
	stdout, _, err = ith.RunCommand()
	if err != nil {
		t.Fatalf("List command failed: %v", err)
	}
	t.Logf("List output after rename: %s", stdout)
	// Check for exact word "dev" (not part of "development")
	lines := strings.Split(stdout, "\n")
	foundOldName := false
	foundNewName := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		// Remove the * and whitespace for current context
		trimmed = strings.TrimSpace(strings.TrimPrefix(trimmed, "*"))
		if trimmed == "dev" {
			foundOldName = true
		}
		if trimmed == "development" {
			foundNewName = true
		}
	}
	if foundOldName {
		t.Error("Old context name should not exist after rename")
	}
	if !foundNewName {
		t.Error("New context name should exist after rename")
	}

	// Test deleting context (should fail if current)
	_, _, err = ith.RunCommand("-d", "staging")
	if err == nil {
		t.Error("Should not be able to delete current context")
	}

	// Switch and then delete
	_, _, err = ith.RunCommand("development")
	if err != nil {
		t.Fatalf("Failed to switch context: %v", err)
	}

	_, _, err = ith.RunCommand("-d", "staging")
	if err != nil {
		t.Fatalf("Failed to delete context: %v", err)
	}

	// Verify deletion
	stdout, _, err = ith.RunCommand()
	if err != nil {
		t.Fatalf("List command failed: %v", err)
	}
	if strings.Contains(stdout, "staging") {
		t.Error("Deleted context should not appear in list")
	}
}

func TestIntegration_ShowAndExport(t *testing.T) {
	// Skip integration tests on Windows due to path and binary execution complexities
	if runtime.GOOS == "windows" {
		t.Skip("Integration tests skipped on Windows")
	}

	ith := NewIntegrationTestHelper(t)
	defer ith.Cleanup()

	ith.CreateSampleConfig()

	// Create context
	_, _, err := ith.RunCommand("-n", "test-context")
	if err != nil {
		t.Fatalf("Failed to create context: %v", err)
	}

	// Test show context
	stdout, _, err := ith.RunCommand("-s", "test-context")
	if err != nil {
		t.Fatalf("Show command failed: %v", err)
	}

	// Verify JSON content
	var config map[string]interface{}
	if err := json.Unmarshal([]byte(stdout), &config); err != nil {
		t.Fatalf("Show output is not valid JSON: %v", err)
	}

	if config["theme"] != "default" {
		t.Error("Expected theme 'default' in show output")
	}

	// Test export context
	stdout, _, err = ith.RunCommand("--export", "test-context")
	if err != nil {
		t.Fatalf("Export command failed: %v", err)
	}

	// Verify export output is valid JSON
	if err := json.Unmarshal([]byte(stdout), &config); err != nil {
		t.Fatalf("Export output is not valid JSON: %v", err)
	}
}

func TestIntegration_ErrorHandling(t *testing.T) {
	// Skip integration tests on Windows due to path and binary execution complexities
	if runtime.GOOS == "windows" {
		t.Skip("Integration tests skipped on Windows")
	}

	ith := NewIntegrationTestHelper(t)
	defer ith.Cleanup()

	// Test without config file
	_, stderr, err := ith.RunCommand("-n", "test")
	if err == nil {
		t.Error("Should fail when no config file exists")
	}
	if !strings.Contains(stderr, "no active opencode.json found") {
		t.Error("Expected missing config error message")
	}

	ith.CreateSampleConfig()

	// Test invalid context name
	_, stderr, err = ith.RunCommand("-n", ".invalid")
	if err == nil {
		t.Error("Should fail with invalid context name")
	}

	// Test switching to non-existent context
	_, stderr, err = ith.RunCommand("non-existent")
	if err == nil {
		t.Error("Should fail when switching to non-existent context")
	}

	// Test deleting non-existent context
	_, stderr, err = ith.RunCommand("-d", "non-existent")
	if err == nil {
		t.Error("Should fail when deleting non-existent context")
	}
}

func TestIntegration_StateManagement(t *testing.T) {
	// Skip integration tests on Windows due to path and binary execution complexities
	if runtime.GOOS == "windows" {
		t.Skip("Integration tests skipped on Windows")
	}

	ith := NewIntegrationTestHelper(t)
	defer ith.Cleanup()

	ith.CreateSampleConfig()

	// Create contexts
	_, _, err := ith.RunCommand("-n", "context1")
	if err != nil {
		t.Fatalf("Failed to create context1: %v", err)
	}
	_, _, err = ith.RunCommand("-n", "context2")
	if err != nil {
		t.Fatalf("Failed to create context2: %v", err)
	}

	// Switch contexts and test state persistence
	_, _, err = ith.RunCommand("context1")
	if err != nil {
		t.Fatalf("Failed to switch to context1: %v", err)
	}

	_, _, err = ith.RunCommand("context2")
	if err != nil {
		t.Fatalf("Failed to switch to context2: %v", err)
	}

	// Test previous context switch
	stdout, _, err := ith.RunCommand("-")
	if err != nil {
		t.Fatalf("Failed to switch to previous: %v", err)
	}
	if !strings.Contains(stdout, "context1") {
		t.Error("Expected to switch back to context1")
	}

	// Test unset
	_, _, err = ith.RunCommand("-u")
	if err != nil {
		t.Fatalf("Failed to unset context: %v", err)
	}

	// Verify no current context
	stdout, _, err = ith.RunCommand("-c")
	if err != nil {
		t.Fatalf("Current command failed: %v", err)
	}
	if !strings.Contains(stdout, "No current context set") {
		t.Error("Expected no current context message")
	}
}
