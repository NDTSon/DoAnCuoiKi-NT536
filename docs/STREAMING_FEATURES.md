# Hệ Thống Live Streaming Tương Tác Thời Gian Thực

Dự án này mở rộng **LiveKit Server** với các tính năng live streaming tương tác theo yêu cầu của đề bài.

## 📋 Tính Năng Đã Triển Khai

### ✅ Tính Năng Cơ Bản (Chiếm 70%)

#### 1. **Xác thực người dùng**
- ✅ Đăng ký, đăng nhập tài khoản
- ✅ Phát live stream từ webcam hoặc màn hình
- ✅ **Stream Key Management** - Mỗi streamer có stream key duy nhất để xác thực
  - File: `pkg/streaming/streamkey.go`
  - Tính năng: Generate, validate, revoke, expire stream keys
  - API: `/api/streaming/keys/*`

#### 2. **Xem live stream**
- ✅ Người xem có thể truy cập stream qua URL
- ✅ Video player với các controls cơ bản (play, pause, điều chỉnh âm lượng)
- Tích hợp sẵn trong LiveKit client SDKs

#### 3. **Trò chuyện trực tiếp (Live Chat)**
- ✅ **Chat Service** - Real-time chat với WebSocket
  - File: `pkg/streaming/chat.go`
  - Tính năng: 
    - Gửi tin nhắn text, emoji
    - Tag/mention users (@username)
    - Rate limiting, slow mode
    - Moderation: mute, ban, delete messages
    - Bad word filtering
  - API: `/api/streaming/chat/*`

### ✅ Tính Năng Nâng Cao (Chiếm 30%)

#### 1. **Thả biểu cảm (Live Reactions)**
- ✅ **Reaction System** - Tương tác real-time với animations
  - File: `pkg/streaming/reactions.go`
  - Tính năng:
    - Nhiều loại reactions: like 👍, heart ❤️, wow 😮, laugh 😂, fire 🔥, etc.
    - Vị trí reactions trên màn hình
    - Rate limiting
    - Thống kê reactions
    - Leaderboard top reactors
  - API: `/api/streaming/reactions/*`

#### 2. **Hiện thị số người đang xem theo thời gian thực**
- ✅ **Analytics Service** - Real-time viewer tracking
  - File: `pkg/streaming/analytics.go`
  - Tính năng:
    - Current viewers, peak viewers, unique viewers
    - Viewer timeline (time-series data)
    - Geographic distribution
    - Platform/device breakdown
    - Engagement metrics

#### 3. **Streaming với nhiều chất lượng (ABR - Adaptive Bitrate)**
- ✅ Đã tích hợp sẵn trong LiveKit SFU
  - File: `pkg/sfu/streamallocator/*`
  - Tự động điều chỉnh chất lượng dựa trên bandwidth
  - Hỗ trợ simulcast (multi-quality encoding)

#### 4. **Chia sẻ màn hình (Screen Sharing)**
- ✅ Đã có trong LiveKit
  - Streamer có thể share screen
  - Hỗ trợ audio từ màn hình

### ✅ Tính Năng Mở Rộng (Điểm Cộng)

#### 1. **Lưu trữ và phát lại (VOD - Video on Demand)**
- ✅ **VOD Service** - Recording và playback
  - File: `pkg/streaming/vod.go`
  - Tính năng:
    - Tự động recording live streams
    - Video storage và management
    - Playback sessions
    - Thumbnail generation
    - Transcoding multiple qualities
    - View analytics
  - API: `/api/streaming/vod/*`

#### 2. **Thông báo (Notifications)**
- ✅ **Notification Service** - Multi-channel notifications
  - File: `pkg/streaming/notifications.go`
  - Tính năng:
    - Subscribe/follow streamers
    - Thông báo khi stream bắt đầu/kết thúc
    - Notifications cho mentions, replies
    - WebSocket, Email, Push notifications
    - Notification preferences
  - API: `/api/streaming/notifications/*`

#### 3. **Analytics Dashboard**
- ✅ **Analytics Service** với dashboard data
  - File: `pkg/streaming/analytics.go`
  - Tính năng:
    - Real-time metrics
    - Viewer statistics
    - Chat activity
    - Reaction statistics
    - Technical metrics (bitrate, latency, buffering)
    - Time-series charts
    - Export analytics data

---

## 🏗️ Kiến Trúc Hệ Thống

