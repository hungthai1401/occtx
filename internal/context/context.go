package context

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/hungthai1401/occtx/internal/config"
)

// Context represents an opencode context
type Context struct {
	Name     string                 `json:"-"` // Name is derived from filename
	Data     map[string]interface{} `json:"-"` // Raw JSON data
	FilePath string                 `json:"-"` // Full path to the context file
}

// Manager handles context operations
type Manager struct {
	paths      *config.Paths
	useProject bool
}

// GetPaths returns the paths configuration
func (m *Manager) GetPaths() *config.Paths {
	return m.paths
}

// NewManager creates a new context manager
func NewManager(useProject bool) (*Manager, error) {
	paths, err := config.NewPaths()
	if err != nil {
		return nil, err
	}

	return &Manager{
		paths:      paths,
		useProject: useProject,
	}, nil
}

// ListContexts returns all available contexts
func (m *Manager) ListContexts() ([]*Context, error) {
	contextsDir := m.paths.GetContextsDir(m.useProject)

	// Check if directory exists
	if _, err := os.Stat(contextsDir); os.IsNotExist(err) {
		return []*Context{}, nil
	}

	entries, err := os.ReadDir(contextsDir)
	if err != nil {
		return nil, err
	}

	var contexts []*Context
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		// Skip state file
		if entry.Name() == config.StateFileName {
			continue
		}

		// Check for both .json and .jsonc files
		var name string
		if strings.HasSuffix(entry.Name(), ".json") {
			name = strings.TrimSuffix(entry.Name(), ".json")
		} else if strings.HasSuffix(entry.Name(), ".jsonc") {
			name = strings.TrimSuffix(entry.Name(), ".jsonc")
		} else {
			continue // Skip non-JSON files
		}

		contextPath := filepath.Join(contextsDir, entry.Name())

		context := &Context{
			Name:     name,
			FilePath: contextPath,
		}

		contexts = append(contexts, context)
	}

	return contexts, nil
}

// GetContext loads a specific context by name
func (m *Manager) GetContext(name string) (*Context, error) {
	if err := validateContextName(name); err != nil {
		return nil, err
	}

	contextsDir := m.paths.GetContextsDir(m.useProject)

	// Try .json first, then .jsonc
	var contextPath string
	jsonPath := filepath.Join(contextsDir, name+".json")
	jsoncPath := filepath.Join(contextsDir, name+".jsonc")

	if _, err := os.Stat(jsonPath); err == nil {
		contextPath = jsonPath
	} else if _, err := os.Stat(jsoncPath); err == nil {
		contextPath = jsoncPath
	} else {
		return nil, fmt.Errorf("context '%s' not found", name)
	}

	data, err := os.ReadFile(contextPath)
	if err != nil {
		return nil, err
	}

	// For JSONC, we need to strip comments before parsing
	var contextData map[string]interface{}
	if strings.HasSuffix(contextPath, ".jsonc") {
		// Simple comment removal for JSONC (remove lines starting with //)
		lines := strings.Split(string(data), "\n")
		var cleanLines []string
		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			if !strings.HasPrefix(trimmed, "//") {
				cleanLines = append(cleanLines, line)
			}
		}
		cleanData := strings.Join(cleanLines, "\n")

		if err := json.Unmarshal([]byte(cleanData), &contextData); err != nil {
			return nil, fmt.Errorf("invalid JSON in context '%s': %v", name, err)
		}
	} else {
		if err := json.Unmarshal(data, &contextData); err != nil {
			return nil, fmt.Errorf("invalid JSON in context '%s': %v", name, err)
		}
	}

	return &Context{
		Name:     name,
		Data:     contextData,
		FilePath: contextPath,
	}, nil
}

// CreateContext creates a new context from current active config (JSON format)
func (m *Manager) CreateContext(name string) error {
	return m.CreateContextWithFormat(name, FormatJSON)
}

// CreateContextWithFormat creates a new context with specified format
func (m *Manager) CreateContextWithFormat(name string, format ContextFormat) error {

	if err := validateContextName(name); err != nil {
		return err
	}

	// Ensure directories exist
	if err := m.paths.EnsureDirectories(m.useProject); err != nil {
		return err
	}

	// Determine file extension using enum
	fileExt := format.FileExtension()

	// Check if context already exists (check both .json and .jsonc)
	contextsDir := m.paths.GetContextsDir(m.useProject)
	contextPath := filepath.Join(contextsDir, name+fileExt)

	// Check if context exists in any format
	for _, f := range GetAllFormats() {
		existingPath := filepath.Join(contextsDir, name+f.FileExtension())
		if _, err := os.Stat(existingPath); err == nil {
			return fmt.Errorf("context '%s' already exists (%s format)", name, f.DisplayName())
		}
	}

	// Read current active config
	activeConfigPath := m.paths.GetActiveConfigPath(m.useProject)
	if _, err := os.Stat(activeConfigPath); os.IsNotExist(err) {
		return fmt.Errorf("no active opencode.json found at %s", activeConfigPath)
	}

	data, err := os.ReadFile(activeConfigPath)
	if err != nil {
		return err
	}

	// Validate JSON
	var jsonData map[string]interface{}
	if err := json.Unmarshal(data, &jsonData); err != nil {
		return fmt.Errorf("current opencode.json is not valid JSON: %v", err)
	}

	// Format content based on format type
	var formattedData []byte
	switch format {
	case FormatJSONC:
		// For JSONC, add a comment header and format nicely
		formattedJSON, err := json.MarshalIndent(jsonData, "", "  ")
		if err != nil {
			return err
		}

		comment := fmt.Sprintf("// opencode context: %s\n// Format: %s\n// Created: %s\n",
			name,
			format.DisplayName(),
			time.Now().Format("2006-01-02 15:04:05"))

		formattedData = append([]byte(comment), formattedJSON...)
	case FormatJSON:
		// Standard JSON formatting
		var err error
		formattedData, err = json.MarshalIndent(jsonData, "", "  ")
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}

	// Write atomically
	tempPath := contextPath + ".tmp"
	if err := os.WriteFile(tempPath, formattedData, 0644); err != nil {
		return err
	}

	return os.Rename(tempPath, contextPath)
}

