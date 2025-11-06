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

// NotificationType defines the type of notification
type NotificationType string

const (
	NotificationTypeStreamStarted  NotificationType = "stream_started"
	NotificationTypeStreamEnded    NotificationType = "stream_ended"
	NotificationTypeNewFollower    NotificationType = "new_follower"
	NotificationTypeMention        NotificationType = "mention"
	NotificationTypeReply          NotificationType = "reply"
	NotificationTypeModerator      NotificationType = "moderator"
	NotificationTypeGift           NotificationType = "gift"
	NotificationTypeSystem         NotificationType = "system"
)

// Notification represents a single notification
type Notification struct {
	ID        string                      `json:"id"`
	UserID    livekit.ParticipantIdentity `json:"user_id"`
	Type      NotificationType            `json:"type"`
	Title     string                      `json:"title"`
	Body      string                      `json:"body"`
	ImageURL  string                      `json:"image_url,omitempty"`
	ActionURL string                      `json:"action_url,omitempty"`
	Data      map[string]string           `json:"data,omitempty"`
	Priority  NotificationPriority        `json:"priority"`
	CreatedAt time.Time                   `json:"created_at"`
	ReadAt    *time.Time                  `json:"read_at,omitempty"`
	IsRead    bool                        `json:"is_read"`
	ExpiresAt *time.Time                  `json:"expires_at,omitempty"`
}

// NotificationPriority defines notification priority
type NotificationPriority string

const (
	PriorityLow    NotificationPriority = "low"
	PriorityMedium NotificationPriority = "medium"
	PriorityHigh   NotificationPriority = "high"
	PriorityUrgent NotificationPriority = "urgent"
)

// NotificationSubscription represents a user's notification preferences
type NotificationSubscription struct {
	UserID            livekit.ParticipantIdentity `json:"user_id"`
	StreamerID        livekit.ParticipantIdentity `json:"streamer_id"`
	StreamerName      string                      `json:"streamer_name"`
	EnableStreamStart bool                        `json:"enable_stream_start"`
	EnableStreamEnd   bool                        `json:"enable_stream_end"`
	EnableChat        bool                        `json:"enable_chat"`
	EnableMentions    bool                        `json:"enable_mentions"`
	CreatedAt         time.Time                   `json:"created_at"`
}

// NotificationChannel defines how notifications are delivered
type NotificationChannel string

const (
	ChannelWebSocket NotificationChannel = "websocket"
	ChannelEmail     NotificationChannel = "email"
	ChannelPush      NotificationChannel = "push"
	ChannelSMS       NotificationChannel = "sms"
)

// NotificationService manages notifications
type NotificationService struct {
	mu                     sync.RWMutex
	notifications          map[livekit.ParticipantIdentity][]*Notification // userID -> notifications
	subscriptions          map[livekit.ParticipantIdentity][]*NotificationSubscription // userID -> subscriptions
	streamerFollowers      map[livekit.ParticipantIdentity][]livekit.ParticipantIdentity // streamerID -> followerIDs
	onlineUsers            map[livekit.ParticipantIdentity]bool
	notificationHandlers   map[NotificationChannel][]NotificationHandler
	logger                 logger.Logger
	config                 *NotificationConfig
}

// NotificationConfig defines notification service configuration
type NotificationConfig struct {
	MaxNotificationsPerUser int           `json:"max_notifications_per_user"`
	NotificationTTL         time.Duration `json:"notification_ttl"`
	EnableWebSocket         bool          `json:"enable_websocket"`
	EnableEmail             bool          `json:"enable_email"`
	EnablePush              bool          `json:"enable_push"`
	EmailFromAddress        string        `json:"email_from_address"`
	SMTPServer              string        `json:"smtp_server"`
	SMTPPort                int           `json:"smtp_port"`
}

// NotificationHandler is a callback for sending notifications
type NotificationHandler func(notification *Notification)

