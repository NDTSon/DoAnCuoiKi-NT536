# 🚀 Quick Start - LiveKit Streaming Server
# Chạy server nhanh nhất có thể!

Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  LiveKit Quick Start" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# Check Go
Write-Host "Checking Go..." -ForegroundColor Yellow
$goCheck = Get-Command go -ErrorAction SilentlyContinue
if (-not $goCheck) {
    Write-Host "❌ Go not found!" -ForegroundColor Red
    Write-Host "Please install Go from: https://go.dev/dl/" -ForegroundColor Yellow
    Write-Host "Then open a NEW PowerShell window and try again." -ForegroundColor Yellow
    pause
    exit 1
}

$goVersion = go version
Write-Host "✅ $goVersion" -ForegroundColor Green
Write-Host ""

# Run directly without building (faster for testing)
Write-Host "Starting server in development mode..." -ForegroundColor Yellow
Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  Server Starting..." -ForegroundColor Green  
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "📍 URLs:" -ForegroundColor Yellow
Write-Host "   Server: http://localhost:7880" -ForegroundColor White
Write-Host "   Demo UI: file:///c:/da-NT536/livekit/examples/streaming-demo.html" -ForegroundColor White
Write-Host ""
Write-Host "🎯 To test:" -ForegroundColor Yellow
Write-Host "   1. Wait for server to start (you'll see 'starting LiveKit server')" -ForegroundColor White
Write-Host "   2. Open examples/streaming-demo.html in your browser" -ForegroundColor White
Write-Host "   3. Click 'Start Stream' button" -ForegroundColor White
Write-Host ""
Write-Host "Press Ctrl+C to stop" -ForegroundColor Gray
Write-Host ""

# Run without building (go run is faster for development)
go run ./cmd/server --dev
