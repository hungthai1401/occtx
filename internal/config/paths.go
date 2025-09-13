package config

import (
	"os"
	"path/filepath"
)

const (
	// OpenCodeConfigDir is the default directory for opencode configurations
	OpenCodeConfigDir = ".config/opencode"
	// SettingsSubDir is the subdirectory where contexts are stored
	SettingsSubDir = "settings"
	// StateFileName is the hidden state file that tracks current/previous contexts
	StateFileName = ".occtx-state.json"
	// ActiveConfigFileName is the active opencode.json file
	ActiveConfigFileName = "opencode.json"
	// ProjectConfigFileName is the project-level config file
	ProjectConfigFileName = "opencode.json"
	// ProjectConfigDir is the project-level config directory
	ProjectConfigDir = "opencode"
)

// Paths holds all the important file paths for occtx
type Paths struct {
	// Global level paths (default)
	GlobalConfigDir    string // ~/.config/opencode/
	GlobalSettingsDir  string // ~/.config/opencode/settings/
	GlobalActiveConfig string // ~/.config/opencode/opencode.json
	GlobalStateFile    string // ~/.config/opencode/settings/.occtx-state.json

	// Project level paths
	ProjectConfigDir    string // ./opencode/
	ProjectSettingsDir  string // ./opencode/settings/
	ProjectActiveConfig string // ./opencode.json
	ProjectStateFile    string // ./opencode/settings/.occtx-state.json
}

// NewPaths creates a new Paths struct with all paths initialized
func NewPaths() (*Paths, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	globalConfigDir := filepath.Join(homeDir, OpenCodeConfigDir)
	globalSettingsDir := filepath.Join(globalConfigDir, SettingsSubDir)

	currentDir, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	projectConfigDir := filepath.Join(currentDir, ProjectConfigDir)
	projectSettingsDir := filepath.Join(projectConfigDir, SettingsSubDir)

	return &Paths{
		GlobalConfigDir:    globalConfigDir,
		GlobalSettingsDir:  globalSettingsDir,
		GlobalActiveConfig: filepath.Join(globalConfigDir, ActiveConfigFileName),
		GlobalStateFile:    filepath.Join(globalSettingsDir, StateFileName),

		ProjectConfigDir:    projectConfigDir,
		ProjectSettingsDir:  projectSettingsDir,
		ProjectActiveConfig: filepath.Join(currentDir, ProjectConfigFileName),
		ProjectStateFile:    filepath.Join(projectSettingsDir, StateFileName),
	}, nil
}

// GetContextsDir returns the appropriate contexts directory based on level
func (p *Paths) GetContextsDir(useProject bool) string {
	if useProject {
		return p.ProjectSettingsDir
	}
	return p.GlobalSettingsDir
}

// GetActiveConfigPath returns the appropriate active config path based on level
func (p *Paths) GetActiveConfigPath(useProject bool) string {
	if useProject {
		return p.ProjectActiveConfig
	}
	return p.GlobalActiveConfig
}

// GetStateFilePath returns the appropriate state file path based on level
func (p *Paths) GetStateFilePath(useProject bool) string {
	if useProject {
		return p.ProjectStateFile
	}
	return p.GlobalStateFile
}

// EnsureDirectories creates all necessary directories
func (p *Paths) EnsureDirectories(useProject bool) error {
	var dirs []string

	if useProject {
		dirs = []string{p.ProjectConfigDir, p.ProjectSettingsDir}
	} else {
		dirs = []string{p.GlobalConfigDir, p.GlobalSettingsDir}
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	return nil
}

// ProjectContextsExist checks if project-level contexts exist
func (p *Paths) ProjectContextsExist() bool {
	if _, err := os.Stat(p.ProjectSettingsDir); os.IsNotExist(err) {
		return false
	}

	entries, err := os.ReadDir(p.ProjectSettingsDir)
	if err != nil {
		return false
	}

	// Check if there are any .json files (excluding state file)
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".json" && entry.Name() != StateFileName {
			return true
		}
	}

	return false
}
