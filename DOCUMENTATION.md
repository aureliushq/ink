# Ink Documentation

Ink is a small static site generator. You write content in Markdown, supply (or
inherit) HTML+CSS templates, and Ink renders a static site you can host
anywhere.

This guide covers everything you need to build a site with Ink today. For
installation instructions, see the [README](README.md).

## Table of contents

- [Quick start](#quick-start)
- [Project structure](#project-structure)
- [Configuration (`ink.toml`)](#configuration-inktoml)
- [Writing content](#writing-content)
  - [Frontmatter](#frontmatter)
  - [Markdown](#markdown)
  - [Drafts](#drafts)
- [Collections](#collections)
- [Series](#series)
- [Themes and templates](#themes-and-templates)
  - [Reserved template names](#reserved-template-names)
  - [Template data](#template-data)
  - [Overriding templates with `layouts/`](#overriding-templates-with-layouts)
- [Static assets](#static-assets)
- [CLI reference](#cli-reference)
- [Building and deploying](#building-and-deploying)

## Quick start

1. Create a project directory with an `ink.toml` config file at its root.
2. Add Markdown content under your content directory (e.g. `content/`).
3. Build the site:

   ```sh
   ink build
   ```

4. Preview locally:

   ```sh
   ink serve
   ```

   Then open <http://localhost:8782>.

The `demo/` directory in this repository is a complete, working example site —
the fastest way to see how everything fits together.

## Project structure

A typical Ink site looks like this:

```
my-site/
├── ink.toml            # site configuration
├── content/            # your Markdown content
│   ├── index.md        # home page
│   ├── articles/       # a collection
│   │   ├── index.md    # collection listing page
│   │   └── my-post.md
│   └── series/         # series (see below)
│       └── my-series/
│           ├── index.md
│           ├── part-1.md
│           └── part-2.md
├── layouts/            # optional template overrides
│   └── single.html
└── public/             # build output (generated)
```

Only `ink.toml` and a content directory are strictly required. If you don't
provide any templates, Ink uses its built-in default theme.

## Configuration (`ink.toml`)

Ink reads a single `ink.toml` file from the directory you run it in. It has
three sections:

```toml
[site]
title       = "My Site"
subtitle    = "A short tagline"
description = "What this site is about"
base_url    = "https://example.com"
author      = "Your Name"

[theme]
name = "default"        # selects an embedded built-in theme

[build]
collections = ["articles", "projects/*"]
content     = "content"
output      = "public"
drafts      = false
```

### `[site]`

| Key           | Description                                              |
| ------------- | -------------------------------------------------------- |
| `title`       | Site title (used in `<title>` and meta tags).            |
| `subtitle`    | Optional tagline.                                        |
| `description` | Default meta description.                                |
| `base_url`    | Absolute base URL; used to build canonical / page URLs.  |
| `author`      | Site author.                                             |

### `[theme]`

| Key    | Description                                          |
| ------ | --------------------------------------------------- |
| `name` | Name of the embedded built-in theme (e.g. `default`). |

### `[build]`

| Key           | Description                                                            |
| ------------- | --------------------------------------------------------------------- |
| `collections` | Glob patterns for directories treated as [collections](#collections). |
| `content`     | Path to the content directory.                                        |
| `output`      | Output directory for the generated site (default `public`).           |
| `drafts`      | When `true`, include draft content in the build.                      |

> Note: `base_url` is also used to compute the `BasePath` for asset links, so
> set it to the full URL where the site will be served (including any subpath).

## Writing content

Content lives in your configured `content` directory as Markdown (`.md`) files.
Each file becomes a page. A file's path determines its URL:

- `content/index.md` → `/`
- `content/now.md` → `/now/`
- `content/articles/my-post.md` → `/articles/my-post/`
- `content/articles/index.md` → `/articles/` (a collection listing page)

### Frontmatter

Each file may start with a YAML frontmatter block delimited by `---`:

```markdown
---
title: My First Post
subtitle: An optional subtitle
description: A short summary used in listings and meta tags.
tags:
  - go
  - tutorial
status: published
created_at: 2025-01-06
published_at: 2025-01-06
updated_at: 2025-01-06
series_id: my-series
series_order: 1
---

Your Markdown content goes here.
```

Supported frontmatter fields:

| Field          | Type       | Description                                              |
| -------------- | ---------- | -------------------------------------------------------- |
| `title`        | string     | Page title.                                              |
| `subtitle`     | string     | Optional subtitle.                                       |
| `description`  | string     | Summary shown in listings and meta tags.                |
| `tags`         | string[]   | Tags for the page.                                       |
| `status`       | string     | `published` or `draft`.                                  |
| `created_at`   | date       | Creation date.                                           |
| `updated_at`   | date       | Last-updated date.                                       |
| `published_at` | date       | Publish date.                                            |
| `series_id`    | string     | Series this page belongs to (see [Series](#series)).     |
| `series_order` | integer    | Position of this page within its series.                 |

### Markdown

Ink renders Markdown with [goldmark](https://github.com/yuin/goldmark) and the
GitHub Flavored Markdown (GFM) extension, so you get:

- CommonMark
- Tables
- Task lists
- Strikethrough
- Autolinks

Raw HTML in Markdown is allowed (rendered as-is).

### Drafts

Set `status: draft` to mark a page as a draft. Drafts are excluded from the
build unless `drafts = true` in `ink.toml`'s `[build]` section.

## Collections

A collection is a directory of related pages that also gets a listing page.
Register a collection by adding its directory (or a glob) to `collections` in
`ink.toml`:

```toml
[build]
collections = ["articles", "projects/*"]
```

- Pages inside a collection directory render with the `single.html` template.
- The collection's `index.md` renders with the `list.html` template and
  automatically receives all the collection's pages as `Items` to list.

For example, with `articles` registered:

```
content/articles/
├── index.md     → /articles/   (list page, lists all articles)
├── post-a.md    → /articles/post-a/
└── post-b.md    → /articles/post-b/
```

Globs like `projects/*` treat each matching subdirectory as a collection.

## Series

A series is ordered, multi-part content. Any content under the `series/`
directory is treated as a series. The convention is one directory per series:

```
content/series/
├── index.md                    → /series/   (lists every series)
└── build-your-own-redis/
    ├── index.md                → /series/build-your-own-redis/  (lists this series' posts)
    ├── redis-1-....md          → a part
    ├── redis-2-....md
    └── redis-3-....md
```

How it works:

- Each part sets `series_id` (the series identifier) and `series_order` (its
  position). Parts render with `single.html`.
- A series' `index.md` sets the same `series_id` and renders with `list.html`,
  automatically receiving that series' parts as `Items`.
- The top-level `series/index.md` renders with `list.html` and receives every
  series' index page as `Items`, producing an index of all series.

A minimal series index page:

```markdown
---
title: Build Your Own Redis in Go
description: Rebuild a minimal Redis-compatible server in Go, step by step.
series_id: build-your-own-redis-in-go
---

A short description of what the series covers.
```

And a part:

```markdown
---
title: "Part 1: Listening on a TCP Port"
description: Open a TCP socket and accept connections.
series_id: build-your-own-redis-in-go
series_order: 1
---

Part content here.
```

## Themes and templates

Ink ships with a built-in `default` theme, so a site needs zero templates to
build. Select a theme with `theme.name` in `ink.toml`. Themes are plain
`html/template` files with a small, fixed set of names.

### Reserved template names

A theme provides these templates:

| Template       | Used for                                                              |
| -------------- | -------------------------------------------------------------------- |
| `base.html`    | The HTML shell (`<head>`, layout) that every page is rendered into.  |
| `index.html`   | The site home page (`content/index.md`).                             |
| `page.html`    | Standalone pages not in a collection (e.g. `content/now.md`).        |
| `list.html`    | Listing pages: collection indexes and series indexes.               |
| `single.html`  | Individual collection pages and series parts.                       |

Each non-base template defines a `content` block, which `base.html` renders via
`{{block "content" .}}`. Shared snippets live in a theme's `partials/`
directory and are included with `{{template "partials/header" .}}`.

Which template a page uses is decided automatically:

| Page                                    | Template      |
| --------------------------------------- | ------------- |
| `content/index.md`                      | `index.html`  |
| Standalone page (no collection)         | `page.html`   |
| Collection `index.md`                   | `list.html`   |
| Collection page                         | `single.html` |
| `series/index.md`                       | `list.html`   |
| A series' `index.md`                    | `list.html`   |
| A series part                           | `single.html` |

### Template data

Every template receives a `TemplateData` value as `.`:

| Field         | Description                                                       |
| ------------- | ---------------------------------------------------------------- |
| `.Config`     | The full parsed config (`.Config.Site.Title`, etc.).            |
| `.Title`      | Page title.                                                      |
| `.Subtitle`   | Page subtitle.                                                   |
| `.Description`| Page description.                                                |
| `.PageURL`    | Absolute URL of the page.                                        |
| `.BasePath`   | URL path prefix derived from `base_url` (for asset links).      |
| `.Slug`       | The page's slug / path.                                          |
| `.Tags`       | The page's tags (on `single.html`).                             |
| `.Content`    | Rendered HTML body of the page.                                  |
| `.Items`      | Child pages to list (on `list.html`); each is a `TemplateData`. |

Example listing loop (from `list.html`):

```html
{{range .Items}}
  <a href="/{{.Slug}}">
    <span>{{.Title}}</span>
    <span>{{.Description}}</span>
  </a>
{{end}}
```

### Overriding templates with `layouts/`

To customize a theme without forking it, drop a file with a reserved name into a
`layouts/` directory at your site root. It overrides that template:

```
my-site/
└── layouts/
    └── single.html     # overrides the theme's single.html
```

Only files matching a reserved name take effect. If a file in `layouts/` matches
no reserved name, Ink logs a warning and ignores it (this usually means a typo,
e.g. `signle.html`).

## Static assets

Static files (CSS, images, fonts, favicons) are copied into the output
directory. If your site has its own static directory, it is used; otherwise the
selected theme's static assets are used. Reference them from templates relative
to `.BasePath`, e.g.:

```html
<link rel="stylesheet" href="{{.BasePath}}/static/main.css" />
```

## CLI reference

### `ink build`

Builds the static site. Reads the content directory, applies templates, copies
assets, and writes the result to the output directory (default `public/`). The
output directory is cleaned before each build.

```sh
ink build
```

### `ink serve`

Serves the built site locally.

```sh
ink serve
ink serve --host 0.0.0.0 --port 3000
```

| Flag     | Default     | Description           |
| -------- | ----------- | --------------------- |
| `--host` | `localhost` | Host to bind to.      |
| `--port` | `8782`      | Port to listen on.    |

### `ink version`

Prints the version, commit, and build date stamped into the binary.

```sh
ink version
ink --version   # short version line (also: ink -v)
```

## Building and deploying

`ink build` produces a fully static site in the output directory. Because the
output is just HTML, CSS, and assets, you can host it on any static host —
GitHub Pages, Netlify, Cloudflare Pages, S3, or your own server.

To build in CI, use the bundled GitHub Action:

```yaml
- uses: aureliushq/ink@v1
  with:
    version: latest          # release tag (e.g. v1.2.3) or "latest"
    method: binary           # "binary" (default) or "go-install"
    args: ""                 # extra args passed to `ink build`
    working-directory: .     # site root
```

Then deploy the generated output directory with your platform's preferred
deploy step.
