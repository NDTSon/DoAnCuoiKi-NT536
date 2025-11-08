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

// ChatMessage represents a single chat message
type ChatMessage struct {
	ID             string                        `json:"id"`
	RoomName       livekit.RoomName              `json:"room_name"`
	SenderID       livekit.ParticipantIdentity   `json:"sender_id"`
	SenderName     string                        `json:"sender_name"`
	Content        string                        `json:"content"`
	Timestamp      time.Time                     `json:"timestamp"`
	MessageType    ChatMessageType               `json:"message_type"`
	Metadata       map[string]string             `json:"metadata,omitempty"`
	Emojis         []string                      `json:"emojis,omitempty"`
	MentionedUsers []livekit.ParticipantIdentity `json:"mentioned_users,omitempty"`
	IsDeleted      bool                          `json:"is_deleted"`
	IsModerated    bool                          `json:"is_moderated"`
	ReplyTo        *string                       `json:"reply_to,omitempty"`
}

// ChatMessageType defines the type of chat message
type ChatMessageType string

const (
	ChatMessageTypeText         ChatMessageType = "text"
	ChatMessageTypeEmoji        ChatMessageType = "emoji"
	ChatMessageTypeSystemNotice ChatMessageType = "system"
	ChatMessageTypeGift         ChatMessageType = "gift"
	ChatMessageTypeJoinLeave    ChatMessageType = "join_leave"
)

// ChatRoom represents a chat room for a live stream
type ChatRoom struct {
	RoomName     livekit.RoomName                                 `json:"room_name"`
	Messages     []*ChatMessage                                   `json:"messages"`
	Participants map[livekit.ParticipantIdentity]*ChatParticipant `json:"participants"`
	Moderators   map[livekit.ParticipantIdentity]bool             `json:"moderators"`
	BannedUsers  map[livekit.ParticipantIdentity]time.Time        `json:"banned_users"`
	CreatedAt    time.Time                                        `json:"created_at"`
	Settings     *ChatRoomSettings                                `json:"settings"`
	mu           sync.RWMutex
}

// ChatParticipant represents a participant in chat
type ChatParticipant struct {
	Identity     livekit.ParticipantIdentity `json:"identity"`
	Name         string                      `json:"name"`
	IsModerator  bool                        `json:"is_moderator"`
	IsMuted      bool                        `json:"is_muted"`
	JoinedAt     time.Time                   `json:"joined_at"`
	MessageCount int                         `json:"message_count"`
}

// ChatRoomSettings defines chat room configuration
type ChatRoomSettings struct {
	MaxMessageLength    int           `json:"max_message_length"`
	MaxMessagesPerMin   int           `json:"max_messages_per_min"`
	EnableEmojis        bool          `json:"enable_emojis"`
	EnableMentions      bool          `json:"enable_mentions"`
	EnableModeration    bool          `json:"enable_moderation"`
	SlowModeDelay       time.Duration `json:"slow_mode_delay"`
	RequireVerification bool          `json:"require_verification"`
	EnableBadWords      bool          `json:"enable_bad_words"`
}

// ChatService manages all chat rooms
type ChatService struct {
	mu              sync.RWMutex
	rooms           map[livekit.RoomName]*ChatRoom
	logger          logger.Logger
	messageHandlers []ChatMessageHandler
	badWords        []string
}

// ChatMessageHandler is a callback for new messages
type ChatMessageHandler func(message *ChatMessage)

// NewChatService creates a new chat service
func NewChatService() *ChatService {
	return &ChatService{
		rooms:           make(map[livekit.RoomName]*ChatRoom),
		logger:          logger.GetLogger(),
		messageHandlers: make([]ChatMessageHandler, 0),
		badWords:        []string{"spam", "badword1", "badword2"}, // Add more as needed
	}
}