```
┌─────────────────────────────────────────────────────────┐
│                    Client Applications                   │
│  (Web Browser, Mobile Apps, Desktop Apps)               │
└────────────────┬────────────────────────────────────────┘
                 │
                 ├─ WebRTC (Video/Audio)
                 ├─ WebSocket (Chat, Reactions, Notifications)
                 └─ HTTP/REST API
                 │
┌────────────────┴────────────────────────────────────────┐
│                    API Gateway Layer                     │
│               pkg/service/streaming_api.go               │
└────────────────┬────────────────────────────────────────┘
                 │
     ┌───────────┼───────────┬───────────┬────────────┐
     │           │           │           │            │
┌────┴────┐ ┌───┴────┐ ┌────┴────┐ ┌───┴─────┐ ┌───┴──────┐
│ Stream  │ │  Chat  │ │Reaction │ │   VOD   │ │Notifica- │
│   Key   │ │Service │ │ Service │ │ Service │ │tion      │
│ Manager │ │        │ │         │ │         │ │ Service  │
└─────────┘ └────────┘ └─────────┘ └─────────┘ └──────────┘
     │           │           │           │            │
     └───────────┴───────────┴───────────┴────────────┘
                 │
         ┌───────┴──────┐
         │  Analytics   │
         │   Service    │
         └──────────────┘
                 │
     ┌───────────┴───────────┐
     │                       │
┌────┴─────┐         ┌──────┴──────┐
│  LiveKit │         │   Storage   │
│  Server  │         │  (Redis,    │
│   Core   │         │   Database) │
└──────────┘         └─────────────┘
```

---

## 📁 Cấu Trúc Dự Án

```
livekit/
├── pkg/
│   ├── streaming/               # 🆕 Module streaming mới
│   │   ├── streamkey.go        # Quản lý stream keys
│   │   ├── chat.go             # Live chat service
│   │   ├── reactions.go        # Reaction system
│   │   ├── vod.go              # Video on demand
│   │   ├── notifications.go    # Notification service
│   │   └── analytics.go        # Analytics & metrics
│   │
│   ├── service/
│   │   ├── streaming_api.go    # 🆕 HTTP/WebSocket APIs
│   │   └── ... (các service hiện có)
│   │
│   ├── rtc/                     # WebRTC core (có sẵn)
│   ├── sfu/                     # SFU với ABR (có sẵn)
│   └── config/                  # Configuration
│
├── cmd/server/                  # Server entry point
└── docs/
    └── STREAMING_API.md         # 🆕 API documentation
```

---

## 🚀 Cài Đặt và Chạy

### Prerequisites
- Go 1.24+
- Redis (cho distributed mode)

### Build từ source

```bash
# Clone repository
git clone https://github.com/livekit/livekit
cd livekit

# Build server
go build -o livekit-server ./cmd/server

# Chạy development mode
./livekit-server --dev
```

### Cấu Hình

Tạo file `config.yaml`:

```yaml
port: 7880

# API Keys
keys:
  devkey: secret

# Redis (optional - cho distributed mode)
redis:
  address: localhost:6379

# RTC Configuration
rtc:
  port_range_start: 50000
  port_range_end: 60000
  tcp_port: 7881
  use_external_ip: true

# Streaming Features Configuration
streaming:
  chat:
    max_message_length: 500
    max_messages_per_minute: 20
    enable_moderation: true
  
  reactions:
    max_reactions_per_minute: 60
    enable_animation: true
    enable_leaderboard: true
  
  vod:
    storage_path: /var/livekit/recordings
    auto_publish: false
    generate_thumbnails: true
    enable_transcoding: true
  
  notifications:
    enable_websocket: true
    enable_email: false
    enable_push: false
  
  analytics:
    enable_real_time: true
    update_interval: 10s
    retention_days: 90
```

---

## 📡 API Documentation

### Stream Key Management

#### Generate Stream Key
```bash
POST /api/streaming/keys/generate
Content-Type: application/json

{
  "streamer_id": "user123",
  "room_name": "my-stream",
  "expires_in": 86400
}
```

#### Validate Stream Key
```bash
POST /api/streaming/keys/validate
Content-Type: application/json

{
  "key": "stream_key_here"
}
```

### Live Chat

#### Send Message
```bash
POST /api/streaming/chat/send
Content-Type: application/json

{
  "room_name": "my-stream",
  "sender_id": "user123",
  "content": "Hello everyone!",
  "message_type": "text",
  "mentioned_users": ["@user456"]
}
```

#### Get Messages
```bash
GET /api/streaming/chat/messages?room_name=my-stream&limit=50
```

#### Mute/Ban Participant
```bash
POST /api/streaming/chat/mute
Content-Type: application/json

{
  "room_name": "my-stream",
  "participant_id": "spammer123",
  "moderator_id": "mod456",
  "duration_secs": 300
}
```

### Reactions

#### Send Reaction
```bash
POST /api/streaming/reactions/send
Content-Type: application/json

{
  "room_name": "my-stream",
  "user_id": "user123",
  "user_name": "John Doe",
  "reaction_type": "heart",
  "x": 0.5,
  "y": 0.8
}
```

#### Get Reaction Stats
```bash
GET /api/streaming/reactions/stats?room_name=my-stream
```

### Analytics

#### Get Stream Analytics
```bash
GET /api/streaming/analytics/stream?room_name=my-stream
```

