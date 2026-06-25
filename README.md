---
author: Ilango
title: Ink - Yet another super simple static site generator
---

Ink is yet another super simple static site generator. Write content in
markdown, bring your own HTML+CSS templates, deploy anywhere. Supports
CommonMark and GFM. Comes with syntax highlighting, footnotes and margin
notes, and more out-of-the-box.

# Installation

## Install script (Linux / macOS)

```sh
curl -fsSL https://raw.githubusercontent.com/aureliushq/ink/main/install.sh | sh
```

Install a specific version or change the install directory with
environment variables:

```sh
INK_VERSION=v1.2.3 INK_INSTALL="$HOME/.local/bin" \
  curl -fsSL https://raw.githubusercontent.com/aureliushq/ink/main/install.sh | sh
```

## Prebuilt binaries

Download the binary for your platform from the [releases
page](https://github.com/aureliushq/ink/releases), then make it
executable and put it on your `PATH`{.verbatim}:

```sh
chmod +x ink_*_linux_amd64
mv ink_*_linux_amd64 /usr/local/bin/ink
```

On Windows, download the `.exe`{.verbatim} and place it somewhere on
your `PATH`{.verbatim}.

## Go install

```sh
go install github.com/aureliushq/ink@latest
```

# Usage

## Build a site

Run from your site\'s root directory (where `ink.toml`{.verbatim}
lives):

```sh
ink build
```

This reads `content/`{.verbatim}, applies your templates, and writes the
static site to the output directory (`public/`{.verbatim} by default).

## Serve locally with live reload

```sh
ink serve
```

Serves the site at `http://localhost:8782`{.verbatim} with live
reloading. Override the host and port:

```sh
ink serve --host 0.0.0.0 --port 3000
```

## Version

```sh
ink version      # version, commit, and build date
ink --version    # short version line (also: ink -v)
```

# GitHub Action

Build your site in CI with a single step. The action downloads the
prebuilt binary for the runner and runs `ink build`{.verbatim} from your
repo root.

``` yaml
- uses: aureliushq/ink@v1
  with:
    version: latest          # release tag (e.g. v1.2.3) or "latest"
    method: binary           # "binary" (default) or "go-install"
    args: ""                 # extra args passed to `ink build`
    working-directory: .     # site root
```

See [docs/DISTRIBUTION.md](docs/DISTRIBUTION.md) for the release and
distribution details.
