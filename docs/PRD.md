# Product Requirements Document — Static Site Generator

**Working name:** `yourssg` (placeholder)
**Status:** Draft
**Scope of this document:** Requirements and phased delivery plan for a small, single-binary static site generator written in Go. A companion implementation guide covers the *how*; this document defines the *what* and *when*.

---

## 1. Summary

A static site generator (SSG) written in Go that converts markdown content, HTML templates, and CSS into a static website at build time. It is intentionally smaller and simpler than general-purpose generators such as Hugo: it targets content-driven sites (blogs, project pages, documentation) and is driven entirely by configuration so that a single binary can build any number of independent sites with no code changes.

The tool ships as one self-contained, cross-platform binary with no runtime dependencies. It is distributable as a CLI for use in CI (GitHub Actions) and produces output deployable to any static host.

---

## 2. Problem statement

General-purpose static site generators are powerful but large, with broad feature surfaces that most content sites never use. Teams and individuals who want a minimal, understandable tool — markdown in, static HTML out, with sensible theming and code highlighting — are left choosing between heavyweight generators and hand-rolled scripts.

This project fills that gap with a purpose-built generator that covers the common needs of a content site (markdown, templates, CSS, tags, series, syntax highlighting, theming) while remaining a single dependency-free binary that is easy to reason about, distribute, and reuse across projects.

---

## 3. Goals

| ID | Goal |
|----|------|
| G1 | Build static HTML sites from markdown content, HTML templates, and CSS at build time. |
| G2 | Ship as a single, self-contained, cross-platform static binary with no runtime dependencies. |
| G3 | Be fully config-driven, so one binary builds many sites with no code changes. |
| G4 | Provide selectable built-in themes, with per-template site overrides. |
| G5 | Support config-driven theming: fonts (heading/body/mono), color palette, default light/dark mode, and syntax-highlight theme. |
| G6 | Support tags (unordered) and series (ordered, with prev/next) taxonomies, plus computed reading time. |
| G7 | Provide built-in syntax highlighting with configurable light/dark themes. |
| G8 | Provide a development server with live reload. |
| G9 | Be distributable as a CLI usable in GitHub Actions across multiple site repositories. |
| G10 | Deploy cleanly to GitHub Pages, Cloudflare Pages, Vercel, and generic static hosts. |

---

## 4. Non-goals (v1)

The following are explicitly out of scope for the first version. Several are candidate future work (Section 11).

- Dynamic or server-rendered features; any runtime process or database.
- Full feature parity with general-purpose generators.
- Pagination of list, tag, or series pages.
- Incremental or cached builds.
- Shortcodes or content render hooks.
- Multiple built-in themes with *differing layouts* (theme **selection** is in scope; divergent layouts per theme are deferred).
- Internationalization / multi-language content.
- A plugin system.
- Multiple series membership per page (one series per page in v1).

---

## 5. Target users and use cases

**Users:** Developers building content-driven sites who want a minimal, predictable generator they can fully understand and reuse.

**Primary use cases:**
- Technical blogs requiring code syntax highlighting and multi-part series.
- Small project, marketing, or landing sites.
- Lightweight documentation sites.

---

## 6. Architecture overview

The system separates an **engine** (library) from a thin **CLI** (binary) that calls it.

**Build pipeline:** discover content → parse (frontmatter + markdown) → build content model → render templates → process assets → write output. An optional development server adds file watching, rebuild, and live reload over the same pipeline.

**Package layout (indicative):**

```
yourssg/
  cmd/yourssg/        # CLI entrypoint, command/flag parsing
  internal/
    config/           # load + validate configuration (Viper)
    content/          # discover, parse, build the content model
    render/           # template loading + execution
    assets/           # static + generated asset handling
    build/            # pipeline orchestrator
    server/           # dev server: watch + live reload
  themes/             # built-in themes, embedded via go:embed themes/*
    default/
    minimal/
```

**Model-before-render barrier:** Series require each member page to know its neighbours (prev/next), so the complete content model must be assembled and sorted before any page is rendered. This barrier is a hard sequencing rule between the model and render stages. It also makes the model effectively read-only during rendering, which permits safe parallel rendering as a later optimization.

---

## 7. Functional requirements

### 7.1 Content pipeline

| ID | Requirement |
|----|-------------|
| FR-1 | Discover all markdown files under `content/`. |
| FR-2 | Parse frontmatter (TOML or YAML) into a typed struct. Schema: `title`, `date`, `draft`, `tags`, `slug`, `description`, `series`, `series_order`. |
| FR-3 | Convert the markdown body to HTML using goldmark with GFM extensions. |
| FR-4 | Compute derived fields per page: word count and reading time (word count ÷ ~200 wpm, rounded up to whole minutes, minimum 1). |
| FR-5 | Exclude `draft: true` pages from production builds (included in dev mode unless flagged); exclude future-dated pages from production builds. |