// NewNotificationService creates a new notification service
func NewNotificationService(config *NotificationConfig) *NotificationService {
	if config == nil {
		config = &NotificationConfig{
			MaxNotificationsPerUser: 1000,
			NotificationTTL:         30 * 24 * time.Hour, // 30 days
			EnableWebSocket:         true,
			EnableEmail:             false,
			EnablePush:              false,
		}
	}

	return &NotificationService{
		notifications:        make(map[livekit.ParticipantIdentity][]*Notification),
		subscriptions:        make(map[livekit.ParticipantIdentity][]*NotificationSubscription),
		streamerFollowers:    make(map[livekit.ParticipantIdentity][]livekit.ParticipantIdentity),
		onlineUsers:          make(map[livekit.ParticipantIdentity]bool),
		notificationHandlers: make(map[NotificationChannel][]NotificationHandler),
		logger:               logger.GetLogger(),
		config:               config,
	}
}

// Subscribe allows a user to follow a streamer
func (ns *NotificationService) Subscribe(
	ctx context.Context,
	userID livekit.ParticipantIdentity,
	streamerID livekit.ParticipantIdentity,
	streamerName string,
	preferences *NotificationSubscription,
) error {
	ns.mu.Lock()
	defer ns.mu.Unlock()

	// Check if already subscribed
	userSubs, exists := ns.subscriptions[userID]
	if exists {
		for _, sub := range userSubs {
			if sub.StreamerID == streamerID {
				return fmt.Errorf("already subscribed to this streamer")
			}
		}
	}

	// Create subscription with default preferences if not provided
	if preferences == nil {
		preferences = &NotificationSubscription{
			EnableStreamStart: true,
			EnableStreamEnd:   false,
			EnableChat:        false,
			EnableMentions:    true,
		}
	}

	subscription := &NotificationSubscription{
		UserID:            userID,
		StreamerID:        streamerID,
		StreamerName:      streamerName,
		EnableStreamStart: preferences.EnableStreamStart,
		EnableStreamEnd:   preferences.EnableStreamEnd,
		EnableChat:        preferences.EnableChat,
		EnableMentions:    preferences.EnableMentions,
		CreatedAt:         time.Now(),
	}

	ns.subscriptions[userID] = append(ns.subscriptions[userID], subscription)
	ns.streamerFollowers[streamerID] = append(ns.streamerFollowers[streamerID], userID)

	ns.logger.Infow("user subscribed to streamer",
		"userID", userID,
		"streamerID", streamerID,
	)

	return nil
}

// Unsubscribe removes a subscription
func (ns *NotificationService) Unsubscribe(
	ctx context.Context,
	userID livekit.ParticipantIdentity,
	streamerID livekit.ParticipantIdentity,
) error {
	ns.mu.Lock()
	defer ns.mu.Unlock()

	// Remove from user subscriptions
	userSubs := ns.subscriptions[userID]
	for i, sub := range userSubs {
		if sub.StreamerID == streamerID {
			ns.subscriptions[userID] = append(userSubs[:i], userSubs[i+1:]...)
			break
		}
	}

	// Remove from streamer followers
	followers := ns.streamerFollowers[streamerID]
	for i, followerID := range followers {
		if followerID == userID {
			ns.streamerFollowers[streamerID] = append(followers[:i], followers[i+1:]...)
			break
		}
	}

	ns.logger.Infow("user unsubscribed from streamer",
		"userID", userID,
		"streamerID", streamerID,
	)

	return nil
}

// NotifyStreamStarted notifies followers when a stream starts
func (ns *NotificationService) NotifyStreamStarted(
	ctx context.Context,
	streamerID livekit.ParticipantIdentity,
	streamerName string,
	roomName livekit.RoomName,
	streamTitle string,
) error {
	ns.mu.RLock()
	followers, exists := ns.streamerFollowers[streamerID]
	ns.mu.RUnlock()

	if !exists || len(followers) == 0 {
		return nil
	}

	ns.logger.Infow("notifying stream started",
		"streamerID", streamerID,
		"followerCount", len(followers),
	)

	// Send notifications to all followers
	for _, followerID := range followers {
		// Check subscription preferences
		ns.mu.RLock()
		userSubs := ns.subscriptions[followerID]
		shouldNotify := false
		for _, sub := range userSubs {
			if sub.StreamerID == streamerID && sub.EnableStreamStart {
				shouldNotify = true
				break
			}
		}
		ns.mu.RUnlock()

		if !shouldNotify {
			continue
		}

		notification := &Notification{
			ID:        fmt.Sprintf("notif-%d-%s", time.Now().UnixNano(), followerID),
			UserID:    followerID,
			Type:      NotificationTypeStreamStarted,
			Title:     fmt.Sprintf("%s is live!", streamerName),
			Body:      streamTitle,
			ActionURL: fmt.Sprintf("/watch/%s", roomName),
			Priority:  PriorityHigh,
			CreatedAt: time.Now(),
			IsRead:    false,
			Data: map[string]string{
				"streamer_id":   string(streamerID),
				"streamer_name": streamerName,
				"room_name":     string(roomName),
			},
		}

		ns.addNotification(followerID, notification)
		ns.sendNotification(notification, ChannelWebSocket)
	}

	return nil
}

