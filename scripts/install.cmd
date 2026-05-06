@echo off
setlocal enabledelayedexpansion

REM MobiAI CLI installer (Windows cmd)
REM Usage: curl.exe -fsSL https://mobiai.dev/install.cmd -o "%TEMP%\i.cmd" && "%TEMP%\i.cmd"

if "%MOBIAI_INSTALL_BASE%"=="" (
    set "INSTALL_BASE=https://github.com/ArisGuimera/MobiAI-Core/releases"
) else (
    set "INSTALL_BASE=%MOBIAI_INSTALL_BASE%"
)
if "%MOBIAI_INSTALL_DIR%"=="" (
    set "INSTALL_DIR=%USERPROFILE%\.mobiai\bin"
) else (
    set "INSTALL_DIR=%MOBIAI_INSTALL_DIR%"
)

REM Detect arch
set "ARCH=amd64"
if "%PROCESSOR_ARCHITECTURE%"=="ARM64" set "ARCH=arm64"

REM Resolve version
if "%MOBIAI_VERSION%"=="" (
    for /f "delims=" %%i in ('powershell -NoProfile -Command "(Invoke-RestMethod 'https://api.github.com/repos/ArisGuimera/MobiAI-Core/releases?per_page=20') ^| Where-Object { $_.tag_name -like 'cli-v*' } ^| Select-Object -First 1 -ExpandProperty tag_name"') do set "LATEST_TAG=%%i"
    if "!LATEST_TAG!"=="" (
        echo Error: no pude detectar la ultima version de MobiAI CLI.
        echo Configura MOBIAI_VERSION manualmente o revisa el repo.
        exit /b 1
    )
    set "MOBIAI_VERSION=!LATEST_TAG:cli-v=!"
) else (
    set "LATEST_TAG=cli-v%MOBIAI_VERSION%"
)

echo Version: !MOBIAI_VERSION!

set "ARCHIVE=mobiai-!MOBIAI_VERSION!-windows-%ARCH%.zip"
set "URL=%INSTALL_BASE%/download/!LATEST_TAG!/!ARCHIVE!"

set "TMP=%TEMP%\mobiai-install-%RANDOM%"
mkdir "%TMP%" 2>nul

echo Instalador de MobiAI CLI
echo Detectado: windows %ARCH%
echo Descargando desde: !URL!

curl.exe -fsSL "!URL!" -o "%TMP%\!ARCHIVE!"
if errorlevel 1 (
    echo Error: fallo la descarga
    rmdir /s /q "%TMP%"
    exit /b 1
)

powershell -NoProfile -Command "Expand-Archive -Force \"%TMP%\!ARCHIVE!\" \"%TMP%\""
if errorlevel 1 (
    echo Error: fallo la extraccion
    rmdir /s /q "%TMP%"
    exit /b 1
)

if not exist "%INSTALL_DIR%" mkdir "%INSTALL_DIR%"
move /y "%TMP%\mobiai.exe" "%INSTALL_DIR%\mobiai.exe" >nul
rmdir /s /q "%TMP%"

echo.
echo Instalado en: %INSTALL_DIR%\mobiai.exe
powershell -NoProfile -Command "$p = [Environment]::GetEnvironmentVariable('Path', 'User'); if ($p -notlike '*%INSTALL_DIR%*') { [Environment]::SetEnvironmentVariable('Path', '%INSTALL_DIR%;' + $p, 'User'); Write-Host 'Agregado al PATH del usuario (reinicia la terminal).' }"

echo.
echo Proximo paso: mobiai --version
