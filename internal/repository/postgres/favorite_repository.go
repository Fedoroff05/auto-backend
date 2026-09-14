package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Fedoroff05/auto-backend/internal/domain"
)

type FavoriteRepository struct {
	pool *pgxpool.Pool
}

func NewFavoriteRepository(pool *pgxpool.Pool) *FavoriteRepository {
	return &FavoriteRepository{pool: pool}
}

func (r *FavoriteRepository) Add(ctx context.Context, userID, listingID uuid.UUID) (bool, error) {
	query := `
		INSERT INTO favorites (user_id, listing_id, created_at)
		VALUES ($1, $2, NOW())
		ON CONFLICT (user_id, listing_id) DO NOTHING
	`

	tag, err := r.pool.Exec(ctx, query, userID, listingID)
	if err != nil {
		return false, fmt.Errorf("failed to add favorite: %w", err)
	}

	return tag.RowsAffected() > 0, nil
}

func (r *FavoriteRepository) Remove(ctx context.Context, userID, listingID uuid.UUID) (bool, error) {
	query := `DELETE FROM favorites WHERE user_id = $1 AND listing_id = $2`

	tag, err := r.pool.Exec(ctx, query, userID, listingID)
	if err != nil {
		return false, fmt.Errorf("failed to remove favorite: %w", err)
	}

	return tag.RowsAffected() > 0, nil
}

func (r *FavoriteRepository) IsFavorite(ctx context.Context, userID, listingID uuid.UUID) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM favorites WHERE user_id = $1 AND listing_id = $2)`
	var exists bool
	err := r.pool.QueryRow(ctx, query, userID, listingID).Scan(&exists)
	return exists, err
}

func (r *FavoriteRepository) GetUserFavorites(ctx context.Context, userID uuid.UUID) ([]domain.Listing, error) {
	query := `
		SELECT 
			l.id, l.user_id, l.brand_id, b.name AS brand_name,
			l.model_id, m.name AS model_name,
			l.year, l.price, l.mileage, l.vin, l.description, l.status,
			l.created_at, l.updated_at
		FROM favorites f
		JOIN listings l ON f.listing_id = l.id
		JOIN car_brands b ON l.brand_id = b.id
		JOIN car_models m ON l.model_id = m.id
		WHERE f.user_id = $1
		ORDER BY f.created_at DESC
	`
	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user favorites: %w", err)
	}
	defer rows.Close()

	var listings []domain.Listing
	for rows.Next() {
		var l domain.Listing
		if err := rows.Scan(
			&l.ID, &l.UserID, &l.BrandID, &l.BrandName,
			&l.ModelID, &l.ModelName, &l.Year, &l.Price,
			&l.Mileage, &l.VIN, &l.Description, &l.Status,
			&l.CreatedAt, &l.UpdatedAt,
		); err != nil {
			return nil, err
		}
		listings = append(listings, l)
	}
	return listings, nil
}
