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
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"github.com/livekit/protocol/livekit"
	"github.com/livekit/protocol/logger"
)

// StreamKey represents a unique key for a streamer
type StreamKey struct {
	Key         string                      `json:"key"`
	StreamerID  livekit.ParticipantIdentity `json:"streamer_id"`
	RoomName    livekit.RoomName            `json:"room_name"`
	IsActive    bool                        `json:"is_active"`
	CreatedAt   time.Time                   `json:"created_at"`
	ExpiresAt   *time.Time                  `json:"expires_at,omitempty"`
	Metadata    map[string]string           `json:"metadata,omitempty"`
	UsageCount  int                         `json:"usage_count"`
	LastUsedAt  *time.Time                  `json:"last_used_at,omitempty"`
	Permissions *StreamPermissions          `json:"permissions,omitempty"`
}

// StreamPermissions defines what a stream key can do
type StreamPermissions struct {
	CanPublishVideo  bool `json:"can_publish_video"`
	CanPublishAudio  bool `json:"can_publish_audio"`
	CanScreenShare   bool `json:"can_screen_share"`
	CanRecord        bool `json:"can_record"`
	MaxViewers       int  `json:"max_viewers"`
	MaxDurationMins  int  `json:"max_duration_mins"`
	EnableChat       bool `json:"enable_chat"`
	EnableReactions  bool `json:"enable_reactions"`
	EnableModeration bool `json:"enable_moderation"`
}

// StreamKeyManager manages stream keys for all streamers
type StreamKeyManager struct {
	mu   sync.RWMutex
	keys map[string]*StreamKey // key -> StreamKey
	// streamerID -> []keys for quick lookup
	streamerKeys map[livekit.ParticipantIdentity][]string
	logger       logger.Logger
}

// NewStreamKeyManager creates a new stream key manager
func NewStreamKeyManager() *StreamKeyManager {
	return &StreamKeyManager{
		keys:         make(map[string]*StreamKey),
		streamerKeys: make(map[livekit.ParticipantIdentity][]string),
		logger:       logger.GetLogger(),
	}
}

// GenerateStreamKey creates a new unique stream key for a streamer
func (m *StreamKeyManager) GenerateStreamKey(
	ctx context.Context,
	streamerID livekit.ParticipantIdentity,
	roomName livekit.RoomName,
	permissions *StreamPermissions,
	expiresIn *time.Duration,
) (*StreamKey, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Generate a cryptographically secure random key
	keyBytes := make([]byte, 32)
	if _, err := rand.Read(keyBytes); err != nil {
		return nil, fmt.Errorf("failed to generate random key: %w", err)
	}
	key := hex.EncodeToString(keyBytes)

	// Set default permissions if not provided
	if permissions == nil {
		permissions = &StreamPermissions{
			CanPublishVideo:  true,
			CanPublishAudio:  true,
			CanScreenShare:   true,
			CanRecord:        false,
			MaxViewers:       10000,
			MaxDurationMins:  180, // 3 hours
			EnableChat:       true,
			EnableReactions:  true,
			EnableModeration: true,
		}
	}

	streamKey := &StreamKey{
		Key:         key,
		StreamerID:  streamerID,
		RoomName:    roomName,
		IsActive:    true,
		CreatedAt:   time.Now(),
		Metadata:    make(map[string]string),
		UsageCount:  0,
		Permissions: permissions,
	}

	if expiresIn != nil {
		expiresAt := time.Now().Add(*expiresIn)
		streamKey.ExpiresAt = &expiresAt
	}

	// Store the key
	m.keys[key] = streamKey
	m.streamerKeys[streamerID] = append(m.streamerKeys[streamerID], key)

	m.logger.Infow("generated new stream key",
		"streamerID", streamerID,
		"roomName", roomName,
		"key", key[:8]+"...",
	)

	return streamKey, nil
}

// ValidateStreamKey checks if a stream key is valid and can be used
func (m *StreamKeyManager) ValidateStreamKey(ctx context.Context, key string) (*StreamKey, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	streamKey, exists := m.keys[key]
	if !exists {
		return nil, fmt.Errorf("stream key not found")
	}

	if !streamKey.IsActive {
		return nil, fmt.Errorf("stream key is inactive")
	}

	// Check expiration
	if streamKey.ExpiresAt != nil && time.Now().After(*streamKey.ExpiresAt) {
		return nil, fmt.Errorf("stream key has expired")
	}

	return streamKey, nil
}

