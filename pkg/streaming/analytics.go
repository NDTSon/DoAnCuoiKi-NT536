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

// StreamAnalytics contains detailed analytics for a live stream
type StreamAnalytics struct {
	RoomName   livekit.RoomName            `json:"room_name"`
	StreamerID livekit.ParticipantIdentity `json:"streamer_id"`
	StartTime  time.Time                   `json:"start_time"`
	EndTime    *time.Time                  `json:"end_time,omitempty"`
	Duration   time.Duration               `json:"duration"`

	// Viewer metrics
	TotalViewers     int           `json:"total_viewers"`
	PeakViewers      int           `json:"peak_viewers"`
	CurrentViewers   int           `json:"current_viewers"`
	AverageViewers   float64       `json:"average_viewers"`
	UniqueViewers    int           `json:"unique_viewers"`
	ViewerRetention  float64       `json:"viewer_retention"` // percentage
	AverageWatchTime time.Duration `json:"average_watch_time"`

	// Chat metrics
	TotalMessages     int     `json:"total_messages"`
	UniqueMessagers   int     `json:"unique_messagers"`
	MessagesPerMinute float64 `json:"messages_per_minute"`

	// Reaction metrics
	TotalReactions     int                  `json:"total_reactions"`
	ReactionsPerMinute float64              `json:"reactions_per_minute"`
	ReactionBreakdown  map[ReactionType]int `json:"reaction_breakdown"`

	// Technical metrics
	AverageBitrate  int           `json:"average_bitrate"` // kbps
	PeakBitrate     int           `json:"peak_bitrate"`    // kbps
	AverageLatency  time.Duration `json:"average_latency"`
	BufferingEvents int           `json:"buffering_events"`
	QualityChanges  int           `json:"quality_changes"`
	DroppedFrames   int           `json:"dropped_frames"`

	// Engagement metrics
	ShareCount  int `json:"share_count"`
	LikeCount   int `json:"like_count"`
	FollowCount int `json:"follow_count"`

	// Geographic distribution
	ViewersByCountry map[string]int `json:"viewers_by_country"`
	ViewersByRegion  map[string]int `json:"viewers_by_region"`

	// Device/Platform metrics
	ViewersByPlatform map[string]int `json:"viewers_by_platform"`
	ViewersByDevice   map[string]int `json:"viewers_by_device"`

	// Time-series data points
	ViewerTimeline   []TimeSeriesDataPoint `json:"viewer_timeline"`
	ChatTimeline     []TimeSeriesDataPoint `json:"chat_timeline"`
	ReactionTimeline []TimeSeriesDataPoint `json:"reaction_timeline"`
	BitrateTimeline  []TimeSeriesDataPoint `json:"bitrate_timeline"`

	LastUpdated time.Time `json:"last_updated"`
}

// TimeSeriesDataPoint represents a metric at a point in time
type TimeSeriesDataPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value"`
}

// ViewerSession represents a single viewer's session
type ViewerSession struct {
	ViewerID          livekit.ParticipantIdentity `json:"viewer_id"`
	RoomName          livekit.RoomName            `json:"room_name"`
	JoinedAt          time.Time                   `json:"joined_at"`
	LeftAt            *time.Time                  `json:"left_at,omitempty"`
	WatchDuration     time.Duration               `json:"watch_duration"`
	MessagesSent      int                         `json:"messages_sent"`
	ReactionsSent     int                         `json:"reactions_sent"`
	Platform          string                      `json:"platform"`
	Device            string                      `json:"device"`
	Country           string                      `json:"country"`
	Region            string                      `json:"region"`
	AvgBitrate        int                         `json:"avg_bitrate"`
	BufferingCount    int                         `json:"buffering_count"`
	QualityLevel      string                      `json:"quality_level"`
	ConnectionQuality string                      `json:"connection_quality"` // excellent, good, fair, poor
}

// AnalyticsService manages stream analytics
type AnalyticsService struct {
	mu              sync.RWMutex
	streamAnalytics map[livekit.RoomName]*StreamAnalytics
	viewerSessions  map[livekit.RoomName]map[livekit.ParticipantIdentity]*ViewerSession
	logger          logger.Logger
	config          *AnalyticsConfig
}

