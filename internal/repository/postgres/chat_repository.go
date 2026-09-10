package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Fedoroff05/auto-backend/internal/domain"
)

type ChatRepository struct {
	pool *pgxpool.Pool
}

func NewChatRepository(pool *pgxpool.Pool) *ChatRepository {
	return &ChatRepository{pool: pool}
}

// находит существующий чат или создает новый
func (r *ChatRepository) GetOrCreateChat(ctx context.Context, listingID, buyerID, sellerID uuid.UUID) (*domain.Chat, error) {
	queryFind := `SELECT id, listing_id, buyer_id, seller_id, created_at FROM chats WHERE listing_id = $1 AND buyer_id = $2`
	c := &domain.Chat{}
	err := r.pool.QueryRow(ctx, queryFind, listingID, buyerID).Scan(&c.ID, &c.ListingID, &c.BuyerID, &c.SellerID, &c.CreatedAt)
	if err == nil {
		return c, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("failed to find chat: %w", err)
	}

	//чат не найден - создаем
	queryInsert := `
		INSERT INTO chats (id, listing_id, buyer_id, seller_id, created_at)
		VALUES ($1, $2, $3, $4, NOW())
		RETURNING id, listing_id, buyer_id, seller_id, created_at
	`
	newID := uuid.New()
	err = r.pool.QueryRow(ctx, queryInsert, newID, listingID, buyerID, sellerID).Scan(&c.ID, &c.ListingID, &c.BuyerID, &c.SellerID, &c.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create chat: %w", err)
	}

	return c, nil
}

func (r *ChatRepository) GetChatByID(ctx context.Context, chatID uuid.UUID) (*domain.Chat, error) {
	query := `SELECT id, listing_id, buyer_id, seller_id, created_at FROM chats WHERE id = $1`
	c := &domain.Chat{}
	err := r.pool.QueryRow(ctx, query, chatID).Scan(&c.ID, &c.ListingID, &c.BuyerID, &c.SellerID, &c.CreatedAt)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (r *ChatRepository) GetUserChats(ctx context.Context, userID uuid.UUID) ([]domain.Chat, error) {
	query := `
		SELECT id, listing_id, buyer_id, seller_id, created_at
		FROM chats
		WHERE buyer_id = $1 OR seller_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user chats: %w", err)
	}
	defer rows.Close()

	var chats []domain.Chat
	for rows.Next() {
		var c domain.Chat
		if err := rows.Scan(&c.ID, &c.ListingID, &c.BuyerID, &c.SellerID, &c.CreatedAt); err != nil {
			return nil, err
		}
		chats = append(chats, c)
	}
	return chats, nil
}

func (r *ChatRepository) SaveMessage(ctx context.Context, msg *domain.Message) error {
	query := `
		INSERT INTO messages (id, chat_id, sender_id, content, is_read, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.pool.Exec(ctx, query, msg.ID, msg.ChatID, msg.SenderID, msg.Content, msg.IsRead, msg.CreatedAt)
	return err
}

func (r *ChatRepository) GetChatMessages(ctx context.Context, chatID uuid.UUID, limit, offset int) ([]domain.Message, error) {
	query := `
		SELECT id, chat_id, sender_id, content, is_read, created_at
		FROM messages
		WHERE chat_id = $1
		ORDER BY created_at ASC
		LIMIT $2 OFFSET $3
	`
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	rows, err := r.pool.Query(ctx, query, chatID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query messages: %w", err)
	}
	defer rows.Close()

	var messages []domain.Message
	for rows.Next() {
		var m domain.Message
		if err := rows.Scan(&m.ID, &m.ChatID, &m.SenderID, &m.Content, &m.IsRead, &m.CreatedAt); err != nil {
			return nil, err
		}
		messages = append(messages, m)
	}
	return messages, nil
}
