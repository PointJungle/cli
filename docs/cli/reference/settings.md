# CLI settings

## Configuration directory

The CLI keeps its files in `codeforge` in your user configuration directory, or in `CODEFORGE_CONFIG_DIR`:

| System | Directory |
|---|---|
| Linux | `~/.config/codeforge` (or `$XDG_CONFIG_HOME/codeforge`) |
| macOS | `~/Library/Application Support/codeforge` |
| Windows | `%AppData%\codeforge` |

## Environment variables

| Variable | |
|---|---|
| `CODEFORGE_CONFIG_DIR` | the configuration directory |
| `CODEFORGE_TOKEN_STORE` | `file` to keep tokens in a file instead of the keychain (see [Token storage](token_storage.md)) |
