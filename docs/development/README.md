# Developing the codeforge CLI

How the CLI is built, tested and released. For how to use it, see [the CLI guide](../README.md).

## Prerequisites

- Go, at the version in `go.mod` or later.
- [GoReleaser](https://goreleaser.com) v2, only to check the release configuration or build every platform locally: `go install github.com/goreleaser/goreleaser/v2@latest`.

## Checks

Run them before opening a pull request; CI runs the same ones.

```bash
gofmt -l .                     # must print nothing
go vet ./...
go test -race ./...
sh -n install.sh
goreleaser check               # when .goreleaser.yaml changes
goreleaser release --snapshot --clean --skip=publish   # builds every platform into dist/
```

## Layout

```
cmd/codeforge/          entry point; version, commit and date injected by GoReleaser
internal/cli/           commands, output, exit codes
internal/config/        settings resolution, host normalization
internal/tokenstore/    keychain / file token storage
internal/signin/        OIDC discovery, loopback + PKCE, device flow, refresh, revoke
internal/api/           authenticated API calls with automatic refresh
internal/httpclient/    HTTP client: User-Agent, *.localhost to this machine
internal/browser/       open a URL in the browser (WSL included)
.goreleaser.yaml        release builds
install.sh              install script served from main
```

## How sign-in works

1. `GET <host>/api/cli/config` returns the API URL, the OIDC issuer, the client ID (`codeforge-cli`), the scopes, the redirect URI pattern and `signed_in_url`, the page the browser ends on.
2. `<issuer>/.well-known/openid-configuration` gives the authorization, token, device authorization and revocation endpoints.
3. **Loopback flow** (default): the CLI listens on `127.0.0.1:<random port>`, opens the authorization URL with a `state`, a PKCE S256 challenge and `prompt=select_account` (the account site shows its account chooser), receives the code on the redirect URI's path (`/callback`) and exchanges it with the verifier. Then it redirects the browser to `signed_in_url`, with `?error=<code>` when the sign-in failed (`access_denied`, `invalid_state`, `token_exchange_failed`…); a host without `signed_in_url` gets the CLI's own page. The client is public: no secret, credentials sent in the form body.
4. **Device flow** (`--device`): device authorization request, then polling the token endpoint. A server that refuses the grant (`unauthorized_client`, `unsupported_grant_type`, `invalid_client`) gives the `device_unavailable` error.
5. The token (access, refresh, ID token, expiry, client ID, issuer, API URL) is stored per host, then `GET /api/profiles/me/` confirms who signed in.

API calls send `Authorization: Bearer <access token>`. The token is refreshed when it expires within 30 seconds, and once more if the API answers 401; refresh tokens are rotated by the server, so the new one is stored every time. If the refresh fails, the command exits with code 3. `auth logout` revokes the refresh token (RFC 7009, best effort) and deletes the stored token.

## Rules for new commands

- **Output contract**: with `--json`, exactly one JSON document on stdout — the result, or `{"error": {"code", "message"}}` — and messages for people on stderr. Return a `cliError` with the right exit code: 1 error, 2 usage, 3 not signed in, 4 denied or not found. Scripts and AI agents rely on it, so changing a field or an exit code is a breaking change.
- **Settings** come from `internal/config`; do not read environment variables elsewhere.
- **Tokens** never go to stdout, logs or files other than the token store.
- **No CGO**: the binaries are static (`CGO_ENABLED=0`).

## Tests

Tests in `internal/cli/` run commands end to end against a fake codeforge (`fake_test.go`): an `httptest` server playing both the API and the identity provider. It checks what a real server checks — PKCE S256, the loopback `redirect_uri`, refresh token rotation, revocation — so a test that passes against it exercises the real protocol. The browser is replaced by a function that follows the authorization URL. Add a scenario to the fake rather than mocking `signin` or `api`.

`internal/tokenstore` uses `keyring.MockInit()`, so tests never touch your keychain.

## Releases

Maintainers publish releases. Each pull request that changes the CLI adds a changelog entry; a release gathers them in [CHANGELOG.md](https://github.com/pointjungle/codeforge-cli/blob/main/CHANGELOG.md) and is tagged `vX.Y.Z`. Then:

- [GoReleaser](https://goreleaser.com) builds Linux, macOS and Windows binaries for amd64 and arm64 (`.goreleaser.yaml`), with the version from the tag;
- the [GitHub release](https://github.com/pointjungle/codeforge-cli/releases) gets the archives and `checksums.txt`, where `install.sh` downloads them, and the release's changelog entries as notes.

`go install github.com/pointjungle/codeforge-cli/cmd/codeforge@vX.Y.Z` installs a given release.
