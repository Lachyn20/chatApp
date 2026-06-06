package repository

import (
	"chatApp/internal/model"
	"context"
	"database/sql"
	"time"
)

type ChatRepository interface {
	CreateChat(ctx context.Context, chat model.Chat) (int, error)
	GetChatByID(ctx context.Context, chatID int) (*model.Chat, error)
	GetChatsByUserID(ctx context.Context, userID int) ([]*model.Chat, error)
	AddParticipant(ctx context.Context, chatID int, userID int) error
	GetParticipantsByChatID(ctx context.Context, chatID int) ([]*model.User, error)
}

type chatRepository struct {
	db *sql.DB
}

func NewChatRepository(db *sql.DB) ChatRepository {
	return &chatRepository{db: db}
}

func (r *chatRepository) CreateChat(ctx context.Context, chat model.Chat) (int, error) {
	query := `
		INSERT INTO chats (name, is_group, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`
	var id int
	if chat.CreatedAt.IsZero() {
		chat.CreatedAt = time.Now()
	}
	if chat.UpdatedAt.IsZero() {
		chat.UpdatedAt = chat.CreatedAt
	}
	err := r.db.QueryRowContext(ctx, query, chat.Name, chat.IsGroup, chat.CreatedBy, chat.CreatedAt, chat.UpdatedAt).Scan(&id)
	return id, err
}

func (r *chatRepository) GetChatByID(ctx context.Context, chatID int) (*model.Chat, error) {
	query := `
		SELECT id, name, is_group, created_by, created_at, updated_at
		FROM chats
		WHERE id = $1
	`

	var chat model.Chat
	if err := r.db.QueryRowContext(ctx, query, chatID).Scan(
		&chat.ID,
		&chat.Name,
		&chat.IsGroup,
		&chat.CreatedBy,
		&chat.CreatedAt,
		&chat.UpdatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &chat, nil
}

func (r *chatRepository) GetChatsByUserID(ctx context.Context, userID int) ([]*model.Chat, error) {
	query := `
		SELECT c.id, c.name, c.is_group, c.created_by, c.created_at, c.updated_at
		FROM chats c
		JOIN chats_participants cp ON cp.chat_id = c.id
		WHERE cp.user_id = $1
		ORDER BY c.updated_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var chats []*model.Chat
	for rows.Next() {
		var chat model.Chat
		if err := rows.Scan(
			&chat.ID,
			&chat.Name,
			&chat.IsGroup,
			&chat.CreatedBy,
			&chat.CreatedAt,
			&chat.UpdatedAt,
		); err != nil {
			return nil, err
		}
		chats = append(chats, &chat)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return chats, nil
}

func (r *chatRepository) AddParticipant(ctx context.Context, chatID int, userID int) error {
	query := `
		INSERT INTO chats_participants (chat_id, user_id, joined_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (chat_id, user_id) DO NOTHING
	`
	_, err := r.db.ExecContext(ctx, query, chatID, userID, time.Now())
	return err
}

func (r *chatRepository) GetParticipantsByChatID(ctx context.Context, chatID int) ([]*model.User, error) {
	query := `
		SELECT u.id, u.username, u.email
		FROM users u
		JOIN chats_participants cp ON cp.user_id = u.id
		WHERE cp.chat_id = $1
		ORDER BY u.username
	`

	rows, err := r.db.QueryContext(ctx, query, chatID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var participants []*model.User
	for rows.Next() {
		var user model.User
		if err := rows.Scan(&user.ID, &user.Username, &user.Email); err != nil {
			return nil, err
		}
		participants = append(participants, &user)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return participants, nil
}
