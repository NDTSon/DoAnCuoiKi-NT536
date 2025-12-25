// Copyright 2025 LiveKit, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/livekit/protocol/auth"
	"github.com/livekit/protocol/livekit"
	"github.com/livekit/protocol/logger"

	"github.com/livekit/livekit-server/pkg/streaming"
)

// StreamingAPIService provides HTTP/WebSocket APIs for the streaming features
type StreamingAPIService struct {
	streamKeyManager    *streaming.StreamKeyManager
	chatService         *streaming.ChatService
	reactionService     *streaming.ReactionService
	vodService          *streaming.VODService
	notificationService *streaming.NotificationService
	analyticsService    *streaming.AnalyticsService
	egressService       *EgressService
	logger              logger.Logger
	upgrader            websocket.Upgrader
	apiKey              string
	apiSecret           string
}

// NewStreamingAPIService creates a new streaming API service
func NewStreamingAPIService(egressService *EgressService) *StreamingAPIService {
	return &StreamingAPIService{
		streamKeyManager:    streaming.NewStreamKeyManager(),
		chatService:         streaming.NewChatService(),
		reactionService:     streaming.NewReactionService(nil),
		vodService:          streaming.NewVODService(nil),
		notificationService: streaming.NewNotificationService(nil),
		analyticsService:    streaming.NewAnalyticsService(nil),
		egressService:       egressService,
		logger:              logger.GetLogger(),
		apiKey:              "devkey", // Default dev key - should load from config
		apiSecret:           "secret", // Default dev secret - should load from config
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // Configure properly for production
			},
		},
	}
}

// RegisterHTTPHandlers registers all HTTP handlers
func (s *StreamingAPIService) RegisterHTTPHandlers(mux *http.ServeMux) {
	// LiveKit Token Generation (NEW)
	mux.HandleFunc("/api/streaming/token", s.handleGetToken)

	// Stream Key Management
	mux.HandleFunc("/api/streaming/keys/generate", s.handleGenerateStreamKey)
	mux.HandleFunc("/api/streaming/keys/validate", s.handleValidateStreamKey)
	mux.HandleFunc("/api/streaming/keys/revoke", s.handleRevokeStreamKey)
	mux.HandleFunc("/api/streaming/keys/list", s.handleListStreamKeys)

	// Chat
	mux.HandleFunc("/api/streaming/chat/create", s.handleCreateChatRoom)
	mux.HandleFunc("/api/streaming/chat/send", s.handleSendChatMessage)
	mux.HandleFunc("/api/streaming/chat/messages", s.handleGetChatMessages)
	mux.HandleFunc("/api/streaming/chat/mute", s.handleMuteParticipant)
	mux.HandleFunc("/api/streaming/chat/ban", s.handleBanParticipant)
	mux.HandleFunc("/api/streaming/chat/ws", s.handleChatWebSocket)

	// Reactions
	mux.HandleFunc("/api/streaming/reactions/send", s.handleSendReaction)
	mux.HandleFunc("/api/streaming/reactions/stats", s.handleGetReactionStats)
	mux.HandleFunc("/api/streaming/reactions/recent", s.handleGetRecentReactions)
	mux.HandleFunc("/api/streaming/reactions/ws", s.handleReactionsWebSocket)

	// VOD
	mux.HandleFunc("/api/streaming/vod/start", s.handleStartRecording)
	mux.HandleFunc("/api/streaming/vod/stop", s.handleStopRecording)
	mux.HandleFunc("/api/streaming/vod/publish", s.handlePublishRecording)
	mux.HandleFunc("/api/streaming/vod/list", s.handleListRecordings)
	mux.HandleFunc("/api/streaming/vod/play", s.handlePlayRecording)

	// Notifications
	mux.HandleFunc("/api/streaming/notifications/subscribe", s.handleSubscribe)
	mux.HandleFunc("/api/streaming/notifications/unsubscribe", s.handleUnsubscribe)
	mux.HandleFunc("/api/streaming/notifications/list", s.handleGetNotifications)
	mux.HandleFunc("/api/streaming/notifications/read", s.handleMarkAsRead)
	mux.HandleFunc("/api/streaming/notifications/ws", s.handleNotificationsWebSocket)

	// Analytics
	mux.HandleFunc("/api/streaming/analytics/stream", s.handleGetStreamAnalytics)
	mux.HandleFunc("/api/streaming/analytics/dashboard", s.handleGetDashboard)
	mux.HandleFunc("/api/streaming/analytics/export", s.handleExportAnalytics)

	s.logger.Infow("registered streaming API handlers")
}

