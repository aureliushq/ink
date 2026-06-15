package cmd

import (
	"github.com/aureliushq/ink/internal/content"
	"github.com/spf13/cobra"
)

// buildCmd represents the build command
func newBuildCommand(app *App) *cobra.Command {
	buildCmd := &cobra.Command{
		Use:   "build",
		Short: "Build the static site from content files",
		Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			paths, err := content.DiscoverFiles(app.Config.Build.ContentDir, app.Logger)
			if err != nil {
				return err
			}
			allContent := []content.Content{}
			for _, path := range paths {
				content := content.NewContent()
				content.SourcePath = path

				err := content.Unmarshal(app.Config.Build)
				if err != nil {
					return err
				}
				allContent = append(allContent, content)
			}
			return nil
		},
	}
	return buildCmd
}