// MarkKeyAsUsed updates the usage statistics of a stream key
func (m *StreamKeyManager) MarkKeyAsUsed(ctx context.Context, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	streamKey, exists := m.keys[key]
	if !exists {
		return fmt.Errorf("stream key not found")
	}

	now := time.Now()
	streamKey.UsageCount++
	streamKey.LastUsedAt = &now

	m.logger.Debugw("stream key used",
		"key", key[:8]+"...",
		"usageCount", streamKey.UsageCount,
	)

	return nil
}

// RevokeStreamKey deactivates a stream key
func (m *StreamKeyManager) RevokeStreamKey(ctx context.Context, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	streamKey, exists := m.keys[key]
	if !exists {
		return fmt.Errorf("stream key not found")
	}

	streamKey.IsActive = false

	m.logger.Infow("stream key revoked",
		"key", key[:8]+"...",
		"streamerID", streamKey.StreamerID,
	)

	return nil
}

// GetStreamKeysByStreamer returns all stream keys for a specific streamer
func (m *StreamKeyManager) GetStreamKeysByStreamer(
	ctx context.Context,
	streamerID livekit.ParticipantIdentity,
) ([]*StreamKey, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	keyIDs, exists := m.streamerKeys[streamerID]
	if !exists {
		return []*StreamKey{}, nil
	}

	keys := make([]*StreamKey, 0, len(keyIDs))
	for _, keyID := range keyIDs {
		if streamKey, ok := m.keys[keyID]; ok {
			keys = append(keys, streamKey)
		}
	}

	return keys, nil
}

// DeleteStreamKey permanently removes a stream key
func (m *StreamKeyManager) DeleteStreamKey(ctx context.Context, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	streamKey, exists := m.keys[key]
	if !exists {
		return fmt.Errorf("stream key not found")
	}

	// Remove from main map
	delete(m.keys, key)

	// Remove from streamer keys
	streamerKeys := m.streamerKeys[streamKey.StreamerID]
	for i, k := range streamerKeys {
		if k == key {
			m.streamerKeys[streamKey.StreamerID] = append(streamerKeys[:i], streamerKeys[i+1:]...)
			break
		}
	}

	m.logger.Infow("stream key deleted",
		"key", key[:8]+"...",
		"streamerID", streamKey.StreamerID,
	)

	return nil
}

// UpdateStreamKeyMetadata updates metadata for a stream key
func (m *StreamKeyManager) UpdateStreamKeyMetadata(
	ctx context.Context,
	key string,
	metadata map[string]string,
) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	streamKey, exists := m.keys[key]
	if !exists {
		return fmt.Errorf("stream key not found")
	}

	if streamKey.Metadata == nil {
		streamKey.Metadata = make(map[string]string)
	}

	for k, v := range metadata {
		streamKey.Metadata[k] = v
	}

	return nil
}

// CleanupExpiredKeys removes expired stream keys
func (m *StreamKeyManager) CleanupExpiredKeys(ctx context.Context) int {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	count := 0

	for key, streamKey := range m.keys {
		if streamKey.ExpiresAt != nil && now.After(*streamKey.ExpiresAt) {
			delete(m.keys, key)

			// Remove from streamer keys
			streamerKeys := m.streamerKeys[streamKey.StreamerID]
			for i, k := range streamerKeys {
				if k == key {
					m.streamerKeys[streamKey.StreamerID] = append(streamerKeys[:i], streamerKeys[i+1:]...)
					break
				}
			}

			count++
		}
	}

	if count > 0 {
		m.logger.Infow("cleaned up expired stream keys", "count", count)
	}

	return count
}

// GetActiveStreamCount returns the number of currently active streams
func (m *StreamKeyManager) GetActiveStreamCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	count := 0
	for _, key := range m.keys {
		if key.IsActive && key.LastUsedAt != nil {
			// Consider active if used in the last 5 minutes
			if time.Since(*key.LastUsedAt) < 5*time.Minute {
				count++
			}
		}
	}

	return count
}