// SwitchToContext switches to the specified context
func (m *Manager) SwitchToContext(name string) error {
	// Get the context to ensure it exists and is valid
	context, err := m.GetContext(name)
	if err != nil {
		return err
	}

	// Ensure active config directory exists
	activeConfigPath := m.paths.GetActiveConfigPath(m.useProject)
	if err := os.MkdirAll(filepath.Dir(activeConfigPath), 0755); err != nil {
		return err
	}

	// Copy context file to active config (atomic operation)
	data, err := os.ReadFile(context.FilePath)
	if err != nil {
		return err
	}

	tempPath := activeConfigPath + ".tmp"
	if err := os.WriteFile(tempPath, data, 0644); err != nil {
		return err
	}

	if err := os.Rename(tempPath, activeConfigPath); err != nil {
		return err
	}

	// Update state
	stateFilePath := m.paths.GetStateFilePath(m.useProject)
	state, err := LoadState(stateFilePath)
	if err != nil {
		return err
	}

	state.SetCurrent(name)
	return state.SaveState(stateFilePath)
}

// DeleteContext deletes the specified context
func (m *Manager) DeleteContext(name string) error {
	if err := validateContextName(name); err != nil {
		return err
	}

	// Check if context exists
	context, err := m.GetContext(name)
	if err != nil {
		return err
	}

	// Check if it's the current context
	stateFilePath := m.paths.GetStateFilePath(m.useProject)
	state, err := LoadState(stateFilePath)
	if err != nil {
		return err
	}

	if state.Current == name {
		return fmt.Errorf("cannot delete current context '%s'. Switch to another context first", name)
	}

	// Delete the file
	return os.Remove(context.FilePath)
}

// RenameContext renames a context
func (m *Manager) RenameContext(oldName, newName string) error {
	if err := validateContextName(oldName); err != nil {
		return fmt.Errorf("invalid old name: %v", err)
	}
	if err := validateContextName(newName); err != nil {
		return fmt.Errorf("invalid new name: %v", err)
	}

	// Check if old context exists
	oldContext, err := m.GetContext(oldName)
	if err != nil {
		return err
	}

	// Check if new name already exists
	contextsDir := m.paths.GetContextsDir(m.useProject)
	newContextPath := filepath.Join(contextsDir, newName+".json")

	if _, err := os.Stat(newContextPath); err == nil {
		return fmt.Errorf("context '%s' already exists", newName)
	}

	// Rename the file
	if err := os.Rename(oldContext.FilePath, newContextPath); err != nil {
		return err
	}

	// Update state if the renamed context is current or previous
	stateFilePath := m.paths.GetStateFilePath(m.useProject)
	state, err := LoadState(stateFilePath)
	if err != nil {
		return err
	}

	updated := false
	if state.Current == oldName {
		state.Current = newName
		updated = true
	}
	if state.Previous == oldName {
		state.Previous = newName
		updated = true
	}

	if updated {
		return state.SaveState(stateFilePath)
	}

	return nil
}

// GetCurrentContext returns the current context name
func (m *Manager) GetCurrentContext() (string, error) {
	stateFilePath := m.paths.GetStateFilePath(m.useProject)
	state, err := LoadState(stateFilePath)
	if err != nil {
		return "", err
	}

	return state.Current, nil
}

// SwitchToPrevious switches to the previous context
func (m *Manager) SwitchToPrevious() error {
	stateFilePath := m.paths.GetStateFilePath(m.useProject)
	state, err := LoadState(stateFilePath)
	if err != nil {
		return err
	}

	if !state.SwitchToPrevious() {
		return fmt.Errorf("no previous context available")
	}

	// Verify the previous context still exists
	if _, err := m.GetContext(state.Current); err != nil {
		return fmt.Errorf("previous context '%s' no longer exists", state.Current)
	}

	// Switch to the context
	if err := m.SwitchToContext(state.Current); err != nil {
		return err
	}

	return nil
}

// UnsetCurrentContext removes the current context
func (m *Manager) UnsetCurrentContext() error {
	activeConfigPath := m.paths.GetActiveConfigPath(m.useProject)

	// Remove active config file if it exists
	if _, err := os.Stat(activeConfigPath); err == nil {
		if err := os.Remove(activeConfigPath); err != nil {
			return err
		}
	}

	// Update state
	stateFilePath := m.paths.GetStateFilePath(m.useProject)
	state, err := LoadState(stateFilePath)
	if err != nil {
		return err
	}

	state.Unset()
	return state.SaveState(stateFilePath)
}

// validateContextName validates that a context name is safe
func validateContextName(name string) error {
	if name == "" {
		return fmt.Errorf("context name cannot be empty")
	}

	if strings.Contains(name, "/") || strings.Contains(name, "\\") {
		return fmt.Errorf("context name cannot contain path separators")
	}

	if name == "." || name == ".." {
		return fmt.Errorf("context name cannot be '.' or '..'")
	}

	if strings.HasPrefix(name, ".") {
		return fmt.Errorf("context name cannot start with '.'")
	}

	return nil
}
