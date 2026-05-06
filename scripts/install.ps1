# MobiAI CLI installer (Windows PowerShell)
# Usage: iwr -useb https://mobiai.dev/install.ps1 | iex
$ErrorActionPreference = "Stop"

$InstallBase = if ($env:MOBIAI_INSTALL_BASE) { $env:MOBIAI_INSTALL_BASE } else { "https://github.com/ArisGuimera/MobiAI-Core/releases" }
$InstallDir  = if ($env:MOBIAI_INSTALL_DIR)  { $env:MOBIAI_INSTALL_DIR }  else { "$HOME\.mobiai\bin" }

# Detect arch
$Arch = if ($env:PROCESSOR_ARCHITECTURE -eq "ARM64") { "arm64" } else { "amd64" }

# Resolve version
if (-not $env:MOBIAI_VERSION) {
    try {
        $Releases = Invoke-RestMethod "https://api.github.com/repos/ArisGuimera/MobiAI-Core/releases?per_page=20"
        $LatestTag = ($Releases | Where-Object { $_.tag_name -like "cli-v*" } | Select-Object -First 1).tag_name
    } catch {
        Write-Host "Error: no pude consultar la API de GitHub: $_" -ForegroundColor Red
        exit 1
    }
    if (-not $LatestTag) {
        Write-Host "Error: no pude detectar la última versión de MobiAI CLI. Configurá `$env:MOBIAI_VERSION manualmente." -ForegroundColor Red
        exit 1
    }
    $Version = $LatestTag -replace '^cli-v', ''
} else {
    $Version = $env:MOBIAI_VERSION
    $LatestTag = "cli-v$Version"
}

Write-Host "Versión: $Version"

$Archive = "mobiai-$Version-windows-$Arch.zip"
$Url = "$InstallBase/download/$LatestTag/$Archive"

$Tmp = Join-Path $env:TEMP ("mobiai-install-" + (Get-Random))
New-Item -ItemType Directory -Force -Path $Tmp | Out-Null
New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null

Write-Host "Instalador de MobiAI CLI"
Write-Host "Detectado: windows $Arch"
Write-Host "Descargando desde: $Url"

try {
    Invoke-WebRequest -Uri $Url -OutFile (Join-Path $Tmp $Archive) -UseBasicParsing
} catch {
    Write-Host "Error: falló la descarga: $_" -ForegroundColor Red
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
Write-Host "Instalado en: $InstallDir\mobiai.exe"

# PATH update
$UserPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($UserPath -notlike "*$InstallDir*") {
    [Environment]::SetEnvironmentVariable("Path", "$InstallDir;$UserPath", "User")
    Write-Host "Agregado al PATH del usuario (reiniciá la terminal)."
}

Write-Host ""
Write-Host "Próximo paso: mobiai --version"
