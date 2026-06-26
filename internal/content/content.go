package content

import (
	"bytes"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/aureliushq/ink/internal/config"
	"github.com/charmbracelet/log"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	fm "go.abhg.dev/goldmark/frontmatter"
)

type Frontmatter struct {
	Title       string    `yaml:"title"`
	Subtitle    string    `yaml:"subtitle"`
	Description string    `yaml:"description"`
	Tags        []string  `yaml:"tags"`
	Status      string    `yaml:"status"`
	CreatedAt   time.Time `yaml:"created_at"`
	UpdatedAt   time.Time `yaml:"updated_at"`
	PublishedAt time.Time `yaml:"published_at"`
	SeriesID    string    `yaml:"series_id"`
	SeriesOrder int       `yaml:"series_order"`
}

func NewFrontmatter() Frontmatter {
	return Frontmatter{}
}

type Content struct {
	Frontmatter     Frontmatter
	Collection      string
	DestinationPath string
	HTMLBody        string
	SourcePath      string
	Slug            string
	SeriesID        string
	IsIndex         bool
	IsSeries        bool
	ShouldBuild     bool
}

func NewContent() Content {
	return Content{
		ShouldBuild: true,
	}
}

func (content *Content) Unmarshal(buildConfig config.BuildConfig) error {
	file, err := os.ReadFile(content.SourcePath)
	if err != nil {
		return err
	}

	dir := path.Dir(content.SourcePath)
	fileName := path.Base(content.SourcePath)
	contentDir := strings.Replace(dir, buildConfig.ContentDir, "", 1)
	contentDir = strings.Replace(path.Clean(contentDir), "/", "", 1)

	content.IsSeries = strings.HasPrefix(contentDir, "series")

	for _, collection := range buildConfig.Collections {
		match, err := filepath.Match(collection, contentDir)
		if err != nil {
			return err
		}
		if match {
			content.Collection = contentDir
		}
	}

	fileExt := path.Ext(content.SourcePath)
	name := strings.TrimSuffix(fileName, fileExt)
	content.IsIndex = name == "index"

	relDir := contentDir
	if relDir == "." {
		relDir = ""
	}

	var slug string
	if content.Collection != "" {
		if content.IsIndex {
			slug = content.Collection
		} else {
			slug = path.Join(content.Collection, name)
		}
	} else {
		if content.IsIndex {
			slug = relDir
		} else {
			slug = path.Join(relDir, name)
		}
	}
	content.Slug = slug

	if slug == "" {
		content.DestinationPath = path.Join(buildConfig.OutputDir, "index.html")
	} else {
		content.DestinationPath = path.Join(buildConfig.OutputDir, slug, "index.html")
	}

	gm := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			&fm.Extender{},
		),
		goldmark.WithRendererOptions(
			html.WithUnsafe(),
		),
	)

	var buf bytes.Buffer
	ctx := parser.NewContext()
	if err := gm.Convert(file, &buf, parser.WithContext(ctx)); err != nil {
		return err
	}

	frontmatterData := fm.Get(ctx)
	var frontmatter Frontmatter
	if err = frontmatterData.Decode(&frontmatter); err != nil {
		return err
	}

	content.Frontmatter = frontmatter
	content.HTMLBody = buf.String()

	if content.Frontmatter.Status != "" && content.Frontmatter.Status == "draft" && !buildConfig.Drafts {
		content.ShouldBuild = false
	}

	return nil
}

func DiscoverFiles(contentDir string, logger *log.Logger) ([]string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		logger.Fatal("Error when getting working directory", "dir", cwd)
		return []string{}, err
	}
	fileSystem := os.DirFS(cwd)

	paths := []string{}
	fs.WalkDir(fileSystem, contentDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			logger.Fatal("Error while reading files", "dir", cwd)
			return err
		}
		if !d.IsDir() {
			paths = append(paths, path)
		}
		return nil
	})

	return paths, nil
}
