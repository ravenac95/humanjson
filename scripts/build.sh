#!/usr/bin/env bash
# Cross-compile and package the hujson CLI for a target platform.
#
# Usage:
#   scripts/build.sh <goos> <goarch> [version] [output-dir]
#
# Examples:
#   scripts/build.sh linux amd64
#   scripts/build.sh darwin arm64 v0.1.0 dist
#
# Produces:
#   <output-dir>/hujson_<version>_<goos>_<goarch>.tar.gz
#   <output-dir>/hujson_<version>_<goos>_<goarch>.tar.gz.sha256

set -euo pipefail

if [ $# -lt 2 ]; then
  echo "usage: $0 <goos> <goarch> [version] [output-dir]" >&2
  exit 1
fi

GOOS="$1"
GOARCH="$2"
VERSION="${3:-dev}"
OUTDIR="${4:-dist}"

REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"

mkdir -p "$OUTDIR"
OUTDIR_ABS="$(cd "$OUTDIR" && pwd)"

WORKDIR="$(mktemp -d)"
trap 'rm -rf "$WORKDIR"' EXIT

BIN_NAME="hujson"
if [ "$GOOS" = "windows" ]; then
  BIN_NAME="hujson.exe"
fi
BIN_PATH="$WORKDIR/$BIN_NAME"

echo "Building hujson ${VERSION} for ${GOOS}/${GOARCH}..."
(
  cd "$REPO_ROOT"
  CGO_ENABLED=0 GOOS="$GOOS" GOARCH="$GOARCH" \
    go build \
      -trimpath \
      -ldflags="-s -w -X main.version=${VERSION}" \
      -o "$BIN_PATH" \
      .
)

SAFE_VERSION="$(printf '%s' "$VERSION" | tr -c 'A-Za-z0-9._-' '_')"
ARCHIVE_NAME="hujson_${SAFE_VERSION}_${GOOS}_${GOARCH}.tar.gz"
ARCHIVE_PATH="${OUTDIR_ABS}/${ARCHIVE_NAME}"

tar -czf "$ARCHIVE_PATH" -C "$WORKDIR" "$BIN_NAME"

(
  cd "$OUTDIR_ABS"
  if command -v sha256sum >/dev/null 2>&1; then
    sha256sum "$ARCHIVE_NAME" > "${ARCHIVE_NAME}.sha256"
  else
    shasum -a 256 "$ARCHIVE_NAME" > "${ARCHIVE_NAME}.sha256"
  fi
)

echo "Built: $ARCHIVE_PATH"
