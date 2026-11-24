# Install the PointJungle CLI on Linux

For amd64 and arm64. On Windows with WSL, follow this page inside WSL.

## Homebrew

```bash
brew install pointjungle/tap/pj
```

`brew upgrade pj` installs new releases.

## Install script

```bash
curl -fsSL https://raw.githubusercontent.com/PointJungle/cli/main/install.sh | sh
```

The script downloads the latest release from the [GitHub releases](https://github.com/PointJungle/cli/releases), checks its SHA-256 against `checksums.txt` and installs `pj` in `~/.local/bin`. If that directory is not in your `PATH`, the script prints the command that adds it for your shell (`~/.bashrc`, `~/.zshrc`, fish or `~/.profile`), for example:

```bash
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc
```

Then open a new terminal.

| Variable | |
|---|---|
| `PJ_INSTALL_DIR` | where to install (default: `~/.local/bin`) |
| `PJ_VERSION` | a given version, e.g. `0.1.0` (default: the latest) |

## Download an archive

Download `pj_<version>_linux_amd64.tar.gz` (or `_arm64`) from the [releases](https://github.com/PointJungle/cli/releases), check it against `checksums.txt`, then:

```bash
tar -xzf pj_0.4.0_linux_amd64.tar.gz pj
install -m 0755 pj ~/.local/bin/pj
```

## Go

```bash
go install github.com/pointjungle/cli/cmd/pj@latest
```

The binary goes to `$(go env GOPATH)/bin`.

## Check the installation

```console
$ pj version
pj 0.4.0 (1a2b3c4, 2026-11-25T10:00:00Z)
```

Next: [sign in](../README.md#sign-in).
