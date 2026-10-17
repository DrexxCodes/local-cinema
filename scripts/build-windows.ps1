<#
.SYNOPSIS
  Builds Local Cinema on Windows: UI, Go programs and (optionally) the Inno Setup installer.

.EXAMPLE
  .\scripts\build-windows.ps1                      # UI + local-cinema.exe + dist\local-cinema-gui.exe
  .\scripts\build-windows.ps1 -Installer           # also builds dist\installer\LocalCinema-Setup-*.exe
  .\scripts\build-windows.ps1 -Installer -Version 0.5.0 -Repo you/local-cinema
  .\scripts\build-windows.ps1 -SkipUi              # reuse the existing ui\out folder

  If scripts are blocked, run once:  Set-ExecutionPolicy -Scope Process Bypass
#>
[CmdletBinding()]
param(
  [string]$Version = "0.0.0-local",
  [string]$Repo = "OWNER/local-cinema",
  [switch]$Installer,
  [switch]$SkipUi,
  [switch]$SkipIcon
)

$ErrorActionPreference = "Stop"
$root = Resolve-Path (Join-Path $PSScriptRoot "..")
Set-Location $root

function Step($msg) { Write-Host "`n==> $msg" -ForegroundColor Cyan }
function Need($cmd, $hint) {
  if (-not (Get-Command $cmd -ErrorAction SilentlyContinue)) { throw "$cmd not found. $hint" }
}
function Run($exe, [string[]]$arguments) {
  & $exe @arguments
  if ($LASTEXITCODE -ne 0) { throw "$exe failed (exit code $LASTEXITCODE)" }
}

Need go "Install Go from https://go.dev/dl/"

# 1. UI (static export that gets embedded into the Go program)
if (-not $SkipUi) {
  Step "Building the UI"
  Push-Location ui
  try {
    if (Get-Command pnpm -ErrorAction SilentlyContinue) {
      Run pnpm @("install", "--frozen-lockfile")
      Run pnpm @("build")
    } else {
      Need npm "Install Node.js from https://nodejs.org/ (or run: corepack enable)"
      Run npm @("install")
      Run npm @("run", "build")
    }
  } finally { Pop-Location }
}
if (-not (Test-Path "ui\out\_next")) { throw "ui\out looks empty. Build the UI first (do not use -SkipUi)." }

# 2. Windows icon and version info (optional, makes the .exe show the logo)
Remove-Item "rsrc_*.syso" -ErrorAction SilentlyContinue
if (-not $SkipIcon) {
  Step "Embedding icon and version info"
  $fv = ($Version -split "-")[0]
  while (($fv -split "\.").Count -lt 4) { $fv = "$fv.0" }
  $env:GOOS = ""; $env:GOARCH = ""
  Run go @("install", "github.com/tc-hib/go-winres@latest")
  $winres = Join-Path (& go env GOPATH) "bin\go-winres.exe"
  Run $winres @("make", "--in", "packaging/windows/winres.json", "--arch", "amd64", "--file-version", $fv, "--product-version", $fv)
}

# 3. Go programs
Step "Building the Go programs"
Run go @("mod", "tidy")
$env:CGO_ENABLED = "0"; $env:GOOS = "windows"; $env:GOARCH = "amd64"
$pkg = "local-cinema/internal/version"
$bundle = (Get-Content packaging\ffmpeg-bundle.txt -Raw).Trim()
$ld = "-s -w -X $pkg.Version=$Version -X $pkg.Repo=$Repo -X $pkg.FFmpegBundle=$bundle"
New-Item -ItemType Directory -Force dist | Out-Null
# Console build: handy for seeing logs.
Run go @("build", "-trimpath", "-ldflags=$ld", "-o", "local-cinema.exe", ".")
# Windowless build: this is the file the installer packs.
Run go @("build", "-trimpath", "-ldflags=$ld -H=windowsgui", "-o", "dist\local-cinema-gui.exe", ".")
Remove-Item "rsrc_*.syso" -ErrorAction SilentlyContinue
Write-Host "Built local-cinema.exe and dist\local-cinema-gui.exe" -ForegroundColor Green

# 4. Installer (optional)
# if ($Installer) {
#   Step "Building the installer"
#   $candidates = @(
#     "${env:ProgramFiles(x86)}\Inno Setup 6\ISCC.exe",
#     "$env:ProgramFiles\Inno Setup 6\ISCC.exe",
#     "$env:LOCALAPPDATA\Programs\Inno Setup 6\ISCC.exe"
#   )
#   $iscc = $candidates | Where-Object { Test-Path $_ } | Select-Object -First 1
#   if (-not $iscc) { throw "Inno Setup 6 not found. Install it from https://jrsoftware.org/isdl.php" }
#   # No FFmpeg bundle or checksum here, so the installer skips the FFmpeg download.
#   # Install FFmpeg later from Settings, or run: winget install Gyan.FFmpeg
#   Run $iscc @("/DAppVersion=$Version", "/DRepo=$Repo", "packaging\windows\LocalCinema.iss")
#   Write-Host "Installer: dist\installer\LocalCinema-Setup-$Version.exe" -ForegroundColor Green
# }
