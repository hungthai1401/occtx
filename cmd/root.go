package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/hungthai1401/occtx/internal/context"
	"github.com/hungthai1401/occtx/internal/ui"
	"github.com/spf13/cobra"
)

var (
	// Global flags
	inProject bool
	verbose   bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:                "occtx",
	Short:              "opencode context switcher",
	Version:            "0.1.0",
	RunE:               runRoot,
	DisableFlagParsing: false,
	DisableAutoGenTag:  true,
	SilenceUsage:       true,
	SilenceErrors:      false,
	Args:               cobra.ArbitraryArgs, // Allow arbitrary args for context names
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().BoolVar(&inProject, "in-project", false, "Use project-level contexts (./opencode.json)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")

	// Local flags for root command
	rootCmd.Flags().BoolP("current", "c", false, "Show current context name")
	rootCmd.Flags().BoolP("unset", "u", false, "Unset current context")
	rootCmd.Flags().StringP("new", "n", "", "Create new context from current settings")
	rootCmd.Flags().StringP("format", "f", "json", fmt.Sprintf("Format for new context (%s)", context.GetSupportedFormats()))
	rootCmd.Flags().StringP("delete", "d", "", "Delete context")
	rootCmd.Flags().StringP("edit", "e", "", "Edit context with $EDITOR")
	rootCmd.Flags().StringP("show", "s", "", "Show context content")
	rootCmd.Flags().StringP("export", "", "", "Export context to stdout")
	rootCmd.Flags().StringP("import", "", "", "Import context from stdin")
	rootCmd.Flags().BoolP("interactive", "i", false, "Interactive context selection")

	// Rename requires two arguments, will handle in runRoot
	rootCmd.Flags().BoolP("rename", "r", false, "Rename context (usage: occtx -r old new)")
}

func runRoot(cmd *cobra.Command, args []string) error {
	// Handle different command modes

	// Interactive mode
	if interactive, _ := cmd.Flags().GetBool("interactive"); interactive {
		return runInteractiveSelection()
	}

	// Show current context
	if current, _ := cmd.Flags().GetBool("current"); current {
		return showCurrentContext()
	}

	// Unset current context
	if unset, _ := cmd.Flags().GetBool("unset"); unset {
		return unsetCurrentContext()
	}

	// Create new context
	if newName, _ := cmd.Flags().GetString("new"); newName != "" {
		format, _ := cmd.Flags().GetString("format")
		return createNewContext(newName, format)
	}

	// Delete context
	if deleteName, _ := cmd.Flags().GetString("delete"); deleteName != "" {
		return deleteContext(deleteName)
	}

	// Edit context
	if editName, _ := cmd.Flags().GetString("edit"); editName != "" {
		return editContext(editName)
	}

	// Show context
	if showName, _ := cmd.Flags().GetString("show"); showName != "" {
		return showContext(showName)
	}

	// Export context
	if exportName, _ := cmd.Flags().GetString("export"); exportName != "" {
		return exportContext(exportName)
	}

	// Import context
	if importName, _ := cmd.Flags().GetString("import"); importName != "" {
		return importContext(importName)
	}

	// Handle rename (requires special parsing)
	if rename, _ := cmd.Flags().GetBool("rename"); rename {
		if len(args) != 2 {
			return fmt.Errorf("rename requires exactly 2 arguments: old_name new_name")
		}
		return renameContext(args[0], args[1])
	}

	// Handle context switching and listing
	switch len(args) {
	case 0:
		// List contexts
		return listContexts()
	case 1:
		if args[0] == "-" {
			// Switch to previous context
			return switchToPreviousContext()
		}
		// Switch to named context
		return switchToContext(args[0])
	default:
		return fmt.Errorf("too many arguments")
	}
}

// Implementation functions using context manager
func showCurrentContext() error {
	manager, err := context.NewManager(inProject)
	if err != nil {
		return err
	}

	current, err := manager.GetCurrentContext()
	if err != nil {
		return err
	}

	if current == "" {
		fmt.Println("No current context set")
		return nil
	}

	fmt.Println(current)
	return nil
}

func unsetCurrentContext() error {
	manager, err := context.NewManager(inProject)
	if err != nil {
		return err
	}

	if err := manager.UnsetCurrentContext(); err != nil {
		return err
	}

	fmt.Println("Current context unset")
	return nil
}

func createNewContext(name, formatStr string) error {
	// Parse and validate format
	format, err := context.ParseFormat(formatStr)
	if err != nil {
		return err
	}

	manager, err := context.NewManager(inProject)
	if err != nil {
		return err
	}

	if err := manager.CreateContextWithFormat(name, format); err != nil {
		return err
	}

	printer := ui.NewColorPrinter()
	printer.PrintSuccess("Context '%s' created successfully (%s format)\n", name, format.DisplayName())
	return nil
}

