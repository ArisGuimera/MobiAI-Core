# MobiAI CLI installer (Windows PowerShell)
# Usage: iwr -useb https://mobiai.dev/install.ps1 | iex
$ErrorActionPreference = "Stop"

$InstallBase = if ($env:MOBIAI_INSTALL_BASE) { $env:MOBIAI_INSTALL_BASE } else { "https://github.com/ArisGuimera/MobiAI-Core/releases/latest/download" }
$InstallDir  = if ($env:MOBIAI_INSTALL_DIR)  { $env:MOBIAI_INSTALL_DIR }  else { "$HOME\.mobiai\bin" }

# Detect arch
$Arch = if ($env:PROCESSOR_ARCHITECTURE -eq "ARM64") { "arm64" } else { "amd64" }
$Archive = "mobiai-latest-windows-$Arch.zip"
$Url = "$InstallBase/$Archive"

$Tmp = Join-Path $env:TEMP ("mobiai-install-" + (Get-Random))
New-Item -ItemType Directory -Force -Path $Tmp | Out-Null
New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null

Write-Host "MobiAI CLI installer"
Write-Host "Detected: windows $Arch"
Write-Host "Downloading from: $Url"

try {
    Invoke-WebRequest -Uri $Url -OutFile (Join-Path $Tmp $Archive) -UseBasicParsing
} catch {
    Write-Host "Error: download failed: $_" -ForegroundColor Red
    Remove-Item -Recurse -Force $Tmp
    exit 1
}

try {
    Expand-Archive -Force -Path (Join-Path $Tmp $Archive) -DestinationPath $Tmp
    Move-Item -Force -Path (Join-Path $Tmp "mobiai.exe") -Destination (Join-Path $InstallDir "mobiai.exe")
} finally {
    if (Test-Path $Tmp) {
        Remove-Item -Recurse -Force $Tmp
    }
}

Write-Host ""
Write-Host "Installed to: $InstallDir\mobiai.exe"

# PATH update
$UserPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($UserPath -notlike "*$InstallDir*") {
    [Environment]::SetEnvironmentVariable("Path", "$InstallDir;$UserPath", "User")
    Write-Host "Added to user PATH (restart terminal to apply)."
}

Write-Host ""
Write-Host "Next step: mobiai --version"