// AnalyticsConfig defines analytics service configuration
type AnalyticsConfig struct {
	EnableRealTime        bool          `json:"enable_real_time"`
	UpdateInterval        time.Duration `json:"update_interval"`
	TimelineResolution    time.Duration `json:"timeline_resolution"`
	MaxTimelinePoints     int           `json:"max_timeline_points"`
	RetentionDays         int           `json:"retention_days"`
	EnableGeoIP           bool          `json:"enable_geoip"`
	EnableDeviceDetection bool          `json:"enable_device_detection"`
}

// NewAnalyticsService creates a new analytics service
func NewAnalyticsService(config *AnalyticsConfig) *AnalyticsService {
	if config == nil {
		config = &AnalyticsConfig{
			EnableRealTime:        true,
			UpdateInterval:        10 * time.Second,
			TimelineResolution:    1 * time.Minute,
			MaxTimelinePoints:     1000,
			RetentionDays:         90,
			EnableGeoIP:           true,
			EnableDeviceDetection: true,
		}
	}

	return &AnalyticsService{
		streamAnalytics: make(map[livekit.RoomName]*StreamAnalytics),
		viewerSessions:  make(map[livekit.RoomName]map[livekit.ParticipantIdentity]*ViewerSession),
		logger:          logger.GetLogger(),
		config:          config,
	}
}

// StartStreamAnalytics initializes analytics for a new stream
func (as *AnalyticsService) StartStreamAnalytics(
	ctx context.Context,
	roomName livekit.RoomName,
	streamerID livekit.ParticipantIdentity,
) (*StreamAnalytics, error) {
	as.mu.Lock()
	defer as.mu.Unlock()

	if _, exists := as.streamAnalytics[roomName]; exists {
		return nil, fmt.Errorf("analytics already started for this stream")
	}

	analytics := &StreamAnalytics{
		RoomName:          roomName,
		StreamerID:        streamerID,
		StartTime:         time.Now(),
		ReactionBreakdown: make(map[ReactionType]int),
		ViewersByCountry:  make(map[string]int),
		ViewersByRegion:   make(map[string]int),
		ViewersByPlatform: make(map[string]int),
		ViewersByDevice:   make(map[string]int),
		ViewerTimeline:    make([]TimeSeriesDataPoint, 0),
		ChatTimeline:      make([]TimeSeriesDataPoint, 0),
		ReactionTimeline:  make([]TimeSeriesDataPoint, 0),
		BitrateTimeline:   make([]TimeSeriesDataPoint, 0),
		LastUpdated:       time.Now(),
	}

	as.streamAnalytics[roomName] = analytics
	as.viewerSessions[roomName] = make(map[livekit.ParticipantIdentity]*ViewerSession)

	as.logger.Infow("started stream analytics",
		"roomName", roomName,
		"streamerID", streamerID,
	)

	// Start real-time updates if enabled
	if as.config.EnableRealTime {
		go as.updateAnalyticsLoop(ctx, roomName)
	}

	return analytics, nil
}

// StopStreamAnalytics finalizes analytics for a stream
func (as *AnalyticsService) StopStreamAnalytics(
	ctx context.Context,
	roomName livekit.RoomName,
) error {
	as.mu.Lock()
	defer as.mu.Unlock()

	analytics, exists := as.streamAnalytics[roomName]
	if !exists {
		return fmt.Errorf("analytics not found for this stream")
	}

	now := time.Now()
	analytics.EndTime = &now
	analytics.Duration = now.Sub(analytics.StartTime)

	// Close all viewer sessions
	if sessions, ok := as.viewerSessions[roomName]; ok {
		for _, session := range sessions {
			if session.LeftAt == nil {
				session.LeftAt = &now
				session.WatchDuration = now.Sub(session.JoinedAt)
			}
		}
	}

	// Final update
	as.calculateMetrics(analytics, roomName)

	as.logger.Infow("stopped stream analytics",
		"roomName", roomName,
		"duration", analytics.Duration,
		"totalViewers", analytics.TotalViewers,
		"peakViewers", analytics.PeakViewers,
	)

	return nil
}

