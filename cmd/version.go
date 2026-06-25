package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// newVersionCommand prints the stamped build metadata. Its output mirrors the
// `ink --version` flag handled by fang, with the commit and build date added.
func newVersionCommand(app *App) *cobra.Command {
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Print the version information",
		Long:  `Print the version, commit, and build date stamped into the binary at build time.`,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "%s version %s\n", cmd.Root().Name(), app.Build.Version)
			fmt.Fprintf(out, "commit: %s\n", app.Build.Commit)
			fmt.Fprintf(out, "built:  %s\n", app.Build.Date)
			return nil
		},
	}
	return versionCmd
}
