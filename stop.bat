@echo off
setlocal enabledelayedexpansion

REM GAIOL Web Server Stop Script (Batch)
REM Change to script directory
cd /d "%~dp0"

echo.
echo ========================================
echo Stopping GAIOL Web Server
echo ========================================
echo.

REM Run PowerShell script and capture exit code
powershell.exe -ExecutionPolicy Bypass -File "%~dp0stop.ps1" %*
set EXIT_CODE=%ERRORLEVEL%

REM Always pause to show results
pause
exit /b %EXIT_CODE%
