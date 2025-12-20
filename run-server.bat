@echo off
echo ========================================
echo   LiveKit Server - Super Quick Start
echo ========================================
echo.

REM Check if Go is installed
where go >nul 2>nul
if %errorlevel% neq 0 (
    echo ERROR: Go is not found!
    echo.
    echo Please:
    echo 1. Close this window
    echo 2. Open a NEW PowerShell window
    echo 3. Run from the project root directory
    echo 4. Run: .\run-server.bat
    echo.
    pause
    exit /b 1
)

echo [OK] Go found: 
go version
echo.

echo Starting server...
echo.
echo Server will be available at: http://localhost:7880
echo Press Ctrl+C to stop
echo.
echo ========================================
echo.

REM Run server directly without building
go run ./cmd/server --dev --redis-host localhost:6379 --bind 0.0.0.0

pause
