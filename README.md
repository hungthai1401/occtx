# occtx

![GitHub release (latest by date)](https://img.shields.io/github/v/release/hungthai1401/occtx)
![GitHub](https://img.shields.io/github/license/hungthai1401/occtx)
![Go version](https://img.shields.io/github/go-mod/go-version/hungthai1401/occtx)

**occtx** is a command-line tool for switching between different opencode contexts quickly and easily.

Inspired by [kubectx](https://github.com/ahmetb/kubectx) and [cctx](https://github.com/nwiizo/cctx).

## Features

- 🚀 **Fast context switching** - Switch between opencode configurations instantly
- 📁 **Multi-level support** - Global (`~/.config/opencode/`) and project-level (`./opencode/`) contexts
- 🎨 **Interactive mode** - Built-in fuzzy finder with fzf integration
- 📝 **Multiple formats** - Support for JSON and JSONC (JSON with Comments)
- 🎯 **Type-safe** - Enum-based format validation
- 💾 **State persistence** - Remembers current and previous contexts
- 🛡️ **Safe operations** - Atomic file operations prevent corruption

## Installation

### From source

```bash
git clone https://github.com/hungthai1401/occtx.git
cd occtx
go build -o occtx .
sudo mv occtx /usr/local/bin/
```

### Using Go

```bash
go install github.com/hungthai1401/occtx@latest
```

## Usage

### Basic Commands

```bash
# List all contexts
occtx

# Switch to a context
occtx work

# Switch to previous context
occtx -

# Show current context
occtx -c
```

### Context Management

```bash
# Create new context from current settings
occtx -n personal

# Create context with specific format
occtx -n work -f jsonc

# Delete a context
occtx -d old-context

# Rename a context
occtx -r old-name new-name

# Unset current context
occtx -u
```

### Interactive Mode

```bash
# Interactive selection (flag form)
occtx -i

# Interactive selection (command form)
occtx interactive
```

### Context Content

```bash
# Show context content
occtx -s work

# Edit context with $EDITOR
occtx -e work

# Export context to stdout
occtx --export work

# Import context from stdin
echo '{"apiKey": "key"}' | occtx --import new-context
```

### Project-Level Contexts

```bash
# Use project-level contexts
occtx --in-project

# Create project-level context
occtx --in-project -n local-dev

# List project contexts
occtx --in-project
```

## Format Support

### JSON (Default)

```bash
occtx -n my-context          # Creates my-context.json
```

Standard JSON format:
```json
{
  "theme": "default",
  "provider": {
    "anthropic": {
      "api": "https://api.anthropic.com",
      "options": {
        "apiKey": "your-api-key",
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
}
```

### JSONC (JSON with Comments)

```bash
occtx -n my-context -f jsonc  # Creates my-context.jsonc
```

JSONC format with metadata:
```jsonc
// opencode context: my-context
// Format: JSONC
// Created: 2025-09-13 15:35:19
{
  "theme": "default",
  "keybinds": {
    "leader": "ctrl+x",
    "app_help": "h",
    "app_exit": "ctrl+c,q"
  },
  "provider": {
    "anthropic": {
      "api": "https://api.anthropic.com",
      "options": {
        "apiKey": "your-api-key",
        "timeout": 30000
      }
    }
  },
  "agent": {
    "default": {
      "provider": "anthropic",
      "model": "claude-4-sonnet"
    }
  },
  "mcp": {
    "filesystem": {
      "type": "local",
      "command": ["npx", "@modelcontextprotocol/server-filesystem"],
      "enabled": true
    }
  }
}
```

## Examples

### Daily Workflow

```bash
# Create contexts for different environments
occtx -n development -f json
occtx -n production -f jsonc
occtx -n staging

# List all contexts
occtx
👤 Global contexts:
  development
* production
  staging

# Switch between contexts
occtx development
occtx production
occtx -                    # Back to development

# Interactive selection
occtx -i                   # Opens fuzzy finder
```

### Project-Specific Contexts

```bash
# Work with project-level contexts
cd my-project
occtx --in-project -n local
occtx --in-project -n test

# List shows both levels
occtx
👤 Global contexts:
  development
* production

💡 Hint: Found 2 project-level contexts. Use --in-project to see them.
```

## Configuration

### Context Storage

- **Global contexts**: `~/.config/opencode/settings/*.json`
- **Project contexts**: `./opencode/settings/*.json`
- **Active config**: `~/.config/opencode/opencode.json` or `./opencode.json`
- **State file**: `.occtx-state.json` (tracks current/previous contexts)

### Interactive Features

- **fzf integration**: Auto-detects and uses `fzf` if available
- **Built-in finder**: Fallback fuzzy finder using promptui
- **Color coding**: Current context highlighted in green
- **Visual indicators**: Emojis for different context levels

## Requirements

- Go 1.21 or later
- Optional: [fzf](https://github.com/junegunn/fzf) for enhanced interactive mode

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## Development

### Building

```bash
go build -o occtx .
```

### Testing

```bash
# Run unit tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests with verbose output
go test -v ./...

# Run specific test
go test -run TestContextManager ./internal/context

# Lint and vet
go vet ./...
go fmt ./...

# Test with race detection
go test -race ./...
```

### Integration Testing

```bash
# Test with sample opencode config
cat > ~/.config/opencode/opencode.json << EOF
{
  "theme": "default",
  "provider": {
    "anthropic": {
      "api": "https://api.anthropic.com",
      "options": {
        "apiKey": "your-key-here"
      }
    }
  },
  "agent": {
    "default": {
      "provider": "anthropic",
      "model": "claude-4-sonnet"
    }
  }
}
EOF

# Test context creation and switching
./occtx -n development
./occtx -n production -f jsonc
./occtx development
./occtx -c

# Test interactive mode
./occtx -i

# Test project-level contexts
./occtx --in-project -n local-dev
./occtx --in-project

# Clean up test contexts
./occtx -d development
./occtx -d production
./occtx --in-project -d local-dev
```

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- Inspired by [kubectx](https://github.com/ahmetb/kubectx) - Kubernetes context switcher
- Inspired by [cctx](https://github.com/nwiizo/cctx) - Another context management tool
- Built with [Cobra](https://github.com/spf13/cobra) CLI framework
- Interactive features powered by [promptui](https://github.com/manifoldco/promptui)
