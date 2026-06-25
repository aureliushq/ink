#!/bin/sh
# install.sh — install the ink static site generator.
#
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/aureliushq/ink/main/install.sh | sh
#
# Environment variables:
#   INK_VERSION   release tag to install (e.g. v1.2.3). Default: latest release.
#   INK_INSTALL   directory to install into. Default: /usr/local/bin
#                 (falls back to $HOME/.local/bin if not writable).

set -eu

REPO="aureliushq/ink"
BINARY="ink"

err() {
	echo "error: $*" >&2
	exit 1
}

# --- detect OS -------------------------------------------------------------
os="$(uname -s)"
case "$os" in
	Linux) os="linux" ;;
	Darwin) os="darwin" ;;
	*) err "unsupported OS: $os (use the GitHub releases page for Windows)" ;;
esac

# --- detect arch -----------------------------------------------------------
arch="$(uname -m)"
case "$arch" in
	x86_64 | amd64) arch="amd64" ;;
	arm64 | aarch64) arch="arm64" ;;
	*) err "unsupported architecture: $arch" ;;
esac

# --- pick a downloader -----------------------------------------------------
if command -v curl >/dev/null 2>&1; then
	dl() { curl -fsSL "$1"; }
	dlo() { curl -fsSL -o "$2" "$1"; }
elif command -v wget >/dev/null 2>&1; then
	dl() { wget -qO- "$1"; }
	dlo() { wget -qO "$2" "$1"; }
else
	err "need curl or wget to download ink"
fi

# --- resolve version -------------------------------------------------------
version="${INK_VERSION:-}"
if [ -z "$version" ]; then
	echo "Resolving latest release..."
	version="$(dl "https://api.github.com/repos/${REPO}/releases/latest" |
		grep '"tag_name"' | head -n1 | cut -d'"' -f4)"
	[ -n "$version" ] || err "could not determine the latest release"
fi
num="${version#v}"

asset="${BINARY}_${num}_${os}_${arch}"
url="https://github.com/${REPO}/releases/download/${version}/${asset}"

# --- download --------------------------------------------------------------
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT

echo "Downloading ${asset} (${version})..."
dlo "$url" "$tmp/$BINARY" || err "download failed: $url"
chmod +x "$tmp/$BINARY"

# --- verify checksum (best effort) -----------------------------------------
if command -v sha256sum >/dev/null 2>&1 || command -v shasum >/dev/null 2>&1; then
	if dlo "https://github.com/${REPO}/releases/download/${version}/checksums.txt" "$tmp/checksums.txt" 2>/dev/null; then
		expected="$(grep " ${asset}\$" "$tmp/checksums.txt" | awk '{print $1}')"
		if [ -n "$expected" ]; then
			if command -v sha256sum >/dev/null 2>&1; then
				actual="$(sha256sum "$tmp/$BINARY" | awk '{print $1}')"
			else
				actual="$(shasum -a 256 "$tmp/$BINARY" | awk '{print $1}')"
			fi
			[ "$expected" = "$actual" ] || err "checksum mismatch for $asset"
			echo "Checksum verified."
		fi
	fi
fi

# --- install ---------------------------------------------------------------
dest="${INK_INSTALL:-/usr/local/bin}"
if [ ! -d "$dest" ] || [ ! -w "$dest" ]; then
	if [ "${INK_INSTALL:-}" = "" ]; then
		dest="$HOME/.local/bin"
		mkdir -p "$dest"
	else
		err "install dir not writable: $dest"
	fi
fi

mv "$tmp/$BINARY" "$dest/$BINARY"
echo "Installed ink to $dest/$BINARY"

if ! command -v ink >/dev/null 2>&1; then
	echo
	echo "Note: $dest is not on your PATH. Add it with:"
	echo "  export PATH=\"$dest:\$PATH\""
fi

"$dest/$BINARY" version || true
