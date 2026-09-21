#Requires -Version 5.1
<#
.SYNOPSIS
  Installs or updates eng from GitHub Releases (remote install/update).
.DESCRIPTION
  Downloads the windows/amd64 binary for the requested tag (default latest),
  verifies its SHA256 against the release checksums.txt, and atomically
  replaces the binary in the install dir (default $HOME\.local\bin as
  eng.exe). Idempotent: re-running installs the newest release.

  Install (latest):
    irm https://github.com/JhonMA82/engineering-platform/releases/latest/download/install.ps1 | iex
  Pin a version:
    & install.ps1 -Version 1.4.0
.PARAMETER Version
  X.Y.Z, vX.Y.Z or latest (default latest; or $env:ENG_VERSION).
.PARAMETER InstallDir
  Destination directory (default $HOME\.local\bin; or $env:ENG_INSTALL_DIR).
.PARAMETER Repo
  GitHub OWNER/NAME slug (default JhonMA82/engineering-platform; or $env:ENG_GITHUB_REPO).
.PARAMETER AddToPath
  Append the install dir to the user PATH when missing.
#>
[CmdletBinding()]
param(
  [string]$Version = $(if ($env:ENG_VERSION) { $env:ENG_VERSION } else { 'latest' }),
  [string]$InstallDir = $(if ($env:ENG_INSTALL_DIR) { $env:ENG_INSTALL_DIR } else { Join-Path $HOME '.local\bin' }),
  [string]$Repo = $(if ($env:ENG_GITHUB_REPO) { $env:ENG_GITHUB_REPO } else { 'JhonMA82/engineering-platform' }),
  [switch]$AddToPath
)

$ErrorActionPreference = 'Stop'

if ($Repo -notmatch '^[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+$') {
  throw "invalid repo '$Repo' (want OWNER/NAME)"
}

$Asset = 'eng-windows-amd64.exe'
$Tag = $Version
if ($Tag -eq 'latest') {
  # No API token needed: the /releases/latest redirect carries the tag.
  $req = [System.Net.WebRequest]::Create("https://github.com/$Repo/releases/latest")
  $req.AllowAutoRedirect = $false
  try {
    $resp = $req.GetResponse()
    $resp.Close()
    throw 'expected a redirect from /releases/latest'
  } catch [System.Net.WebException] {
    $webResp = $_.Exception.Response
    if (-not $webResp) { throw "could not reach github.com ($($_.Exception.Message))" }
    $loc = $webResp.Headers['Location']
    if (-not $loc) { throw 'could not resolve latest release' }
    $Tag = $loc.Substring($loc.LastIndexOf('/') + 1)
  }
} elseif ($Tag -notmatch '^v') {
  $Tag = "v$Tag"
}

$Base = "https://github.com/$Repo/releases/download/$Tag"
$tmp = Join-Path ([System.IO.Path]::GetTempPath()) ("eng-install-" + [System.Guid]::NewGuid().ToString('N'))
New-Item -ItemType Directory -Path $tmp | Out-Null
try {
  Write-Host "installing eng $Tag ($Asset) from $Repo..."
  Invoke-WebRequest -Uri "$Base/$Asset" -OutFile (Join-Path $tmp $Asset)
  Invoke-WebRequest -Uri "$Base/checksums.txt" -OutFile (Join-Path $tmp 'checksums.txt')

  $want = $null
  foreach ($line in (Get-Content (Join-Path $tmp 'checksums.txt'))) {
    $parts = ($line.Trim() -split '\s+')
    if ($parts.Count -eq 2 -and ($parts[1] -eq $Asset -or $parts[1] -eq "*$Asset")) { $want = $parts[0]; break }
  }
  if (-not $want) { throw 'checksums.txt does not cover ' + $Asset }
  $got = (Get-FileHash -Path (Join-Path $tmp $Asset) -Algorithm SHA256).Hash.ToLower()
  if ($got -ne $want.ToLower()) { throw "checksum mismatch for $Asset (download corrupted or tampered)" }

  if (-not (Test-Path $InstallDir)) { New-Item -ItemType Directory -Path $InstallDir | Out-Null }
  $dest = Join-Path $InstallDir 'eng.exe'
  $staged = Join-Path $InstallDir '.eng.new.exe'
  Copy-Item -Path (Join-Path $tmp $Asset) -Destination $staged -Force
  Move-Item -Path $staged -Destination $dest -Force
  Write-Host "installed $dest ($Tag)"

  $onPath = ($env:PATH -split ';') -contains $InstallDir
  if (-not $onPath) {
    if ($AddToPath) {
      $userPath = [Environment]::GetEnvironmentVariable('Path', 'User')
      if (($userPath -split ';') -notcontains $InstallDir) {
        [Environment]::SetEnvironmentVariable('Path', "$userPath;$InstallDir", 'User')
        Write-Host "added $InstallDir to the user PATH (restart the terminal)"
      }
    } else {
      Write-Host "add $InstallDir to your PATH, or re-run with -AddToPath"
    }
  }
  & $dest version
} finally {
  Remove-Item -Recurse -Force $tmp -ErrorAction SilentlyContinue
}
