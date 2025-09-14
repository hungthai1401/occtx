package test

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hungthai1401/occtx/internal/context"
)

func BenchmarkContextCreation(b *testing.B) {
	// Setup
	tempDir, err := os.MkdirTemp("", "occtx-bench-*")
	if err != nil {
		b.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	bh := setupBenchmarkHelper(b, tempDir)
	manager := bh.CreateManagerWithTempDir()

	b.ResetTimer()

	// Benchmark context creation
	for i := 0; i < b.N; i++ {
		contextName := fmt.Sprintf("bench-context-%d", i)
		if err := manager.CreateContext(contextName); err != nil {
			b.Fatalf("CreateContext failed: %v", err)
		}
	}
}

func BenchmarkContextSwitch(b *testing.B) {
	// Setup
	tempDir, err := os.MkdirTemp("", "occtx-bench-*")
	if err != nil {
		b.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	bh := setupBenchmarkHelper(b, tempDir)
	manager := bh.CreateManagerWithTempDir()

	// Create test contexts
	contexts := []string{"bench1", "bench2", "bench3", "bench4", "bench5"}
	for _, ctx := range contexts {
		if err := manager.CreateContext(ctx); err != nil {
			b.Fatalf("Failed to create context %s: %v", ctx, err)
		}
	}

	b.ResetTimer()

	// Benchmark context switching
	for i := 0; i < b.N; i++ {
		contextName := contexts[i%len(contexts)]
		if err := manager.SwitchToContext(contextName); err != nil {
			b.Fatalf("SwitchToContext failed: %v", err)
		}
	}
}

func BenchmarkContextList(b *testing.B) {
	// Setup with many contexts
	tempDir, err := os.MkdirTemp("", "occtx-bench-*")
	if err != nil {
		b.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	bh := setupBenchmarkHelper(b, tempDir)
	manager := bh.CreateManagerWithTempDir()

	// Create 100 contexts
	for i := 0; i < 100; i++ {
		contextName := fmt.Sprintf("context-%03d", i)
		if err := manager.CreateContext(contextName); err != nil {
			b.Fatalf("Failed to create context: %v", err)
		}
	}

	b.ResetTimer()

	// Benchmark listing contexts
	for i := 0; i < b.N; i++ {
		_, err := manager.ListContexts()
		if err != nil {
			b.Fatalf("ListContexts failed: %v", err)
		}
	}
}

func BenchmarkJSONParsing(b *testing.B) {
	// Create sample JSON data
	sampleData := map[string]interface{}{
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

	jsonData, err := json.Marshal(sampleData)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()

	// Benchmark JSON parsing
	for i := 0; i < b.N; i++ {
		var parsed map[string]interface{}
		if err := json.Unmarshal(jsonData, &parsed); err != nil {
			b.Fatalf("JSON parsing failed: %v", err)
		}
	}
}

func BenchmarkJSONCCommentStripping(b *testing.B) {
	// Create sample JSONC data
	jsoncData := `// opencode context: test
// Format: JSONC
// Created: 2025-09-13 15:35:19
{
  "theme": "default",
  // This is a comment
  "provider": {
    "anthropic": {
      "api": "https://api.anthropic.com",
      // Another comment
      "options": {
        "apiKey": "test-key",
        "timeout": 30000
      }
    }
  },
  "agent": {
    "default": {
      "provider": "anthropic",
      "model": "claude-4-sonnet"
    }
  }
}`

	b.ResetTimer()

	// Benchmark comment stripping
	for i := 0; i < b.N; i++ {
		lines := strings.Split(jsoncData, "\n")
		var cleanLines []string
		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			if !strings.HasPrefix(trimmed, "//") {
				cleanLines = append(cleanLines, line)
			}
		}
		cleanData := strings.Join(cleanLines, "\n")

		var parsed map[string]interface{}
		if err := json.Unmarshal([]byte(cleanData), &parsed); err != nil {
			b.Fatalf("JSONC parsing failed: %v", err)
		}
	}
}

func BenchmarkStateOperations(b *testing.B) {
	tempDir, err := os.MkdirTemp("", "occtx-bench-*")
	if err != nil {
		b.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	stateFile := filepath.Join(tempDir, "state.json")

	b.ResetTimer()

	// Benchmark state save/load operations
	for i := 0; i < b.N; i++ {
		state := &context.State{
			Current:  fmt.Sprintf("current-%d", i),
			Previous: fmt.Sprintf("previous-%d", i),
		}

		// Save state
		if err := state.SaveState(stateFile); err != nil {
			b.Fatalf("SaveState failed: %v", err)
		}

		// Load state
		_, err := context.LoadState(stateFile)
		if err != nil {
			b.Fatalf("LoadState failed: %v", err)
		}
	}
}

func BenchmarkFormatParsing(b *testing.B) {
	formats := []string{"json", "jsonc", "yaml", "xml", "invalid"}

	b.ResetTimer()

	// Benchmark format parsing
	for i := 0; i < b.N; i++ {
		format := formats[i%len(formats)]
		_, _ = context.ParseFormat(format)
	}
}

// BenchmarkHelper provides utilities for benchmark testing
type BenchmarkHelper struct {
	TempDir     string
	ConfigDir   string
	SettingsDir string
	StateFile   string
}

func (bh *BenchmarkHelper) CreateSampleConfig() {
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
	activeConfigPath := filepath.Join(bh.ConfigDir, "opencode.json")
	os.WriteFile(activeConfigPath, data, 0644)
}

func (bh *BenchmarkHelper) CreateManagerWithTempDir() *context.Manager {
	// Set HOME environment to our temp directory for testing
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", bh.TempDir)

	manager, err := context.NewManager(false)
	if err != nil {
		// Restore HOME and panic
		os.Setenv("HOME", oldHome)
		panic(fmt.Sprintf("Failed to create manager: %v", err))
	}

	// Restore HOME
	os.Setenv("HOME", oldHome)
	return manager
}

// Benchmark helper setup
func setupBenchmarkHelper(b *testing.B, tempDir string) *BenchmarkHelper {
	configDir := filepath.Join(tempDir, ".config", "opencode")
	settingsDir := filepath.Join(configDir, "settings")
	stateFile := filepath.Join(settingsDir, ".occtx-state.json")

	// Create directories
	if err := os.MkdirAll(settingsDir, 0755); err != nil {
		b.Fatal(err)
	}

	bh := &BenchmarkHelper{
		TempDir:     tempDir,
		ConfigDir:   configDir,
		SettingsDir: settingsDir,
		StateFile:   stateFile,
	}

	bh.CreateSampleConfig()
	return bh
}
