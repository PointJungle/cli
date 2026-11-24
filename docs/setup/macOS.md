# Install the PointJungle CLI on macOS

For Apple silicon (arm64) and Intel (amd64).

## Homebrew

```bash
brew install pointjungle/tap/pj
```

`brew upgrade pj` installs new releases.

## Install script

```bash
curl -fsSL https://raw.githubusercontent.com/PointJungle/cli/main/install.sh | sh
```

```console
Downloading pj 0.4.0 (darwin/arm64)…
Installed /Users/ada/.local/bin/pj
```

The script downloads the latest release from the [GitHub releases](https://github.com/PointJungle/cli/releases), checks its SHA-256 against `checksums.txt` and installs `pj` in `~/.local/bin`. If that directory is not in your `PATH`, the script prints the command that adds it for your shell (`~/.zshrc` by default on macOS, `~/.bash_profile` for bash, fish), for example:

```bash
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.zshrc
```

Then open a new terminal.

| Variable | |
|---|---|
| `PJ_INSTALL_DIR` | where to install (default: `~/.local/bin`) |
| `PJ_VERSION` | a given version, e.g. `0.1.0` (default: the latest) |

## Download an archive

Download `pj_<version>_darwin_arm64.tar.gz` (Apple silicon) or `_darwin_amd64.tar.gz` (Intel) from the [releases](https://github.com/PointJungle/cli/releases), check it against `checksums.txt`, then:

```bash
tar -xzf pj_0.4.0_darwin_arm64.tar.gz pj
mkdir -p ~/.local/bin && mv pj ~/.local/bin/
```

## Go

```bash
go install github.com/pointjungle/cli/cmd/pj@latest
```

The binary goes to `$(go env GOPATH)/bin`.

## If macOS refuses to run pj

A binary downloaded with a browser is quarantined. Run once:

```bash
xattr -d com.apple.quarantine ~/.local/bin/pj
```

## Check the installation

```console
$ pj version
pj 0.4.0 (1a2b3c4, 2026-11-25T10:00:00Z)
```

Next: [sign in](../README.md#sign-in).