### 7.2 Templating and themes

| ID | Requirement |
|----|-------------|
| FR-6 | Define reserved template names (`base`, `single`, `list`, `tag`, `series`, `404`) and document the data each receives. |
| FR-7 | Ship one or more built-in themes embedded in the binary; `theme.name` selects the active theme (default `"default"`). |
| FR-8 | Allow a site to override any reserved template by placing a same-named file in `layouts/`. Resolution order per template: site `layouts/` → selected built-in theme → error only if the theme itself lacks it. |
| FR-9 | Load the selected theme's templates first, then the site's `layouts/` into the same template set, so same-named templates (including named blocks) override. |
| FR-10 | Validate `theme.name` against the embedded set; validate that every reserved name resolves; warn (not fail) on a `layouts/` file matching no reserved name. |
| FR-11 | Each built-in theme bundles a default stylesheet; a site `static/` file at the same path overrides it. |

### 7.3 Theming configuration

| ID | Requirement |
|----|-------------|
| FR-12 | Expose the full parsed configuration to templates as site-level data (e.g. `.Site`). |
| FR-13 | Generate a `:root` CSS custom-property block from configuration (font families, colors, default mode); built-in theme CSS consumes these variables. |
| FR-14 | Implement light/dark via `:root` defaults, `[data-theme="…"]` overrides, and `prefers-color-scheme`, unified across page palette, syntax highlighting, and default mode. |
| FR-15 | Support font sources: `system` (map presets to system stacks; default), `google` (generate the stylesheet link), `self` (reference user-provided files in `static/fonts/` via generated `@font-face`). |

### 7.4 Syntax highlighting

| ID | Requirement |
|----|-------------|
| FR-16 | Provide built-in code highlighting via goldmark-highlighting (chroma), using **class-based** output (required for theme switching). |
| FR-17 | Support configurable light and dark highlight themes (`highlight.theme_light`, `highlight.theme_dark`). |
| FR-18 | Generate the highlight stylesheet (light + dark, scoped) at build time and inject it automatically into each page's `<head>`; require no template authoring by the user. |

### 7.5 Taxonomies: tags and series

| ID | Requirement |
|----|-------------|
| FR-19 | Group pages by `tags`; generate a listing page per tag rendered with `tag.html`. |
| FR-20 | Group pages by `series`; sort members by `series_order`; generate an ordered listing page per series rendered with `series.html` at `/series/<slug>/`. |
| FR-21 | Assign per-page series context: position, total, and prev/next pointers, exposed nil-safely (e.g. `HasPrev`/`HasNext`). |
| FR-22 | Filter drafts and future-dated pages **before** assigning series adjacency, so numbering stays contiguous and no link points to an unpublished page. |
| FR-23 | Ensure deterministic ordering: explicit integer `series_order` with tie-break by date then filename; sort map keys before iterating to render index pages. |
| FR-24 | Support exactly one series per page in v1. |

### 7.6 Assets and output

| ID | Requirement |
|----|-------------|
| FR-25 | Copy `static/` to the output verbatim; emit generated CSS (theme variables, highlight) as assets. |
| FR-26 | Produce pretty URLs (`/path/` served via `index.html`); support a configurable base URL including a path prefix (e.g. project pages under `/<repo>/`). |
| FR-27 | Generate `404.html`, `sitemap.xml`, `robots.txt`, and an RSS/Atom feed. |
| FR-28 | Emit SEO and Open Graph metadata using the absolute base URL. |

### 7.7 Development server (`serve`)

| ID | Requirement |
|----|-------------|
| FR-29 | Serve the output directory over HTTP with pretty-URL support. |
| FR-30 | Watch `content/`, `layouts/`, `static/`, and the config file — never the output directory; add subdirectories recursively, including newly created ones. |
| FR-31 | Debounce filesystem events and ignore editor swap files, dotfiles, and output-directory paths. |
| FR-32 | Perform a full rebuild on any relevant change (v1). |
| FR-33 | Implement live reload via a Server-Sent Events endpoint and a client broker; inject the reload script in dev only and in-flight (never into build output); send `Cache-Control: no-store` in dev. |
| FR-34 | Complete the rebuild before broadcasting reload; on build failure, keep serving the last good output and surface the error rather than reloading. |

### 7.8 CLI and configuration

| ID | Requirement |
|----|-------------|
| FR-35 | Provide commands: `build`, `serve`, `new post`, `new site`, `eject`, `themes` (list built-in themes), highlight-style listing (`highlight themes` or `themes --highlight`), and `version`. |
| FR-36 | Load configuration with **Viper**: support TOML/YAML, defaults via `SetDefault`, environment-variable overrides, and flag binding; unmarshal into a typed struct. |
| FR-37 | Fail validation with actionable messages that list valid options. |

