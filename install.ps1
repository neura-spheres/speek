# Speek Installer for Windows
# Run with: irm https://raw.githubusercontent.com/neura-spheres/speek/main/install.ps1 | iex

$ErrorActionPreference = "Stop"
$Repo = "neura-spheres/speek"
$InstallDir = "$env:USERPROFILE\.speek"
$BinaryName = "speek.exe"
$dest = "$InstallDir\$BinaryName"

Write-Host ""
Write-Host "  Speek Language Installer" -ForegroundColor Cyan
Write-Host "  ========================" -ForegroundColor DarkGray
Write-Host ""

# ── 1. Download latest release ──────────────────────────────────────────────
Write-Host "  [1/3] Fetching latest release..." -ForegroundColor Cyan

$release = Invoke-RestMethod "https://api.github.com/repos/$Repo/releases/latest"
$version = $release.tag_name

# Find the Windows zip asset (GoReleaser produces speek_VERSION_windows_amd64.zip)
$arch = if ([System.Environment]::Is64BitOperatingSystem) { "amd64" } else { "386" }
$zipAsset = $release.assets | Where-Object { $_.name -like "*windows*$arch*.zip" } | Select-Object -First 1

if (-not $zipAsset) {
    # Fallback: look for any windows zip
    $zipAsset = $release.assets | Where-Object { $_.name -like "*windows*.zip" } | Select-Object -First 1
}

if (-not $zipAsset) {
    # Last resort: try a plain exe (older releases or alternate build config)
    $exeAsset = $release.assets | Where-Object { $_.name -like "*windows*$arch*.exe" } | Select-Object -First 1
    if (-not $exeAsset) {
        $exeAsset = $release.assets | Where-Object { $_.name -like "*windows*.exe" } | Select-Object -First 1
    }
    if ($exeAsset) {
        Write-Host "  Downloading $($exeAsset.name)..." -ForegroundColor Cyan
        New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null
        Invoke-WebRequest -Uri $exeAsset.browser_download_url -OutFile $dest -UseBasicParsing
    } else {
        Write-Host ""
        Write-Host "  ERROR: No Windows binary found in release $version" -ForegroundColor Red
        Write-Host "  Check: https://github.com/$Repo/releases/latest" -ForegroundColor DarkGray
        exit 1
    }
} else {
    # Download and extract zip
    $tmpZip = "$env:TEMP\speek-install.zip"
    $tmpDir = "$env:TEMP\speek-install"

    Write-Host "  Downloading $($zipAsset.name)..." -ForegroundColor Cyan
    Invoke-WebRequest -Uri $zipAsset.browser_download_url -OutFile $tmpZip -UseBasicParsing

    Write-Host "  Extracting..." -ForegroundColor Cyan
    if (Test-Path $tmpDir) { Remove-Item $tmpDir -Recurse -Force }
    Expand-Archive -Path $tmpZip -DestinationPath $tmpDir -Force

    # Find speek.exe inside the extracted folder
    $exeFile = Get-ChildItem -Path $tmpDir -Filter "speek.exe" -Recurse | Select-Object -First 1
    if (-not $exeFile) {
        Write-Host "  ERROR: speek.exe not found inside the downloaded archive." -ForegroundColor Red
        exit 1
    }

    New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null
    Copy-Item -Path $exeFile.FullName -Destination $dest -Force

    # Clean up temp files
    Remove-Item $tmpZip -Force -ErrorAction SilentlyContinue
    Remove-Item $tmpDir -Recurse -Force -ErrorAction SilentlyContinue
}

Write-Host "  Speek $version installed to $dest" -ForegroundColor Green

# ── 2. Add to PATH ───────────────────────────────────────────────────────────
Write-Host "  [2/3] Updating PATH..." -ForegroundColor Cyan

$currentPath = [Environment]::GetEnvironmentVariable("PATH", "User")
if ($currentPath -notlike "*$InstallDir*") {
    [Environment]::SetEnvironmentVariable("PATH", "$currentPath;$InstallDir", "User")
    Write-Host "  Added $InstallDir to user PATH." -ForegroundColor Green
} else {
    Write-Host "  PATH already up to date." -ForegroundColor DarkGray
}

# Refresh current session so speek works immediately without reopening terminal
$env:PATH = [Environment]::GetEnvironmentVariable("PATH", "User") + ";" + [Environment]::GetEnvironmentVariable("PATH", "Machine")

# ── 3. Install editor extensions ─────────────────────────────────────────────
Write-Host "  [3/3] Installing editor extensions..." -ForegroundColor Cyan
Write-Host ""

# speek install-vscode copies the embedded extension to all detected editors
# --force overwrites any previously installed version so updates always apply
& $dest install-vscode --force

# Register .spk file association with the Speek icon (Windows only, no admin required)
& $dest register-filetype

# ── Done ──────────────────────────────────────────────────────────────────────
Write-Host ""
Write-Host "  All done! Speek $version is ready." -ForegroundColor Green
Write-Host ""
Write-Host "  IMPORTANT: Restart VS Code / Cursor to activate syntax highlighting." -ForegroundColor Yellow
Write-Host ""
Write-Host "  Quick start:" -ForegroundColor White
Write-Host "    speek repl                 # interactive shell" -ForegroundColor DarkGray
Write-Host "    speek run yourfile.spk     # run a program" -ForegroundColor DarkGray
Write-Host "    speek help                 # show all commands" -ForegroundColor DarkGray
Write-Host ""
