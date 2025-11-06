# 🚀 Hướng Dẫn Cài Đặt và Chạy LiveKit Streaming

## Bước 1: Cài đặt Go (Golang) 1.24+

### Cách 1: Download trực tiếp
1. Truy cập: https://go.dev/dl/
2. Download **go1.24.x.windows-amd64.msi** (hoặc phiên bản mới hơn)
3. Chạy file .msi và làm theo hướng dẫn
4. Mở **PowerShell mới** và kiểm tra:
   ```powershell
   go version
   ```

### Cách 2: Dùng Chocolatey (nếu đã cài)
```powershell
choco install golang
```

### Cách 3: Dùng winget (Windows 11)
```powershell
winget install GoLang.Go
```

## Bước 2: Kiểm tra cài đặt

Mở PowerShell mới và chạy:
```powershell
go version
# Kết quả mong đợi: go version go1.24.x windows/amd64
```

## Bước 3: Cài đặt dependencies

Trong thư mục dự án, chạy:
```powershell
cd c:\da-NT536\livekit
go mod download
```

## Bước 4: Build project

```powershell
# Build server
go build -o livekit-server.exe ./cmd/server

# Hoặc dùng mage (nếu đã cài)
# go install github.com/magefile/mage@latest
# mage
```

## Bước 5: Chạy server

### Cách 1: Development mode (đơn giản nhất)
```powershell
go run ./cmd/server --dev
```

### Cách 2: Chạy file build
```powershell
.\livekit-server.exe --dev
```

### Cách 3: Với config file
```powershell
# Tạo config file
Copy-Item config-sample.yaml config.yaml

# Chỉnh sửa config.yaml theo nhu cầu

# Chạy với config
.\livekit-server.exe --config config.yaml
```

## Bước 6: Test server

Mở browser và truy cập:
- **Server info**: http://localhost:7880
- **Demo UI**: Mở file `examples/streaming-demo.html` trong browser
- **Metrics**: http://localhost:6789/metrics (nếu bật Prometheus)

## Bước 7: Test APIs

### Generate Stream Key
```powershell
$body = @{
    streamer_id = "user123"
    room_name = "my-stream"
} | ConvertTo-Json

Invoke-RestMethod -Uri "http://localhost:7880/api/streaming/keys/generate" `
    -Method POST `
    -Body $body `
    -ContentType "application/json"
```

### Create Chat Room
```powershell
$body = @{
    room_name = "my-stream"
} | ConvertTo-Json

Invoke-RestMethod -Uri "http://localhost:7880/api/streaming/chat/create" `
    -Method POST `
    -Body $body `
    -ContentType "application/json"
```

### Send Chat Message
```powershell
$body = @{
    room_name = "my-stream"
    sender_id = "user123"
    content = "Hello from PowerShell!"
    message_type = "text"
} | ConvertTo-Json

Invoke-RestMethod -Uri "http://localhost:7880/api/streaming/chat/send" `
    -Method POST `
    -Body $body `
    -ContentType "application/json"
```

### Send Reaction
```powershell
$body = @{
    room_name = "my-stream"
    user_id = "user123"
    user_name = "John Doe"
    reaction_type = "heart"
    x = 0.5
    y = 0.8
} | ConvertTo-Json

Invoke-RestMethod -Uri "http://localhost:7880/api/streaming/reactions/send" `
    -Method POST `
    -Body $body `
    -ContentType "application/json"
```

### Get Analytics
```powershell
Invoke-RestMethod -Uri "http://localhost:7880/api/streaming/analytics/stream?room_name=my-stream" `
    -Method GET
```

## Bước 8: Mở Demo UI

```powershell
# Mở trực tiếp trong browser
Start-Process "c:\da-NT536\livekit\examples\streaming-demo.html"
```

## ⚠️ Troubleshooting

### Lỗi: "go: command not found"
➡️ Cài Go như hướng dẫn ở Bước 1

### Lỗi: "cannot find module"
```powershell
go mod tidy
go mod download
```

### Lỗi: Port 7880 đã được sử dụng
```powershell
# Tìm process đang dùng port
netstat -ano | findstr :7880

