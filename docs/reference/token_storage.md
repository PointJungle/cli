# CLI token storage

The CLI keeps your sign-in — an access token, a refresh token and when they expire — and never prints it.

| Where | When |
|---|---|
| macOS Keychain, Windows Credential Manager, or the Secret Service on Linux (GNOME Keyring, KWallet) | by default, under the service name `codeforge` |
| `tokens.json` in the [configuration directory](settings.md#configuration-directory), readable by you only (mode `0600`) | when no keychain is available — a server, a container, WSL without a desktop session — or with `CODEFORGE_TOKEN_STORE=file` |

`codeforge auth logout` deletes it. Deleting the entry or the file yourself signs you out of the CLI, but leaves the session open on codeforge until it expires.
