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

// ReactionType defines the type of reaction
type ReactionType string

const (
	ReactionTypeLike  ReactionType = "like"  // 👍
	ReactionTypeHeart ReactionType = "heart" // ❤️
	ReactionTypeWow   ReactionType = "wow"   // 😮
	ReactionTypeLaugh ReactionType = "laugh" // 😂
	ReactionTypeSad   ReactionType = "sad"   // 😢
	ReactionTypeFire  ReactionType = "fire"  // 🔥
	ReactionTypeClap  ReactionType = "clap"  // 👏
	ReactionTypeParty ReactionType = "party" // 🎉
)

// Reaction represents a single reaction from a user
type Reaction struct {
	ID        string                      `json:"id"`
	RoomName  livekit.RoomName            `json:"room_name"`
	UserID    livekit.ParticipantIdentity `json:"user_id"`
	UserName  string                      `json:"user_name"`
	Type      ReactionType                `json:"type"`
	Timestamp time.Time                   `json:"timestamp"`
	Metadata  map[string]string           `json:"metadata,omitempty"`
	Position  *ReactionPosition           `json:"position,omitempty"` // For animated reactions on screen
}

// ReactionPosition defines where a reaction appears on screen
type ReactionPosition struct {
	X float64 `json:"x"` // 0-1, percentage of screen width
	Y float64 `json:"y"` // 0-1, percentage of screen height
}

// ReactionStats tracks reaction statistics for a stream
type ReactionStats struct {
	RoomName        livekit.RoomName     `json:"room_name"`
	TotalReactions  int                  `json:"total_reactions"`
	ReactionCounts  map[ReactionType]int `json:"reaction_counts"`
	TopReactors     []*TopReactor        `json:"top_reactors"`
	RecentReactions []*Reaction          `json:"recent_reactions"`
	LastUpdated     time.Time            `json:"last_updated"`
}

// TopReactor represents a user who has reacted the most
type TopReactor struct {
	UserID        livekit.ParticipantIdentity `json:"user_id"`
	UserName      string                      `json:"user_name"`
	ReactionCount int                         `json:"reaction_count"`
}

// ReactionRoom manages reactions for a live stream room
type ReactionRoom struct {
	RoomName      livekit.RoomName                            `json:"room_name"`
	Reactions     []*Reaction                                 `json:"reactions"`
	Stats         *ReactionStats                              `json:"stats"`
	UserReactions map[livekit.ParticipantIdentity][]*Reaction `json:"user_reactions"`
	RateLimits    map[livekit.ParticipantIdentity]*RateLimit  `json:"rate_limits"`
	CreatedAt     time.Time                                   `json:"created_at"`
	mu            sync.RWMutex
}

// RateLimit tracks reaction rate limiting per user
type RateLimit struct {
	LastReaction time.Time
	Count        int
	WindowStart  time.Time
}

// ReactionService manages reactions across all stream rooms
type ReactionService struct {
	mu               sync.RWMutex
	rooms            map[livekit.RoomName]*ReactionRoom
	logger           logger.Logger
	reactionHandlers []ReactionHandler
	config           *ReactionConfig
}

// ReactionConfig defines reaction service configuration
type ReactionConfig struct {
	MaxReactionsPerMinute int           `json:"max_reactions_per_minute"`
	MaxReactionsPerSecond int           `json:"max_reactions_per_second"`
	ReactionTTL           time.Duration `json:"reaction_ttl"`
	EnableRateLimit       bool          `json:"enable_rate_limit"`
	EnableAnimation       bool          `json:"enable_animation"`
	MaxRecentReactions    int           `json:"max_recent_reactions"`
	EnableLeaderboard     bool          `json:"enable_leaderboard"`
}

// ReactionHandler is a callback for new reactions
type ReactionHandler func(reaction *Reaction)

// NewReactionService creates a new reaction service
func NewReactionService(config *ReactionConfig) *ReactionService {
	if config == nil {
		config = &ReactionConfig{
			MaxReactionsPerMinute: 60,
			MaxReactionsPerSecond: 3,
			ReactionTTL:           5 * time.Minute,
			EnableRateLimit:       true,
			EnableAnimation:       true,
			MaxRecentReactions:    100,
			EnableLeaderboard:     true,
		}
	}

	return &ReactionService{
		rooms:            make(map[livekit.RoomName]*ReactionRoom),
		logger:           logger.GetLogger(),
		reactionHandlers: make([]ReactionHandler, 0),
		config:           config,
	}
}

// CreateReactionRoom creates a new reaction room for a stream
func (rs *ReactionService) CreateReactionRoom(
	ctx context.Context,
	roomName livekit.RoomName,
) (*ReactionRoom, error) {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	if _, exists := rs.rooms[roomName]; exists {
		return nil, fmt.Errorf("reaction room already exists")
	}

	room := &ReactionRoom{
		RoomName:      roomName,
		Reactions:     make([]*Reaction, 0),
		UserReactions: make(map[livekit.ParticipantIdentity][]*Reaction),
		RateLimits:    make(map[livekit.ParticipantIdentity]*RateLimit),
		CreatedAt:     time.Now(),
		Stats: &ReactionStats{
			RoomName:        roomName,
			TotalReactions:  0,
			ReactionCounts:  make(map[ReactionType]int),
			TopReactors:     make([]*TopReactor, 0),
			RecentReactions: make([]*Reaction, 0),
			LastUpdated:     time.Now(),
		},
	}

	rs.rooms[roomName] = room

	rs.logger.Infow("created reaction room", "roomName", roomName)

	return room, nil
}