func deleteContext(name string) error {
	manager, err := context.NewManager(inProject)
	if err != nil {
		return err
	}

	if err := manager.DeleteContext(name); err != nil {
		return err
	}

	fmt.Printf("Context '%s' deleted\n", name)
	return nil
}

func editContext(name string) error {
	manager, err := context.NewManager(inProject)
	if err != nil {
		return err
	}

	// Get context to ensure it exists
	ctx, err := manager.GetContext(name)
	if err != nil {
		return err
	}

	// Get editor from environment
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vi" // fallback to vi
	}

	// Show progress
	progress := ui.NewProgressIndicator("Opening editor")
	progress.Show()

	// Open editor
	cmd := exec.Command(editor, ctx.FilePath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		progress.Error("Failed to edit context")
		return fmt.Errorf("failed to run editor: %v", err)
	}

	progress.Success(fmt.Sprintf("Context '%s' edited successfully", name))
	return nil
}

func showContext(name string) error {
	manager, err := context.NewManager(inProject)
	if err != nil {
		return err
	}

	ctx, err := manager.GetContext(name)
	if err != nil {
		return err
	}

	// Read and display the raw JSON content
	data, err := os.ReadFile(ctx.FilePath)
	if err != nil {
		return err
	}

	fmt.Print(string(data))
	return nil
}

func exportContext(name string) error {
	manager, err := context.NewManager(inProject)
	if err != nil {
		return err
	}

	ctx, err := manager.GetContext(name)
	if err != nil {
		return err
	}

	// Read and output to stdout
	data, err := os.ReadFile(ctx.FilePath)
	if err != nil {
		return err
	}

	fmt.Print(string(data))
	return nil
}

func importContext(name string) error {
	manager, err := context.NewManager(inProject)
	if err != nil {
		return err
	}

	// Read from stdin
	var input strings.Builder
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		input.WriteString(scanner.Text())
		input.WriteString("\n")
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("failed to read from stdin: %v", err)
	}

	jsonData := input.String()
	if jsonData == "" {
		return fmt.Errorf("no input provided")
	}

	// Validate JSON
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(jsonData), &data); err != nil {
		return fmt.Errorf("invalid JSON: %v", err)
	}

	// Ensure directories exist
	if err := manager.GetPaths().EnsureDirectories(inProject); err != nil {
		return err
	}

	// Write to context file
	contextsDir := manager.GetPaths().GetContextsDir(inProject)
	contextPath := filepath.Join(contextsDir, name+".json")

	// Check if context already exists
	if _, err := os.Stat(contextPath); err == nil {
		return fmt.Errorf("context '%s' already exists", name)
	}

	// Format and write JSON
	formattedData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	// Write atomically
	tempPath := contextPath + ".tmp"
	if err := os.WriteFile(tempPath, formattedData, 0644); err != nil {
		return err
	}

	if err := os.Rename(tempPath, contextPath); err != nil {
		return err
	}

	printer := ui.NewColorPrinter()
	printer.PrintSuccess("Context '%s' imported successfully\n", name)
	return nil
}

func renameContext(oldName, newName string) error {
	manager, err := context.NewManager(inProject)
	if err != nil {
		return err
	}

	if err := manager.RenameContext(oldName, newName); err != nil {
		return err
	}

	fmt.Printf("Context '%s' renamed to '%s'\n", oldName, newName)
	return nil
}

func listContexts() error {
	manager, err := context.NewManager(inProject)
	if err != nil {
		return err
	}

	contexts, err := manager.ListContexts()
	if err != nil {
		return err
	}

	// Get current context for highlighting
	currentContext, _ := manager.GetCurrentContext()

	// Use the new formatter
	formatter := ui.NewContextListFormatter()
	formatter.FormatContextList(contexts, currentContext, inProject)

	// Show helpful hints if not using project level
	if !inProject {
		// Check if project contexts exist
		projectManager, _ := context.NewManager(true)
		if projectManager != nil {
			projectContexts, _ := projectManager.ListContexts()
			formatter.ShowHints(inProject, len(projectContexts) > 0)
		}
	}

	return nil
}

func switchToPreviousContext() error {
	manager, err := context.NewManager(inProject)
	if err != nil {
		return err
	}

	if err := manager.SwitchToPrevious(); err != nil {
		return err
	}

	// Show which context we switched to
	current, _ := manager.GetCurrentContext()
	printer := ui.NewColorPrinter()
	printer.PrintSuccess("Switched to context: %s\n", current)
	return nil
}

func switchToContext(name string) error {
	manager, err := context.NewManager(inProject)
	if err != nil {
		return err
	}

	if err := manager.SwitchToContext(name); err != nil {
		return err
	}

	printer := ui.NewColorPrinter()
	printer.PrintSuccess("Switched to context: %s\n", name)
	return nil
}
