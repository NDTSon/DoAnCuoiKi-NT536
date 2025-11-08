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

package streaming

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/livekit/protocol/livekit"
	"github.com/livekit/protocol/logger"
)

// VODRecording represents a recorded live stream
type VODRecording struct {
	ID           string                      `json:"id"`
	RoomName     livekit.RoomName            `json:"room_name"`
	StreamerID   livekit.ParticipantIdentity `json:"streamer_id"`
	StreamerName string                      `json:"streamer_name"`
	Title        string                      `json:"title"`
	Description  string                      `json:"description"`
	ThumbnailURL string                      `json:"thumbnail_url"`
	VideoURL     string                      `json:"video_url"`
	FileSize     int64                       `json:"file_size"` // bytes
	Duration     time.Duration               `json:"duration"`
	Resolution   string                      `json:"resolution"` // e.g., "1920x1080"
	Bitrate      int                         `json:"bitrate"`    // kbps
	Status       VODStatus                   `json:"status"`
	ViewCount    int64                       `json:"view_count"`
	LikeCount    int64                       `json:"like_count"`
	ShareCount   int64                       `json:"share_count"`
	RecordedAt   time.Time                   `json:"recorded_at"`
	PublishedAt  *time.Time                  `json:"published_at,omitempty"`
	ExpiresAt    *time.Time                  `json:"expires_at,omitempty"`
	IsPublic     bool                        `json:"is_public"`
	Tags         []string                    `json:"tags,omitempty"`
	Category     string                      `json:"category,omitempty"`
	Language     string                      `json:"language,omitempty"`
	Metadata     map[string]string           `json:"metadata,omitempty"`
	// Analytics
	AverageViewDuration time.Duration `json:"average_view_duration"`
	PeakViewers         int           `json:"peak_viewers"`
	ChatMessageCount    int           `json:"chat_message_count"`
	ReactionCount       int           `json:"reaction_count"`
}

// VODStatus represents the status of a VOD recording
type VODStatus string

const (
	VODStatusRecording  VODStatus = "recording"
	VODStatusProcessing VODStatus = "processing"
	VODStatusReady      VODStatus = "ready"
	VODStatusFailed     VODStatus = "failed"
	VODStatusArchived   VODStatus = "archived"
	VODStatusDeleted    VODStatus = "deleted"
)

// VODPlaybackSession represents a user watching a VOD
type VODPlaybackSession struct {
	ID              string                      `json:"id"`
	RecordingID     string                      `json:"recording_id"`
	UserID          livekit.ParticipantIdentity `json:"user_id"`
	StartedAt       time.Time                   `json:"started_at"`
	LastHeartbeat   time.Time                   `json:"last_heartbeat"`
	CurrentPosition time.Duration               `json:"current_position"`
	WatchDuration   time.Duration               `json:"watch_duration"`
	Completed       bool                        `json:"completed"`
	Quality         string                      `json:"quality"`
}

// VODService manages video on demand recordings
type VODService struct {
	mu                 sync.RWMutex
	recordings         map[string]*VODRecording                 // recordingID -> Recording
	streamerRecordings map[livekit.ParticipantIdentity][]string // streamerID -> []recordingIDs
	playbackSessions   map[string]*VODPlaybackSession           // sessionID -> Session
	logger             logger.Logger
	config             *VODConfig
}

// VODConfig defines VOD service configuration
type VODConfig struct {
	StoragePath          string        `json:"storage_path"`
	MaxRecordingSize     int64         `json:"max_recording_size"` // bytes
	DefaultRetentionDays int           `json:"default_retention_days"`
	AutoPublish          bool          `json:"auto_publish"`
	GenerateThumbnails   bool          `json:"generate_thumbnails"`
	EnableTranscoding    bool          `json:"enable_transcoding"`
	TranscodingQualities []string      `json:"transcoding_qualities"` // e.g., ["1080p", "720p", "480p"]
	SessionTimeout       time.Duration `json:"session_timeout"`
	EnableAnalytics      bool          `json:"enable_analytics"`
}

