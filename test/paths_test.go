package test

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/hungthai1401/occtx/internal/config"
)

func TestNewPaths(t *testing.T) {
	paths, err := config.NewPaths()
	if err != nil {
		t.Fatalf("NewPaths failed: %v", err)
	}

	// Check that paths are not empty
	if paths.GlobalConfigDir == "" {
		t.Error("GlobalConfigDir should not be empty")
	}
	if paths.GlobalSettingsDir == "" {
		t.Error("GlobalSettingsDir should not be empty")
	}
	if paths.GlobalActiveConfig == "" {
		t.Error("GlobalActiveConfig should not be empty")
	}
	if paths.GlobalStateFile == "" {
		t.Error("GlobalStateFile should not be empty")
	}

	// Check path relationships - use constants from config package
	expectedSettingsDir := filepath.Join(paths.GlobalConfigDir, "settings")
	if paths.GlobalSettingsDir != expectedSettingsDir {
		t.Errorf("Expected settings dir %s, got %s", expectedSettingsDir, paths.GlobalSettingsDir)
	}

	expectedActiveConfig := filepath.Join(paths.GlobalConfigDir, "opencode.json")
	if paths.GlobalActiveConfig != expectedActiveConfig {
		t.Errorf("Expected active config %s, got %s", expectedActiveConfig, paths.GlobalActiveConfig)
	}

	expectedStateFile := filepath.Join(paths.GlobalSettingsDir, ".occtx-state.json")
	if paths.GlobalStateFile != expectedStateFile {
		t.Errorf("Expected state file %s, got %s", expectedStateFile, paths.GlobalStateFile)
	}
}

func TestPaths_GetContextsDir(t *testing.T) {
	paths, err := config.NewPaths()
	if err != nil {
		t.Fatal(err)
	}

	// Test global level
	globalDir := paths.GetContextsDir(false)
	if globalDir != paths.GlobalSettingsDir {
		t.Errorf("Expected global settings dir %s, got %s", paths.GlobalSettingsDir, globalDir)
	}

	// Test project level
	projectDir := paths.GetContextsDir(true)
	if projectDir != paths.ProjectSettingsDir {
		t.Errorf("Expected project settings dir %s, got %s", paths.ProjectSettingsDir, projectDir)
	}
}

func TestPaths_GetActiveConfigPath(t *testing.T) {
	paths, err := config.NewPaths()
	if err != nil {
		t.Fatal(err)
	}

	// Test global level
	globalConfig := paths.GetActiveConfigPath(false)
	if globalConfig != paths.GlobalActiveConfig {
		t.Errorf("Expected global active config %s, got %s", paths.GlobalActiveConfig, globalConfig)
	}

	// Test project level
	projectConfig := paths.GetActiveConfigPath(true)
	if projectConfig != paths.ProjectActiveConfig {
		t.Errorf("Expected project active config %s, got %s", paths.ProjectActiveConfig, projectConfig)
	}
}

func TestPaths_GetStateFilePath(t *testing.T) {
	paths, err := config.NewPaths()
	if err != nil {
		t.Fatal(err)
	}

	// Test global level
	globalState := paths.GetStateFilePath(false)
	if globalState != paths.GlobalStateFile {
		t.Errorf("Expected global state file %s, got %s", paths.GlobalStateFile, globalState)
	}

	// Test project level
	projectState := paths.GetStateFilePath(true)
	if projectState != paths.ProjectStateFile {
		t.Errorf("Expected project state file %s, got %s", paths.ProjectStateFile, projectState)
	}
}