// SendReaction sends a reaction to a stream
func (rs *ReactionService) SendReaction(
	ctx context.Context,
	roomName livekit.RoomName,
	userID livekit.ParticipantIdentity,
	userName string,
	reactionType ReactionType,
	position *ReactionPosition,
) (*Reaction, error) {
	rs.mu.RLock()
	room, exists := rs.rooms[roomName]
	rs.mu.RUnlock()

	// Auto-create room if not exists
	if !exists {
		rs.mu.Lock()
		// Double-check after acquiring write lock
		room, exists = rs.rooms[roomName]
		if !exists {
			room = &ReactionRoom{
				RoomName:      roomName,
				Reactions:     make([]*Reaction, 0),
				UserReactions: make(map[livekit.ParticipantIdentity][]*Reaction),
				RateLimits:    make(map[livekit.ParticipantIdentity]*RateLimit),
				CreatedAt:     time.Now(),
				Stats: &ReactionStats{
					RoomName:        roomName,
					TotalReactions:  0,
					ReactionCounts:  make(map[ReactionType]int),
					TopReactors:     make([]*TopReactor, 0),
					RecentReactions: make([]*Reaction, 0),
					LastUpdated:     time.Now(),
				},
			}
			rs.rooms[roomName] = room
		}
		rs.mu.Unlock()
	}

	// Check rate limit
	if rs.config.EnableRateLimit {
		if err := rs.checkRateLimit(room, userID); err != nil {
			return nil, err
		}
	}

	room.mu.Lock()
	defer room.mu.Unlock()

	// Create reaction
	reaction := &Reaction{
		ID:        fmt.Sprintf("reaction-%d-%s", time.Now().UnixNano(), userID),
		RoomName:  roomName,
		UserID:    userID,
		UserName:  userName,
		Type:      reactionType,
		Timestamp: time.Now(),
		Position:  position,
		Metadata:  make(map[string]string),
	}

	// Add to room
	room.Reactions = append(room.Reactions, reaction)
	room.UserReactions[userID] = append(room.UserReactions[userID], reaction)

	// Update stats
	room.Stats.TotalReactions++
	room.Stats.ReactionCounts[reactionType]++
	room.Stats.RecentReactions = append([]*Reaction{reaction}, room.Stats.RecentReactions...)
	if len(room.Stats.RecentReactions) > rs.config.MaxRecentReactions {
		room.Stats.RecentReactions = room.Stats.RecentReactions[:rs.config.MaxRecentReactions]
	}
	room.Stats.LastUpdated = time.Now()

	// Update top reactors
	rs.updateTopReactors(room)

	// Update rate limit
	rs.updateRateLimit(room, userID)

	rs.logger.Debugw("reaction sent",
		"roomName", roomName,
		"userID", userID,
		"type", reactionType,
	)

	// Notify handlers
	rs.notifyHandlers(reaction)

	return reaction, nil
}

// GetReactionStats returns reaction statistics for a room
func (rs *ReactionService) GetReactionStats(
	ctx context.Context,
	roomName livekit.RoomName,
) (*ReactionStats, error) {
	rs.mu.RLock()
	room, exists := rs.rooms[roomName]
	rs.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("reaction room not found")
	}

	room.mu.RLock()
	defer room.mu.RUnlock()

	return room.Stats, nil
}

// GetRecentReactions returns recent reactions from a room
func (rs *ReactionService) GetRecentReactions(
	ctx context.Context,
	roomName livekit.RoomName,
	limit int,
) ([]*Reaction, error) {
	rs.mu.RLock()
	room, exists := rs.rooms[roomName]
	rs.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("reaction room not found")
	}

	room.mu.RLock()
	defer room.mu.RUnlock()

	reactions := make([]*Reaction, 0)
	count := 0
	for i := len(room.Reactions) - 1; i >= 0 && count < limit; i-- {
		reactions = append(reactions, room.Reactions[i])
		count++
	}

	return reactions, nil
}

// GetUserReactions returns all reactions from a specific user
func (rs *ReactionService) GetUserReactions(
	ctx context.Context,
	roomName livekit.RoomName,
	userID livekit.ParticipantIdentity,
) ([]*Reaction, error) {
	rs.mu.RLock()
	room, exists := rs.rooms[roomName]
	rs.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("reaction room not found")
	}

	room.mu.RLock()
	defer room.mu.RUnlock()

	userReactions, exists := room.UserReactions[userID]
	if !exists {
		return []*Reaction{}, nil
	}

	return userReactions, nil
}

