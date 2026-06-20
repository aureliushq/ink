package cmd

import (
	"embed"
	"html/template"
	"os"
	"path"

	"github.com/aureliushq/ink/internal/build"
	"github.com/aureliushq/ink/internal/renderer"
	"github.com/spf13/cobra"
)

// buildCmd represents the build command
func newBuildCommand(app *App, themesFS embed.FS) *cobra.Command {
	buildCmd := &cobra.Command{
		Use:   "build",
		Short: "Build the static site from content files",
		Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			allContent, err := build.ReadContent(app.Config.Build, app.Logger)
			if err != nil {
				return err
			}

			for _, content := range allContent {
				templateData := renderer.NewTemplateData(app.Config)
				templateData.Title = content.Frontmatter.Title
				templateData.Subtitle = content.Frontmatter.Subtitle
				templateData.Description = content.Frontmatter.Description
				templateData.Content = template.HTML(content.HTMLBody)

				var templateName string
				if content.Collection != "" {
					templateName = path.Base(content.DestinationPath)
					if templateName != "index.html" {
						templateName = "single.html"
					} else {
						templateName = "list.html"
					}
				} else {
					templateName = path.Base(content.DestinationPath)
					if templateName != "index.html" {
						templateName = "page.html"
					}
				}

				html, err := app.TemplateCache.Execute(templateName, templateData)
				if err != nil {
					return err
				}
				dir := path.Dir(content.DestinationPath)
				err = os.MkdirAll(dir, 0755)
				if err != nil {
					app.Logger.Infof("directory exists: %s", dir)
				}
				f, err := os.OpenFile(content.DestinationPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
				if err != nil {
					return err
				}
				defer f.Close()

				if _, err = f.WriteString(html); err != nil {
					return err
				}
			}

			return nil
		},
	}
	return buildCmd
}
