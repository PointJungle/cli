# Install the codeforge CLI on macOS

For Apple silicon (arm64) and Intel (amd64).

## Install script

```bash
curl -fsSL https://raw.githubusercontent.com/pointjungle/codeforge-cli/main/install.sh | sh
```

```console
Downloading codeforge 0.1.0 (darwin/arm64)…
Installed /Users/ada/.local/bin/codeforge
```

The script downloads the latest release from the [GitHub releases](https://github.com/pointjungle/codeforge-cli/releases), checks its SHA-256 against `checksums.txt` and installs `codeforge` in `~/.local/bin`. If that directory is not in your `PATH`, add it to `~/.zshrc`:

```bash
export PATH="$HOME/.local/bin:$PATH"
```

| Variable | |
|---|---|
| `CODEFORGE_INSTALL_DIR` | where to install (default: `~/.local/bin`) |
| `CODEFORGE_VERSION` | a given version, e.g. `0.1.0` (default: the latest) |

## Download an archive

Download `codeforge_<version>_darwin_arm64.tar.gz` (Apple silicon) or `_darwin_amd64.tar.gz` (Intel) from the [releases](https://github.com/pointjungle/codeforge-cli/releases), check it against `checksums.txt`, then:

```bash
tar -xzf codeforge_0.1.0_darwin_arm64.tar.gz codeforge
mkdir -p ~/.local/bin && mv codeforge ~/.local/bin/
```

## Go

```bash
go install github.com/pointjungle/codeforge-cli/cmd/codeforge@latest
```

The binary goes to `$(go env GOPATH)/bin`.

## If macOS refuses to run codeforge

A binary downloaded with a browser is quarantined. Run once:

```bash
xattr -d com.apple.quarantine ~/.local/bin/codeforge
```

## Check the installation

```console
$ codeforge version
codeforge 0.1.0 (a55150b, 2026-11-21T10:00:00Z)
```

Next: [sign in](../README.md#sign-in).
