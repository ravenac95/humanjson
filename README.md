# humanjson

A stupidly simple CLI to convert JSONC / HuJSON (JSON with comments and
trailing commas) into normal JSON on stdout. Available as a single static
binary for macOS and Linux.

Built on top of [`tailscale/hujson`](https://github.com/tailscale/hujson).

## Install

Download the archive for your platform from the
[Releases](../../releases) page, untar it, and put the `hujson` binary on
your `PATH`:

```sh
tar -xzf hujson_v0.1.0_linux_amd64.tar.gz
sudo install -m 0755 hujson /usr/local/bin/hujson
```

Each release ships archives for `linux/amd64`, `linux/arm64`,
`darwin/amd64`, and `darwin/arm64`, plus per-archive `.sha256` files and
a combined `SHA256SUMS` file.

## Usage

```sh
hujson [file ...]
```

With no arguments, `hujson` reads from stdin. Pass `-` to mean stdin
explicitly. Pass multiple files to convert each in order.

```sh
# pipe a JSONC file through to jq
cat config.jsonc | hujson | jq .

# convert a file to JSON
hujson config.jsonc > config.json
```

Comments and trailing commas are removed; the output is valid JSON. The
original whitespace structure is preserved (comments are replaced with
equivalent runs of whitespace), so diffs against the source stay readable.
Pipe through `jq` if you want it pretty-printed or compacted.

## Building

You need Go installed (matching the version pinned in `go.mod`).

```sh
go build -o hujson .
```

To produce a release-style cross-compiled tarball for any target:

```sh
./scripts/build.sh <goos> <goarch> [version] [output-dir]

./scripts/build.sh linux amd64
./scripts/build.sh darwin arm64 v0.1.0 dist
```

Output goes to `dist/` by default as
`hujson_<version>_<goos>_<goarch>.tar.gz` plus a `.sha256` companion.

## Cutting a release

Releases are cut manually via the **Release** GitHub Actions workflow.

1. Open the **Actions** tab and select the **Release** workflow.
2. Click **Run workflow** and provide:
   - **release_tag** — the tag to publish under (e.g. `v0.1.0`).
   - **draft / prerelease** — optional flags.
3. The workflow runs tests, builds for all four target platforms, and
   publishes a GitHub release with the resulting archives and checksums.
