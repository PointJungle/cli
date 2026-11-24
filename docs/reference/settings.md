# CLI settings

## Configuration directory

The CLI keeps its files in `pj` in your user configuration directory, or in `PJ_CONFIG_DIR`:

| System | Directory |
|---|---|
| Linux | `~/.config/pointjungle` (or `$XDG_CONFIG_HOME/pointjungle`) |
| macOS | `~/Library/Application Support/pointjungle` |
| Windows | `%AppData%\pointjungle` |

## Environment variables

| Variable | |
|---|---|
| `PJ_CONFIG_DIR` | the configuration directory |
| `PJ_TOKEN_STORE` | `file` to keep tokens in a file instead of the keychain (see [Token storage](token_storage.md)) |
