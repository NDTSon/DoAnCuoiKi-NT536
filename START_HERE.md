# 🚀 CHẠY NHANH - 3 BƯỚC

## ⚠️ BẠN ĐANG GẶP LỖI "ERR_CONNECTION_REFUSED"?

**Nguyên nhân:** Server chưa chạy!

---

## ✅ GIẢI PHÁP:

### Bước 1: Mở PowerShell MỚI
- Nhấn `Windows + X`
- Chọn "Windows PowerShell"

### Bước 2: Chạy lệnh
```powershell
cd c:\da-NT536\livekit
.\start-server.bat
```

### Bước 3: Đợi server start
Khi thấy dòng này:
```
starting LiveKit server
```
= Server đã chạy! ✅

---

## 🎯 SAU ĐÓ MỚI MỞ DEMO:

1. **Test connection trước:**
   - Mở: `examples\test-connection.html`
   - Click "Test Connection"
   - Phải thấy ✅

2. **Rồi mới mở demo:**
   - Mở: `examples\streaming-demo.html`
   - Click "Start Stream"
   - Cho phép camera/mic
   - Bắt đầu stream! 🎉

---

## 📚 Hướng dẫn chi tiết:

Mở file này trong browser:
```
examples\index.html
```

Hoặc đọc:
- `FIX_ERROR.md` - Chi tiết lỗi và cách sửa
- `QUICK_START.md` - Hướng dẫn đầy đủ

---

## 🐛 Vẫn lỗi?

Kiểm tra:
1. PowerShell MỚI đã mở chưa?
2. `go version` có hoạt động không?
3. Server có chạy không? (PowerShell window còn mở?)
4. `http://localhost:7880` có mở được không?

**Nếu tất cả OK mà vẫn lỗi, báo lại cho tôi!**

---

**Good luck! 🚀**
