package context

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// State represents the current state of occtx (current and previous context)
type State struct {
	Current  string `json:"current,omitempty"`
	Previous string `json:"previous,omitempty"`
}

// LoadState loads the state from the state file
func LoadState(stateFilePath string) (*State, error) {
	// If state file doesn't exist, return empty state
	if _, err := os.Stat(stateFilePath); os.IsNotExist(err) {
		return &State{}, nil
	}

	data, err := os.ReadFile(stateFilePath)
	if err != nil {
		return nil, err
	}

	var state State
	if err := json.Unmarshal(data, &state); err != nil {
		// If JSON is invalid, return empty state instead of failing
		return &State{}, nil
	}

	return &state, nil
}

// SaveState saves the state to the state file
func (s *State) SaveState(stateFilePath string) error {
	// Ensure the directory exists
	if err := os.MkdirAll(filepath.Dir(stateFilePath), 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}

	// Write atomically by writing to temp file first
	tempFile := stateFilePath + ".tmp"
	if err := os.WriteFile(tempFile, data, 0644); err != nil {
		return err
	}

	// Move temp file to actual location
	return os.Rename(tempFile, stateFilePath)
}

// SetCurrent updates the current context and moves old current to previous
func (s *State) SetCurrent(contextName string) {
	s.Previous = s.Current
	s.Current = contextName
}

// Unset clears the current context but keeps previous
func (s *State) Unset() {
	s.Previous = s.Current
	s.Current = ""
}

// SwitchToPrevious switches current and previous
func (s *State) SwitchToPrevious() bool {
	if s.Previous == "" {
		return false
	}

	current := s.Current
	s.Current = s.Previous
	s.Previous = current
	return true
}