// NotifyStreamEnded notifies followers when a stream ends
func (ns *NotificationService) NotifyStreamEnded(
	ctx context.Context,
	streamerID livekit.ParticipantIdentity,
	streamerName string,
	duration time.Duration,
	viewCount int,
) error {
	ns.mu.RLock()
	followers := ns.streamerFollowers[streamerID]
	ns.mu.RUnlock()

	for _, followerID := range followers {
		// Check preferences
		ns.mu.RLock()
		userSubs := ns.subscriptions[followerID]
		shouldNotify := false
		for _, sub := range userSubs {
			if sub.StreamerID == streamerID && sub.EnableStreamEnd {
				shouldNotify = true
				break
			}
		}
		ns.mu.RUnlock()

		if !shouldNotify {
			continue
		}

		notification := &Notification{
			ID:        fmt.Sprintf("notif-%d-%s", time.Now().UnixNano(), followerID),
			UserID:    followerID,
			Type:      NotificationTypeStreamEnded,
			Title:     fmt.Sprintf("%s's stream ended", streamerName),
			Body:      fmt.Sprintf("Stream lasted %v with %d viewers", duration, viewCount),
			Priority:  PriorityLow,
			CreatedAt: time.Now(),
			IsRead:    false,
		}

		ns.addNotification(followerID, notification)
	}

	return nil
}

// SendNotification sends a custom notification to a user
func (ns *NotificationService) SendNotification(
	ctx context.Context,
	userID livekit.ParticipantIdentity,
	notificationType NotificationType,
	title string,
	body string,
	priority NotificationPriority,
	actionURL string,
	data map[string]string,
) (*Notification, error) {
	notification := &Notification{
		ID:        fmt.Sprintf("notif-%d-%s", time.Now().UnixNano(), userID),
		UserID:    userID,
		Type:      notificationType,
		Title:     title,
		Body:      body,
		ActionURL: actionURL,
		Data:      data,
		Priority:  priority,
		CreatedAt: time.Now(),
		IsRead:    false,
	}

	ns.addNotification(userID, notification)
	ns.sendNotification(notification, ChannelWebSocket)

	return notification, nil
}

// GetNotifications retrieves notifications for a user
func (ns *NotificationService) GetNotifications(
	ctx context.Context,
	userID livekit.ParticipantIdentity,
	unreadOnly bool,
	limit int,
) ([]*Notification, error) {
	ns.mu.RLock()
	defer ns.mu.RUnlock()

	userNotifications, exists := ns.notifications[userID]
	if !exists {
		return []*Notification{}, nil
	}

	notifications := make([]*Notification, 0)
	count := 0

	// Return in reverse order (newest first)
	for i := len(userNotifications) - 1; i >= 0 && count < limit; i-- {
		notif := userNotifications[i]
		if !unreadOnly || !notif.IsRead {
			notifications = append(notifications, notif)
			count++
		}
	}

	return notifications, nil
}

// MarkAsRead marks a notification as read
func (ns *NotificationService) MarkAsRead(
	ctx context.Context,
	userID livekit.ParticipantIdentity,
	notificationID string,
) error {
	ns.mu.Lock()
	defer ns.mu.Unlock()

	userNotifications, exists := ns.notifications[userID]
	if !exists {
		return fmt.Errorf("no notifications found for user")
	}

	for _, notif := range userNotifications {
		if notif.ID == notificationID {
			notif.IsRead = true
			now := time.Now()
			notif.ReadAt = &now
			return nil
		}
	}

	return fmt.Errorf("notification not found")
}

