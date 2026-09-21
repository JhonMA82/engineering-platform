#Requires -Version 5.1
<#
.SYNOPSIS
  Uninstalls eng (removes eng.exe installed by install.ps1).
.DESCRIPTION
  Remote:
    irm https://github.com/JhonMA82/engineering-platform/releases/latest/download/uninstall.ps1 | iex
  Local:
    .\scripts\uninstall.ps1 [-InstallDir DIR]
  Removes exactly one resolved path; never touches projects or
  project-local .engineering state.
#>
[CmdletBinding()]
param(
  [string]$InstallDir = $(if ($env:ENG_INSTALL_DIR) { $env:ENG_INSTALL_DIR } else { Join-Path $HOME '.local\bin' })
)

$ErrorActionPreference = 'Stop'
$target = Join-Path $InstallDir 'eng.exe'
if (-not (Test-Path $target)) {
  Write-Host "uninstall: $target not present — nothing to remove."
  exit 0
}
Remove-Item -Force $target
Write-Host "uninstalled $target"
