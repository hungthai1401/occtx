# 🔄 AGENTS.md - occtx Project Documentation

## 📋 Project Overview

**occtx** (Opencore Context) is a fast, secure, and intuitive command-line tool for managing multiple opencode `opencode.json` configurations. Built with Go for performance and reliability.

## 🏗️ Architecture

### 🎯 Core Concept
- **🔧 Context**: A saved opencode configuration stored as a JSON file or JSONC file
- **⚡ Current Context**: The active configuration (`~/.config/opencode/opencode.json`)
- **📁 Context Storage**: All contexts are stored in `~/.config/opencode/` as individual JSON files
- **📊 State Management**: Current and previous context tracked in `~/.config/opencode/.occtx-state.json`

### 📁 File Structure
```
📁 ~/.config/opencode/
├── ⚙️ opencode.json           # Current active context (managed by occtx)
└── 📁 settings/
    ├── 💼 work.json          # Work context
    ├── 🏠 personal.json      # Personal context
    ├── 🚀 project-alpha.json # Project-specific context
    └── 🔒 .occtx-state.json   # Hidden state file (tracks current/previous)

⚙️ opencode.json 
📁 ./opencode/
└── 📁 settings/
    ├── 💼 work.json          # Work context
    ├── 🏠 personal.json      # Personal context
    ├── 🚀 project-alpha.json # Project-specific context
    └── 🔒 .occtx-state.json   # Hidden state file (tracks current/previous)
```

### 🎯 Key Design Decisions
1. **File-based contexts**: Each context is a separate JSON file or JSONC file, making manual management possible
2. **Simple naming**: Filename (without .json) = context name
3. **Atomic operations**: Context switching is done by copying files
4. **Hidden state file**: Prefixed with `.` to hide from context listings
5. **Predictable UX**: Default behavior always uses user-level contexts for consistency
6. **Progressive disclosure**: Helpful hints show when project/local contexts are available

## 🎯 Command Reference

### 🚀 Basic Commands
- `occtx` - List contexts (defaults to Global-level, shows helpful hints)
- `occtx <name>` - Switch to context
- `occtx -` - Switch to previous context

### 🏗️ Settings Level Management
- `occtx` - Default: Global-level contexts (`~/.config/opencode/opencode.json` or `~/.config/opencode/opencode.jsonc`)
- `occtx --in-project` - Project-level contexts (`./opencode.json` or `./opencode.jsonc`)

### 🛠️ Management Commands
- `occtx -n <name> -f` - Create new context from current settings
- `occtx -d <name>` - Delete context
- `occtx -r <old> <new>` - Rename context
- `occtx -c` - Show current context name
- `occtx -e [name]` - Edit context with $EDITOR
- `occtx -s [name]` - Show context content
- `occtx -u` - Unset current context

### 📥📤 Import/Export
- `occtx --export <name>` - Export to stdout
- `occtx --import <name>` - Import from stdin

## Implementation Details

### Language & Dependencies
- **Language**: Go (edition 2021)
- **Key Dependencies**:
  - `cobra` - Command-line argument parsing
  - `json` - JSON serialization
  - `promptui` - Interactive prompts

### Error Handling
- Use `errors.Wrap` for all functions that can fail
- Provide clear error messages with context
- Validate context names (no `/`, `.`, `..`, or empty)
- Check for active context before deletion

### 🎨 Interactive Features
1. **fzf integration**: Auto-detect and use if available
2. **Built-in fuzzy finder**: Fallback when fzf not available
3. **Color coding**: Current context highlighted in green
4. **Helpful hints**: Shows available project/local contexts when at user level
5. **Visual indicators**: Emojis for different context levels (👤 Global, 📁 Project)

## 🚀 Release Management

### Simplified Release System

The project uses a streamlined release process with one primary tool:

#### **quick-release.sh** - Primary Release Script

A simple, reliable release script that handles the entire release process:

```bash
# One-command release
./quick-release.sh patch      # 0.1.0 -> 0.1.1
./quick-release.sh minor      # 0.1.0 -> 0.2.0  
./quick-release.sh major      # 0.1.0 -> 1.0.0
```

**What it does:**
1. ✅ Validates git state (clean working tree, on main branch)
2. ✅ Runs quality checks (fmt, clippy, test, build)
3. ✅ Updates version in Cargo.toml
4. ✅ Creates git commit and tag
5. ✅ Pushes to GitHub
6. ✅ Triggers GitHub Actions for:
   - Building release binaries for all platforms
   - Creating GitHub release with artifacts
   - Publishing to crates.io

#### **GitHub Actions Workflows**

**CI Pipeline** (`.github/workflows/ci.yml`):
- Multi-platform testing (Ubuntu, macOS, Windows)
- Go stable version only
- Format checking, linting, tests
- Security audit

**Release Pipeline** (`.github/workflows/release.yml`):
- Triggered by version tags (v*.*.*)
- Builds binaries for:
  - Linux x86_64 (glibc and musl)
  - Windows x86_64
  - macOS x86_64 and aarch64
- Creates GitHub release with all artifacts