# Kill process (thay PID bằng số từ lệnh trên)
taskkill /PID <PID> /F
```

### Lỗi: "github.com/gorilla/websocket not found"
```powershell
go get github.com/gorilla/websocket
```

## 📦 Optional: Cài Redis (cho distributed mode)

### Cách 1: Docker
```powershell
docker run -d -p 6379:6379 redis:latest
```

### Cách 2: WSL
```powershell
wsl
sudo apt update
sudo apt install redis-server
redis-server
```

### Cách 3: Windows Native
Download từ: https://github.com/microsoftarchive/redis/releases

## 🎯 Quick Start Script

Tạo file `start.ps1`:
```powershell
# Quick start script
Write-Host "🚀 Starting LiveKit Streaming Server..." -ForegroundColor Green

# Check Go installation
if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
    Write-Host "❌ Go is not installed!" -ForegroundColor Red
    Write-Host "Please install Go from: https://go.dev/dl/" -ForegroundColor Yellow
    exit 1
}

# Check Go version
$goVersion = go version
Write-Host "✅ Found: $goVersion" -ForegroundColor Green

# Download dependencies
Write-Host "📦 Downloading dependencies..." -ForegroundColor Cyan
go mod download

# Run server
Write-Host "🎬 Starting server in development mode..." -ForegroundColor Green
go run ./cmd/server --dev
```

Chạy:
```powershell
.\start.ps1
```

## 🌐 URLs Quan Trọng

| Service | URL | Description |
|---------|-----|-------------|
| **Server** | http://localhost:7880 | Main HTTP endpoint |
| **WebRTC** | tcp://localhost:7881 | ICE/TCP endpoint |
| **UDP** | udp://localhost:50000-60000 | WebRTC UDP ports |
| **Metrics** | http://localhost:6789/metrics | Prometheus metrics |
| **Demo UI** | file:///.../examples/streaming-demo.html | Demo interface |

## 📚 API Endpoints

### Stream Keys
- `POST /api/streaming/keys/generate` - Generate new key
- `POST /api/streaming/keys/validate` - Validate key
- `POST /api/streaming/keys/revoke` - Revoke key
- `GET /api/streaming/keys/list` - List keys

### Chat
- `POST /api/streaming/chat/create` - Create room
- `POST /api/streaming/chat/send` - Send message
- `GET /api/streaming/chat/messages` - Get messages
- `POST /api/streaming/chat/mute` - Mute user
- `POST /api/streaming/chat/ban` - Ban user
- `WS /api/streaming/chat/ws` - WebSocket

### Reactions
- `POST /api/streaming/reactions/send` - Send reaction
- `GET /api/streaming/reactions/stats` - Get stats
- `GET /api/streaming/reactions/recent` - Recent reactions
- `WS /api/streaming/reactions/ws` - WebSocket

### VOD
- `POST /api/streaming/vod/start` - Start recording
- `POST /api/streaming/vod/stop` - Stop recording
- `POST /api/streaming/vod/publish` - Publish video
- `GET /api/streaming/vod/list` - List recordings

### Notifications
- `POST /api/streaming/notifications/subscribe` - Subscribe
- `POST /api/streaming/notifications/unsubscribe` - Unsubscribe
- `GET /api/streaming/notifications/list` - Get notifications
- `POST /api/streaming/notifications/read` - Mark as read
- `WS /api/streaming/notifications/ws` - WebSocket

### Analytics
- `GET /api/streaming/analytics/stream` - Stream analytics
- `GET /api/streaming/analytics/dashboard` - Dashboard data
- `GET /api/streaming/analytics/export` - Export data

## 🎓 Learning Resources

- **LiveKit Docs**: https://docs.livekit.io
- **Go Tutorial**: https://go.dev/tour/
- **WebRTC**: https://webrtc.org/getting-started/overview

## 💡 Tips

1. **Development Mode**: Dùng `--dev` để test nhanh, không cần config
2. **Hot Reload**: Dùng `air` để auto-reload khi code thay đổi
   ```powershell
   go install github.com/cosmtrek/air@latest
   air
   ```
3. **Debug**: Thêm log level:
   ```powershell
   go run ./cmd/server --dev --log-level debug
   ```

## 🔥 Next Steps

1. ✅ Chạy server thành công
2. ✅ Mở demo UI
3. ✅ Test APIs
4. 🚀 Integrate vào app của bạn!

Good luck! 🎉
