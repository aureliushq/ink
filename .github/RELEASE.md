# Distribution (Phase 8)

How ink is released, versioned, and consumed across site repositories.

## Releasing

Releases are cut by pushing a SemVer tag. GoReleaser cross-compiles the matrix
(linux/macOS/windows × amd64/arm64) and attaches the raw binaries plus a
`checksums.txt` to a GitHub Release. Assets are named
`ink_<version>_<os>_<arch>` (Windows binaries get a `.exe` suffix).

```sh
git tag v1.2.3
git push origin v1.2.3
```

The [release workflow](../.github/workflows/release.yml):

1. Runs GoReleaser (`release --clean`) — see [.goreleaser.yaml](../.goreleaser.yaml).
2. Force-moves the matching major tag (e.g. `v1`) to the release commit so
   consumers can pin a major version.

Version, commit, and build date are stamped into the binary via `-ldflags -X`
(`main.version`, `main.commit`, `main.date`) and surfaced through both:

```sh
ink version     # multi-line: version / commit / built
ink --version   # fang-rendered single line (also -v)
```

## Installing

End users install via the [install script](../install.sh) (Linux/macOS), a
prebuilt binary from the releases page, or `go install`. See the
[README](../README.org) for details.

```sh
curl -fsSL https://raw.githubusercontent.com/aureliushq/ink/main/install.sh | sh
```

The script detects OS/arch, downloads the matching binary, verifies it against
`checksums.txt`, and installs to `/usr/local/bin` (or `$HOME/.local/bin`).
Override with `INK_VERSION` and `INK_INSTALL`.

## Using the GitHub Action

Any site repository builds with a single `uses:` line. The composite
[action.yml](../action.yml) downloads the prebuilt release binary matching the
runner OS/arch and runs `ink build` from the site root.

```yaml
- uses: aureliushq/ink@v1
  with:
    version: latest          # release tag (e.g. v1.2.3) or "latest"
    method: binary           # "binary" (default) or "go-install"
    args: ""                 # extra args passed to `ink build`
    working-directory: .     # site root
```

`method: go-install` installs via `go install github.com/aureliushq/ink@<version>`
instead of downloading a release asset (requires Go on the runner).

## Local checks

```sh
go run github.com/goreleaser/goreleaser/v2@latest check                       # validate config
go run github.com/goreleaser/goreleaser/v2@latest build --snapshot --clean    # dry-run build
```
