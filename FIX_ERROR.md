# ⚠️ LỖI VÀ CÁCH SỬA

## 🔴 Lỗi Bạn Đang Gặp:

Từ ảnh bạn gửi, tôi thấy:

1. ❌ **"Error starting stream: Failed to fetch"**
2. ❌ **"ERR_CONNECTION_REFUSED"** trên tất cả WebSocket
3. ❌ **"Disconnected"** status
4. ❌ Các POST requests đều failed

## 🎯 NGUYÊN NHÂN:

**SERVER CHƯA CHẠY!**

Browser đang cố kết nối tới `http://localhost:7880` nhưng không có gì ở đó.

## ✅ GIẢI PHÁP - 3 BƯỚC ĐƠN GIẢN:

---

### BƯỚC 1: MỞ POWERSHELL MỚI ⚠️

**QUAN TRỌNG**: PowerShell hiện tại chưa nhận Go!

1. Đóng tất cả PowerShell cũ
2. Nhấn `Windows + X`
3. Chọn "Windows PowerShell" hoặc "Terminal"

---

### BƯỚC 2: CHẠY SERVER

Trong PowerShell mới, gõ từng lệnh:

```powershell
cd c:\da-NT536\livekit
```

Sau đó chọn 1 trong 2 cách:

#### Cách A: Dùng file .bat (Đơn giản nhất)
```cmd
.\run-server.bat
```

#### Cách B: Dùng PowerShell script
```powershell
.\quick-start.ps1
```

#### Cách C: Chạy trực tiếp
```powershell
go run ./cmd/server --dev
```

---

### BƯỚC 3: ĐỢI SERVER KHỞI ĐỘNG

Bạn sẽ thấy các dòng này:

```
starting LiveKit server
port: 7880
...
```

**KHI THẤY DÒNG NÀY = SERVER ĐÃ CHẠY!** ✅

---

### BƯỚC 4: TEST CONNECTION TRƯỚC

**ĐỪNG MỞ streaming-demo.html ngay!** Hãy test connection trước:

1. Mở file này trong browser:
```
c:\da-NT536\livekit\examples\test-connection.html
```

2. Click nút "🔌 Test Connection"

3. Nếu thấy:
   - ✅ "Server đang hoạt động" → OK, tiếp tục BƯỚC 5
   - ❌ "Không thể kết nối" → Server chưa chạy, quay lại BƯỚC 2

---

### BƯỚC 5: MỞ DEMO

Khi test-connection.html hiện ✅, mới mở demo:

```
c:\da-NT536\livekit\examples\streaming-demo.html
```

Hoặc chạy lệnh:
```powershell
Start-Process "c:\da-NT536\livekit\examples\streaming-demo.html"
```

---

## 🔍 KIỂM TRA NHANH:

### Test 1: Go đã cài chưa?
```powershell
go version
```

✅ Phải thấy: `go version go1.25.3 windows/amd64`
❌ Nếu thấy: "go: command not found" → Mở PowerShell MỚI!

### Test 2: Server có chạy không?

Mở browser, truy cập:
```
http://localhost:7880
```

✅ Phải thấy: Trang LiveKit (có thể trống hoặc có nội dung)
❌ Nếu thấy: "This site can't be reached" → Server chưa chạy!

### Test 3: Port 7880 có free không?

```powershell
netstat -ano | findstr :7880
```

- Nếu có kết quả → Port đang được dùng, kill process:
```powershell
# Lấy PID từ lệnh trên, rồi:
taskkill /PID <PID> /F
```

- Nếu không có kết quả → Port free, OK!

---

## 📋 CHECKLIST:

Trước khi mở demo, đảm bảo:

- [ ] PowerShell MỚI đã mở (không phải cái cũ)
- [ ] `go version` hoạt động
- [ ] Đã cd vào `c:\da-NT536\livekit`
- [ ] Đã chạy `.\run-server.bat` hoặc `go run ./cmd/server --dev`
- [ ] Thấy dòng "starting LiveKit server"
- [ ] `http://localhost:7880` mở được trong browser
- [ ] `test-connection.html` hiển thị ✅

**CHỈ KHI TẤT CẢ ĐỀU ✅ MỚI MỞ streaming-demo.html!**

---

## 🐛 NẾU VẪN LỖI:

### Lỗi: "go: cannot find module"
```powershell
go mod tidy
go mod download
go get github.com/gorilla/websocket
```

### Lỗi: "port 7880 already in use"
```powershell
netstat -ano | findstr :7880
taskkill /PID <PID> /F
```

### Lỗi: Build failed
Không cần build! Dùng `go run` thay vì:
```powershell
go run ./cmd/server --dev
```

---

## 📸 SCREENSHOT CHECKLIST:

Nếu vẫn lỗi, gửi cho tôi screenshot của:

1. PowerShell window sau khi chạy `go version`
2. PowerShell window sau khi chạy `go run ./cmd/server --dev`
3. Browser khi mở `http://localhost:7880`
4. Browser khi mở `test-connection.html`

---

## 🎯 FLOW ĐÚNG:

```
1. Mở PowerShell MỚI
   ↓
2. cd c:\da-NT536\livekit
   ↓
3. .\run-server.bat
   ↓
4. Đợi "starting LiveKit server"
   ↓
5. Mở test-connection.html
   ↓
6. Click "Test Connection" → Phải thấy ✅
   ↓
7. MỚI mở streaming-demo.html
```

**ĐỪNG BỎ QUA BƯỚC NÀO!**

---

## 💡 TÓM TẮT:

**LỖI**: Server chưa chạy
**FIX**: Mở PowerShell mới → `.\run-server.bat` → Đợi start → Test connection → Mới mở demo

**Hãy làm theo đúng thứ tự và báo cho tôi kết quả!** 🚀
