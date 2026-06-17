package renderer

import (
	"bytes"
	"fmt"
	"html/template"
	"io/fs"
	"os"
	"path"
	"path/filepath"

	"github.com/aureliushq/ink/internal/config"
)

type TemplateCache struct {
	Files map[string]*template.Template
}

func NewTemplateCache() *TemplateCache {
	return &TemplateCache{
		Files: map[string]*template.Template{},
	}
}

type TemplateData struct {
	SiteTitle    string
	SiteSubtitle string
	Title        string
	Description  string
	Subtitle     string
	PageURL      string
	Content      template.HTML
}

func NewTemplateData() TemplateData {
	return TemplateData{}
}

func (tc *TemplateCache) Setup(cfg *config.Config) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	fileSystem := os.DirFS(cwd)
	templatePath := path.Join("themes", cfg.Theme.Name, "*.html")
	pages, err := fs.Glob(fileSystem, templatePath)
	if err != nil {
		return err
	}

	for _, page := range pages {
		name := filepath.Base(page)

		patterns := []string{
			path.Join("themes", cfg.Theme.Name, "base.html"),
			path.Join("themes", cfg.Theme.Name, "partials", "header.html"),
			page,
		}

		ts := template.Must(template.New(name).ParseFS(fileSystem, patterns...))

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