### 7.9 Distribution and deployment

| ID | Requirement |
|----|-------------|
| FR-38 | Produce cross-compiled release binaries (linux/macOS/windows × amd64/arm64) via GoReleaser, attached to GitHub Releases. |
| FR-39 | Provide a composite GitHub Action wrapping the binary, versioned with SemVer plus a moving major tag, usable in a site repo with a single `uses:` line. |
| FR-40 | Document and/or provide deployment workflows for GitHub Pages, Cloudflare Pages, Vercel, and generic static hosts. |
| FR-41 | Stamp version/commit/date via build flags and expose them through `version`. |

---

## 8. Non-functional requirements

| ID | Requirement |
|----|-------------|
| NFR-1 | Single static binary, cross-platform, no cgo, no runtime dependencies (no Node or other interpreter). |
| NFR-2 | Deterministic, reproducible output for identical input, so generated sites diff cleanly and support golden-file testing. |
| NFR-3 | Fast full builds for small-to-medium sites (target: sub-second for a typical blog). |
| NFR-4 | Clear, actionable errors; content and template failures report the offending file (and line where feasible). |
| NFR-5 | Testable: golden-file tests for end-to-end build output; table-driven tests for parsing and edge cases. |
| NFR-6 | Zero-config default: a content-only directory builds a presentable site using built-in theme and defaults. |

---

## 9. Configuration specification

Configuration is loaded with Viper. Every key has a built-in default (`SetDefault`) so a zero-config site builds; environment variables override file values (useful for environment-specific base URLs in CI); the result is unmarshalled into a typed struct.

**Representative schema (TOML; design final keys during implementation):**

```toml
[site]
title       = "My Site"
base_url    = "https://example.com"   # override per environment via env var
author      = "Author Name"
description = "Site description"

[fonts]
heading = "sans"      # named preset, family string, or web-font name
body    = "serif"
mono    = "mono"
source  = "system"    # system | google | self

[theme]
name   = "default"    # selects an embedded built-in theme
mode   = "system"     # light | dark | system
accent = "#3b82f6"

[highlight]
theme_light  = "github"
theme_dark   = "github-dark"
line_numbers = false

[build]
output  = "public"
drafts  = false       # include drafts (dev convenience)
```

**Validation at load:** `theme.name` ∈ embedded themes; `theme.mode` ∈ {light, dark, system}; highlight themes ∈ chroma styles; `fonts.source` ∈ {system, google, self}; font presets ∈ the named set. Invalid values fail with a message listing valid options.

**Environment overrides:** at minimum `base_url` must be overridable by environment variable to support production vs. preview vs. project-pages path-prefixed builds in CI.

---

## 10. Milestones and phased roadmap

Phases are carried over from the implementation guide and expanded with objective, deliverables, acceptance criteria, and dependencies. **v1 = Phases 0–7.** Phases are ordered by dependency, not priority: the dev server (a v1 must-have) appears at Phase 4 because it requires a working build to serve.

### Phase 0 — Skeleton
- **Objective:** Establish the project structure and the `build` command's plumbing.
- **Deliverables:** `cmd/yourssg`; a `build` command that loads configuration (Viper) and walks `content/`, logging discovered files. Engine/CLI separation in place.
- **Acceptance criteria:** Running `build` on a sample directory loads config with defaults and lists every content file. No output is produced yet.
- **Dependencies:** None.

### Phase 1 — Markdown and frontmatter
- **Objective:** Parse a single content file end to end.
- **Deliverables:** Frontmatter split and decoding into the typed schema; markdown body conversion to HTML via goldmark (GFM).
- **Acceptance criteria:** A sample file yields correctly decoded metadata and valid HTML for its body. Malformed frontmatter produces a clear, file-attributed error.
- **Dependencies:** Phase 0.

### Phase 2 — Templates and write
- **Objective:** Render a page through templates and write real output.
- **Deliverables:** Built-in theme loading (selected theme) plus site override resolution; render `base`/`single` against a page; write an HTML file to the output directory. Parsed config passed to templates as `.Site` from the outset.
- **Acceptance criteria:** A single content file produces a viewable HTML page in `public/` that reflects site-level config values. With no site templates, the built-in theme renders it.
- **Dependencies:** Phases 1, and theme loading (FR-6 to FR-9).

### Phase 3 — Whole tree and assets
- **Objective:** Build a complete, browsable site.
- **Deliverables:** Render all content to correct output paths; copy `static/` verbatim; base-URL handling and pretty URLs.
- **Acceptance criteria:** A multi-page sample builds to a fully navigable static site with working internal links and assets, both at a root domain and under a path prefix.
- **Dependencies:** Phase 2.

