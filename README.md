# PointJungle CLI

`pj`, the command line tool of **[PointJungle](https://codeforge.pointjungle.com)**, the platform to deploy and run pipelines. It works with PointJungle from your terminal, your scripts and your AI agents. One binary for Linux, macOS and Windows.

- **PointJungle**: https://codeforge.pointjungle.com
- **Documentation**: https://codeforge.pointjungle.com/docs

## Install

```bash
brew install pointjungle/tap/pj        # macOS and Linux, with Homebrew
curl -fsSL https://raw.githubusercontent.com/PointJungle/cli/main/install.sh | sh
```

On Windows, or for a given version, download an archive from the [releases](https://github.com/PointJungle/cli/releases). With Go: `go install github.com/pointjungle/cli/cmd/pj@latest`.

Step by step: [Linux](docs/setup/linux.md) · [macOS](docs/setup/macOS.md) · [Windows](docs/setup/windows.md).

## Use

```bash
pj auth login            # sign in through your browser
pj auth status           # who you are
pj auth status --json    # the same, for scripts and agents
```

Every command speaks JSON with `--json` and returns stable exit codes.

| Documentation | |
|---|---|
| [CLI guide](docs/README.md) | prerequisites, setup, sign in, commands, exit codes (also on [PointJungle's docs](https://codeforge.pointjungle.com/docs/cli)) |
| [Settings](docs/reference/settings.md) | configuration directory, environment variables |
| [Token storage](docs/reference/token_storage.md) | where the CLI keeps your sign-in |
| [Changelog](CHANGELOG.md) | what changed in each release |

Client libraries to write pipelines live in their own repositories; the first, for Python, is coming soon.

## Contributing

Issues and pull requests are welcome: see [CONTRIBUTING.md](CONTRIBUTING.md) and the [developer docs](docs/development/README.md).

## License

[MIT](LICENSE)
