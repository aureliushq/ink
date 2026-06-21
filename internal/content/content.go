package content

import (
	"bufio"
	"bytes"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/aureliushq/ink/internal/config"
	"github.com/charmbracelet/log"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/renderer/html"
)

type Content struct {
	Frontmatter     Frontmatter
	Collection      string
	DestinationPath string
	HTMLBody        string
	SourcePath      string
	Slug            string
	IsIndex         bool
	ShouldBuild     bool
}

func NewContent() Content {
	return Content{
		ShouldBuild: true,
	}
}

func (content *Content) Unmarshal(buildConfig config.BuildConfig) error {
	frontmatterLines := []string{}
	bodyLines := []string{}

	file, err := os.OpenFile(content.SourcePath, os.O_RDONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	dir := path.Dir(content.SourcePath)
	fileName := path.Base(content.SourcePath)
	contentDir := strings.Replace(dir, buildConfig.ContentDir, "", 1)
	contentDir = strings.Replace(path.Clean(contentDir), "/", "", 1)
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

	scanner := bufio.NewScanner(file)
	seenHR := false
	contentStart := false
	for scanner.Scan() {
		line := scanner.Text()
		if line == "---" && seenHR {
			contentStart = true
			continue
		}
		if line == "---" && !seenHR {
			seenHR = true
			continue
		}
		if !contentStart {
			frontmatterLines = append(frontmatterLines, line)
		} else {
			bodyLines = append(bodyLines, line)
		}
	}
	if err = scanner.Err(); err != nil {
		return err
	}

	frontmatter := NewFrontmatter()
	if err = frontmatter.Parse(frontmatterLines); err != nil {
		return err
	}

	body := strings.Join(bodyLines, "\n")

	content.Frontmatter = frontmatter
	html, err := convertToHTML(body)
	if err != nil {
		return err
	}
	content.HTMLBody = html

	if content.Frontmatter.Status != StatusNil && content.Frontmatter.Status == StatusDraft && !buildConfig.Drafts {
		content.ShouldBuild = false
	}

	return nil
}

func convertToHTML(md string) (string, error) {
	gm := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
		),
		goldmark.WithRendererOptions(
			html.WithUnsafe(),
		),
	)
	var buf bytes.Buffer
	if err := gm.Convert([]byte(md), &buf); err != nil {
		return "", err
	}
	return buf.String(), nil
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
