package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// сущность диалога между покупателем и продавцом по конкретному авто
type Chat struct {
	ID        uuid.UUID `json:"id"`
	ListingID uuid.UUID `json:"listing_id"`
	BuyerID   uuid.UUID `json:"buyer_id"`
	SellerID  uuid.UUID `json:"seller_id"`
	CreatedAt time.Time `json:"created_at"`
}

// сообщение внутри чата
type Message struct {
	ID        uuid.UUID `json:"id"`
	ChatID    uuid.UUID `json:"chat_id"`
	SenderID  uuid.UUID `json:"sender_id"`
	Content   string    `json:"content"`
	IsRead    bool      `json:"is_read"`
	CreatedAt time.Time `json:"created_at"`
}

// контракт хранилища сообщений и чатов в бд
type ChatRepository interface {
	GetOrCreateChat(ctx context.Context, listingID, buyerID, sellerID uuid.UUID) (*Chat, error)
	GetChatByID(ctx context.Context, chatID uuid.UUID) (*Chat, error)
	GetUserChats(ctx context.Context, userID uuid.UUID) ([]Chat, error)
	SaveMessage(ctx context.Context, msg *Message) error
	GetChatMessages(ctx context.Context, chatID uuid.UUID, limit, offset int) ([]Message, error)
}