// LiveKit Token Generation Handler
func (s *StreamingAPIService) handleGetToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request
	var req struct {
		RoomName    string `json:"room_name"`
		Identity    string `json:"identity"`
		IsPublisher bool   `json:"is_publisher"` // true for streamer, false for viewer
	}

	if r.Method == http.MethodPost {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
	} else {
		req.RoomName = r.URL.Query().Get("room_name")
		req.Identity = r.URL.Query().Get("identity")
		req.IsPublisher = r.URL.Query().Get("is_publisher") == "true"
	}

	if req.RoomName == "" || req.Identity == "" {
		http.Error(w, "room_name and identity required", http.StatusBadRequest)
		return
	}

	// Create video grant
	grant := &auth.VideoGrant{
		RoomJoin: true,
		Room:     req.RoomName,
	}

	if req.IsPublisher {
		// Streamer permissions
		grant.SetCanPublish(true)
		grant.SetCanPublishData(true)
		grant.SetCanSubscribe(true)
		grant.RoomRecord = true // Allow recording
	} else {
		// Viewer permissions
		grant.SetCanPublish(false)
		grant.SetCanPublishData(true) // Allow chat/reactions
		grant.SetCanSubscribe(true)
	}

	// Create access token
	at := auth.NewAccessToken(s.apiKey, s.apiSecret)
	at.AddGrant(grant).
		SetIdentity(req.Identity).
		SetValidFor(24 * time.Hour) // Valid for 24 hours

	token, err := at.ToJWT()
	if err != nil {
		s.logger.Errorw("failed to generate token", err)
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	// Return token
	response := map[string]string{
		"token": token,
		"url":   "ws://localhost:7880", // Should be configurable
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Stream Key Management Handlers

func (s *StreamingAPIService) handleGenerateStreamKey(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		StreamerID  string                 `json:"streamer_id"`
		RoomName    string                 `json:"room_name"`
		ExpiresIn   *int64                 `json:"expires_in,omitempty"` // seconds
		Permissions map[string]interface{} `json:"permissions,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var expiresIn *time.Duration
	if req.ExpiresIn != nil {
		d := time.Duration(*req.ExpiresIn) * time.Second
		expiresIn = &d
	}

	streamKey, err := s.streamKeyManager.GenerateStreamKey(
		r.Context(),
		livekit.ParticipantIdentity(req.StreamerID),
		livekit.RoomName(req.RoomName),
		nil, // permissions
		expiresIn,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(streamKey)
}

func (s *StreamingAPIService) handleValidateStreamKey(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Key string `json:"key"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	streamKey, err := s.streamKeyManager.ValidateStreamKey(r.Context(), req.Key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	// Mark as used
	s.streamKeyManager.MarkKeyAsUsed(r.Context(), req.Key)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"valid": true,
		"key":   streamKey,
	})
}

func (s *StreamingAPIService) handleRevokeStreamKey(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Key string `json:"key"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err := s.streamKeyManager.RevokeStreamKey(r.Context(), req.Key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
	})
}

func (s *StreamingAPIService) handleListStreamKeys(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	streamerID := r.URL.Query().Get("streamer_id")
	if streamerID == "" {
		http.Error(w, "streamer_id required", http.StatusBadRequest)
		return
	}

	keys, err := s.streamKeyManager.GetStreamKeysByStreamer(
		r.Context(),
		livekit.ParticipantIdentity(streamerID),
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(keys)
}

// Chat Handlers

func (s *StreamingAPIService) handleCreateChatRoom(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		RoomName string `json:"room_name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	room, err := s.chatService.CreateChatRoom(
		r.Context(),
		livekit.RoomName(req.RoomName),
		nil,
	)

	// If room already exists, return success anyway
	if err != nil && err.Error() == "chat room already exists" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"room_name": req.RoomName,
			"success":   true,
			"message":   "Room already exists",
		})
		return
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return simplified response to avoid JSON serialization issues
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"room_name":         string(room.RoomName),
		"created_at":        room.CreatedAt,
		"message_count":     len(room.Messages),
		"participant_count": len(room.Participants),
		"success":           true,
	})
}

