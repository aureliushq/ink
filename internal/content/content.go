package content

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"strings"

	"github.com/charmbracelet/log"
)

type Content struct {
	Frontmatter Frontmatter
	Body        string
	Path        string
}

func NewContent() Content {
	return Content{}
}

func (content *Content) ReadFile() error {
	frontmatterLines := []string{}
	bodyLines := []string{}

	file, err := os.OpenFile(content.Path, os.O_RDONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

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
	content.Body = body

	return nil
}

func DiscoverContentFiles(contentDir string, logger *log.Logger) []string {
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
	}
	fileSystem := os.DirFS(cwd)

	paths := []string{}
	fs.WalkDir(fileSystem, contentDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			logger.Fatal("Error while reading files", "dir", cwd)
		}
		if !d.IsDir() {
			paths = append(paths, path)
		}
		return nil
	})

	return paths
}
