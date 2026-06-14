package content

import (
	"strings"
	"time"
)

type Status int

const (
	StatusDraft Status = iota
	StatusPublished
)

type Frontmatter struct {
	Title       string
	Description string
	Tags        []string
	Status      Status
	CreatedAt   time.Time
	UpdatedAt   time.Time
	PublishedAt time.Time
}

func NewFrontmatter() Frontmatter {
	return Frontmatter{}
}

func (frontmatter *Frontmatter) Parse(lines []string) error {
	var t time.Time
	var err error

	for _, line := range lines {
		switch {
		case strings.HasPrefix(line, "title"):
			parts := strings.SplitN(line, ":", 2)
			frontmatter.Title = strings.TrimSpace(parts[1])
		case strings.HasPrefix(line, "description"):
			parts := strings.SplitN(line, ":", 2)
			frontmatter.Description = strings.TrimSpace(parts[1])
		case strings.HasPrefix(line, "tags"):
			// TODO: handle tags
			// tags is a special case since we have to parse the lines after the key
			continue
		case strings.HasPrefix(line, "status"):
			parts := strings.SplitN(line, ":", 2)
			switch parts[1] {
			case "draft":
				frontmatter.Status = StatusDraft
			case "published":
				frontmatter.Status = StatusPublished
			}
		case strings.HasPrefix(line, "createdAt"):
			parts := strings.SplitN(line, ":", 2)
			t, err = time.Parse(time.DateOnly, strings.TrimSpace(parts[1]))
			frontmatter.CreatedAt = t
		case strings.HasPrefix(line, "updatedAt"):
			parts := strings.SplitN(line, ":", 2)
			t, err = time.Parse(time.DateOnly, strings.TrimSpace(parts[1]))
			frontmatter.UpdatedAt = t
		case strings.HasPrefix(line, "publishedAt"):
			parts := strings.SplitN(line, ":", 2)
			t, err = time.Parse(time.DateOnly, strings.TrimSpace(parts[1]))
			frontmatter.PublishedAt = t
		default:
			continue
		}
	}
	return err
}
