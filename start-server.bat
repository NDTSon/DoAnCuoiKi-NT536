@echo off
title LiveKit Server with Authentication
color 0A

echo.
echo ================================================
echo    LIVEKIT STREAMING SERVER + AUTH
echo ================================================
echo.

echo [1/4] Checking Go installation...

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

echo [2/4] Checking directory...
if not exist "cmd\server\main.go" (
    color 0C
    echo [ERROR] Wrong directory!
    echo Please run: cd c:\da-NT536\livekit
    pause
    exit /b 1
)
echo [OK] In correct directory
echo.

echo [3/4] Preparing database...
REM Create data directory for SQLite (if PostgreSQL not available)
if not exist "data" (
    mkdir data
    echo [OK] Created data directory for SQLite database
) else (
    echo [OK] Data directory exists
)

REM Check if DATABASE_URL is set
if "%DATABASE_URL%"=="" (
    echo [INFO] DATABASE_URL not set, will use SQLite fallback
    echo [INFO] To use PostgreSQL, set: DATABASE_URL=postgres://...
) else (
    echo [OK] DATABASE_URL is set
)
echo.

echo [4/4] Starting server...
echo.
echo ================================================
echo   SERVER STARTING - Please wait...
echo ================================================
echo.
echo Server URL: http://localhost:7880
echo.
echo FEATURES:
echo - LiveKit Streaming Server
echo - User Registration: /api/register
echo - User Login: /api/login
echo - User Profile: /api/profile (requires auth)
echo - Auth UI: http://localhost:7880/examples/auth.html
echo - Watch Stream: http://localhost:7880/examples/watch-stream.html
echo - Platform Overview: http://localhost:7880/examples/platform-overview.html
echo - Live Streams: http://localhost:7880/examples/live-streams.html
echo.
echo DATABASE:
echo - Primary: PostgreSQL (if DATABASE_URL is set)
echo - Fallback: SQLite (data/dev.db)
echo.
echo IMPORTANT:
echo - Keep this window OPEN while using the app
echo - Press Ctrl+C to stop the server
echo - Make sure to open pages via HTTP (not file://)
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
