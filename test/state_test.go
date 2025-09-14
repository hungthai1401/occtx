package test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/hungthai1401/occtx/internal/context"
)

func TestLoadState_EmptyFile(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "occtx-state-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	stateFile := filepath.Join(tempDir, "state.json")

	// Test loading non-existent file
	state, err := context.LoadState(stateFile)
	if err != nil {
		t.Fatalf("LoadState failed: %v", err)
	}

	if state.Current != "" || state.Previous != "" {
		t.Error("Expected empty state for non-existent file")
	}
}

func TestLoadState_ValidFile(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "occtx-state-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	stateFile := filepath.Join(tempDir, "state.json")

	// Create valid state file
	stateData := context.State{
		Current:  "current-context",
		Previous: "previous-context",
	}

	data, _ := json.MarshalIndent(stateData, "", "  ")
	if err := os.WriteFile(stateFile, data, 0644); err != nil {
		t.Fatal(err)
	}

	// Load state
	state, err := context.LoadState(stateFile)
	if err != nil {
		t.Fatalf("LoadState failed: %v", err)
	}

	if state.Current != "current-context" {
		t.Errorf("Expected current 'current-context', got '%s'", state.Current)
	}
	if state.Previous != "previous-context" {
		t.Errorf("Expected previous 'previous-context', got '%s'", state.Previous)
	}
}

func TestLoadState_InvalidJSON(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "occtx-state-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	stateFile := filepath.Join(tempDir, "state.json")

	// Create invalid JSON file
	if err := os.WriteFile(stateFile, []byte("invalid json"), 0644); err != nil {
		t.Fatal(err)
	}

	// Should return empty state instead of error
	state, err := context.LoadState(stateFile)
	if err != nil {
		t.Fatalf("LoadState should handle invalid JSON gracefully: %v", err)
	}

	if state.Current != "" || state.Previous != "" {
		t.Error("Expected empty state for invalid JSON")
	}
}

func TestSaveState(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "occtx-state-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	stateFile := filepath.Join(tempDir, "subdir", "state.json")

	state := &context.State{
		Current:  "test-current",
		Previous: "test-previous",
	}

	// Save state (should create directory)
	err = state.SaveState(stateFile)
	if err != nil {
		t.Fatalf("SaveState failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(stateFile); os.IsNotExist(err) {
		t.Error("State file was not created")
	}

	// Verify content
	data, err := os.ReadFile(stateFile)
	if err != nil {
		t.Fatalf("Failed to read state file: %v", err)
	}

	var savedState context.State
	if err := json.Unmarshal(data, &savedState); err != nil {
		t.Fatalf("State file contains invalid JSON: %v", err)
	}

	if savedState.Current != "test-current" {
		t.Errorf("Expected current 'test-current', got '%s'", savedState.Current)
	}
	if savedState.Previous != "test-previous" {
		t.Errorf("Expected previous 'test-previous', got '%s'", savedState.Previous)
	}
}

func TestState_SetCurrent(t *testing.T) {
	state := &context.State{
		Current:  "old-current",
		Previous: "old-previous",
	}

	state.SetCurrent("new-current")

	if state.Current != "new-current" {
		t.Errorf("Expected current 'new-current', got '%s'", state.Current)
	}
	if state.Previous != "old-current" {
		t.Errorf("Expected previous 'old-current', got '%s'", state.Previous)
	}
}

func TestState_Unset(t *testing.T) {
	state := &context.State{
		Current:  "current-context",
		Previous: "previous-context",
	}

	state.Unset()

	if state.Current != "" {
		t.Errorf("Expected empty current, got '%s'", state.Current)
	}
	if state.Previous != "current-context" {
		t.Errorf("Expected previous 'current-context', got '%s'", state.Previous)
	}
}

func TestState_SwitchToPrevious(t *testing.T) {
	tests := []struct {
		name            string
		initialCurrent  string
		initialPrevious string
		expectedResult  bool
		finalCurrent    string
		finalPrevious   string
	}{
		{
			name:            "successful switch",
			initialCurrent:  "context1",
			initialPrevious: "context2",
			expectedResult:  true,
			finalCurrent:    "context2",
			finalPrevious:   "context1",
		},
		{
			name:            "no previous context",
			initialCurrent:  "context1",
			initialPrevious: "",
			expectedResult:  false,
			finalCurrent:    "context1",
			finalPrevious:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := &context.State{
				Current:  tt.initialCurrent,
				Previous: tt.initialPrevious,
			}

			result := state.SwitchToPrevious()

			if result != tt.expectedResult {
				t.Errorf("Expected result %v, got %v", tt.expectedResult, result)
			}
			if state.Current != tt.finalCurrent {
				t.Errorf("Expected final current '%s', got '%s'", tt.finalCurrent, state.Current)
			}
			if state.Previous != tt.finalPrevious {
				t.Errorf("Expected final previous '%s', got '%s'", tt.finalPrevious, state.Previous)
			}
		})
	}
}

func TestState_AtomicSave(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "occtx-atomic-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	stateFile := filepath.Join(tempDir, "state.json")

	// Create initial state
	initialState := &context.State{Current: "initial"}
	if err := initialState.SaveState(stateFile); err != nil {
		t.Fatal(err)
	}

	// Verify atomic operation by checking temp file doesn't exist after save
	tempFile := stateFile + ".tmp"
	if _, err := os.Stat(tempFile); !os.IsNotExist(err) {
		t.Error("Temporary file should not exist after successful save")
	}

	// Verify final file exists and has correct content
	if _, err := os.Stat(stateFile); os.IsNotExist(err) {
		t.Error("State file should exist after save")
	}

	// Load and verify content
	loadedState, err := context.LoadState(stateFile)
	if err != nil {
		t.Fatalf("Failed to load saved state: %v", err)
	}

	if loadedState.Current != "initial" {
		t.Errorf("Expected 'initial', got '%s'", loadedState.Current)
	}
}
