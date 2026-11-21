# Install the codeforge CLI on Linux

For amd64 and arm64. On Windows with WSL, follow this page inside WSL.

## Install script

```bash
curl -fsSL https://raw.githubusercontent.com/pointjungle/codeforge-cli/main/install.sh | sh
```

The script downloads the latest release from the [GitHub releases](https://github.com/pointjungle/codeforge-cli/releases), checks its SHA-256 against `checksums.txt` and installs `codeforge` in `~/.local/bin`. If that directory is not in your `PATH`, the script prints the command that adds it for your shell (`~/.bashrc`, `~/.zshrc`, fish or `~/.profile`), for example:

```bash
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc
```

Then open a new terminal.

| Variable | |
|---|---|
| `CODEFORGE_INSTALL_DIR` | where to install (default: `~/.local/bin`) |
| `CODEFORGE_VERSION` | a given version, e.g. `0.1.0` (default: the latest) |

## Download an archive

Download `codeforge_<version>_linux_amd64.tar.gz` (or `_arm64`) from the [releases](https://github.com/pointjungle/codeforge-cli/releases), check it against `checksums.txt`, then:

```bash
tar -xzf codeforge_0.1.0_linux_amd64.tar.gz codeforge
install -m 0755 codeforge ~/.local/bin/codeforge
```

## Go

```bash
go install github.com/pointjungle/codeforge-cli/cmd/codeforge@latest
```

The binary goes to `$(go env GOPATH)/bin`.

## Check the installation

```console
$ codeforge version
codeforge 0.1.0 (a55150b, 2026-11-21T10:00:00Z)
```

Next: [sign in](../README.md#sign-in).
