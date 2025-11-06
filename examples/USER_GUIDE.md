# 🎥 Hướng Dẫn Sử Dụng Hệ Thống Live Streaming

## 📺 Cách Xem Stream (Viewer)

### Bước 1: Mở Danh Sách Live Streams
1. Mở browser và vào: **`live-streams.html`**
2. Bạn sẽ thấy danh sách các streamer đang phát sóng trực tiếp
3. Mỗi stream hiển thị:
   - Tên stream và streamer
   - Số lượng người đang xem
   - Số lượng messages và reactions
   - Thời gian đã stream

### Bước 2: Tham Gia Xem Stream
1. Click vào stream bạn muốn xem
2. Tự động chuyển sang trang **`watch-stream.html`**
3. Stream sẽ tự động kết nối và phát

### Bước 3: Tương Tác Trong Stream
**💬 Chat:**
- Nhập tin nhắn ở ô chat bên phải
- Click "📤" hoặc nhấn Enter để gửi
- Tin nhắn của bạn sẽ hiển thị cho tất cả mọi người

**❤️ Reactions:**
- Click tab "Reactions"
- Chọn emoji bạn thích: 👍❤️😮😂😢🔥👏🎉
- Reaction sẽ bay lên màn hình video

**ℹ️ Xem Thông Tin:**
- Click tab "Info" để xem:
  - Thông tin streamer
  - Số viewers hiện tại
  - Thời gian stream
  - Tổng số messages và reactions

**🎚️ Điều Chỉnh:**
- Click 🔊/🔇 để bật/tắt âm thanh
- Click ⛶ để xem fullscreen
- Chọn chất lượng: Auto, 1080p, 720p, 480p

---

## 🎤 Cách Phát Stream (Streamer)

### Bước 1: Mở Streamer Dashboard
1. Mở browser và vào: **`streamer-advanced.html`**
2. Bạn sẽ thấy dashboard quản lý stream

### Bước 2: Tạo Stream Key
1. Click nút **"🔄 Generate Key"**
2. Stream key sẽ xuất hiện trong ô
3. (Tùy chọn) Click **"📋 Copy"** để copy key

### Bước 3: Bắt Đầu Stream
1. Click nút **"▶️ Start Stream"**
2. Browser sẽ hỏi quyền truy cập:
   - ✅ Cho phép Microphone (bắt buộc)
   - ✅ Cho phép Screen Share (tự động hiện)
3. Chọn màn hình hoặc cửa sổ để share
4. Click **"Share"**
5. Stream bắt đầu! 🔴

### Bước 4: Quản Lý Stream
**Điều khiển:**
- 🎤 Bật/tắt microphone
- 📹 Bật/tắt camera (nếu có)
- 🖥️ Share màn hình mới
- ⏹️ Stop stream

**Xem Thống Kê:**
- 👥 Số viewers đang xem
- 📈 Peak viewers (cao nhất)
- 💬 Tổng số messages
- ❤️ Tổng số reactions
- ⏱️ Thời gian đã stream

**Tương Tác:**
- Đọc và trả lời chat từ viewers
- Xem danh sách viewers đang online
- Nhận notifications về sự kiện

---

## 🔄 Quy Trình Hoàn Chỉnh

### Scenario 1: Viewer muốn xem stream
```
1. Mở live-streams.html
   ↓
2. Xem danh sách streams đang live
   ↓
3. Click vào stream muốn xem
   ↓
4. Tự động chuyển đến watch-stream.html
   ↓
5. Xem stream + Chat + Reactions
```

### Scenario 2: Streamer muốn phát sóng
```
1. Mở streamer-advanced.html
   ↓
2. Generate stream key
   ↓
3. Click "Start Stream"
   ↓
4. Cho phép Mic + Screen Share
   ↓
5. Stream đang live! Viewers có thể vào xem
   ↓
6. Chat với viewers + Xem thống kê
   ↓
7. Click "Stop Stream" khi xong
```

---

## 🌐 Kết Nối Thực Giữa Streamer và Viewer

### Cách Hoạt Động:

1. **Streamer bắt đầu stream:**
   - Stream được đăng ký vào hệ thống với room ID
   - Thông tin stream xuất hiện trong `live-streams.html`

2. **Viewer thấy stream:**
   - `live-streams.html` hiển thị danh sách streams đang active
   - Mỗi stream có room ID riêng

3. **Viewer join vào stream:**
   - Click vào stream → Chuyển đến `watch-stream.html?stream=ROOM_ID`
   - Kết nối WebSocket đến cùng room với streamer
   - Chat và reactions được sync real-time

4. **Real-time Communication:**
   ```
   Streamer (Room: abc123)
        ↓
   WebSocket Server
        ↓
   Viewers (Room: abc123)
   ```

---

## 📝 Các File Quan Trọng

1. **`platform-overview.html`** - Trang chủ với links đến tất cả tính năng
2. **`live-streams.html`** - Danh sách streams đang live (cho viewer)
3. **`watch-stream.html`** - Trang xem stream cụ thể (cho viewer)
4. **`streamer-advanced.html`** - Dashboard cho streamer
5. **`vod-library.html`** - Thư viện video đã lưu

---

## 🚀 Bắt Đầu Nhanh

### Cho Viewer:
```
1. Mở: platform-overview.html hoặc live-streams.html
2. Click vào stream muốn xem
3. Chat và thả reactions!
```

### Cho Streamer:
```
1. Mở: streamer-advanced.html
2. Generate Key → Start Stream
3. Cho phép Mic + Screen
4. Stream!
```

---

## 💡 Tips

**Viewer:**
- Refresh trang `live-streams.html` để cập nhật danh sách streams
- Bạn có thể mở nhiều stream cùng lúc (nhiều tab)
- Sử dụng fullscreen để trải nghiệm tốt hơn

**Streamer:**
- Test mic và screen trước khi stream
- Theo dõi viewer count để biết mức độ quan tâm
- Đọc và phản hồi chat để tăng tương tác
- Chọn quality phù hợp với băng thông của bạn

---

## ❓ FAQ

**Q: Tôi không thấy stream nào trong live-streams.html?**
A: Có thể chưa có streamer nào đang live. Thử bắt đầu stream của bạn hoặc refresh lại trang.

**Q: Làm sao viewer tìm được stream của tôi?**
A: Khi bạn Start Stream, nó tự động xuất hiện trong danh sách `live-streams.html`.

**Q: Chat có real-time không?**
A: Có! Chat sử dụng WebSocket nên tin nhắn xuất hiện ngay lập tức.

**Q: Có giới hạn số viewers không?**
A: Không, hệ thống hỗ trợ unlimited viewers (phụ thuộc server capacity).

---

**Chúc bạn streaming vui vẻ! 🎉**
