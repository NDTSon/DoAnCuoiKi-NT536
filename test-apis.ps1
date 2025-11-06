# 🧪 Test LiveKit Streaming APIs

$baseUrl = "http://localhost:7880/api/streaming"

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  Testing LiveKit Streaming APIs" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# Test 1: Generate Stream Key
Write-Host "[Test 1] Generate Stream Key..." -ForegroundColor Yellow
$body1 = @{
    streamer_id = "user123"
    room_name = "test-stream-room"
} | ConvertTo-Json

try {
    $response1 = Invoke-RestMethod -Uri "$baseUrl/keys/generate" `
        -Method POST `
        -Body $body1 `
        -ContentType "application/json"
    
    Write-Host "✅ Stream Key Generated:" -ForegroundColor Green
    Write-Host "   Key: $($response1.stream_key)" -ForegroundColor White
    Write-Host "   Room: $($response1.room_name)" -ForegroundColor White
    $streamKey = $response1.stream_key
} catch {
    Write-Host "❌ Failed: $_" -ForegroundColor Red
}
Write-Host ""

Start-Sleep -Seconds 1

# Test 2: Create Chat Room
Write-Host "[Test 2] Create Chat Room..." -ForegroundColor Yellow
$body2 = @{
    room_name = "test-stream-room"
} | ConvertTo-Json

try {
    $response2 = Invoke-RestMethod -Uri "$baseUrl/chat/create" `
        -Method POST `
        -Body $body2 `
        -ContentType "application/json"
    
    Write-Host "✅ Chat Room Created" -ForegroundColor Green
} catch {
    Write-Host "❌ Failed: $_" -ForegroundColor Red
}
Write-Host ""

Start-Sleep -Seconds 1

# Test 3: Send Chat Message
Write-Host "[Test 3] Send Chat Message..." -ForegroundColor Yellow
$body3 = @{
    room_name = "test-stream-room"
    sender_id = "user123"
    sender_name = "John Doe"
    content = "Hello from PowerShell! 👋"
    message_type = "text"
} | ConvertTo-Json

try {
    $response3 = Invoke-RestMethod -Uri "$baseUrl/chat/send" `
        -Method POST `
        -Body $body3 `
        -ContentType "application/json"
    
    Write-Host "✅ Message Sent" -ForegroundColor Green
} catch {
    Write-Host "❌ Failed: $_" -ForegroundColor Red
}
Write-Host ""

Start-Sleep -Seconds 1

# Test 4: Get Chat Messages
Write-Host "[Test 4] Get Chat Messages..." -ForegroundColor Yellow
try {
    $response4 = Invoke-RestMethod -Uri "$baseUrl/chat/messages?room_name=test-stream-room&limit=10" `
        -Method GET
    
    Write-Host "✅ Messages Retrieved: $($response4.messages.Count) messages" -ForegroundColor Green
    foreach ($msg in $response4.messages) {
        Write-Host "   [$($msg.sender_name)]: $($msg.content)" -ForegroundColor White
    }
} catch {
    Write-Host "❌ Failed: $_" -ForegroundColor Red
}
Write-Host ""

Start-Sleep -Seconds 1

# Test 5: Send Reaction
Write-Host "[Test 5] Send Reaction..." -ForegroundColor Yellow
$body5 = @{
    room_name = "test-stream-room"
    user_id = "user456"
    user_name = "Jane Smith"
    reaction_type = "heart"
    x = 0.5
    y = 0.8
} | ConvertTo-Json

try {
    $response5 = Invoke-RestMethod -Uri "$baseUrl/reactions/send" `
        -Method POST `
        -Body $body5 `
        -ContentType "application/json"
    
    Write-Host "✅ Reaction Sent: ❤️" -ForegroundColor Green
} catch {
    Write-Host "❌ Failed: $_" -ForegroundColor Red
}
Write-Host ""

Start-Sleep -Seconds 1

# Test 6: Get Reaction Stats
Write-Host "[Test 6] Get Reaction Stats..." -ForegroundColor Yellow
try {
    $response6 = Invoke-RestMethod -Uri "$baseUrl/reactions/stats?room_name=test-stream-room" `
        -Method GET
    
    Write-Host "✅ Reaction Stats:" -ForegroundColor Green
    Write-Host "   Total: $($response6.total_count)" -ForegroundColor White
    foreach ($stat in $response6.type_counts.PSObject.Properties) {
        Write-Host "   $($stat.Name): $($stat.Value)" -ForegroundColor White
    }
} catch {
    Write-Host "❌ Failed: $_" -ForegroundColor Red
}
Write-Host ""