func (s *StreamingAPIService) handleSendChatMessage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		RoomName       string   `json:"room_name"`
		SenderID       string   `json:"sender_id"`
		SenderName     string   `json:"sender_name"`
		Content        string   `json:"content"`
		MessageType    string   `json:"message_type"`
		MentionedUsers []string `json:"mentioned_users,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	mentioned := make([]livekit.ParticipantIdentity, len(req.MentionedUsers))
	for i, u := range req.MentionedUsers {
		mentioned[i] = livekit.ParticipantIdentity(u)
	}

	message, err := s.chatService.SendMessage(
		r.Context(),
		livekit.RoomName(req.RoomName),
		livekit.ParticipantIdentity(req.SenderID),
		req.Content,
		streaming.ChatMessageType(req.MessageType),
		mentioned,
		nil,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(message)
}

func (s *StreamingAPIService) handleGetChatMessages(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	roomName := r.URL.Query().Get("room_name")
	if roomName == "" {
		http.Error(w, "room_name required", http.StatusBadRequest)
		return
	}

	messages, err := s.chatService.GetMessages(
		r.Context(),
		livekit.RoomName(roomName),
		50,
		nil,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(messages)
}

func (s *StreamingAPIService) handleMuteParticipant(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		RoomName      string `json:"room_name"`
		ParticipantID string `json:"participant_id"`
		ModeratorID   string `json:"moderator_id"`
		DurationSecs  int64  `json:"duration_secs"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	duration := time.Duration(req.DurationSecs) * time.Second

	err := s.chatService.MuteParticipant(
		r.Context(),
		livekit.RoomName(req.RoomName),
		livekit.ParticipantIdentity(req.ParticipantID),
		livekit.ParticipantIdentity(req.ModeratorID),
		duration,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

func (s *StreamingAPIService) handleBanParticipant(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		RoomName      string `json:"room_name"`
		ParticipantID string `json:"participant_id"`
		ModeratorID   string `json:"moderator_id"`
		DurationSecs  int64  `json:"duration_secs"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	duration := time.Duration(req.DurationSecs) * time.Second

	err := s.chatService.BanParticipant(
		r.Context(),
		livekit.RoomName(req.RoomName),
		livekit.ParticipantIdentity(req.ParticipantID),
		livekit.ParticipantIdentity(req.ModeratorID),
		duration,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

// WebSocket Handlers

func (s *StreamingAPIService) handleChatWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		s.logger.Errorw("failed to upgrade websocket", err)
		return
	}
	defer conn.Close()

	// Handle chat WebSocket connection
	// Implementation would handle real-time chat messages
	s.logger.Infow("chat websocket connected")
}

func (s *StreamingAPIService) handleReactionsWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		s.logger.Errorw("failed to upgrade websocket", err)
		return
	}
	defer conn.Close()

	// Handle reactions WebSocket connection
	s.logger.Infow("reactions websocket connected")
}

func (s *StreamingAPIService) handleNotificationsWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		s.logger.Errorw("failed to upgrade websocket", err)
		return
	}
	defer conn.Close()

	// Handle notifications WebSocket connection
	s.logger.Infow("notifications websocket connected")
}

// Reaction Handlers

func (s *StreamingAPIService) handleSendReaction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		RoomName     string  `json:"room_name"`
		UserID       string  `json:"user_id"`
		UserName     string  `json:"user_name"`
		ReactionType string  `json:"reaction_type"`
		X            float64 `json:"x,omitempty"`
		Y            float64 `json:"y,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var position *streaming.ReactionPosition
	if req.X != 0 || req.Y != 0 {
		position = &streaming.ReactionPosition{X: req.X, Y: req.Y}
	}

	reaction, err := s.reactionService.SendReaction(
		r.Context(),
		livekit.RoomName(req.RoomName),
		livekit.ParticipantIdentity(req.UserID),
		req.UserName,
		streaming.ReactionType(req.ReactionType),
		position,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(reaction)
}

func (s *StreamingAPIService) handleGetReactionStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	roomName := r.URL.Query().Get("room_name")
	if roomName == "" {
		http.Error(w, "room_name required", http.StatusBadRequest)
		return
	}

	stats, err := s.reactionService.GetReactionStats(
		r.Context(),
		livekit.RoomName(roomName),
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

func (s *StreamingAPIService) handleGetRecentReactions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	roomName := r.URL.Query().Get("room_name")
	if roomName == "" {
		http.Error(w, "room_name required", http.StatusBadRequest)
		return
	}

	reactions, err := s.reactionService.GetRecentReactions(
		r.Context(),
		livekit.RoomName(roomName),
		50,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(reactions)
}

// VOD Handlers - Simplified implementations
func (s *StreamingAPIService) handleStartRecording(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RoomName     string `json:"room_name"`
		StreamerID   string `json:"streamer_id"`
		StreamerName string `json:"streamer_name"`
		Title        string `json:"title"`
		// Tracks
		VideoTrackID string `json:"video_track_id"`
		AudioTrackID string `json:"audio_track_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.RoomName == "" || req.StreamerID == "" {
		http.Error(w, "room_name and streamer_id required", http.StatusBadRequest)
		return
	}

	// 1. Create VOD record
	rec, err := s.vodService.StartRecording(
		r.Context(),
		livekit.RoomName(req.RoomName),
		livekit.ParticipantIdentity(req.StreamerID),
		req.StreamerName,
		req.Title,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 2. Start Egress
	// We use EncodedFileOutput to save to local disk
	// NOTE: When running Egress in Docker, we must map to the container's path.
	// Our run-egress.bat maps host 'data/recordings' to container '/out'.
	filepath := fmt.Sprintf("/out/%s.mp4", rec.ID)

	var info *livekit.EgressInfo

	// Use RoomCompositeEgress to support source switching (Cam <-> Screen) dynamically
	// This records the room layout, ensuring whatever the streamer publishes is captured.
	egressReq := &livekit.RoomCompositeEgressRequest{
		RoomName:  req.RoomName,
		Layout:    "grid-light",
		AudioOnly: false,
		FileOutputs: []*livekit.EncodedFileOutput{
			{
				FileType: livekit.EncodedFileType_MP4,
				Filepath: filepath,
			},
		},
	}
	info, err = s.egressService.StartRoomCompositeEgress(r.Context(), egressReq)
	if err != nil {
		// Cleanup VOD record if Egress fails
		s.vodService.DeleteRecording(r.Context(), rec.ID)
		http.Error(w, fmt.Sprintf("Failed to start egress: %v", err), http.StatusInternalServerError)
		return
	}

	// Save Egress ID to metadata for stopping later
	s.vodService.UpdateRecordingMetadata(
		r.Context(),
		rec.ID,
		nil,
		nil,
		nil,
		nil,
	)
	// Hack: We should store egressID in the VOD record, but the struct is fixed.
	// We can put it in metadata mapping.
	rec.Metadata["egress_id"] = info.EgressId

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":      true,
		"recording_id": rec.ID,
		"egress_id":    info.EgressId,
		"status":       "recording",
	})
}

func (s *StreamingAPIService) handleStopRecording(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		RecordingID string `json:"recording_id"`
		EgressID    string `json:"egress_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.EgressID != "" {
		// Stop Egress
		_, err := s.egressService.StopEgress(r.Context(), &livekit.StopEgressRequest{
			EgressId: req.EgressID,
		})
		if err != nil {
			s.logger.Errorw("failed to stop egress", err, "egressID", req.EgressID)
			// Continue to stop VOD record anyway
		}
	}

	if req.RecordingID != "" {
		// Stop VOD record
		// We don't know exact duration/size yet, but we'll mark it as processed
		// In a real system, we'd wait for Egress webhooks to update this accuracy.
		err := s.vodService.StopRecording(
			r.Context(),
			req.RecordingID,
			0, // Duration unknown
			0, // Size unknown
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
	})
}

func (s *StreamingAPIService) handlePublishRecording(w http.ResponseWriter, r *http.Request) {
	// Implementation
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}

func (s *StreamingAPIService) handleListRecordings(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	streamerID := r.URL.Query().Get("streamer_id")

	var recordings []*streaming.VODRecording
	var err error

	if streamerID == "" || streamerID == "ALL" {
		recordings, err = s.vodService.ListAllRecordings(
			r.Context(),
			50,
			0,
		)
	} else {
		recordings, err = s.vodService.ListRecordingsByStreamer(
			r.Context(),
			livekit.ParticipantIdentity(streamerID),
			50, // default limit
			0,  // default offset
		)
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(recordings)
}

func (s *StreamingAPIService) handlePlayRecording(w http.ResponseWriter, r *http.Request) {
	// Implementation
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}

// Notification Handlers - Simplified implementations
func (s *StreamingAPIService) handleSubscribe(w http.ResponseWriter, r *http.Request) {
	// Implementation
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}

func (s *StreamingAPIService) handleUnsubscribe(w http.ResponseWriter, r *http.Request) {
	// Implementation
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}

func (s *StreamingAPIService) handleGetNotifications(w http.ResponseWriter, r *http.Request) {
	// Implementation
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}

func (s *StreamingAPIService) handleMarkAsRead(w http.ResponseWriter, r *http.Request) {
	// Implementation
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}

// Analytics Handlers
func (s *StreamingAPIService) handleGetStreamAnalytics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	roomName := r.URL.Query().Get("room_name")
	if roomName == "" {
		http.Error(w, "room_name required", http.StatusBadRequest)
		return
	}

	analytics, err := s.analyticsService.GetStreamAnalytics(
		r.Context(),
		livekit.RoomName(roomName),
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(analytics)
}

func (s *StreamingAPIService) handleGetDashboard(w http.ResponseWriter, r *http.Request) {
	// Implementation for dashboard data
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}

func (s *StreamingAPIService) handleExportAnalytics(w http.ResponseWriter, r *http.Request) {
	// Implementation for exporting analytics
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}
