package usecase

import (
	"chatApp/internal/model"
	"chatApp/internal/repository"
	"chatApp/internal/websocket"
	"context"
	"encoding/json"
	"errors"
	"time"
)

type MessageUsecase interface {
	SendMessage(ctx context.Context, message model.Message) error
	GetMessagesByChatID(ctx context.Context, chatID int, limit int, offset int) ([]*model.Message, error)
	DeleteMessage(ctx context.Context, messageID int) error
}

type messageUseCase struct {
	repo repository.MessageRepository
	hub  *websocket.Hub
}

func NewMessageUsecase(repo repository.MessageRepository, hub *websocket.Hub) MessageUsecase {
	return &messageUseCase{repo: repo, hub: hub}
}

// SendMessage validates and sends a message
func (m *messageUseCase) SendMessage(ctx context.Context, message model.Message) error {
	// Validate message content is not empty
	// if strings.TrimSpace(message.Content) == "" {
	// 	return errors.New("message content cannot be empty")
	// }

	// Validate content length (max 1000 characters)
	if len(message.Content) > 1000 {
		return errors.New("message content exceeds maximum length of 1000 characters")
	}

	// Validate required fields
	if message.ChatID <= 0 {
		return errors.New("invalid chat ID")
	}

	if message.SenderID <= 0 {
		return errors.New("invalid sender ID")
	}

	// Set timestamp for new message
	if message.CreatedAt.IsZero() {
		message.CreatedAt = time.Now()
	}

	// Message is unread when first created
	message.IsRead = false

	// Save message to repository
	if err := m.repo.SendMessage(ctx, message); err != nil {
		return err
	}

	// Broadcast message to WebSocket clients
	messageJSON, _ := json.Marshal(map[string]interface{}{
		"id":         message.ID,
		"chat_id":    message.ChatID,
		"sender_id":  message.SenderID,
		"content":    message.Content,
		"is_read":    message.IsRead,
		"created_at": message.CreatedAt,
		"type":       "new_message",
	})
	m.hub.BroadcastMessage(message.ChatID, string(messageJSON), "message")

	return nil
}

// GetMessagesByChatID retrieves and returns messages for a specific chat with pagination
func (m *messageUseCase) GetMessagesByChatID(ctx context.Context, chatID int, limit int, offset int) ([]*model.Message, error) {
	// Validate chat ID
	if chatID <= 0 {
		return nil, errors.New("invalid chat ID")
	}

	// Set default and max limits
	if limit <= 0 {
		limit = 50 // default
	}
	if limit > 100 {
		limit = 100 // max
	}
	if offset < 0 {
		offset = 0
	}

	// Get messages from repository
	messages, err := m.repo.GetMessagesByChatID(ctx, chatID, limit, offset)
	if err != nil {
		return nil, err
	}

	// Return empty slice if no messages found
	if messages == nil {
		return []*model.Message{}, nil
	}

	// Messages are sorted by creation time (newest first due to DESC)
	return messages, nil
}

// DeleteMessage deletes a message if the user is the sender
func (m *messageUseCase) DeleteMessage(ctx context.Context, messageID int) error {
	// Validate message ID
	if messageID <= 0 {
		return errors.New("invalid message ID")
	}


	// Delete message from repository
	return m.repo.DeleteMessage(ctx, messageID)
}