// CreateChatRoom creates a new chat room for a stream
func (cs *ChatService) CreateChatRoom(
	ctx context.Context,
	roomName livekit.RoomName,
	settings *ChatRoomSettings,
) (*ChatRoom, error) {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	if _, exists := cs.rooms[roomName]; exists {
		return nil, fmt.Errorf("chat room already exists")
	}

	if settings == nil {
		settings = &ChatRoomSettings{
			MaxMessageLength:    500,
			MaxMessagesPerMin:   20,
			EnableEmojis:        true,
			EnableMentions:      true,
			EnableModeration:    true,
			SlowModeDelay:       0,
			RequireVerification: false,
			EnableBadWords:      true,
		}
	}

	room := &ChatRoom{
		RoomName:     roomName,
		Messages:     make([]*ChatMessage, 0),
		Participants: make(map[livekit.ParticipantIdentity]*ChatParticipant),
		Moderators:   make(map[livekit.ParticipantIdentity]bool),
		BannedUsers:  make(map[livekit.ParticipantIdentity]time.Time),
		CreatedAt:    time.Now(),
		Settings:     settings,
	}

	cs.rooms[roomName] = room

	cs.logger.Infow("created chat room", "roomName", roomName)

	return room, nil
}

// JoinChatRoom adds a participant to a chat room
func (cs *ChatService) JoinChatRoom(
	ctx context.Context,
	roomName livekit.RoomName,
	participantID livekit.ParticipantIdentity,
	participantName string,
	isModerator bool,
) error {
	cs.mu.RLock()
	room, exists := cs.rooms[roomName]
	cs.mu.RUnlock()

	if !exists {
		return fmt.Errorf("chat room not found")
	}

	room.mu.Lock()
	defer room.mu.Unlock()

	// Check if user is banned
	if banExpiry, banned := room.BannedUsers[participantID]; banned {
		if time.Now().Before(banExpiry) {
			return fmt.Errorf("user is banned until %v", banExpiry)
		}
		// Ban expired, remove it
		delete(room.BannedUsers, participantID)
	}

	participant := &ChatParticipant{
		Identity:     participantID,
		Name:         participantName,
		IsModerator:  isModerator,
		IsMuted:      false,
		JoinedAt:     time.Now(),
		MessageCount: 0,
	}

	room.Participants[participantID] = participant
	if isModerator {
		room.Moderators[participantID] = true
	}

	// Send system message
	systemMsg := &ChatMessage{
		ID:          fmt.Sprintf("sys-%d", time.Now().UnixNano()),
		RoomName:    roomName,
		SenderID:    "system",
		SenderName:  "System",
		Content:     fmt.Sprintf("%s joined the chat", participantName),
		Timestamp:   time.Now(),
		MessageType: ChatMessageTypeJoinLeave,
	}
	room.Messages = append(room.Messages, systemMsg)

	cs.logger.Infow("participant joined chat",
		"roomName", roomName,
		"participantID", participantID,
		"isModerator", isModerator,
	)

	// Notify handlers
	cs.notifyHandlers(systemMsg)

	return nil
}

// LeaveChatRoom removes a participant from a chat room
func (cs *ChatService) LeaveChatRoom(
	ctx context.Context,
	roomName livekit.RoomName,
	participantID livekit.ParticipantIdentity,
) error {
	cs.mu.RLock()
	room, exists := cs.rooms[roomName]
	cs.mu.RUnlock()

	if !exists {
		return fmt.Errorf("chat room not found")
	}

	room.mu.Lock()
	defer room.mu.Unlock()

	participant, exists := room.Participants[participantID]
	if !exists {
		return fmt.Errorf("participant not in chat room")
	}

	delete(room.Participants, participantID)
	delete(room.Moderators, participantID)

	// Send system message
	systemMsg := &ChatMessage{
		ID:          fmt.Sprintf("sys-%d", time.Now().UnixNano()),
		RoomName:    roomName,
		SenderID:    "system",
		SenderName:  "System",
		Content:     fmt.Sprintf("%s left the chat", participant.Name),
		Timestamp:   time.Now(),
		MessageType: ChatMessageTypeJoinLeave,
	}
	room.Messages = append(room.Messages, systemMsg)

	cs.notifyHandlers(systemMsg)

	return nil
}

