package usecase

import (
	"chatApp/internal/model"
	"chatApp/internal/repository"
	"context"
	"errors"
	"time"
)

type ChatUsecase interface {
	CreateChat(ctx context.Context, chat model.Chat, participantIDs []int) (*model.Chat, error)
	GetChatsByUserID(ctx context.Context, userID int) ([]*model.Chat, error)
	GetChatByID(ctx context.Context, chatID int) (*model.Chat, error)
	AddParticipant(ctx context.Context, chatID int, userID int) error
	GetParticipantsByChatID(ctx context.Context, chatID int) ([]*model.User, error)
}

type chatUsecase struct {
	repo repository.ChatRepository
}

func NewChatUsecase(repo repository.ChatRepository) ChatUsecase {
	return &chatUsecase{repo: repo}
}

func (u *chatUsecase) CreateChat(ctx context.Context, chat model.Chat, participantIDs []int) (*model.Chat, error) {
	if chat.CreatedBy <= 0 {
		return nil, errors.New("invalid creator user id")
	}

	if len(participantIDs) == 0 {
		return nil, errors.New("at least one participant is required")
	}

	participantSet := map[int]struct{}{}
	for _, id := range participantIDs {
		if id <= 0 {
			return nil, errors.New("invalid participant id")
		}
		if id == chat.CreatedBy {
			continue
		}
		participantSet[id] = struct{}{}
	}

	if chat.IsGroup {
		if len(participantSet) == 0 {
			return nil, errors.New("group chat requires at least one other participant")
		}
	} else {
		if len(participantSet) != 1 {
			return nil, errors.New("direct chat requires exactly one other participant")
		}
	}

	if chat.CreatedAt.IsZero() {
		chat.CreatedAt = time.Now()
	}
	if chat.UpdatedAt.IsZero() {
		chat.UpdatedAt = chat.CreatedAt
	}

	chatID, err := u.repo.CreateChat(ctx, chat)
	if err != nil {
		return nil, err
	}

	participantsToAdd := map[int]struct{}{
		chat.CreatedBy: {},
	}
	for _, id := range participantIDs {
		participantsToAdd[id] = struct{}{}
	}

	for id := range participantsToAdd {
		if err := u.repo.AddParticipant(ctx, chatID, id); err != nil {
			return nil, err
		}
	}

	chat.ID = chatID
	return &chat, nil
}

func (u *chatUsecase) GetChatsByUserID(ctx context.Context, userID int) ([]*model.Chat, error) {
	if userID <= 0 {
		return nil, errors.New("invalid user id")
	}

	return u.repo.GetChatsByUserID(ctx, userID)
}

func (u *chatUsecase) GetChatByID(ctx context.Context, chatID int) (*model.Chat, error) {
	if chatID <= 0 {
		return nil, errors.New("invalid chat id")
	}

	return u.repo.GetChatByID(ctx, chatID)
}

func (u *chatUsecase) AddParticipant(ctx context.Context, chatID int, userID int) error {
	if chatID <= 0 {
		return errors.New("invalid chat id")
	}
	if userID <= 0 {
		return errors.New("invalid user id")
	}

	chat, err := u.repo.GetChatByID(ctx, chatID)
	if err != nil {
		return err
	}
	if chat == nil {
		return errors.New("chat not found")
	}
	if !chat.IsGroup {
		return errors.New("cannot add participants to a direct chat")
	}

	participants, err := u.repo.GetParticipantsByChatID(ctx, chatID)
	if err != nil {
		return err
	}
	for _, participant := range participants {
		if participant.ID == userID {
			return errors.New("user is already a participant in the chat")
		}
	}

	return u.repo.AddParticipant(ctx, chatID, userID)
}

func (u *chatUsecase) GetParticipantsByChatID(ctx context.Context, chatID int) ([]*model.User, error) {
	if chatID <= 0 {
		return nil, errors.New("invalid chat id")
	}

	return u.repo.GetParticipantsByChatID(ctx, chatID)
}