// NewVODService creates a new VOD service
func NewVODService(config *VODConfig) *VODService {
	if config == nil {
		config = &VODConfig{
			StoragePath:          "/var/livekit/recordings",
			MaxRecordingSize:     10 * 1024 * 1024 * 1024, // 10 GB
			DefaultRetentionDays: 30,
			AutoPublish:          false,
			GenerateThumbnails:   true,
			EnableTranscoding:    true,
			TranscodingQualities: []string{"1080p", "720p", "480p", "360p"},
			SessionTimeout:       5 * time.Minute,
			EnableAnalytics:      true,
		}
	}

	return &VODService{
		recordings:         make(map[string]*VODRecording),
		streamerRecordings: make(map[livekit.ParticipantIdentity][]string),
		playbackSessions:   make(map[string]*VODPlaybackSession),
		logger:             logger.GetLogger(),
		config:             config,
	}
}

// StartRecording initiates a new VOD recording
func (vs *VODService) StartRecording(
	ctx context.Context,
	roomName livekit.RoomName,
	streamerID livekit.ParticipantIdentity,
	streamerName string,
	title string,
) (*VODRecording, error) {
	vs.mu.Lock()
	defer vs.mu.Unlock()

	recordingID := fmt.Sprintf("rec-%d-%s", time.Now().UnixNano(), streamerID)

	recording := &VODRecording{
		ID:           recordingID,
		RoomName:     roomName,
		StreamerID:   streamerID,
		StreamerName: streamerName,
		Title:        title,
		Status:       VODStatusRecording,
		RecordedAt:   time.Now(),
		IsPublic:     vs.config.AutoPublish,
		Metadata:     make(map[string]string),
		Tags:         make([]string, 0),
	}

	// Set expiration if retention is configured
	if vs.config.DefaultRetentionDays > 0 {
		expiresAt := time.Now().AddDate(0, 0, vs.config.DefaultRetentionDays)
		recording.ExpiresAt = &expiresAt
	}

	vs.recordings[recordingID] = recording
	vs.streamerRecordings[streamerID] = append(vs.streamerRecordings[streamerID], recordingID)

	vs.logger.Infow("started VOD recording",
		"recordingID", recordingID,
		"roomName", roomName,
		"streamerID", streamerID,
	)

	return recording, nil
}

// StopRecording stops an active recording
func (vs *VODService) StopRecording(
	ctx context.Context,
	recordingID string,
	duration time.Duration,
	fileSize int64,
) error {
	vs.mu.Lock()
	defer vs.mu.Unlock()

	recording, exists := vs.recordings[recordingID]
	if !exists {
		return fmt.Errorf("recording not found")
	}

	if recording.Status != VODStatusRecording {
		return fmt.Errorf("recording is not in recording status")
	}

	recording.Duration = duration
	recording.FileSize = fileSize
	recording.Status = VODStatusProcessing

	vs.logger.Infow("stopped VOD recording",
		"recordingID", recordingID,
		"duration", duration,
		"fileSize", fileSize,
	)

	// Start post-processing
	go vs.processRecording(context.Background(), recordingID)

	return nil
}

// processRecording handles post-processing of a recording
func (vs *VODService) processRecording(ctx context.Context, recordingID string) {
	vs.mu.Lock()
	recording, exists := vs.recordings[recordingID]
	vs.mu.Unlock()

	if !exists {
		return
	}

	vs.logger.Infow("processing VOD recording", "recordingID", recordingID)

	// Simulate processing time
	time.Sleep(5 * time.Second)

	vs.mu.Lock()
	defer vs.mu.Unlock()

	// Generate thumbnail
	if vs.config.GenerateThumbnails {
		recording.ThumbnailURL = fmt.Sprintf("/thumbnails/%s.jpg", recordingID)
	}

	// Set video URL
	recording.VideoURL = fmt.Sprintf("/videos/%s.mp4", recordingID)

	// Mark as ready
	recording.Status = VODStatusReady
	if vs.config.AutoPublish {
		now := time.Now()
		recording.PublishedAt = &now
	}

	vs.logger.Infow("VOD recording ready",
		"recordingID", recordingID,
		"videoURL", recording.VideoURL,
	)
}