Start-Sleep -Seconds 1

# Test 7: Start VOD Recording
Write-Host "[Test 7] Start VOD Recording..." -ForegroundColor Yellow
$body7 = @{
    room_name = "test-stream-room"
    streamer_id = "user123"
    title = "Test Stream Recording"
    description = "Testing VOD functionality"
} | ConvertTo-Json

try {
    $response7 = Invoke-RestMethod -Uri "$baseUrl/vod/start" `
        -Method POST `
        -Body $body7 `
        -ContentType "application/json"
    
    Write-Host "✅ Recording Started" -ForegroundColor Green
    Write-Host "   ID: $($response7.recording_id)" -ForegroundColor White
    $recordingId = $response7.recording_id
} catch {
    Write-Host "❌ Failed: $_" -ForegroundColor Red
}
Write-Host ""

Start-Sleep -Seconds 2

# Test 8: Stop VOD Recording
if ($recordingId) {
    Write-Host "[Test 8] Stop VOD Recording..." -ForegroundColor Yellow
    $body8 = @{
        recording_id = $recordingId
    } | ConvertTo-Json

    try {
        $response8 = Invoke-RestMethod -Uri "$baseUrl/vod/stop" `
            -Method POST `
            -Body $body8 `
            -ContentType "application/json"
        
        Write-Host "✅ Recording Stopped" -ForegroundColor Green
    } catch {
        Write-Host "❌ Failed: $_" -ForegroundColor Red
    }
    Write-Host ""
}

Start-Sleep -Seconds 1

# Test 9: Subscribe to Notifications
Write-Host "[Test 9] Subscribe to Notifications..." -ForegroundColor Yellow
$body9 = @{
    user_id = "user123"
    streamer_id = "user456"
} | ConvertTo-Json

try {
    $response9 = Invoke-RestMethod -Uri "$baseUrl/notifications/subscribe" `
        -Method POST `
        -Body $body9 `
        -ContentType "application/json"
    
    Write-Host "✅ Subscribed to Notifications" -ForegroundColor Green
} catch {
    Write-Host "❌ Failed: $_" -ForegroundColor Red
}
Write-Host ""

Start-Sleep -Seconds 1

# Test 10: Get Stream Analytics
Write-Host "[Test 10] Get Stream Analytics..." -ForegroundColor Yellow
try {
    $response10 = Invoke-RestMethod -Uri "$baseUrl/analytics/stream?room_name=test-stream-room" `
        -Method GET
    
    Write-Host "✅ Analytics Retrieved:" -ForegroundColor Green
    Write-Host "   Room: $($response10.room_name)" -ForegroundColor White
    Write-Host "   Current Viewers: $($response10.current_viewers)" -ForegroundColor White
    Write-Host "   Peak Viewers: $($response10.peak_viewers)" -ForegroundColor White
    Write-Host "   Total Messages: $($response10.total_messages)" -ForegroundColor White
    Write-Host "   Total Reactions: $($response10.total_reactions)" -ForegroundColor White
} catch {
    Write-Host "❌ Failed: $_" -ForegroundColor Red
}
Write-Host ""

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  All Tests Completed! ✅" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "Next steps:" -ForegroundColor Yellow
Write-Host "  1. Open examples\streaming-demo.html in browser" -ForegroundColor White
Write-Host "  2. Test WebSocket connections (Chat, Reactions, Notifications)" -ForegroundColor White
Write-Host "  3. Integrate APIs into your application" -ForegroundColor White
Write-Host ""
