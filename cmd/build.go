package cmd

import (
	"html/template"
	"os"
	"path"
	"strings"

	"github.com/aureliushq/ink/internal/content"
	"github.com/aureliushq/ink/internal/renderer"
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

			for _, content := range allContent {
				// TODO: remove this later when we're handling collections correctly
				if !strings.HasPrefix(content.Slug, "/posts") {
					templateData := renderer.NewTemplateData()
					templateData.SiteTitle = app.Config.Site.Title
					templateData.SiteSubtitle = app.Config.Site.Subtitle
					templateData.Title = content.Frontmatter.Title
					templateData.Subtitle = content.Frontmatter.Subtitle
					templateData.Description = content.Frontmatter.Description
					templateData.Content = template.HTML(content.HTMLBody)

					// TODO: use the correct template file for different content types
					templateName := path.Base(content.DestinationPath)
					if templateName != "index.html" {
						templateName = "page.html"
					}

					html, err := app.TemplateCache.Execute(templateName, templateData)
					if err != nil {
						return err
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
			}

			return nil
		},
	}
	return buildCmd
}
