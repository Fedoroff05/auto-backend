package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// результат алгоритма рыночной оценки
type MarketValuationResult struct {
	BrandID        int     `json:"brand_id"`
	ModelID        int     `json:"model_id"`
	TargetYear     int     `json:"target_year"`
	EstimatedPrice float64 `json:"estimated_price"`
	Method         string  `json:"method"`
	SampleCount    int     `json:"sample_count"`
}

// сущность избранного
type Favorite struct {
	UserID    uuid.UUID `json:"user_id"`
	ListingID uuid.UUID `json:"listing_id"`
	CreatedAt time.Time `json:"created_at"`
}

// контракт для избранного
type FavoriteRepository interface {
	Add(ctx context.Context, userID, listingID uuid.UUID) error
	Remove(ctx context.Context, userID, listingID uuid.UUID) error
	IsFavorite(ctx context.Context, userID, listingID uuid.UUID) (bool, error)
	GetUserFavorites(ctx context.Context, userID uuid.UUID) ([]Listing, error)
}

// контракт для расчета цены
type ValuationRepository interface {
	GetAveragePrice(ctx context.Context, brandID, modelID, year int) (float64, int, error)
	GetFirstPublishedPrice(ctx context.Context, brandID, modelID int) (float64, error)
}
