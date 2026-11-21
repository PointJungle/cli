# Install the codeforge CLI on Linux

For amd64 and arm64. On Windows with WSL, follow this page inside WSL.

## Install script

```bash
curl -fsSL https://raw.githubusercontent.com/pointjungle/codeforge-sdk/main/cli/install.sh | sh
```

The script downloads the latest release from the [GitHub releases](https://github.com/pointjungle/codeforge-sdk/releases), checks its SHA-256 against `checksums.txt` and installs `codeforge` in `~/.local/bin`. If that directory is not in your `PATH`, the script tells you to add it — for example in `~/.bashrc` or `~/.zshrc`:

```bash
export PATH="$HOME/.local/bin:$PATH"
```

| Variable | |
|---|---|
| `CODEFORGE_INSTALL_DIR` | where to install (default: `~/.local/bin`) |
| `CODEFORGE_VERSION` | a given version, e.g. `0.1.0` (default: the latest) |

## Download an archive

Download `codeforge_<version>_linux_amd64.tar.gz` (or `_arm64`) from the [releases](https://github.com/pointjungle/codeforge-sdk/releases), check it against `checksums.txt`, then:

```bash
tar -xzf codeforge_0.1.0_linux_amd64.tar.gz codeforge
install -m 0755 codeforge ~/.local/bin/codeforge
```

## Go

```bash
go install github.com/pointjungle/codeforge-sdk/cli/cmd/codeforge@latest
```

The binary goes to `$(go env GOPATH)/bin`.

## Check the installation

```console
$ codeforge version
codeforge 0.1.0 (a55150b, 2026-11-21T10:00:00Z)
```

Next: [sign in](../README.md#sign-in).
