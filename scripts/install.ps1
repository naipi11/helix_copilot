param(
  [string]$Version = "latest",
  [string]$BinDir = "$HOME\bin"
)

$ErrorActionPreference = "Stop"
$Repo = "naipi11/helix_copilot"
$Arch = if ([System.Runtime.InteropServices.RuntimeInformation]::ProcessArchitecture -eq "Arm64") { "arm64" } else { "amd64" }
$Asset = "helix-copilot_windows_$Arch.zip"
if ($Version -eq "latest") {
  $Url = "https://github.com/$Repo/releases/latest/download/$Asset"
} else {
  $Url = "https://github.com/$Repo/releases/download/$Version/$Asset"
}

New-Item -ItemType Directory -Force -Path $BinDir | Out-Null
$Tmp = Join-Path ([System.IO.Path]::GetTempPath()) ([System.Guid]::NewGuid().ToString())
New-Item -ItemType Directory -Force -Path $Tmp | Out-Null
try {
  $Zip = Join-Path $Tmp $Asset
  Write-Host "Downloading $Url"
  Invoke-WebRequest -Uri $Url -OutFile $Zip
  Expand-Archive -Path $Zip -DestinationPath $Tmp -Force
  Copy-Item -Path (Join-Path $Tmp "helix-copilot.exe") -Destination (Join-Path $BinDir "helix-copilot.exe") -Force
  $Hx = Join-Path $Tmp "hx.exe"
  if (Test-Path $Hx) { Copy-Item -Path $Hx -Destination (Join-Path $BinDir "hx.exe") -Force }
  Write-Host "Installed helix-copilot to $BinDir"
  if (-not (Get-Command node -ErrorAction SilentlyContinue)) {
    Write-Warning "Node.js is required for @github/copilot-language-server"
  }

  # Install copilot-green theme
  $HelixConfigDir = Join-Path $env:APPDATA "helix"
  $ThemeDir = Join-Path $HelixConfigDir "themes"
  $ThemeFile = Join-Path $Tmp "themes\copilot-green.toml"
  if (Test-Path $ThemeFile) {
    New-Item -ItemType Directory -Force -Path $ThemeDir | Out-Null
    Copy-Item -Path $ThemeFile -Destination (Join-Path $ThemeDir "copilot-green.toml") -Force
    Write-Host ""
    Write-Host "✓ Installed copilot-green theme to $ThemeDir" -ForegroundColor Green
    Write-Host ""
    Write-Host "To enable the theme (recommended for better Ghost Text visibility):"
    Write-Host "  1. Edit $(Join-Path $HelixConfigDir 'config.toml')"
    Write-Host "  2. Add at the top: theme = `"copilot-green`""
    Write-Host "  3. Restart Helix or run :reload-config"
    Write-Host ""
  }
} finally {
  Remove-Item -Recurse -Force $Tmp -ErrorAction SilentlyContinue
}
