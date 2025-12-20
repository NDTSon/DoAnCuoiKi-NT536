@echo off
echo Starting Redis...

rem 1. Check if container exists and is running
docker ps -q -f name=livekit-redis | findstr . >nul
if %errorlevel% equ 0 (
    echo Redis is already running.
    goto :end
)

rem 2. Check if container exists but is stopped
docker ps -a -q -f name=livekit-redis | findstr . >nul
if %errorlevel% equ 0 (
    echo Starting existing 'livekit-redis' container...
    docker start livekit-redis
    goto :end
)

rem 3. Create and run new container
echo Creating and running new 'livekit-redis' container...
docker run -d -p 6379:6379 --name livekit-redis redis:alpine

:end
echo Redis is ready on localhost:6379
pause
