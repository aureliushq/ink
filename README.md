# Ink

Ink is yet another static site generator, cuz what's one more. Write content in
markdown, bring your own HTML+CSS templates, deploy anywhere. Supports
CommonMark and GFM. Comes with syntax highlighting, footnotes and margin notes,
and more out-of-the-box.

# Features

- [x] **Markdown content**
- [x] **TOML/YAML frontmatter**
- [x] **Bring your own templates**
- [x] **Built-in themes**
- [x] **Collections**
- [x] **Drafts**
- [x] **Local dev server**
- [x] **Single dependency-free binary**
- [x] **CI-friendly**
- [ ] **Series**
- [ ] **Tags**
- [ ] **Syntax highlighting**
- [ ] **Footnotes and margin notes**
- [ ] **Live reload**
- [ ] **RSS feed**
- [ ] **SEO meta tags**
- [ ] **Open Graph & Twitter cards**
- [ ] **Sitemap**
- [ ] **robots.txt**
- [ ] **Structured data**

# Installation

## Install script (Linux / macOS)

```sh
curl -fsSL https://raw.githubusercontent.com/aureliushq/ink/main/install.sh | sh
```

Install a specific version or change the install directory with environment
variables:

```sh
INK_VERSION=v1.2.3 INK_INSTALL="$HOME/.local/bin" \
  curl -fsSL https://raw.githubusercontent.com/aureliushq/ink/main/install.sh | sh
```

## Prebuilt binaries

Download the binary for your platform from the
[releases page](https://github.com/aureliushq/ink/releases), then make it
executable and put it on your `PATH`:

```sh
chmod +x ink_*_linux_amd64
mv ink_*_linux_amd64 /usr/local/bin/ink
```

On Windows, download the `.exe` and place it somewhere on your `PATH`.

## Go install

```sh
go install github.com/aureliushq/ink@latest
```

# Usage

## Build a site

Run from your site's root directory (where `ink.toml` lives):

```sh
ink build
```

This reads `content/`, applies your templates, and writes the static site to the
output directory (`public/` by default).

## Serve locally (live reload coming soon)

```sh
ink serve
```

Serves the site at `http://localhost:8782`. Override the host and port:

```sh
ink serve --host 0.0.0.0 --port 3000
```

## Version

```sh
ink version      # version, commit, and build date
ink --version    # short version line (also: ink -v)
```

# GitHub Action

Build your site in CI with a single step. The action downloads the prebuilt
binary for the runner and runs `ink build` from your repo root.

```yaml
- uses: aureliushq/ink@v1
  with:
    version: latest # release tag (e.g. v1.2.3) or "latest"
    method: binary # "binary" (default) or "go-install"
    args: "" # extra args passed to `ink build`
    working-directory: . # site root
```

See [RELEASE.md](.github/RELEASE.md) for the release and distribution details.

# Documentation

See [DOCUMENTATION.md](DOCUMENTATION.md) for the full guide on configuring Ink,
writing content, collections, series, themes/templates, and deployment.