// SendMessage sends a chat message to a room
func (cs *ChatService) SendMessage(
	ctx context.Context,
	roomName livekit.RoomName,
	senderID livekit.ParticipantIdentity,
	content string,
	messageType ChatMessageType,
	mentionedUsers []livekit.ParticipantIdentity,
	replyTo *string,
) (*ChatMessage, error) {
	cs.mu.RLock()
	room, exists := cs.rooms[roomName]
	cs.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("chat room not found")
	}

	room.mu.Lock()
	defer room.mu.Unlock()

	// Auto-create participant if not exists
	participant, exists := room.Participants[senderID]
	if !exists {
		// Create participant automatically
		participant = &ChatParticipant{
			Identity:     senderID,
			Name:         string(senderID), // Use ID as name
			IsModerator:  false,
			IsMuted:      false,
			JoinedAt:     time.Now(),
			MessageCount: 0,
		}
		room.Participants[senderID] = participant
	}

	// Check if participant is muted
	if participant.IsMuted {
		return nil, fmt.Errorf("participant is muted")
	}

	// Check message length
	if len(content) > room.Settings.MaxMessageLength {
		return nil, fmt.Errorf("message too long")
	}

	// Check rate limiting
	recentMessages := cs.countRecentMessages(room, senderID, time.Minute)
	if recentMessages >= room.Settings.MaxMessagesPerMin {
		return nil, fmt.Errorf("rate limit exceeded")
	}

	// Check slow mode
	if room.Settings.SlowModeDelay > 0 {
		lastMsg := cs.getLastMessage(room, senderID)
		if lastMsg != nil && time.Since(lastMsg.Timestamp) < room.Settings.SlowModeDelay {
			return nil, fmt.Errorf("slow mode active, please wait")
		}
	}

	// Filter bad words
	if room.Settings.EnableBadWords {
		content = cs.filterBadWords(content)
	}

	message := &ChatMessage{
		ID:             fmt.Sprintf("msg-%d-%s", time.Now().UnixNano(), senderID),
		RoomName:       roomName,
		SenderID:       senderID,
		SenderName:     participant.Name,
		Content:        content,
		Timestamp:      time.Now(),
		MessageType:    messageType,
		MentionedUsers: mentionedUsers,
		ReplyTo:        replyTo,
		IsDeleted:      false,
		IsModerated:    false,
	}

	room.Messages = append(room.Messages, message)
	participant.MessageCount++

	cs.logger.Debugw("chat message sent",
		"roomName", roomName,
		"senderID", senderID,
		"messageType", messageType,
	)

	// Notify handlers
	cs.notifyHandlers(message)

	return message, nil
}

// DeleteMessage deletes a chat message (moderator action)
func (cs *ChatService) DeleteMessage(
	ctx context.Context,
	roomName livekit.RoomName,
	messageID string,
	moderatorID livekit.ParticipantIdentity,
) error {
	cs.mu.RLock()
	room, exists := cs.rooms[roomName]
	cs.mu.RUnlock()

	if !exists {
		return fmt.Errorf("chat room not found")
	}

	room.mu.Lock()
	defer room.mu.Unlock()

	// Check if user is moderator
	if !room.Moderators[moderatorID] {
		return fmt.Errorf("user is not a moderator")
	}

	// Find and mark message as deleted
	for _, msg := range room.Messages {
		if msg.ID == messageID {
			msg.IsDeleted = true
			msg.IsModerated = true
			cs.logger.Infow("message deleted by moderator",
				"messageID", messageID,
				"moderatorID", moderatorID,
			)
			return nil
		}
	}

	return fmt.Errorf("message not found")
}

