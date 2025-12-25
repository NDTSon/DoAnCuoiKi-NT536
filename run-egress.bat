@echo off
echo Starting LiveKit Egress...

rem Create recordings directory if not exists
if not exist "data\recordings" mkdir "data\recordings"

echo Pulling latest Egress image...
docker pull livekit/egress

echo Running Egress...
rem Note: host.docker.internal requires Docker Desktop for Windows
docker run -d ^
    --name livekit-egress ^
    --restart unless-stopped ^
    -e EGRESS_CONFIG_FILE=/etc/egress.yaml ^
    -v "%CD%\config\egress.yaml:/etc/egress.yaml" ^
    -v "%CD%\data\recordings:/out" ^
    livekit/egress

if %errorlevel% equ 0 (
    echo [SUCCESS] Egress service started!
    echo Logs: docker logs -f livekit-egress
) else (
    echo [ERROR] Failed to start Egress. Make sure Docker is running.
)
pause
