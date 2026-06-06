package repository

import (
	"chatApp/internal/model"
	"context"
	"database/sql"
)

type MessageRepository interface {
	SendMessage(ctx context.Context, message model.Message) error
	GetMessagesByChatID(ctx context.Context, chatID int, limit int, offset int) ([]*model.Message, error)
	GetMessageByID(ctx context.Context, messageID int) (*model.Message, error)
	DeleteMessage(ctx context.Context, messageID int) error
}

type messageRepository struct {
	db *sql.DB
}

func NewMessageRepository(db *sql.DB) MessageRepository {
	return &messageRepository{db: db}
}

func (r *messageRepository) SendMessage(ctx context.Context, message model.Message) error {
	query := `
		INSERT INTO messages (chat_id, sender_id, content, is_read, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := r.db.ExecContext(ctx, query, message.ChatID, message.SenderID, message.Content, message.IsRead, message.CreatedAt)
	return err
}

func (r *messageRepository) GetMessagesByChatID(ctx context.Context, chatID int, limit int, offset int) ([]*model.Message, error) {
	query := `
		SELECT id, chat_id, sender_id, content, is_read, created_at
		FROM messages
		WHERE chat_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, chatID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []*model.Message
	for rows.Next() {
		var message model.Message
		if err := rows.Scan(
			&message.ID,
			&message.ChatID,
			&message.SenderID,
			&message.Content,
			&message.IsRead,
			&message.CreatedAt,
		); err != nil {
			return nil, err
		}
		messages = append(messages, &message)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return messages, nil
}

func (r *messageRepository) GetMessageByID(ctx context.Context, messageID int) (*model.Message, error) {
	query := `
		SELECT id, chat_id, sender_id, content, is_read, created_at
		FROM messages
		WHERE id = $1
	`

	var message model.Message
	err := r.db.QueryRowContext(ctx, query, messageID).Scan(
		&message.ID,
		&message.ChatID,
		&message.SenderID,
		&message.Content,
		&message.IsRead,
		&message.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &message, nil
}

func (r *messageRepository) DeleteMessage(ctx context.Context, messageID int) error {
	query := `
		DELETE FROM messages
		WHERE id = $1
	`
	_, err := r.db.ExecContext(ctx, query, messageID)
	return err
}