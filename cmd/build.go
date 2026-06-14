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
			paths := content.DiscoverContentFiles(app.Config.Build.ContentDir, app.Logger)
			allContent := []content.Content{}
			for _, path := range paths {
				content := content.NewContent()
				content.Path = path

				err := content.ReadFile()
				if err != nil {
					app.Logger.Error(err)
					return err
				}
				allContent = append(allContent, content)
			}
			return nil
		},
	}
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// buildCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// buildCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	return buildCmd
}