### Phase 4 — Development server with live reload (v1 must-have)
- **Objective:** Provide fast local iteration.
- **Deliverables:** `serve` command: HTTP file server, file watcher (debounced, excluding output dir, recursive incl. new dirs), full rebuild on change, SSE-based live reload with dev-only in-flight script injection and `no-store` caching.
- **Acceptance criteria:** Editing content, templates, static files, or config triggers exactly one rebuild and a browser reload after the rebuild completes. A failed build keeps the last good output and shows the error without reloading.
- **Dependencies:** Phase 3.

### Phase 5 — Content model: tags, series, reading time
- **Objective:** Implement taxonomies and computed page data behind the model-before-render barrier.
- **Deliverables:** Full content model assembled before rendering; tag grouping and tag listing pages; series grouping, ordering, adjacency (prev/next, position/total) and series listing pages; reading time and word count; draft/future filtering applied before adjacency.
- **Acceptance criteria:** Tag and series pages render correctly; series member pages show accurate prev/next and "Part N of M"; excluding a draft mid-series renumbers the visible set with no dangling links; output is deterministic across repeated builds.
- **Dependencies:** Phase 3 (rendering); independent of Phase 4.

### Phase 6 — Highlighting, config-driven theming, web essentials
- **Objective:** Complete presentation and standard site outputs.
- **Deliverables:** Built-in syntax highlighting (chroma, class-based) with generated, auto-injected light/dark stylesheet; generated `:root` theme-variable block (fonts, colors, mode) and font loading; heading anchors and TOC; RSS/Atom, `sitemap.xml`, `robots.txt`, `404.html`; SEO/Open Graph metadata.
- **Acceptance criteria:** Code blocks are highlighted and switch with light/dark; changing fonts/colors/mode/highlight theme in config changes the rendered site with no template edits; feed, sitemap, robots, and 404 are generated and valid.
- **Dependencies:** Phases 3 and 5.

### Phase 7 — CLI polish (v1 complete)
- **Objective:** Round out the command surface and scaffolding.
- **Deliverables:** `new post` and `new site` scaffolding; `eject` (copies the selected theme's templates into `layouts/`); theme/highlight listing commands with disambiguated naming; improved help and flag handling.
- **Acceptance criteria:** A new site can be scaffolded and built; `eject` yields editable copies of the active theme; listing commands report available built-in themes and highlight styles distinctly.
- **Dependencies:** Phases 2, 6.

### Phase 8 — Distribution
- **Objective:** Make the tool consumable across repositories.
- **Deliverables:** GoReleaser cross-compiled binaries and GitHub Releases; a composite GitHub Action wrapping the binary; SemVer + moving major tag; version stamping.
- **Acceptance criteria:** A site repository can build via a single-line `uses:` reference to the action; released binaries run on all target OS/arch combinations; `version` reports the stamped build.
- **Dependencies:** Phase 7 (stable CLI).

### Phase 9 — Deployment targets
- **Objective:** Document and enable deployment to common hosts.
- **Deliverables:** Reference workflows for GitHub Pages, Cloudflare Pages, and Vercel; guidance for generic static hosts; base-URL handling per environment.
- **Acceptance criteria:** A site builds in CI and deploys successfully to each documented host, including correct path-prefix handling for project pages.
- **Dependencies:** Phase 8.

### Phase 10 — Optimizations (post-v1)
- **Objective:** Improve performance and output where measurement justifies it.
- **Deliverables (candidates):** Concurrent rendering over the frozen model; CSS/HTML minification; asset fingerprinting/cache-busting; incremental builds.
- **Acceptance criteria:** Each optimization preserves identical (or intentionally improved) output and is justified by measured gains; concurrency introduces no data races.
- **Dependencies:** A complete, correct v1 (Phases 0–9).

---

## 11. Open questions and future work

- **Multiple themes with differing layouts.** The selection mechanism is in v1; authoring multiple distinct theme layouts is deferred. How many built-in themes ship initially is open (one vs. a small set).
- **Pagination** for list, tag, and series pages.
- **Incremental builds**, which require modelling the content/template dependency graph.
- **Asset pipeline**: minification and fingerprinting.
- **Shortcodes / render hooks** for reusable in-content components.
- **Multiple series membership** per page.
- **Default stylesheet polish level** for built-in themes (minimal vs. opinionated).

---

## 12. Success criteria (v1 acceptance)

The first version is considered complete when all of the following hold:

1. A content-only directory builds a presentable site with zero templates and zero configuration.
2. A site can override any template, select a built-in theme, and configure fonts, color palette, default mode, and syntax-highlight theme through configuration alone.
3. Tags and series render correctly, with accurate series adjacency and correct draft/future-dated exclusion.
4. `serve` live-reloads on changes to content, templates, static files, and configuration, and degrades gracefully on build errors.
5. The tool is distributable via a single-line GitHub Action and deploys to GitHub Pages, Cloudflare Pages, and Vercel.
6. The deliverable is a single static, cross-platform binary with no runtime dependencies.

