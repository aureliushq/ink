package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// newVersionCommand prints the stamped build metadata. Its output mirrors the
// `ink --version` flag handled by fang, with the commit and build date added.
func newVersionCommand(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the version information",
		Long:  `Print the version, commit, and build date stamped into the binary at build time.`,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("%s version: %s\n", cmd.Root().Name(), app.Build.Version)
			fmt.Printf("commit: %s\n", app.Build.Commit)
			fmt.Printf("built:  %s\n", app.Build.Date)
			return nil
		},
	}
}