// GetTopReactors returns the top reactors in a room
func (rs *ReactionService) GetTopReactors(
	ctx context.Context,
	roomName livekit.RoomName,
	limit int,
) ([]*TopReactor, error) {
	rs.mu.RLock()
	room, exists := rs.rooms[roomName]
	rs.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("reaction room not found")
	}

	room.mu.RLock()
	defer room.mu.RUnlock()

	topReactors := room.Stats.TopReactors
	if len(topReactors) > limit {
		topReactors = topReactors[:limit]
	}

	return topReactors, nil
}

// CleanupOldReactions removes reactions older than the TTL
func (rs *ReactionService) CleanupOldReactions(ctx context.Context) int {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	totalCleaned := 0
	cutoff := time.Now().Add(-rs.config.ReactionTTL)

	for _, room := range rs.rooms {
		room.mu.Lock()

		// Filter reactions
		validReactions := make([]*Reaction, 0)
		for _, reaction := range room.Reactions {
			if reaction.Timestamp.After(cutoff) {
				validReactions = append(validReactions, reaction)
			} else {
				totalCleaned++
			}
		}
		room.Reactions = validReactions

		// Update user reactions
		for userID, reactions := range room.UserReactions {
			validUserReactions := make([]*Reaction, 0)
			for _, reaction := range reactions {
				if reaction.Timestamp.After(cutoff) {
					validUserReactions = append(validUserReactions, reaction)
				}
			}
			room.UserReactions[userID] = validUserReactions
		}

		room.mu.Unlock()
	}

	if totalCleaned > 0 {
		rs.logger.Infow("cleaned up old reactions", "count", totalCleaned)
	}

	return totalCleaned
}

// Helper functions

func (rs *ReactionService) checkRateLimit(room *ReactionRoom, userID livekit.ParticipantIdentity) error {
	room.mu.RLock()
	rateLimit, exists := room.RateLimits[userID]
	room.mu.RUnlock()

	if !exists {
		return nil
	}

	now := time.Now()

	// Check per-second limit
	if now.Sub(rateLimit.LastReaction) < time.Second/time.Duration(rs.config.MaxReactionsPerSecond) {
		return fmt.Errorf("rate limit exceeded: too many reactions per second")
	}

	// Check per-minute limit
	if now.Sub(rateLimit.WindowStart) < time.Minute {
		if rateLimit.Count >= rs.config.MaxReactionsPerMinute {
			return fmt.Errorf("rate limit exceeded: too many reactions per minute")
		}
	}

	return nil
}

func (rs *ReactionService) updateRateLimit(room *ReactionRoom, userID livekit.ParticipantIdentity) {
	now := time.Now()

	rateLimit, exists := room.RateLimits[userID]
	if !exists {
		rateLimit = &RateLimit{
			WindowStart: now,
			Count:       0,
		}
		room.RateLimits[userID] = rateLimit
	}

	// Reset window if needed
	if now.Sub(rateLimit.WindowStart) >= time.Minute {
		rateLimit.WindowStart = now
		rateLimit.Count = 0
	}

	rateLimit.LastReaction = now
	rateLimit.Count++
}

func (rs *ReactionService) updateTopReactors(room *ReactionRoom) {
	if !rs.config.EnableLeaderboard {
		return
	}

	// Count reactions per user
	userCounts := make(map[livekit.ParticipantIdentity]int)
	userNames := make(map[livekit.ParticipantIdentity]string)

	for userID, reactions := range room.UserReactions {
		userCounts[userID] = len(reactions)
		if len(reactions) > 0 {
			userNames[userID] = reactions[0].UserName
		}
	}

	// Create sorted list
	topReactors := make([]*TopReactor, 0, len(userCounts))
	for userID, count := range userCounts {
		topReactors = append(topReactors, &TopReactor{
			UserID:        userID,
			UserName:      userNames[userID],
			ReactionCount: count,
		})
	}

	// Sort by reaction count (descending)
	for i := 0; i < len(topReactors); i++ {
		for j := i + 1; j < len(topReactors); j++ {
			if topReactors[j].ReactionCount > topReactors[i].ReactionCount {
				topReactors[i], topReactors[j] = topReactors[j], topReactors[i]
			}
		}
	}

	// Keep top 10
	if len(topReactors) > 10 {
		topReactors = topReactors[:10]
	}

	room.Stats.TopReactors = topReactors
}

func (rs *ReactionService) notifyHandlers(reaction *Reaction) {
	for _, handler := range rs.reactionHandlers {
		go handler(reaction)
	}
}

// RegisterReactionHandler adds a callback for new reactions
func (rs *ReactionService) RegisterReactionHandler(handler ReactionHandler) {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	rs.reactionHandlers = append(rs.reactionHandlers, handler)
}

// DeleteReactionRoom removes a reaction room
func (rs *ReactionService) DeleteReactionRoom(ctx context.Context, roomName livekit.RoomName) error {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	delete(rs.rooms, roomName)
	rs.logger.Infow("deleted reaction room", "roomName", roomName)
	return nil
}
