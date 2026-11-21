# codeforge SDK

The command line tool and the client libraries of **[codeforge](https://codeforge.pointjungle.com)**, PointJungle's platform to deploy and run pipelines.

- **codeforge**: https://codeforge.pointjungle.com
- **Documentation**: https://codeforge.pointjungle.com/docs

## The codeforge CLI

`codeforge` works with codeforge from your terminal, your scripts and your AI agents. It is one binary for Linux, macOS and Windows.

```bash
curl -fsSL https://raw.githubusercontent.com/pointjungle/codeforge-sdk/main/cli/install.sh | sh
```

On Windows, or for a given version, download an archive from the [releases](https://github.com/pointjungle/codeforge-sdk/releases). With Go: `go install github.com/pointjungle/codeforge-sdk/cli/cmd/codeforge@latest`.

Step by step: [Linux](docs/cli/setup/linux.md) · [macOS](docs/cli/setup/macOS.md) · [Windows](docs/cli/setup/windows.md).

```bash
codeforge auth login            # sign in through your browser
codeforge auth status           # who you are
codeforge auth status --json    # the same, for scripts and agents
```

Every command speaks JSON with `--json` and returns stable exit codes.

| Documentation | |
|---|---|
| [CLI guide](docs/cli/README.md) | prerequisites, setup, sign in, commands, exit codes (also on [codeforge's docs](https://codeforge.pointjungle.com/docs/cli)) |
| [Settings](docs/cli/reference/settings.md) | configuration directory, environment variables |
| [Token storage](docs/cli/reference/token_storage.md) | where the CLI keeps your sign-in |

## Client libraries for pipeline development

A pipeline is your code; a client library connects it to codeforge: it reads the job's inputs, reports results, sends logs and reads variables. Each library comes with its documentation and examples.

| Language | Library | Status |
|---|---|---|
| Python | [`pointjungle-codeforge-sdk`](python/README.md) | coming soon |

## Contributing

Issues and pull requests are welcome: see [CONTRIBUTING.md](CONTRIBUTING.md) and the [CLI developer docs](docs/cli/development/README.md).

## License

[MIT](LICENSE)