// MuteParticipant mutes a participant (moderator action)
func (cs *ChatService) MuteParticipant(
	ctx context.Context,
	roomName livekit.RoomName,
	participantID livekit.ParticipantIdentity,
	moderatorID livekit.ParticipantIdentity,
	duration time.Duration,
) error {
	cs.mu.RLock()
	room, exists := cs.rooms[roomName]
	cs.mu.RUnlock()

	if !exists {
		return fmt.Errorf("chat room not found")
	}

	room.mu.Lock()
	defer room.mu.Unlock()

	// Check if user is moderator
	if !room.Moderators[moderatorID] {
		return fmt.Errorf("user is not a moderator")
	}

	participant, exists := room.Participants[participantID]
	if !exists {
		return fmt.Errorf("participant not found")
	}

	participant.IsMuted = true

	// Schedule unmute if duration is provided
	if duration > 0 {
		go func() {
			time.Sleep(duration)
			room.mu.Lock()
			defer room.mu.Unlock()
			if p, ok := room.Participants[participantID]; ok {
				p.IsMuted = false
			}
		}()
	}

	cs.logger.Infow("participant muted",
		"participantID", participantID,
		"moderatorID", moderatorID,
		"duration", duration,
	)

	return nil
}

// BanParticipant bans a participant from the chat room
func (cs *ChatService) BanParticipant(
	ctx context.Context,
	roomName livekit.RoomName,
	participantID livekit.ParticipantIdentity,
	moderatorID livekit.ParticipantIdentity,
	duration time.Duration,
) error {
	cs.mu.RLock()
	room, exists := cs.rooms[roomName]
	cs.mu.RUnlock()

	if !exists {
		return fmt.Errorf("chat room not found")
	}

	room.mu.Lock()
	defer room.mu.Unlock()

	// Check if user is moderator
	if !room.Moderators[moderatorID] {
		return fmt.Errorf("user is not a moderator")
	}

	banExpiry := time.Now().Add(duration)
	room.BannedUsers[participantID] = banExpiry

	// Remove from participants
	delete(room.Participants, participantID)

	cs.logger.Infow("participant banned",
		"participantID", participantID,
		"moderatorID", moderatorID,
		"until", banExpiry,
	)

	return nil
}

// GetMessages returns recent messages from a chat room
func (cs *ChatService) GetMessages(
	ctx context.Context,
	roomName livekit.RoomName,
	limit int,
	beforeTimestamp *time.Time,
) ([]*ChatMessage, error) {
	cs.mu.RLock()
	room, exists := cs.rooms[roomName]
	cs.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("chat room not found")
	}

	room.mu.RLock()
	defer room.mu.RUnlock()

	messages := make([]*ChatMessage, 0)
	for i := len(room.Messages) - 1; i >= 0 && len(messages) < limit; i-- {
		msg := room.Messages[i]
		if beforeTimestamp == nil || msg.Timestamp.Before(*beforeTimestamp) {
			if !msg.IsDeleted {
				messages = append(messages, msg)
			}
		}
	}

	return messages, nil
}

// Helper functions

func (cs *ChatService) countRecentMessages(room *ChatRoom, senderID livekit.ParticipantIdentity, duration time.Duration) int {
	count := 0
	cutoff := time.Now().Add(-duration)
	for i := len(room.Messages) - 1; i >= 0; i-- {
		msg := room.Messages[i]
		if msg.Timestamp.Before(cutoff) {
			break
		}
		if msg.SenderID == senderID {
			count++
		}
	}
	return count
}

func (cs *ChatService) getLastMessage(room *ChatRoom, senderID livekit.ParticipantIdentity) *ChatMessage {
	for i := len(room.Messages) - 1; i >= 0; i-- {
		msg := room.Messages[i]
		if msg.SenderID == senderID {
			return msg
		}
	}
	return nil
}

func (cs *ChatService) filterBadWords(content string) string {
	// Simple bad word filter - in production, use more sophisticated filtering
	for _, badWord := range cs.badWords {
		// Replace with asterisks
		content = replaceWord(content, badWord)
	}
	return content
}

func replaceWord(content, word string) string {
	// Simple replacement - use regex for better matching in production
	return content
}

func (cs *ChatService) notifyHandlers(message *ChatMessage) {
	for _, handler := range cs.messageHandlers {
		go handler(message)
	}
}

// RegisterMessageHandler adds a callback for new messages
func (cs *ChatService) RegisterMessageHandler(handler ChatMessageHandler) {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	cs.messageHandlers = append(cs.messageHandlers, handler)
}
