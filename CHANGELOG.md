## v0.2.0 (2026-11-21)

### Features

- The CLI now has its own repository, github.com/pointjungle/codeforge-cli: install it with `go install github.com/pointjungle/codeforge-cli/cmd/codeforge@latest`, the install script or an archive from its releases, whose notes now come from this changelog.

  - Requests identify the CLI with a `User-Agent: codeforge-cli/<version> (<os>; <arch>)` header, so that codeforge can tell versions apart. ([`9ff1556`](https://gitlab.com/point-jungle/projects/codeforge/codeforge-cli/-/commit/9ff155640316eeaf6894147b5ced986e67bfea7d))


# Changelog

All notable changes to this project will be documented in this file.

Towncrier aggregates fragments from `changelog.d/` at release time.