Response:
```json
{
  "room_name": "my-stream",
  "streamer_id": "user123",
  "start_time": "2025-11-06T10:00:00Z",
  "duration": 3600000000000,
  "total_viewers": 1250,
  "peak_viewers": 850,
  "current_viewers": 420,
  "average_viewers": 523.5,
  "unique_viewers": 1050,
  "total_messages": 3420,
  "total_reactions": 8950,
  "reaction_breakdown": {
    "like": 3200,
    "heart": 2800,
    "fire": 1500,
    "wow": 1450
  },
  "viewers_by_country": {
    "VN": 450,
    "US": 280,
    "JP": 150
  }
}
```

---

## 🔌 WebSocket Connections

### Chat WebSocket
```javascript
const ws = new WebSocket('ws://localhost:7880/api/streaming/chat/ws?room_name=my-stream&user_id=user123');

ws.onmessage = (event) => {
  const message = JSON.parse(event.data);
  console.log('New message:', message);
};

// Send message
ws.send(JSON.stringify({
  type: 'message',
  content: 'Hello!',
  mentioned_users: []
}));
```

### Reactions WebSocket
```javascript
const ws = new WebSocket('ws://localhost:7880/api/streaming/reactions/ws?room_name=my-stream');

ws.onmessage = (event) => {
  const reaction = JSON.parse(event.data);
  // Display reaction animation
  showReactionAnimation(reaction);
};
```

### Notifications WebSocket
```javascript
const ws = new WebSocket('ws://localhost:7880/api/streaming/notifications/ws?user_id=user123');

ws.onmessage = (event) => {
  const notification = JSON.parse(event.data);
  showNotification(notification.title, notification.body);
};
```

---

## 🧪 Testing

### Functional Testing
```bash
# Run all tests
go test ./pkg/streaming/...

# Test specific service
go test ./pkg/streaming -run TestStreamKeyManager
go test ./pkg/streaming -run TestChatService
go test ./pkg/streaming -run TestReactionService
```

### Integration Testing
```bash
# Test with real LiveKit server
go test ./test/... -tags=integration
```

### Load Testing
```bash
# Use LiveKit CLI for load testing
lk load-test \
  --room test-room \
  --duration 5m \
  --publishers 10 \
  --subscribers 100
```

---

## 📊 Monitoring & Metrics

LiveKit exposes Prometheus metrics tại `/metrics`:

- `livekit_stream_viewers_current` - Số viewer hiện tại
- `livekit_stream_messages_total` - Tổng số tin nhắn chat
- `livekit_stream_reactions_total` - Tổng số reactions
- `livekit_stream_bitrate_avg` - Bitrate trung bình
- `livekit_stream_duration_seconds` - Thời lượng stream

Xem real-time tại: `http://localhost:6789/metrics`

---

## 🎯 Các Tính Năng Đặc Biệt

### 1. **Adaptive Bitrate Streaming (ABR)**
- Tự động điều chỉnh chất lượng video theo bandwidth
- Hỗ trợ simulcast (3 layers: low, mid, high)
- Client tự động switch quality levels

### 2. **Smart Rate Limiting**
- Chat: tối đa 20 messages/minute
- Reactions: tối đa 60 reactions/minute
- Slow mode cho chat rooms

### 3. **Advanced Moderation**
- Mute/ban participants
- Delete messages
- Bad word filtering
- Moderator roles

### 4. **Real-time Analytics**
- Updates mỗi 10 giây
- Time-series data cho charts
- Export analytics data

### 5. **Geographic Distribution**
- GeoIP detection
- Viewer distribution by country/region
- Regional CDN optimization (nếu cấu hình)

---

## 🔐 Security Best Practices

1. **API Keys**: Dùng keys > 32 characters
2. **Stream Keys**: Tự động expire, rate limiting
3. **JWT Tokens**: Cho client authentication
4. **CORS**: Configure properly cho production
5. **Rate Limiting**: Tránh spam và abuse
6. **Input Validation**: Sanitize tất cả user input

---

## 📈 Performance Optimization

- **Connection Pooling**: Redis connections
- **Caching**: In-memory cache cho hot data
- **Batch Processing**: Group updates để giảm network calls
- **Lazy Loading**: Load data khi cần
- **Compression**: WebSocket message compression

---

## 🤝 Contributing

Xem [CONTRIBUTING.md](CONTRIBUTING.md) để biết cách contribute.

---

## 📄 License

Apache License 2.0 - Xem [LICENSE](LICENSE)

---

## 📞 Support

- **Documentation**: https://docs.livekit.io
- **Discord**: https://livekit.io/join-slack
- **GitHub Issues**: https://github.com/livekit/livekit/issues

---

## ✨ Credits

Dự án được xây dựng dựa trên **LiveKit** - Open source WebRTC infrastructure.

**Các tính năng streaming được phát triển bởi**: [Your Name]
