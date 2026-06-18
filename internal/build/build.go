package build

import (
	"embed"

	"github.com/aureliushq/ink/internal/config"
	"github.com/aureliushq/ink/internal/content"
	"github.com/aureliushq/ink/internal/renderer"
	"github.com/charmbracelet/log"
)

func InitConfig() (*config.Config, error) {
	cfg := config.NewConfig()
	if err := config.Setup(cfg); err != nil {
		return &config.Config{}, err
	}

	return cfg, nil
}

func InitTemplateCache(cfg *config.Config, logger *log.Logger, themesFS embed.FS) (*renderer.TemplateCache, error) {
	templateCache := renderer.NewTemplateCache()
	if err := templateCache.Setup(cfg, themesFS); err != nil {
		return &renderer.TemplateCache{}, err
	}
	if err := templateCache.Overrides(cfg, logger); err != nil {
		return &renderer.TemplateCache{}, err
	}
	return templateCache, nil
}

func ReadContent(buildConfig config.BuildConfig, logger *log.Logger) ([]content.Content, error) {
	paths, err := content.DiscoverFiles(buildConfig.ContentDir, logger)
	if err != nil {
		return []content.Content{}, err
	}
	allContent := []content.Content{}
	for _, path := range paths {
		data := content.NewContent()
		data.SourcePath = path

		err := data.Unmarshal(buildConfig)
		if err != nil {
			return []content.Content{}, err
		}
		allContent = append(allContent, data)
	}

	return allContent, nil
}