// MarkAllAsRead marks all notifications as read for a user
func (ns *NotificationService) MarkAllAsRead(
	ctx context.Context,
	userID livekit.ParticipantIdentity,
) error {
	ns.mu.Lock()
	defer ns.mu.Unlock()

	userNotifications, exists := ns.notifications[userID]
	if !exists {
		return nil
	}

	now := time.Now()
	for _, notif := range userNotifications {
		if !notif.IsRead {
			notif.IsRead = true
			notif.ReadAt = &now
		}
	}

	return nil
}

// GetUnreadCount returns the count of unread notifications
func (ns *NotificationService) GetUnreadCount(
	ctx context.Context,
	userID livekit.ParticipantIdentity,
) (int, error) {
	ns.mu.RLock()
	defer ns.mu.RUnlock()

	userNotifications, exists := ns.notifications[userID]
	if !exists {
		return 0, nil
	}

	count := 0
	for _, notif := range userNotifications {
		if !notif.IsRead {
			count++
		}
	}

	return count, nil
}

// SetUserOnlineStatus updates user's online status
func (ns *NotificationService) SetUserOnlineStatus(
	ctx context.Context,
	userID livekit.ParticipantIdentity,
	isOnline bool,
) {
	ns.mu.Lock()
	defer ns.mu.Unlock()

	if isOnline {
		ns.onlineUsers[userID] = true
	} else {
		delete(ns.onlineUsers, userID)
	}
}

// Helper functions

func (ns *NotificationService) addNotification(
	userID livekit.ParticipantIdentity,
	notification *Notification,
) {
	ns.mu.Lock()
	defer ns.mu.Unlock()

	userNotifications := ns.notifications[userID]
	userNotifications = append(userNotifications, notification)

	// Limit notifications per user
	if len(userNotifications) > ns.config.MaxNotificationsPerUser {
		userNotifications = userNotifications[len(userNotifications)-ns.config.MaxNotificationsPerUser:]
	}

	ns.notifications[userID] = userNotifications
}

func (ns *NotificationService) sendNotification(
	notification *Notification,
	channel NotificationChannel,
) {
	handlers, exists := ns.notificationHandlers[channel]
	if !exists {
		return
	}

	for _, handler := range handlers {
		go handler(notification)
	}
}

// RegisterNotificationHandler adds a callback for sending notifications
func (ns *NotificationService) RegisterNotificationHandler(
	channel NotificationChannel,
	handler NotificationHandler,
) {
	ns.mu.Lock()
	defer ns.mu.Unlock()

	ns.notificationHandlers[channel] = append(ns.notificationHandlers[channel], handler)
}

// CleanupExpiredNotifications removes old notifications
func (ns *NotificationService) CleanupExpiredNotifications(ctx context.Context) int {
	ns.mu.Lock()
	defer ns.mu.Unlock()

	count := 0
	cutoff := time.Now().Add(-ns.config.NotificationTTL)

	for userID, notifications := range ns.notifications {
		validNotifications := make([]*Notification, 0)
		for _, notif := range notifications {
			if notif.CreatedAt.After(cutoff) {
				validNotifications = append(validNotifications, notif)
			} else {
				count++
			}
		}
		ns.notifications[userID] = validNotifications
	}

	if count > 0 {
		ns.logger.Infow("cleaned up expired notifications", "count", count)
	}

	return count
}

// GetSubscriptions returns all subscriptions for a user
func (ns *NotificationService) GetSubscriptions(
	ctx context.Context,
	userID livekit.ParticipantIdentity,
) ([]*NotificationSubscription, error) {
	ns.mu.RLock()
	defer ns.mu.RUnlock()

	subscriptions, exists := ns.subscriptions[userID]
	if !exists {
		return []*NotificationSubscription{}, nil
	}

	return subscriptions, nil
}

// GetFollowerCount returns the number of followers for a streamer
func (ns *NotificationService) GetFollowerCount(
	ctx context.Context,
	streamerID livekit.ParticipantIdentity,
) (int, error) {
	ns.mu.RLock()
	defer ns.mu.RUnlock()

	followers, exists := ns.streamerFollowers[streamerID]
	if !exists {
		return 0, nil
	}

	return len(followers), nil
}
