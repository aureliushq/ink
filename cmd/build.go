package cmd

import (
	"embed"
	"fmt"
	"html/template"
	"os"
	"path"
	"sort"
	"strings"

	"github.com/aureliushq/ink/internal/assets"
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
			if err := os.RemoveAll(app.Config.Build.OutputDir); err != nil {
				return err
			}

			allContent, err := build.ReadContent(app.Config.Build, app.Logger)
			if err != nil {
				return err
			}

			app.Logger.Infof("Total Content Found: %d", len(allContent))

			collections := map[string][]renderer.TemplateData{}
			collectionTags := map[string][]string{}
			for _, collection := range app.Config.Build.Collections {
				collectionItems := []renderer.TemplateData{}
				seenTags := map[string]struct{}{}
				for _, content := range allContent {
					if content.Collection == collection && !content.IsIndex {
						templateData := renderer.NewTemplateData(app.Config)
						templateData.Title = content.Frontmatter.Title
						templateData.Subtitle = content.Frontmatter.Subtitle
						templateData.Description = content.Frontmatter.Description
						templateData.CreatedAt = content.Frontmatter.CreatedAt
						templateData.UpdatedAt = content.Frontmatter.UpdatedAt
						templateData.PublishedAt = content.Frontmatter.PublishedAt
						templateData.Content = template.HTML(content.HTMLBody)
						templateData.PageURL = renderer.PageURL(app.Config.Site.BaseURL, content.Slug)
						templateData.Slug = path.Join(content.Slug)
						collectionItems = append(collectionItems, templateData)
						for _, tag := range content.Frontmatter.Tags {
							if _, ok := seenTags[tag]; ok {
								continue
							}
							seenTags[tag] = struct{}{}
							collectionTags[collection] = append(collectionTags[collection], tag)
						}
					}
				}
				sort.SliceStable(collectionItems, func(i, j int) bool {
					return collectionItems[i].PublishedAt.After(collectionItems[j].PublishedAt)
				})
				collections[collection] = collectionItems
			}

			series := map[string][]renderer.TemplateData{}
			seriesList := []renderer.TemplateData{}
			for _, content := range allContent {
				templateData := renderer.NewTemplateData(app.Config)
				templateData.Title = content.Frontmatter.Title
				templateData.Subtitle = content.Frontmatter.Subtitle
				templateData.Description = content.Frontmatter.Description
				templateData.CreatedAt = content.Frontmatter.CreatedAt
				templateData.UpdatedAt = content.Frontmatter.UpdatedAt
				templateData.PublishedAt = content.Frontmatter.PublishedAt
				templateData.Content = template.HTML(content.HTMLBody)
				templateData.PageURL = renderer.PageURL(app.Config.Site.BaseURL, content.Slug)
				templateData.Slug = path.Join(content.Slug)
				if content.IsSeries && !content.IsIndex {
					series[content.Frontmatter.SeriesID] = append(series[content.Frontmatter.SeriesID], templateData)
				} else if content.IsSeries && content.IsIndex && !strings.HasSuffix(content.Slug, "series") {
					seriesList = append(seriesList, templateData)
				}
			}

			app.Logger.Infof("Total Series Found: %d", len(series))

			fmt.Println("----------------------------------------")

			for _, content := range allContent {
				templateData := renderer.NewTemplateData(app.Config)
				templateData.Title = content.Frontmatter.Title
				templateData.Subtitle = content.Frontmatter.Subtitle
				templateData.Description = content.Frontmatter.Description
				templateData.CreatedAt = content.Frontmatter.CreatedAt
				templateData.UpdatedAt = content.Frontmatter.UpdatedAt
				templateData.PublishedAt = content.Frontmatter.PublishedAt
				templateData.Content = template.HTML(content.HTMLBody)
				templateData.PageURL = renderer.PageURL(app.Config.Site.BaseURL, content.Slug)

				app.Logger.Info(content.SourcePath)

				var templateName string
				switch {
				case content.IsSeries && content.IsIndex && strings.HasSuffix(content.Slug, "series"):
					templateData.Items = seriesList
					templateName = "series.html"
				case content.IsSeries && content.IsIndex && content.Frontmatter.SeriesID != "":
					templateData.Tags = content.Frontmatter.Tags
					templateData.Items = series[content.Frontmatter.SeriesID]
					templateData.TotalItems = len(series[content.Frontmatter.SeriesID])
					templateName = "series_list.html"
				case content.IsSeries && !content.IsIndex && content.Frontmatter.SeriesID != "":
					templateData.Tags = content.Frontmatter.Tags
					templateData.ItemOrder = content.Frontmatter.SeriesOrder
					templateData.TotalItems = len(series[content.Frontmatter.SeriesID])
					templateName = "single.html"
				case content.Collection != "" && content.IsIndex:
					templateData.Items = collections[content.Collection]
					templateData.Tags = collectionTags[content.Collection]
					templateName = "list.html"
				case content.Collection != "":
					templateData.Tags = content.Frontmatter.Tags
					templateName = "single.html"
				case content.IsIndex:
					templateName = "index.html"
				default:
					templateName = "page.html"
				}

				html, err := app.TemplateCache.Execute(templateName, templateData)
				if err != nil {
					return err
				}
				dir := path.Dir(content.DestinationPath)
				if err = os.MkdirAll(dir, 0755); err != nil {
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

			fmt.Println("----------------------------------------")

			if err := assets.Copy(app.Config, themesFS, app.Logger); err != nil {
				return err
			}

			return nil
		},
	}
	return buildCmd
}
