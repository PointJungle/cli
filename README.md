# codeforge CLI

`codeforge`, the command line tool of **[codeforge](https://codeforge.pointjungle.com)**, PointJungle's platform to deploy and run pipelines. It works with codeforge from your terminal, your scripts and your AI agents. One binary for Linux, macOS and Windows.

- **codeforge**: https://codeforge.pointjungle.com
- **Documentation**: https://codeforge.pointjungle.com/docs

## Install

```bash
curl -fsSL https://raw.githubusercontent.com/pointjungle/codeforge-cli/main/install.sh | sh
```

On Windows, or for a given version, download an archive from the [releases](https://github.com/pointjungle/codeforge-cli/releases). With Go: `go install github.com/pointjungle/codeforge-cli/cmd/codeforge@latest`.

Step by step: [Linux](docs/setup/linux.md) · [macOS](docs/setup/macOS.md) · [Windows](docs/setup/windows.md).

## Use

```bash
codeforge auth login            # sign in through your browser
codeforge auth status           # who you are
codeforge auth status --json    # the same, for scripts and agents
```

Every command speaks JSON with `--json` and returns stable exit codes.

| Documentation | |
|---|---|
| [CLI guide](docs/README.md) | prerequisites, setup, sign in, commands, exit codes (also on [codeforge's docs](https://codeforge.pointjungle.com/docs/cli)) |
| [Settings](docs/reference/settings.md) | configuration directory, environment variables |
| [Token storage](docs/reference/token_storage.md) | where the CLI keeps your sign-in |
| [Changelog](CHANGELOG.md) | what changed in each release |

Client libraries to write pipelines live in their own repositories; the first, for Python, is coming soon.

## Contributing

Issues and pull requests are welcome: see [CONTRIBUTING.md](CONTRIBUTING.md) and the [developer docs](docs/development/README.md).

## License

[MIT](LICENSE)