func TestPaths_EnsureDirectories(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "occtx-paths-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Create paths with temporary directory
	paths := &config.Paths{
		GlobalConfigDir:    filepath.Join(tempDir, "global", "config"),
		GlobalSettingsDir:  filepath.Join(tempDir, "global", "config", "settings"),
		ProjectConfigDir:   filepath.Join(tempDir, "project", "config"),
		ProjectSettingsDir: filepath.Join(tempDir, "project", "config", "settings"),
	}

	// Test global directories
	err = paths.EnsureDirectories(false)
	if err != nil {
		t.Fatalf("EnsureDirectories(false) failed: %v", err)
	}

	// Verify global directories exist
	if _, err := os.Stat(paths.GlobalConfigDir); os.IsNotExist(err) {
		t.Error("Global config directory was not created")
	}
	if _, err := os.Stat(paths.GlobalSettingsDir); os.IsNotExist(err) {
		t.Error("Global settings directory was not created")
	}

	// Test project directories
	err = paths.EnsureDirectories(true)
	if err != nil {
		t.Fatalf("EnsureDirectories(true) failed: %v", err)
	}

	// Verify project directories exist
	if _, err := os.Stat(paths.ProjectConfigDir); os.IsNotExist(err) {
		t.Error("Project config directory was not created")
	}
	if _, err := os.Stat(paths.ProjectSettingsDir); os.IsNotExist(err) {
		t.Error("Project settings directory was not created")
	}
}

func TestPaths_ProjectContextsExist(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "occtx-project-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	projectSettingsDir := filepath.Join(tempDir, "settings")
	paths := &config.Paths{
		ProjectSettingsDir: projectSettingsDir,
	}

	// Test when directory doesn't exist
	if paths.ProjectContextsExist() {
		t.Error("ProjectContextsExist should return false when directory doesn't exist")
	}

	// Create directory but no JSON files
	if err := os.MkdirAll(projectSettingsDir, 0755); err != nil {
		t.Fatal(err)
	}

	if paths.ProjectContextsExist() {
		t.Error("ProjectContextsExist should return false when no JSON files exist")
	}

	// Create state file (should be ignored)
	stateFile := filepath.Join(projectSettingsDir, ".occtx-state.json")
	if err := os.WriteFile(stateFile, []byte("{}"), 0644); err != nil {
		t.Fatal(err)
	}

	if paths.ProjectContextsExist() {
		t.Error("ProjectContextsExist should return false when only state file exists")
	}

	// Create actual context file
	contextFile := filepath.Join(projectSettingsDir, "test-context.json")
	if err := os.WriteFile(contextFile, []byte("{}"), 0644); err != nil {
		t.Fatal(err)
	}

	if !paths.ProjectContextsExist() {
		t.Error("ProjectContextsExist should return true when JSON context files exist")
	}

	// Test with JSONC file
	jsoncFile := filepath.Join(projectSettingsDir, "test-context.jsonc")
	if err := os.WriteFile(jsoncFile, []byte("{}"), 0644); err != nil {
		t.Fatal(err)
	}

	if !paths.ProjectContextsExist() {
		t.Error("ProjectContextsExist should return true when JSONC context files exist")
	}
}

func TestPaths_DirectoryPermissions(t *testing.T) {
	// Skip permission tests on Windows as it has different permission model
	if runtime.GOOS == "windows" {
		t.Skip("Permission tests not applicable on Windows")
	}

	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "occtx-perms-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	paths := &config.Paths{
		GlobalConfigDir:   filepath.Join(tempDir, "config"),
		GlobalSettingsDir: filepath.Join(tempDir, "config", "settings"),
	}

	// Create directories
	err = paths.EnsureDirectories(false)
	if err != nil {
		t.Fatalf("EnsureDirectories failed: %v", err)
	}

	// Check permissions
	configInfo, err := os.Stat(paths.GlobalConfigDir)
	if err != nil {
		t.Fatal(err)
	}

	expectedPerms := os.FileMode(0755)
	if configInfo.Mode().Perm() != expectedPerms {
		t.Errorf("Expected permissions %v, got %v", expectedPerms, configInfo.Mode().Perm())
	}

	settingsInfo, err := os.Stat(paths.GlobalSettingsDir)
	if err != nil {
		t.Fatal(err)
	}

	if settingsInfo.Mode().Perm() != expectedPerms {
		t.Errorf("Expected permissions %v, got %v", expectedPerms, settingsInfo.Mode().Perm())
	}
}
