package cmd

import (
	"fmt"

	"github.com/hungthai1401/occtx/internal/context"
	"github.com/hungthai1401/occtx/internal/ui"
	"github.com/spf13/cobra"
)

// interactiveCmd represents the interactive command for context selection
var interactiveCmd = &cobra.Command{
	Use:   "interactive",
	Short: "Interactive context selection",
	Long: `Interactive mode allows you to select contexts using either fzf (if available) 
or a built-in fuzzy finder. This provides a more user-friendly way to browse 
and select contexts when you have many available.

Examples:
  occtx interactive           # Interactive selection
  occtx -i                    # Flag form (same functionality)`,
	Aliases: []string{"i"},
	RunE: func(cmd *cobra.Command, args []string) error {
		return runInteractiveSelection()
	},
}

func init() {
	rootCmd.AddCommand(interactiveCmd)
}

// runInteractiveSelection is shared between the flag and command forms
func runInteractiveSelection() error {
	manager, err := context.NewManager(inProject)
	if err != nil {
		return err
	}

	selector := ui.NewInteractiveSelector(manager)

	contextName, err := selector.SelectContext()
	if err != nil {
		return fmt.Errorf("interactive selection failed: %v", err)
	}

	// Switch to selected context
	if err := manager.SwitchToContext(contextName); err != nil {
		return err
	}

	// Show success message
	printer := ui.NewColorPrinter()
	printer.PrintSuccess("Switched to context: %s\n", contextName)

	return nil
}
