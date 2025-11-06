# 🚀 LiveKit Streaming Server - Quick Start Script

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  LiveKit Streaming Server Setup" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# Step 1: Check Go installation
Write-Host "[1/5] Checking Go installation..." -ForegroundColor Yellow
try {
    $goVersion = go version 2>&1
    if ($LASTEXITCODE -eq 0) {
        Write-Host "✅ $goVersion" -ForegroundColor Green
    } else {
        throw "Go not found"
    }
} catch {
    Write-Host "❌ Go is not found in PATH!" -ForegroundColor Red
    Write-Host "" 
    Write-Host "Please close this PowerShell window and open a NEW PowerShell window!" -ForegroundColor Yellow
    Write-Host "Then run this script again: .\start.ps1" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "If still not working, manually add Go to PATH:" -ForegroundColor Yellow
    Write-Host "  1. Open: System Properties > Environment Variables" -ForegroundColor White
    Write-Host "  2. Add to PATH: C:\Program Files\Go\bin" -ForegroundColor White
    Write-Host ""
    pause
    exit 1
}

Write-Host ""

# Step 2: Download dependencies
Write-Host "[2/5] Downloading dependencies..." -ForegroundColor Yellow
$env:GOPROXY = "https://proxy.golang.org,direct"
go mod download
if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ Failed to download dependencies!" -ForegroundColor Red
    Write-Host ""
    Write-Host "Trying to fix..." -ForegroundColor Yellow
    go mod tidy
    go mod download
    if ($LASTEXITCODE -ne 0) {
        Write-Host "❌ Still failed. Please check your internet connection." -ForegroundColor Red
        pause
        exit 1
    }
}
Write-Host "✅ Dependencies downloaded" -ForegroundColor Green
Write-Host ""

# Step 3: Add gorilla/websocket if not exists
Write-Host "[3/5] Ensuring gorilla/websocket is installed..." -ForegroundColor Yellow
go get github.com/gorilla/websocket
Write-Host "✅ Dependencies verified" -ForegroundColor Green
Write-Host ""

# Step 4: Build server
Write-Host "[4/5] Building LiveKit server..." -ForegroundColor Yellow
go build -o livekit-server.exe ./cmd/server
if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ Build failed!" -ForegroundColor Red
    pause
    exit 1
}
Write-Host "✅ Server built successfully: livekit-server.exe" -ForegroundColor Green
Write-Host ""

# Step 5: Run server
Write-Host "[5/5] Starting server in development mode..." -ForegroundColor Yellow
Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  Server is starting..." -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "📍 Server URLs:" -ForegroundColor Yellow
Write-Host "   HTTP API: http://localhost:7880" -ForegroundColor White
Write-Host "   WebRTC:   tcp://localhost:7881" -ForegroundColor White
Write-Host ""
Write-Host "📡 Streaming API Endpoints:" -ForegroundColor Yellow
Write-Host "   Stream Keys: http://localhost:7880/api/streaming/keys/*" -ForegroundColor White
Write-Host "   Chat:        http://localhost:7880/api/streaming/chat/*" -ForegroundColor White
Write-Host "   Reactions:   http://localhost:7880/api/streaming/reactions/*" -ForegroundColor White
Write-Host "   VOD:         http://localhost:7880/api/streaming/vod/*" -ForegroundColor White
Write-Host "   Notify:      http://localhost:7880/api/streaming/notifications/*" -ForegroundColor White
Write-Host "   Analytics:   http://localhost:7880/api/streaming/analytics/*" -ForegroundColor White
Write-Host ""
Write-Host "🎨 Demo UI: examples\streaming-demo.html" -ForegroundColor Yellow
Write-Host ""
Write-Host "Press Ctrl+C to stop the server" -ForegroundColor Gray
Write-Host ""

.\livekit-server.exe --dev
