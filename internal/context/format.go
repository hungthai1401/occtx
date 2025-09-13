package context

import "fmt"

// ContextFormat represents the supported context file formats
type ContextFormat int

const (
	// FormatJSON represents standard JSON format
	FormatJSON ContextFormat = iota
	// FormatJSONC represents JSON with Comments format
	FormatJSONC
)

// String returns the string representation of the format
func (f ContextFormat) String() string {
	switch f {
	case FormatJSON:
		return "json"
	case FormatJSONC:
		return "jsonc"
	default:
		return "unknown"
	}
}

// FileExtension returns the file extension for the format
func (f ContextFormat) FileExtension() string {
	switch f {
	case FormatJSON:
		return ".json"
	case FormatJSONC:
		return ".jsonc"
	default:
		return ".json"
	}
}

// DisplayName returns the human-readable format name
func (f ContextFormat) DisplayName() string {
	switch f {
	case FormatJSON:
		return "JSON"
	case FormatJSONC:
		return "JSONC"
	default:
		return "Unknown"
	}
}

// ParseFormat parses a string into a ContextFormat
func ParseFormat(s string) (ContextFormat, error) {
	switch s {
	case "json":
		return FormatJSON, nil
	case "jsonc":
		return FormatJSONC, nil
	default:
		return FormatJSON, fmt.Errorf("invalid format '%s'. Supported formats: %s", s, GetSupportedFormats())
	}
}

// GetSupportedFormats returns a comma-separated list of supported formats
func GetSupportedFormats() string {
	return "json, jsonc"
}

// GetAllFormats returns all supported formats
func GetAllFormats() []ContextFormat {
	return []ContextFormat{FormatJSON, FormatJSONC}
}
