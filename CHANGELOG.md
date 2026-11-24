# Changelog

All notable changes to this project will be documented in this file.

Towncrier aggregates fragments from `changelog.d/` at release time.

<!-- towncrier release notes start -->

## v1.0.0 (2026-11-24)

### Breaking changes

- The command is now `pj`, the PointJungle CLI: install it with `brew install pointjungle/tap/pj`, the install script or `go install github.com/pointjungle/cli/cmd/pj@latest`. The repository moved to https://github.com/PointJungle/cli.

  - Remove the old `codeforge` binary (for example `rm ~/.local/bin/codeforge`).
  - Sign in again with `pj auth login`: the CLI now keeps its settings and tokens under `pointjungle` (configuration directory and keychain service) and reads `PJ_*` environment variables instead of `CODEFORGE_*`.


## v0.3.0 (2026-11-21)

### Features

- When the install directory is not on your `PATH`, the install script now prints the exact command that adds it for your shell (zsh, bash, fish or `~/.profile`).


## v0.2.0 (2026-11-21)

### Features

- The CLI now has its own repository, github.com/pointjungle/codeforge-cli: install it with `go install github.com/pointjungle/codeforge-cli/cmd/codeforge@latest`, the install script or an archive from its releases, whose notes now come from this changelog.

  - Requests identify the CLI with a `User-Agent: codeforge-cli/<version> (<os>; <arch>)` header, so that codeforge can tell versions apart.