// PublishRecording makes a recording publicly available
func (vs *VODService) PublishRecording(
	ctx context.Context,
	recordingID string,
) error {
	vs.mu.Lock()
	defer vs.mu.Unlock()

	recording, exists := vs.recordings[recordingID]
	if !exists {
		return fmt.Errorf("recording not found")
	}

	if recording.Status != VODStatusReady {
		return fmt.Errorf("recording is not ready")
	}

	recording.IsPublic = true
	now := time.Now()
	recording.PublishedAt = &now

	vs.logger.Infow("published VOD recording", "recordingID", recordingID)

	return nil
}

// GetRecording retrieves a recording by ID
func (vs *VODService) GetRecording(
	ctx context.Context,
	recordingID string,
) (*VODRecording, error) {
	vs.mu.RLock()
	defer vs.mu.RUnlock()

	recording, exists := vs.recordings[recordingID]
	if !exists {
		return nil, fmt.Errorf("recording not found")
	}

	return recording, nil
}

// ListRecordingsByStreamer returns all recordings for a streamer
func (vs *VODService) ListRecordingsByStreamer(
	ctx context.Context,
	streamerID livekit.ParticipantIdentity,
	limit int,
	offset int,
) ([]*VODRecording, error) {
	vs.mu.RLock()
	defer vs.mu.RUnlock()

	recordingIDs, exists := vs.streamerRecordings[streamerID]
	if !exists {
		return []*VODRecording{}, nil
	}

	// Apply pagination
	start := offset
	if start >= len(recordingIDs) {
		return []*VODRecording{}, nil
	}

	end := start + limit
	if end > len(recordingIDs) {
		end = len(recordingIDs)
	}

	recordings := make([]*VODRecording, 0, end-start)
	for i := start; i < end; i++ {
		if recording, ok := vs.recordings[recordingIDs[i]]; ok {
			recordings = append(recordings, recording)
		}
	}

	return recordings, nil
}

// StartPlaybackSession starts a new playback session
func (vs *VODService) StartPlaybackSession(
	ctx context.Context,
	recordingID string,
	userID livekit.ParticipantIdentity,
	quality string,
) (*VODPlaybackSession, error) {
	vs.mu.Lock()
	defer vs.mu.Unlock()

	recording, exists := vs.recordings[recordingID]
	if !exists {
		return nil, fmt.Errorf("recording not found")
	}

	if recording.Status != VODStatusReady {
		return nil, fmt.Errorf("recording is not ready for playback")
	}

	if !recording.IsPublic {
		return nil, fmt.Errorf("recording is not public")
	}

	sessionID := fmt.Sprintf("session-%d-%s", time.Now().UnixNano(), userID)

	session := &VODPlaybackSession{
		ID:              sessionID,
		RecordingID:     recordingID,
		UserID:          userID,
		StartedAt:       time.Now(),
		LastHeartbeat:   time.Now(),
		CurrentPosition: 0,
		WatchDuration:   0,
		Completed:       false,
		Quality:         quality,
	}

	vs.playbackSessions[sessionID] = session

	// Increment view count
	recording.ViewCount++

	vs.logger.Debugw("started playback session",
		"sessionID", sessionID,
		"recordingID", recordingID,
		"userID", userID,
	)

	return session, nil
}

// UpdatePlaybackSession updates playback progress
func (vs *VODService) UpdatePlaybackSession(
	ctx context.Context,
	sessionID string,
	position time.Duration,
) error {
	vs.mu.Lock()
	defer vs.mu.Unlock()

	session, exists := vs.playbackSessions[sessionID]
	if !exists {
		return fmt.Errorf("session not found")
	}

	now := time.Now()
	session.LastHeartbeat = now
	session.CurrentPosition = position
	session.WatchDuration = now.Sub(session.StartedAt)

	// Check if completed (watched 95% or more)
	recording, exists := vs.recordings[session.RecordingID]
	if exists && recording.Duration > 0 {
		if float64(position) >= float64(recording.Duration)*0.95 {
			session.Completed = true
		}
	}

	return nil
}

