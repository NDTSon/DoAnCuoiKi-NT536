# Hướng dẫn cấu hình RTMP cho OBS

## Tổng quan

LiveKit hỗ trợ nhận stream RTMP từ OBS (Open Broadcaster Software) thông qua tính năng **Ingress**. Tuy nhiên, LiveKit server chính không tự chạy RTMP server - bạn cần:

1. **Option 1**: Chạy `livekit-ingress` service riêng (khuyến nghị)
2. **Option 2**: Cấu hình RTMPBaseURL trỏ đến ingress service khác

## Kiến trúc

```
┌─────────────┐      RTMP       ┌──────────────┐      WebRTC      ┌─────────────┐
│     OBS     │ ──────────────> │ LiveKit     │ ───────────────> │  Viewers   │
│  Streamer   │                 │ Ingress     │                  │  (Browser) │
└─────────────┘                 │ Service     │                  └─────────────┘
                                └──────────────┘
                                        │
                                        │ gRPC
                                        ▼
                                ┌──────────────┐
                                │ LiveKit      │
                                │ Server       │
                                │ (Main)       │
                                └──────────────┘
```

## Cách 1: Sử dụng LiveKit Ingress Service (Khuyến nghị)

### Bước 1: Cài đặt livekit-ingress

```bash
# Download livekit-ingress binary
# Hoặc dùng Docker:
docker run -p 8080:8080 livekit/livekit-ingress
```

### Bước 2: Cấu hình LiveKit Server

Thêm vào config file (`config.yaml`):

```yaml
ingress:
  rtmp_base_url: rtmp://localhost:1935
```

Hoặc set environment variable:
```bash
export LIVEKIT_INGRESS_RTMP_BASE_URL=rtmp://localhost:1935
```

### Bước 3: Tạo RTMP Ingress

Gọi API để tạo RTMP Ingress:

```bash
curl -X POST http://localhost:7880/twirp/livekit.Ingress/CreateIngress \
  -H "Authorization: Bearer YOUR_API_KEY:YOUR_API_SECRET" \
  -H "Content-Type: application/json" \
  -d '{
    "input_type": 1,
    "name": "My Stream",
    "room_name": "my-room",
    "participant_identity": "streamer-1",
    "participant_name": "Streamer Name"
  }'
```

Response sẽ trả về:
```json
{
  "ingress_id": "IN_xxx",
  "stream_key": "sk_xxx",
  "url": "rtmp://localhost:1935",
  "rtmp_url": "rtmp://localhost:1935/live/sk_xxx"
}
```

### Bước 4: Cấu hình OBS

1. Mở OBS
2. Settings → Stream
3. Service: Custom
4. Server: `rtmp://localhost:1935/live`
5. Stream Key: `sk_xxx` (từ response trên)

## Cách 2: Sử dụng RTMP Server khác (Nginx RTMP, etc.)

Nếu bạn đã có RTMP server (như Nginx RTMP module), chỉ cần:

1. Cấu hình `rtmp_base_url` trỏ đến server đó
2. Tạo Ingress như trên
3. OBS stream đến RTMP server đó với stream key

## Endpoint Helper (Sẽ được thêm)

Để đơn giản hóa, có thể thêm endpoint HTTP:

```
POST /api/streaming/rtmp/create
{
  "room_name": "my-room",
  "streamer_name": "My Name"
}

Response:
{
  "rtmp_url": "rtmp://localhost:1935/live",
  "stream_key": "sk_xxx",
  "full_url": "rtmp://localhost:1935/live/sk_xxx"
}
```

## Lưu ý quan trọng

1. **RTMP Port**: Mặc định RTMP dùng port 1935
2. **Firewall**: Đảm bảo port 1935 được mở
3. **Ingress Service**: Phải chạy riêng, không phải part của main server
4. **Stream Key**: Mỗi ingress có stream key riêng, bảo mật tốt

## Troubleshooting

### Lỗi: "Ingress not connected"
- Kiểm tra Ingress service có đang chạy không
- Kiểm tra RTMPBaseURL có đúng không

### OBS không kết nối được
- Kiểm tra firewall
- Kiểm tra RTMP URL và Stream Key
- Xem logs của Ingress service

### Stream không hiển thị
- Kiểm tra room có được tạo không
- Kiểm tra Ingress state: `ACTIVE` hay `ENDPOINT_INACTIVE`

## Tài liệu tham khảo

- [LiveKit Ingress Docs](https://docs.livekit.io/ingress/)
- [OBS Streaming Guide](https://obsproject.com/wiki/Streaming-With-OBS)

