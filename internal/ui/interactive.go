package ui

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/fatih/color"
	"github.com/hungthai1401/occtx/internal/context"
	"github.com/manifoldco/promptui"
)

// InteractiveSelector handles interactive context selection
type InteractiveSelector struct {
	manager *context.Manager
}

// NewInteractiveSelector creates a new interactive selector
func NewInteractiveSelector(manager *context.Manager) *InteractiveSelector {
	return &InteractiveSelector{
		manager: manager,
	}
}

// SelectContext allows interactive selection of a context
func (s *InteractiveSelector) SelectContext() (string, error) {
	contexts, err := s.manager.ListContexts()
	if err != nil {
		return "", err
	}

	if len(contexts) == 0 {
		return "", fmt.Errorf("no contexts available")
	}

	// Try fzf first if available
	if contextName, err := s.selectWithFzf(contexts); err == nil {
		return contextName, nil
	}

	// Fallback to built-in selector
	return s.selectWithPromptUI(contexts)
}

// selectWithFzf uses fzf for context selection if available
func (s *InteractiveSelector) selectWithFzf(contexts []*context.Context) (string, error) {
	// Check if fzf is available
	if !isFzfAvailable() {
		return "", fmt.Errorf("fzf not available")
	}

	// Get current context for highlighting
	currentContext, _ := s.manager.GetCurrentContext()

	// Prepare input for fzf
	var items []string
	for _, ctx := range contexts {
		if ctx.Name == currentContext {
			items = append(items, fmt.Sprintf("* %s", ctx.Name))
		} else {
			items = append(items, fmt.Sprintf("  %s", ctx.Name))
		}
	}

	input := strings.Join(items, "\n")

	// Run fzf
	cmd := exec.Command("fzf",
		"--height", "40%",
		"--reverse",
		"--border",
		"--prompt", "Select context: ",
		"--header", "Press ESC to cancel",
		"--ansi", // Enable color support
	)

	cmd.Stdin = strings.NewReader(input)
	cmd.Stderr = os.Stderr

	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	selected := strings.TrimSpace(string(output))
	if selected == "" {
		return "", fmt.Errorf("no context selected")
	}

	// Extract context name (remove prefix)
	contextName := strings.TrimSpace(strings.TrimPrefix(selected, "*"))
	contextName = strings.TrimSpace(contextName)

	return contextName, nil
}

// selectWithPromptUI uses the built-in promptui for context selection
func (s *InteractiveSelector) selectWithPromptUI(contexts []*context.Context) (string, error) {
	// Get current context for highlighting
	currentContext, _ := s.manager.GetCurrentContext()

	// Create items for promptui
	items := make([]string, len(contexts))
	for i, ctx := range contexts {
		items[i] = ctx.Name
	}

	// Custom template with colors
	templates := &promptui.SelectTemplates{
		Label:    "{{ . }}?",
		Active:   "▸ {{ . | cyan }}",
		Inactive: "  {{ . }}",
		Selected: "{{ \"✓\" | green }} {{ . | cyan }}",
	}

	// Add current context indicator
	funcMap := promptui.FuncMap
	funcMap["current"] = func(name string) string {
		if name == currentContext {
			return color.GreenString("* %s", name)
		}
		return fmt.Sprintf("  %s", name)
	}

	templates.Active = "▸ {{ . | current }}"
	templates.Inactive = "{{ . | current }}"

	prompt := promptui.Select{
		Label:     "Select context",
		Items:     items,
		Templates: templates,
		Size:      10,
		Searcher: func(input string, index int) bool {
			name := items[index]
			return strings.Contains(strings.ToLower(name), strings.ToLower(input))
		},
	}

	_, result, err := prompt.Run()
	if err != nil {
		return "", err
	}

	return result, nil
}

// isFzfAvailable checks if fzf is available in PATH
func isFzfAvailable() bool {
	_, err := exec.LookPath("fzf")
	return err == nil
}

// ColorPrinter provides consistent color printing across the application
type ColorPrinter struct {
	Success *color.Color
	Error   *color.Color
	Info    *color.Color
	Warning *color.Color
	Current *color.Color
}

// NewColorPrinter creates a new color printer
func NewColorPrinter() *ColorPrinter {
	return &ColorPrinter{
		Success: color.New(color.FgGreen, color.Bold),
		Error:   color.New(color.FgRed, color.Bold),
		Info:    color.New(color.FgBlue),
		Warning: color.New(color.FgYellow),
		Current: color.New(color.FgGreen, color.Bold),
	}
}

// PrintSuccess prints a success message
func (cp *ColorPrinter) PrintSuccess(format string, args ...interface{}) {
	cp.Success.Printf(format, args...)
}

// PrintError prints an error message
func (cp *ColorPrinter) PrintError(format string, args ...interface{}) {
	cp.Error.Printf(format, args...)
}

// PrintInfo prints an info message
func (cp *ColorPrinter) PrintInfo(format string, args ...interface{}) {
	cp.Info.Printf(format, args...)
}

// PrintWarning prints a warning message
func (cp *ColorPrinter) PrintWarning(format string, args ...interface{}) {
	cp.Warning.Printf(format, args...)
}

// PrintCurrent prints current context with highlighting
func (cp *ColorPrinter) PrintCurrent(format string, args ...interface{}) {
	cp.Current.Printf(format, args...)
}

// ContextListFormatter handles formatting of context lists
type ContextListFormatter struct {
	printer *ColorPrinter
}

// NewContextListFormatter creates a new context list formatter
func NewContextListFormatter() *ContextListFormatter {
	return &ContextListFormatter{
		printer: NewColorPrinter(),
	}
}

// FormatContextList formats and prints a list of contexts
func (clf *ContextListFormatter) FormatContextList(contexts []*context.Context, currentContext string, useProject bool) {
	if len(contexts) == 0 {
		levelText := "global"
		if useProject {
			levelText = "project"
		}
		fmt.Printf("No %s contexts found\n", levelText)
		return
	}

	// Show level indicator with emoji
	levelEmoji := "👤"
	levelText := "Global"
	if useProject {
		levelEmoji = "📁"
		levelText = "Project"
	}

	fmt.Printf("%s %s contexts:\n", levelEmoji, levelText)

	// Print contexts with current highlighted
	for _, ctx := range contexts {
		if ctx.Name == currentContext {
			clf.printer.PrintCurrent("* %s\n", ctx.Name)
		} else {
			fmt.Printf("  %s\n", ctx.Name)
		}
	}
}

// ShowHints displays helpful hints to the user
func (clf *ContextListFormatter) ShowHints(useProject bool, hasProjectContexts bool) {
	if !useProject && hasProjectContexts {
		clf.printer.PrintInfo("\n💡 Hint: Found project-level contexts. Use --in-project to see them.\n")
	}
}

// ProgressIndicator shows progress for long-running operations
type ProgressIndicator struct {
	message string
}

// NewProgressIndicator creates a new progress indicator
func NewProgressIndicator(message string) *ProgressIndicator {
	return &ProgressIndicator{message: message}
}

// Show displays the progress message
func (pi *ProgressIndicator) Show() {
	fmt.Printf("⏳ %s...\n", pi.message)
}

// Success shows success message
func (pi *ProgressIndicator) Success(message string) {
	printer := NewColorPrinter()
	printer.PrintSuccess("✓ %s\n", message)
}

// Error shows error message
func (pi *ProgressIndicator) Error(message string) {
	printer := NewColorPrinter()
	printer.PrintError("✗ %s\n", message)
}
