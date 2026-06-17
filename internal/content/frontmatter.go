package content

import (
	"strings"
	"time"
)

type Status int

const (
	StatusNil Status = iota
	StatusDraft
	StatusPublished
)

type ContentType int

const (
	ContentTypePage ContentType = iota
	ContentTypeCollection
	ContentTypeCollectionList
	ContentTypeCollectionItem
)

type Frontmatter struct {
	Title       string
	Subtitle    string
	Description string
	Tags        []string
	Status      Status
	CreatedAt   time.Time
	UpdatedAt   time.Time
	PublishedAt time.Time
	SeriesID    string
	SeriesOrder int
	ContentType ContentType
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
		case strings.HasPrefix(line, "subtitle"):
			parts := strings.SplitN(line, ":", 2)
			frontmatter.Subtitle = strings.TrimSpace(parts[1])
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
			default:
				frontmatter.Status = StatusNil
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
		case strings.HasPrefix(line, "series"):
			// TODO: handle series
			continue
		case strings.HasPrefix(line, "series_order"):
			// TODO: handle series_order
			continue
		case strings.HasPrefix(line, "content_type"):
			parts := strings.SplitN(line, ":", 2)
			switch parts[1] {
			case "page":
				frontmatter.ContentType = ContentTypePage
			case "collection_item":
				frontmatter.ContentType = ContentTypeCollectionItem
			default:
				frontmatter.ContentType = ContentTypeCollection
			}
		default:
			continue
		}
	}
	return err
}
