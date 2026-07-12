package renderer

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"

	"github.com/aureliushq/ink/internal/config"
	"github.com/charmbracelet/log"
)

var TEMPLATE_RESERVED_NAMES = []string{
	"base.html",
	"index.html",
	"page.html",
	"list.html",
	"series.html",
	"series_list.html",
	"single.html",
	"tags.html",
}

type TemplateCache struct {
	Files map[string]*template.Template
}

func NewTemplateCache() *TemplateCache {
	return &TemplateCache{
		Files: map[string]*template.Template{},
	}
}

type TemplateData struct {
	Config      *config.Config
	Title       string
	Description string
	Subtitle    string
	PageURL     string
	BasePath    string
	Slug        string
	Tags        []string
	Content     template.HTML
	Items       []TemplateData
	ItemOrder   int
	TotalItems  int
}

func NewTemplateData(cfg *config.Config) TemplateData {
	return TemplateData{
		Config:   cfg,
		BasePath: BasePath(cfg.Site.BaseURL),
	}
}

func BasePath(path string) string {
	u, err := url.Parse(path)
	if err != nil {
		return ""
	}
	return strings.TrimSuffix(u.Path, "/")
}

func PageURL(baseURL, slug string) string {
	baseURL = strings.TrimSuffix(baseURL, "/")
	slug = strings.Trim(slug, "/")
	if slug == "" {
		return baseURL + "/"
	}
	return baseURL + "/" + slug + "/"
}

func (tc *TemplateCache) Setup(cfg *config.Config, themesFS embed.FS) error {
	templatePath := path.Join("themes", cfg.Theme.Name, "*.html")
	pages, err := fs.Glob(themesFS, templatePath)
	if err != nil {
		return err
	}

	if len(pages) == 0 {
		availableThemes := []string{}
		entries, err := os.ReadDir("themes")
		if err != nil {
			return err
		}
		for _, entry := range entries {
			if entry.IsDir() {
				availableThemes = append(availableThemes, entry.Name())
			}
		}
		return fmt.Errorf("theme \"%s\" not found. Available themes: %v", cfg.Theme.Name, availableThemes)
	}

	for _, page := range pages {
		name := filepath.Base(page)

		if !slices.Contains(TEMPLATE_RESERVED_NAMES, name) {
			return fmt.Errorf("template name doesn't match reserved names: %s", name)
		}

		patterns := []string{
			path.Join("themes", cfg.Theme.Name, "base.html"),
			path.Join("themes", cfg.Theme.Name, "partials", "*.html"),
			page,
		}

		ts, err := template.New(name).ParseFS(themesFS, patterns...)
		if err != nil {
			return err
		}

		tc.Files[name] = ts
	}

	for _, name := range TEMPLATE_RESERVED_NAMES {
		if _, ok := tc.Files[name]; !ok {
			return fmt.Errorf("theme %q is missing, required template: %s", cfg.Theme.Name, name)
		}
	}

	return nil
}

func (tc *TemplateCache) Overrides(cfg *config.Config, logger *log.Logger) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	fileSystem := os.DirFS(cwd)

	templatePath := path.Join("layouts", "*.html")
	pages, err := fs.Glob(fileSystem, templatePath)
	if err != nil {
		return err
	}

	for _, page := range pages {
		name := filepath.Base(page)

		base, ok := tc.Files[name]
		if !ok {
			logger.Warnf("layout is not in template cache: %s", name)
			continue
		}

		if name == "base.html" {
			for pageName, pageTemplate := range tc.Files {
				ts, err := pageTemplate.Clone()
				if err != nil {
					return err
				}

				if _, err := ts.ParseFS(fileSystem, page); err != nil {
					return err
				}

				tc.Files[pageName] = ts
			}
			continue
		}

		ts, err := base.Clone()
		if err != nil {
			return err
		}

		_, err = ts.ParseFS(fileSystem, page)
		if err != nil {
			return err
		}

		tc.Files[name] = ts
	}

	return nil
}

func (tc *TemplateCache) Execute(name string, templateData TemplateData) (string, error) {
	html := new(bytes.Buffer)

	t, ok := tc.Files[name]
	if !ok {
		return "", fmt.Errorf("template not found in cache: %s", name)
	}

	if err := t.ExecuteTemplate(html, "base", templateData); err != nil {
		return "", err
	}

	return html.String(), nil
}