// EndPlaybackSession ends a playback session
func (vs *VODService) EndPlaybackSession(
	ctx context.Context,
	sessionID string,
) error {
	vs.mu.Lock()
	defer vs.mu.Unlock()

	session, exists := vs.playbackSessions[sessionID]
	if !exists {
		return fmt.Errorf("session not found")
	}

	// Update analytics
	if recording, ok := vs.recordings[session.RecordingID]; ok {
		// Update average view duration
		totalSessions := recording.ViewCount
		if totalSessions > 0 {
			totalDuration := recording.AverageViewDuration * time.Duration(totalSessions-1)
			recording.AverageViewDuration = (totalDuration + session.WatchDuration) / time.Duration(totalSessions)
		}
	}

	delete(vs.playbackSessions, sessionID)

	vs.logger.Debugw("ended playback session",
		"sessionID", sessionID,
		"watchDuration", session.WatchDuration,
		"completed", session.Completed,
	)

	return nil
}

// DeleteRecording removes a recording
func (vs *VODService) DeleteRecording(
	ctx context.Context,
	recordingID string,
) error {
	vs.mu.Lock()
	defer vs.mu.Unlock()

	recording, exists := vs.recordings[recordingID]
	if !exists {
		return fmt.Errorf("recording not found")
	}

	recording.Status = VODStatusDeleted

	// Remove from streamer recordings
	streamerRecordings := vs.streamerRecordings[recording.StreamerID]
	for i, id := range streamerRecordings {
		if id == recordingID {
			vs.streamerRecordings[recording.StreamerID] = append(
				streamerRecordings[:i],
				streamerRecordings[i+1:]...,
			)
			break
		}
	}

	delete(vs.recordings, recordingID)

	vs.logger.Infow("deleted VOD recording", "recordingID", recordingID)

	return nil
}

// CleanupExpiredRecordings removes expired recordings
func (vs *VODService) CleanupExpiredRecordings(ctx context.Context) int {
	vs.mu.Lock()
	defer vs.mu.Unlock()

	now := time.Now()
	count := 0

	for recordingID, recording := range vs.recordings {
		if recording.ExpiresAt != nil && now.After(*recording.ExpiresAt) {
			recording.Status = VODStatusDeleted
			delete(vs.recordings, recordingID)
			count++
		}
	}

	if count > 0 {
		vs.logger.Infow("cleaned up expired recordings", "count", count)
	}

	return count
}

// CleanupStaleSessions removes inactive playback sessions
func (vs *VODService) CleanupStaleSessions(ctx context.Context) int {
	vs.mu.Lock()
	defer vs.mu.Unlock()

	now := time.Now()
	count := 0

	for sessionID, session := range vs.playbackSessions {
		if now.Sub(session.LastHeartbeat) > vs.config.SessionTimeout {
			delete(vs.playbackSessions, sessionID)
			count++
		}
	}

	if count > 0 {
		vs.logger.Debugw("cleaned up stale sessions", "count", count)
	}

	return count
}

// UpdateRecordingMetadata updates recording metadata
func (vs *VODService) UpdateRecordingMetadata(
	ctx context.Context,
	recordingID string,
	title *string,
	description *string,
	tags []string,
	category *string,
) error {
	vs.mu.Lock()
	defer vs.mu.Unlock()

	recording, exists := vs.recordings[recordingID]
	if !exists {
		return fmt.Errorf("recording not found")
	}

	if title != nil {
		recording.Title = *title
	}
	if description != nil {
		recording.Description = *description
	}
	if tags != nil {
		recording.Tags = tags
	}
	if category != nil {
		recording.Category = *category
	}

	return nil
}
