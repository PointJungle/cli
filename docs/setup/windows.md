# Install the PointJungle CLI on Windows

For amd64 and arm64. With WSL, you can also install the Linux version inside WSL: see [Linux](linux.md).

## Download an archive

1. Download `pj_<version>_windows_amd64.zip` (or `_arm64.zip`) from the [releases](https://github.com/PointJungle/cli/releases). `checksums.txt` lists the SHA-256 of every archive.
2. Extract `pj.exe` to a folder of your own, for example `%LOCALAPPDATA%\Programs\pj`.
3. Add that folder to your `PATH`. In PowerShell:

```powershell
$dir = "$env:LOCALAPPDATA\Programs\pj"
New-Item -ItemType Directory -Force $dir | Out-Null
Expand-Archive pj_0.1.0_windows_amd64.zip -DestinationPath $dir -Force
[Environment]::SetEnvironmentVariable("Path", "$([Environment]::GetEnvironmentVariable('Path', 'User'));$dir", "User")
```

Open a new terminal for the `PATH` change to apply.

To check the archive: `(Get-FileHash pj_0.1.0_windows_amd64.zip).Hash` must match its line in `checksums.txt` (case aside).

## Go

```powershell
go install github.com/pointjungle/cli/cmd/pj@latest
```

The binary goes to `%USERPROFILE%\go\bin`.

## Check the installation

```console
> pj version
pj 0.4.0 (1a2b3c4, 2026-11-25T10:00:00Z)
```

Next: [sign in](../README.md#sign-in).