**Publish Pipeline** (`.github/workflows/publish.yml`):
- Triggered by version tags
- Runs final quality checks
- Publishes to GitHub Releases

### Release Process

1. **Make your changes and commit them**
2. **Run the release command:**
   ```bash
   ./quick-release.sh patch  # or minor/major
   ```
3. **Confirm when prompted**
4. **Monitor progress at:** https://github.com/hungthai1401/occtx/actions
5. **Release appears at:** https://github.com/hungthai1401/occtx/releases

### Quality Requirements

All releases automatically check:
- ✅ `go fmt` (code formatting)
- ✅ `go vet` (linting)
- ✅ `go test` (unit tests)
- ✅ `go build` (release build)
- ✅ Git working directory is clean
- ✅ On main branch and up-to-date with origin

### CI/CD Configuration
**Key Settings:**
- Platforms: Linux, macOS, Windows
- Release formats: Binary executables

## Development Guidelines

### Before Making Changes

1. **Understand the current implementation**:
   ```bash
   go check
   go vet
   ```

2. **Run existing tests** (if any):
   ```bash
   go test
   ```

### Making Changes

1. **Always run linting** before committing:
   ```bash
   go vet
   ```

2. **Format code** using Go standards:
   ```bash
   go fmt
   ```

3. **Test thoroughly**:
   - Test basic operations: create, switch, delete contexts
   - Test edge cases: empty names, special characters, missing files
   - Test interactive mode with and without fzf
   - Test on different platforms if possible

4. **Validate JSON handling**:
   - Ensure invalid JSON files are rejected
   - Preserve JSON formatting when possible
   - Handle missing or corrupted state files gracefully

### Testing Checklist

When testing changes, verify:

- [ ] `occtx` lists all contexts correctly
- [ ] `occtx <name>` switches context
- [ ] `occtx -` returns to previous context
- [ ] `occtx -n <name>` [-f] creates new context, default is JSON file, -f (--format) jsonc creates JSONC file
- [ ] `occtx -d <name>` deletes context (not if current)
- [ ] `occtx -r <old> <new>` renames context
- [ ] Interactive mode works (both fzf and built-in)
- [ ] Error messages are clear and helpful
- [ ] State persistence works across sessions
- [ ] Hidden files are excluded from listings

### Common Pitfalls

1. **File permissions**: Ensure created files have appropriate permissions
2. **Path handling**: Use PathBuf consistently, avoid string manipulation
3. **JSON validation**: Always validate JSON before writing
4. **State consistency**: Update state file atomically

## Future Considerations

### Potential Enhancements
- Context templates/inheritance
- Context validation against opencode schema
- Backup/restore functionality
- Context history beyond just previous
- Shell completions

### Compatibility
- Maintain backward compatibility with existing contexts
- Keep command-line interface stable
- Preserve kubectx-like user experience

## Code Quality Standards

1. **Every function should**:
   - Have a clear, single responsibility
   - Return `Result` for fallible operations
   - Include error context with `.context()`

2. **User-facing messages**:
   - Error messages should be helpful and actionable
   - Success messages should be concise
   - Use color coding consistently (green=success, red=error)

3. **File operations**:
   - Always check if directories exist before use
   - Handle missing files gracefully
   - Use atomic operations where possible


## 🎯 UX Design Philosophy

### 🏆 Simplified User Experience (v0.1.1+)

**Core Principle**: **Predictable defaults with explicit overrides**

#### ✅ What We Did Right
- **Removed complex auto-detection** that was confusing users
- **Default always uses Global-level** for predictable behavior
- **Clear explicit flags** (`--in-project`) when needed
- **Helpful progressive disclosure** - hints when other contexts available
- **Visual clarity** with emojis and condensed information

#### ❌ What We Avoided
- **Complex flag combinations** (`--global`, `--project`)
- **Unpredictable auto-detection logic** 
- **Verbose technical output** showing file paths
- **Cognitive overhead** from too many options

#### 🎯 UX Goals Achieved
1. **⚡ Speed**: Default behavior is instant and predictable
2. **🧠 Simplicity**: Two explicit flags instead of four confusing ones
3. **🎯 Discoverability**: Helpful hints guide users to advanced features
4. **🔄 Consistency**: Always behaves the same way (Global-level default)

### 📝 Usage Patterns

```bash
# 90% of usage - simple and predictable
occtx                    # List user contexts + helpful hints
occtx work              # Switch to work context

# 10% of usage - explicit when needed  
occtx --in-project staging   # Project-specific contexts
```

## 📚 Notes for AI Assistants

When working on this codebase:

1. **Always run `go vet` and fix warnings** before suggesting code
2. **Test your changes** - don't assume code works
3. **Preserve existing behavior** unless explicitly asked to change it
4. **Follow Go idioms** and best practices
5. **Keep the kubectx-inspired UX** - simple, fast, intuitive
6. **Maintain predictable defaults** - user should never be surprised
7. **Document any new features** in both code and README
8. **Consider edge cases** - empty states, missing files, permissions
9. **Progressive disclosure** - show advanced features only when relevant

Remember: This tool is about speed and simplicity. Every feature should make context switching faster or easier, not more complex. **Predictability beats cleverness.**