// RecordViewerJoin records a viewer joining the stream
func (as *AnalyticsService) RecordViewerJoin(
	ctx context.Context,
	roomName livekit.RoomName,
	viewerID livekit.ParticipantIdentity,
	platform string,
	device string,
	country string,
	region string,
) error {
	as.mu.Lock()
	defer as.mu.Unlock()

	analytics, exists := as.streamAnalytics[roomName]
	if !exists {
		return fmt.Errorf("analytics not found")
	}

	sessions, ok := as.viewerSessions[roomName]
	if !ok {
		sessions = make(map[livekit.ParticipantIdentity]*ViewerSession)
		as.viewerSessions[roomName] = sessions
	}

	// Check if this is a unique viewer
	isUnique := true
	for _, session := range sessions {
		if session.ViewerID == viewerID && session.LeftAt != nil {
			isUnique = false
			break
		}
	}

	session := &ViewerSession{
		ViewerID:     viewerID,
		RoomName:     roomName,
		JoinedAt:     time.Now(),
		Platform:     platform,
		Device:       device,
		Country:      country,
		Region:       region,
		QualityLevel: "auto",
	}

	sessions[viewerID] = session

	// Update analytics
	analytics.CurrentViewers++
	analytics.TotalViewers++
	if isUnique {
		analytics.UniqueViewers++
	}
	if analytics.CurrentViewers > analytics.PeakViewers {
		analytics.PeakViewers = analytics.CurrentViewers
	}

	// Update geographic distribution
	if country != "" {
		analytics.ViewersByCountry[country]++
	}
	if region != "" {
		analytics.ViewersByRegion[region]++
	}

	// Update platform/device distribution
	if platform != "" {
		analytics.ViewersByPlatform[platform]++
	}
	if device != "" {
		analytics.ViewersByDevice[device]++
	}

	as.logger.Debugw("viewer joined",
		"roomName", roomName,
		"viewerID", viewerID,
		"currentViewers", analytics.CurrentViewers,
	)

	return nil
}

// RecordViewerLeave records a viewer leaving the stream
func (as *AnalyticsService) RecordViewerLeave(
	ctx context.Context,
	roomName livekit.RoomName,
	viewerID livekit.ParticipantIdentity,
) error {
	as.mu.Lock()
	defer as.mu.Unlock()

	analytics, exists := as.streamAnalytics[roomName]
	if !exists {
		return fmt.Errorf("analytics not found")
	}

	sessions, ok := as.viewerSessions[roomName]
	if !ok {
		return fmt.Errorf("no sessions found")
	}

	session, exists := sessions[viewerID]
	if !exists {
		return fmt.Errorf("session not found")
	}

	now := time.Now()
	session.LeftAt = &now
	session.WatchDuration = now.Sub(session.JoinedAt)

	analytics.CurrentViewers--
	if analytics.CurrentViewers < 0 {
		analytics.CurrentViewers = 0
	}

	as.logger.Debugw("viewer left",
		"roomName", roomName,
		"viewerID", viewerID,
		"watchDuration", session.WatchDuration,
		"currentViewers", analytics.CurrentViewers,
	)

	return nil
}

// RecordChatMessage records a chat message for analytics
func (as *AnalyticsService) RecordChatMessage(
	ctx context.Context,
	roomName livekit.RoomName,
	senderID livekit.ParticipantIdentity,
) error {
	as.mu.Lock()
	defer as.mu.Unlock()

	analytics, exists := as.streamAnalytics[roomName]
	if !exists {
		return fmt.Errorf("analytics not found")
	}

	analytics.TotalMessages++

	// Update session
	if sessions, ok := as.viewerSessions[roomName]; ok {
		if session, exists := sessions[senderID]; exists {
			session.MessagesSent++
		}
	}

	return nil
}

// RecordReaction records a reaction for analytics
func (as *AnalyticsService) RecordReaction(
	ctx context.Context,
	roomName livekit.RoomName,
	senderID livekit.ParticipantIdentity,
	reactionType ReactionType,
) error {
	as.mu.Lock()
	defer as.mu.Unlock()

	analytics, exists := as.streamAnalytics[roomName]
	if !exists {
		return fmt.Errorf("analytics not found")
	}

	analytics.TotalReactions++
	analytics.ReactionBreakdown[reactionType]++

	// Update session
	if sessions, ok := as.viewerSessions[roomName]; ok {
		if session, exists := sessions[senderID]; exists {
			session.ReactionsSent++
		}
	}

	return nil
}

