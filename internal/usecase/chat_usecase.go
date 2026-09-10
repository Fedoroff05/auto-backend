package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/Fedoroff05/auto-backend/internal/domain"
)

type ChatUsecase struct {
	chatRepo    domain.ChatRepository
	listingRepo domain.ListingRepository
}

func NewChatUsecase(chatRepo domain.ChatRepository, listingRepo domain.ListingRepository) *ChatUsecase {
	return &ChatUsecase{
		chatRepo:    chatRepo,
		listingRepo: listingRepo,
	}
}

// инициирует диалог покупателя с продавцом авто
func (u *ChatUsecase) StartChat(ctx context.Context, buyerID, listingID uuid.UUID) (*domain.Chat, error) {
	listing, err := u.listingRepo.GetByID(ctx, listingID)
	if err != nil {
		return nil, err
	}

	//покупатель не может писать сам себе
	if listing.UserID == buyerID {
		return nil, errors.New("cannot start chat with yourself")
	}
	return u.chatRepo.GetOrCreateChat(ctx, listingID, buyerID, listing.UserID)
}

func (u *ChatUsecase) GetUserChats(ctx context.Context, userID uuid.UUID) ([]domain.Chat, error) {
	return u.chatRepo.GetUserChats(ctx, userID)
}

func (u *ChatUsecase) GetMessages(ctx context.Context, userID, chatID uuid.UUID, limit, offset int) ([]domain.Message, error) {
	chat, err := u.chatRepo.GetChatByID(ctx, chatID)
	if err != nil {
		return nil, err
	}
	//доступ к сообщениям только у участников диалога
	if chat.BuyerID != userID && chat.SellerID != userID {
		return nil, errors.New("access denied to this chat")
	}
	return u.chatRepo.GetChatMessages(ctx, chatID, limit, offset)
}

func (u *ChatUsecase) SaveNewMessage(ctx context.Context, chatID, senderID uuid.UUID, content string) (*domain.Message, error) {
	msg := &domain.Message{
		ID:        uuid.New(),
		ChatID:    chatID,
		SenderID:  senderID,
		Content:   content,
		IsRead:    false,
		CreatedAt: time.Now().UTC(),
	}
	if err := u.chatRepo.SaveMessage(ctx, msg); err != nil {
		return nil, err
	}

	return msg, nil
}
