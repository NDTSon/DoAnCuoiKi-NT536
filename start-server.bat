@echo off
title LiveKit Server
color 0A

echo.
echo ================================================
echo          LIVEKIT STREAMING SERVER
echo ================================================
echo.
echo [1/3] Checking Go installation...

where go >nul 2>nul
if %errorlevel% neq 0 (
    color 0C
    echo.
    echo [ERROR] Go not found in this terminal!
    echo.
    echo SOLUTION:
    echo 1. Close THIS window
    echo 2. Press Windows + X
    echo 3. Select "Windows PowerShell" or "Terminal"  
    echo 4. Run: cd c:\da-NT536\livekit
    echo 5. Run: start-server.bat
    echo.
    pause
    exit /b 1
)

for /f "tokens=*" %%i in ('go version') do set GO_VERSION=%%i
echo [OK] %GO_VERSION%
echo.

echo [2/3] Checking directory...
if not exist "cmd\server\main.go" (
    color 0C
    echo [ERROR] Wrong directory!
    echo Please run: cd c:\da-NT536\livekit
    pause
    exit /b 1
)
echo [OK] In correct directory
echo.

echo [3/3] Starting server...
echo.
echo ================================================
echo   SERVER STARTING - Please wait...
echo ================================================
echo.
echo Server will be at: http://localhost:7880
echo.
echo IMPORTANT:
echo - Keep this window OPEN while using the demo
echo - Press Ctrl+C to stop the server
echo.
echo When you see "starting LiveKit server" below,
echo the server is ready to use!
echo.
echo ================================================
echo.

REM Start the server
go run ./cmd/server --dev

echo.
echo Server stopped.
pause
