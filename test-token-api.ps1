# Test LiveKit Token API
Write-Host "🧪 Testing LiveKit Token API..." -ForegroundColor Cyan
Write-Host ""

# Test data
$testData = @{
    room_name = "test-room-$(Get-Date -Format 'HHmmss')"
    identity = "test-user-123"
    is_publisher = $true
} | ConvertTo-Json

Write-Host "📤 Sending request to: http://localhost:7880/api/streaming/token" -ForegroundColor Yellow
Write-Host "Request body:" -ForegroundColor Yellow
Write-Host $testData -ForegroundColor Gray
Write-Host ""

try {
    $response = Invoke-RestMethod -Method Post -Uri "http://localhost:7880/api/streaming/token" -Body $testData -ContentType "application/json"
    
    Write-Host "✅ SUCCESS!" -ForegroundColor Green
    Write-Host ""
    Write-Host "Response:" -ForegroundColor Green
    Write-Host "URL: $($response.url)" -ForegroundColor White
    Write-Host "Token (first 50 chars): $($response.token.Substring(0, [Math]::Min(50, $response.token.Length)))..." -ForegroundColor White
    Write-Host ""
    Write-Host "🎉 Token API is working! You can now use LiveKit streaming." -ForegroundColor Green
    
} catch {
    Write-Host "❌ FAILED!" -ForegroundColor Red
    Write-Host "Error: $($_.Exception.Message)" -ForegroundColor Red
    Write-Host ""
    Write-Host "Possible causes:" -ForegroundColor Yellow
    Write-Host "1. Server is not running" -ForegroundColor White
    Write-Host "2. Server needs restart after code changes" -ForegroundColor White
    Write-Host "3. Port 7880 is blocked" -ForegroundColor White
    Write-Host ""
    Write-Host "Try running: .\livekit-server.exe --dev --bind 0.0.0.0" -ForegroundColor Cyan
}

Write-Host ""
Write-Host "Press any key to exit..."
$null = $Host.UI.RawUI.ReadKey('NoEcho,IncludeKeyDown')
