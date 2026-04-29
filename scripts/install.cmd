@echo off
setlocal enabledelayedexpansion

REM MobiAI CLI installer (Windows cmd)
REM Usage: curl.exe -fsSL https://mobiai.dev/install.cmd -o "%TEMP%\i.cmd" && "%TEMP%\i.cmd"

if "%MOBIAI_INSTALL_BASE%"=="" (
    set "INSTALL_BASE=https://github.com/ArisGuimera/MobiAI-Core/releases/latest/download"
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

set "ARCHIVE=mobiai-latest-windows-%ARCH%.zip"
set "URL=%INSTALL_BASE%/%ARCHIVE%"

set "TMP=%TEMP%\mobiai-install-%RANDOM%"
mkdir "%TMP%" 2>nul

echo MobiAI CLI installer
echo Detected: windows %ARCH%
echo Downloading from: %URL%

curl.exe -fsSL "%URL%" -o "%TMP%\%ARCHIVE%"
if errorlevel 1 (
    echo Error: download failed
    rmdir /s /q "%TMP%"
    exit /b 1
)

powershell -NoProfile -Command "Expand-Archive -Force \"%TMP%\%ARCHIVE%\" \"%TMP%\""
if errorlevel 1 (
    echo Error: extract failed
    rmdir /s /q "%TMP%"
    exit /b 1
)

if not exist "%INSTALL_DIR%" mkdir "%INSTALL_DIR%"
move /y "%TMP%\mobiai.exe" "%INSTALL_DIR%\mobiai.exe" >nul
rmdir /s /q "%TMP%"

echo.
echo Installed to: %INSTALL_DIR%\mobiai.exe
echo.
echo Add to PATH (run this manually in cmd):
echo   setx PATH "%INSTALL_DIR%;%%PATH%%"
echo.
echo Next step: mobiai --version
