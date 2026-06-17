package content

import (
	"bufio"
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path"
	"strings"

	"github.com/aureliushq/ink/internal/config"
	"github.com/charmbracelet/log"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/renderer/html"
)

type Content struct {
	Frontmatter     Frontmatter
	SourcePath      string
	DestinationPath string
	Slug            string
	ShouldBuild     bool
	HTMLBody        string
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

	dir, fileName := path.Split(content.SourcePath)
	fileExt := path.Ext(content.SourcePath)
	slug := strings.Replace(fileName, fileExt, "", 1)
	if slug == "index" {
		content.Slug = path.Join(strings.Replace(dir, buildConfig.ContentDir, "", 1), "/")
	} else {
		content.Slug = path.Join(strings.Replace(dir, buildConfig.ContentDir, "", 1), slug)
	}
	content.DestinationPath = path.Join(buildConfig.OutputDir, fmt.Sprintf("%s.%s", slug, "html"))

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
		goldmark.WithRendererOptions(
			html.WithHardWraps(),
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