// RecordBitrateUpdate records bitrate information
func (as *AnalyticsService) RecordBitrateUpdate(
	ctx context.Context,
	roomName livekit.RoomName,
	bitrate int,
) error {
	as.mu.Lock()
	defer as.mu.Unlock()

	analytics, exists := as.streamAnalytics[roomName]
	if !exists {
		return fmt.Errorf("analytics not found")
	}

	if bitrate > analytics.PeakBitrate {
		analytics.PeakBitrate = bitrate
	}

	return nil
}

// GetStreamAnalytics retrieves analytics for a stream
func (as *AnalyticsService) GetStreamAnalytics(
	ctx context.Context,
	roomName livekit.RoomName,
) (*StreamAnalytics, error) {
	as.mu.RLock()
	defer as.mu.RUnlock()

	analytics, exists := as.streamAnalytics[roomName]
	if !exists {
		return nil, fmt.Errorf("analytics not found")
	}

	// Calculate latest metrics
	as.calculateMetrics(analytics, roomName)

	return analytics, nil
}

// GetViewerSessions retrieves all viewer sessions for a stream
func (as *AnalyticsService) GetViewerSessions(
	ctx context.Context,
	roomName livekit.RoomName,
) ([]*ViewerSession, error) {
	as.mu.RLock()
	defer as.mu.RUnlock()

	sessions, exists := as.viewerSessions[roomName]
	if !exists {
		return []*ViewerSession{}, nil
	}

	result := make([]*ViewerSession, 0, len(sessions))
	for _, session := range sessions {
		result = append(result, session)
	}

	return result, nil
}

// Helper functions

func (as *AnalyticsService) calculateMetrics(analytics *StreamAnalytics, roomName livekit.RoomName) {
	sessions, ok := as.viewerSessions[roomName]
	if !ok {
		return
	}

	// Calculate average watch time
	totalWatchTime := time.Duration(0)
	completedSessions := 0
	uniqueMessagers := make(map[livekit.ParticipantIdentity]bool)

	for _, session := range sessions {
		if session.LeftAt != nil {
			totalWatchTime += session.WatchDuration
			completedSessions++
		}
		if session.MessagesSent > 0 {
			uniqueMessagers[session.ViewerID] = true
		}
	}

	if completedSessions > 0 {
		analytics.AverageWatchTime = totalWatchTime / time.Duration(completedSessions)
	}

	analytics.UniqueMessagers = len(uniqueMessagers)

	// Calculate rates
	if analytics.EndTime != nil {
		duration := analytics.EndTime.Sub(analytics.StartTime)
		minutes := duration.Minutes()
		if minutes > 0 {
			analytics.MessagesPerMinute = float64(analytics.TotalMessages) / minutes
			analytics.ReactionsPerMinute = float64(analytics.TotalReactions) / minutes
		}
	}

	// Calculate viewer retention
	if analytics.TotalViewers > 0 {
		analytics.ViewerRetention = float64(completedSessions) / float64(analytics.TotalViewers) * 100
	}

	analytics.LastUpdated = time.Now()
}

func (as *AnalyticsService) updateAnalyticsLoop(ctx context.Context, roomName livekit.RoomName) {
	ticker := time.NewTicker(as.config.UpdateInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			as.mu.Lock()
			analytics, exists := as.streamAnalytics[roomName]
			if !exists {
				as.mu.Unlock()
				return
			}

			// Add timeline data points
			now := time.Now()
			analytics.ViewerTimeline = append(analytics.ViewerTimeline, TimeSeriesDataPoint{
				Timestamp: now,
				Value:     float64(analytics.CurrentViewers),
			})

			// Limit timeline points
			if len(analytics.ViewerTimeline) > as.config.MaxTimelinePoints {
				analytics.ViewerTimeline = analytics.ViewerTimeline[1:]
			}

			as.calculateMetrics(analytics, roomName)
			as.mu.Unlock()
		}
	}
}

// CleanupOldAnalytics removes analytics data older than retention period
func (as *AnalyticsService) CleanupOldAnalytics(ctx context.Context) int {
	as.mu.Lock()
	defer as.mu.Unlock()

	count := 0
	cutoff := time.Now().AddDate(0, 0, -as.config.RetentionDays)

	for roomName, analytics := range as.streamAnalytics {
		if analytics.EndTime != nil && analytics.EndTime.Before(cutoff) {
			delete(as.streamAnalytics, roomName)
			delete(as.viewerSessions, roomName)
			count++
		}
	}

	if count > 0 {
		as.logger.Infow("cleaned up old analytics", "count", count)
	}

	return count
}
