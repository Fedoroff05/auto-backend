package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Fedoroff05/auto-backend/internal/domain"
)

type ValuationRepository struct {
	pool *pgxpool.Pool
}

func NewValuationRepository(pool *pgxpool.Pool) *ValuationRepository {
	return &ValuationRepository{pool: pool}
}

// считает среднюю цену авто той же марки/модели и года
func (r *ValuationRepository) GetAveragePrice(ctx context.Context, brandID, modelID, year int) (float64, int, error) {
	query := `
		SELECT COALESCE(AVG(price), 0), COUNT(*)
		FROM listings
		WHERE brand_id = $1 AND model_id = $2 
		  AND year BETWEEN $3 AND $4
		  AND status = 'active'
	`
	var avgPrice float64
	var count int

	err := r.pool.QueryRow(ctx, query, brandID, modelID, year-2, year+2).Scan(&avgPrice, &count)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to calculate avg price: %w", err)
	}

	return avgPrice, count, nil
}

// находит цену самой первой добавленной машины данной марки и модели
func (r *ValuationRepository) GetFirstPublishedPrice(ctx context.Context, brandID, modelID int) (float64, error) {
	query := `
		SELECT price
		FROM listings
		WHERE brand_id = $1 AND model_id = $2
		ORDER BY created_at ASC
		LIMIT 1
	`
	var price float64
	err := r.pool.QueryRow(ctx, query, brandID, modelID).Scan(&price)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, domain.ErrListingNotFound
		}
		return 0, fmt.Errorf("failed to get first published price: %w", err)
	}

	return price, nil
}
