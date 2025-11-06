# 🚀 HƯỚNG DẪN CHẠY NHANH

## ✅ Go đã cài đặt thành công!

Bạn đã có Go 1.25.3. Bây giờ chỉ cần 3 bước đơn giản:

---

## 🎯 CÁCH 1: Chạy Nhanh Nhất (Khuyên dùng)

### Bước 1: Mở PowerShell MỚI
⚠️ **QUAN TRỌNG**: Đóng PowerShell cũ, mở PowerShell mới để nhận diện Go

### Bước 2: Chạy lệnh
```powershell
cd c:\da-NT536\livekit
.\quick-start.ps1
```

### Bước 3: Đợi server khởi động
Bạn sẽ thấy:
```
starting LiveKit server
```

### Bước 4: Mở Demo
Khi server đã chạy, mở file này trong trình duyệt:
```
c:\da-NT536\livekit\examples\streaming-demo.html
```

Hoặc chạy lệnh:
```powershell
Start-Process "c:\da-NT536\livekit\examples\streaming-demo.html"
```

---

## 🎯 CÁCH 2: Build và Chạy (Chậm hơn nhưng ổn định)

### Bước 1: Build server
```powershell
cd c:\da-NT536\livekit
.\start.ps1
```

Script này sẽ:
- ✅ Tải dependencies
- ✅ Build file .exe
- ✅ Chạy server

### Bước 2: Test APIs (mở PowerShell thứ 2)
```powershell
cd c:\da-NT536\livekit
.\test-apis.ps1
```

---

## 📡 Kiểm Tra Server Đang Chạy

Mở browser và truy cập:
```
http://localhost:7880
```

Nếu thấy trang LiveKit = Server đã chạy ✅

---

## 🐛 Gặp Lỗi?

### Lỗi 1: "go: command not found"
**Giải pháp**: 
1. Đóng PowerShell hiện tại
2. Mở PowerShell MỚI
3. Chạy lại lệnh

### Lỗi 2: "ERR_CONNECTION_REFUSED" trong Demo
**Giải pháp**: Server chưa chạy!
- Chạy `.\quick-start.ps1` trong PowerShell
- Đợi đến khi thấy "starting LiveKit server"
- Mới mở demo HTML

### Lỗi 3: "cannot find module github.com/gorilla/websocket"
**Giải pháp**:
```powershell
go get github.com/gorilla/websocket
go mod tidy
```

### Lỗi 4: Port 7880 đang được sử dụng
**Giải pháp**:
```powershell
# Tìm process
netstat -ano | findstr :7880

# Kill process (thay <PID> bằng số từ lệnh trên)
taskkill /PID <PID> /F
```

---

## 🎨 Demo UI Features

Khi mở `examples/streaming-demo.html`:

### 1. Start Stream (Bắt đầu phát)
- Click nút "Start Stream"
- Cho phép truy cập camera/mic
- Stream sẽ bắt đầu

### 2. Live Chat (Chat trực tiếp)
- Gõ tin nhắn ở ô chat
- Nhấn Send hoặc Enter
- Tin nhắn hiện ngay lập tức

### 3. Reactions (Biểu cảm)
- Click vào emoji: 👍 ❤️ 😂 😮
- Animation bay lên màn hình
- Stats tự động cập nhật

### 4. Join as Viewer (Tham gia xem)
- Click "Join as Viewer"
- Số viewer tăng lên
- Có thể chat và gửi reaction

---

## 📊 Test APIs

### Generate Stream Key
```powershell
$body = @{
    streamer_id = "user123"
    room_name = "my-room"
} | ConvertTo-Json

Invoke-RestMethod -Uri "http://localhost:7880/api/streaming/keys/generate" `
    -Method POST -Body $body -ContentType "application/json"
```

### Send Chat Message
```powershell
$body = @{
    room_name = "my-room"
    sender_id = "user123"
    sender_name = "John"
    content = "Hello!"
    message_type = "text"
} | ConvertTo-Json

Invoke-RestMethod -Uri "http://localhost:7880/api/streaming/chat/send" `
    -Method POST -Body $body -ContentType "application/json"
```

### Send Reaction
```powershell
$body = @{
    room_name = "my-room"
    user_id = "user123"
    user_name = "John"
    reaction_type = "heart"
    x = 0.5
    y = 0.8
} | ConvertTo-Json

Invoke-RestMethod -Uri "http://localhost:7880/api/streaming/reactions/send" `
    -Method POST -Body $body -ContentType "application/json"
```

---

## 🌐 API Endpoints

### Stream Keys
- `POST /api/streaming/keys/generate` - Tạo stream key
- `POST /api/streaming/keys/validate` - Validate key
- `GET /api/streaming/keys/list` - Danh sách keys

### Chat
- `POST /api/streaming/chat/create` - Tạo phòng chat
- `POST /api/streaming/chat/send` - Gửi tin nhắn
- `GET /api/streaming/chat/messages` - Lấy tin nhắn
- `WS /api/streaming/chat/ws` - WebSocket chat

### Reactions
- `POST /api/streaming/reactions/send` - Gửi reaction
- `GET /api/streaming/reactions/stats` - Thống kê
- `WS /api/streaming/reactions/ws` - WebSocket reactions

### VOD (Recording)
- `POST /api/streaming/vod/start` - Bắt đầu recording
- `POST /api/streaming/vod/stop` - Dừng recording
- `GET /api/streaming/vod/list` - Danh sách recordings

### Analytics
- `GET /api/streaming/analytics/stream?room_name=xxx` - Thống kê stream
- `GET /api/streaming/analytics/dashboard` - Dashboard

---

## 📚 Tài Liệu Chi Tiết

- **SETUP_GUIDE.md** - Hướng dẫn cài đặt đầy đủ
- **docs/STREAMING_FEATURES.md** - Tài liệu tính năng
- **examples/streaming-demo.html** - Demo UI

---

## 💡 Tips

1. **Development Mode**: Dùng `--dev` để test nhanh
2. **Hot Reload**: Cài `air` để auto-reload:
   ```powershell
   go install github.com/cosmtrek/air@latest
   air
   ```
3. **Debug Mode**: Thêm log level:
   ```powershell
   go run ./cmd/server --dev --log-level debug
   ```

---

## ✅ Checklist

- [ ] Go đã cài (go version)
- [ ] PowerShell MỚI đã mở
- [ ] Server đã chạy (.\quick-start.ps1)
- [ ] Demo UI đã mở (streaming-demo.html)
- [ ] Click "Start Stream" hoặc "Join as Viewer"
- [ ] Test chat và reactions

---

## 🆘 Cần Trợ Giúp?

Nếu vẫn gặp lỗi, hãy:
1. Chụp màn hình lỗi
2. Copy log từ PowerShell
3. Kiểm tra xem server có chạy không: `http://localhost:7880`

**Good luck! 🎉**
