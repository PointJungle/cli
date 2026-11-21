# Install the codeforge CLI on Windows

For amd64 and arm64. With WSL, you can also install the Linux version inside WSL: see [Linux](linux.md).

## Download an archive

1. Download `codeforge_<version>_windows_amd64.zip` (or `_arm64.zip`) from the [releases](https://github.com/pointjungle/codeforge-sdk/releases). `checksums.txt` lists the SHA-256 of every archive.
2. Extract `codeforge.exe` to a folder of your own, for example `%LOCALAPPDATA%\Programs\codeforge`.
3. Add that folder to your `PATH`. In PowerShell:

```powershell
$dir = "$env:LOCALAPPDATA\Programs\codeforge"
New-Item -ItemType Directory -Force $dir | Out-Null
Expand-Archive codeforge_0.1.0_windows_amd64.zip -DestinationPath $dir -Force
[Environment]::SetEnvironmentVariable("Path", "$([Environment]::GetEnvironmentVariable('Path', 'User'));$dir", "User")
```

Open a new terminal for the `PATH` change to apply.

To check the archive: `(Get-FileHash codeforge_0.1.0_windows_amd64.zip).Hash` must match its line in `checksums.txt` (case aside).

## Go

```powershell
go install github.com/pointjungle/codeforge-sdk/cli/cmd/codeforge@latest
```

The binary goes to `%USERPROFILE%\go\bin`.

## Check the installation

```console
> codeforge version
codeforge 0.1.0 (a55150b, 2026-11-21T10:00:00Z)
```

Next: [sign in](../README.md#sign-in).
