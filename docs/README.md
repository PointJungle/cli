# The PointJungle CLI

`pj` works with [PointJungle](https://codeforge.pointjungle.com) from your terminal, your scripts and your AI agents. It is one binary for Linux, macOS and Windows (amd64 and arm64).

## Prerequisites

- A PointJungle account. Create one by signing up on [PointJungle](https://codeforge.pointjungle.com).
- A browser on the machine where you run `pj`, to sign in.

## Setup

### Linux

Install script, archive or `go install`: [setup/linux.md](setup/linux.md).

### macOS

Install script, archive or `go install`: [setup/macOS.md](setup/macOS.md).

### Windows

Archive or `go install`: [setup/windows.md](setup/windows.md).

## Sign in

```console
$ pj auth login
Opening your browser to sign in. If it does not open, visit:

  https://api.codeforge.pointjungle.com/identity/o/authorize?client_id=codeforge-cli&…

Signed in to https://api.codeforge.pointjungle.com as ada@example.com.
```

Your browser opens on the PointJungle sign-in page; once you have signed in, the tab tells you to go back to the terminal. The CLI keeps the tokens it receives and renews them without asking, until you sign out.

Where the tokens are kept: [Token storage](reference/token_storage.md).

## Commands

| Command | Flags | Description |
|---|---|---|
| `pj auth login` | `--no-browser`: print the sign-in URL instead of opening a browser<br>`--device`: sign in with a code on another device (not available on PointJungle yet) | Sign in through your browser |
| `pj auth status` | | Show who you are signed in as; exit code `3` when not signed in |
| `pj auth status --json` | | The same as one JSON document: `host`, `signed_in`, `user` (`id`, `email`, `display_name`), `expires_at` |
| `pj auth logout` | | End the session and forget the tokens; not an error when already signed out |
| `pj version` | | Version, commit and build date |

Every command also takes `--json` (one JSON document on stdout, see below) and `-h`, `--help`.

### Output and exit codes

Output is human-readable by default. With `--json`, every command writes one JSON document to stdout — its result, or `{"error": {"code": "…", "message": "…"}}` — and messages for people go to stderr. Scripts and AI agents should use `--json` and the exit code.

| Exit code | Meaning | `error.code` |
|---|---|---|
| `0` | success | — |
| `1` | error | `error`, `device_unavailable`, … |
| `2` | usage error (unknown command or flag) | `usage` |
| `3` | not signed in, or the sign-in expired and could not be renewed | `not_signed_in` |
| `4` | permission denied or not found | `denied` |

## See also

- [Settings](reference/settings.md): configuration directory, environment variables.
- [Token storage](reference/token_storage.md).
