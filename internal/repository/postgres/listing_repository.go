package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Fedoroff05/auto-backend/internal/domain"
)

// реализует интерфейс domain.ListingRepository
type ListingRepository struct {
	pool *pgxpool.Pool
}

// конструктор репозитория
func NewListingRepository(pool *pgxpool.Pool) *ListingRepository {
	return &ListingRepository{pool: pool}
}

// сохраняет объявление в БД
func (r *ListingRepository) Create(ctx context.Context, listing *domain.Listing) error {
	query := `
		INSERT INTO listings (id, user_id, brand_id, model_id, year, price, mileage, vin, description, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`

	_, err := r.pool.Exec(ctx, query,
		listing.ID,
		listing.UserID,
		listing.BrandID,
		listing.ModelID,
		listing.Year,
		listing.Price,
		listing.Mileage,
		listing.VIN,
		listing.Description,
		listing.Status,
		listing.CreatedAt,
		listing.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to insert listing: %w", err)
	}

	return nil
}

// возвращает подробные данные об объявлении
func (r *ListingRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Listing, error) {
	query := `
		SELECT 
			l.id, l.user_id, l.brand_id, b.name AS brand_name,
			l.model_id, m.name AS model_name,
			l.year, l.price, l.mileage, l.vin, l.description, l.status,
			l.created_at, l.updated_at
		FROM listings l
		JOIN car_brands b ON l.brand_id = b.id
		JOIN car_models m ON l.model_id = m.id
		WHERE l.id = $1
	`

	l := &domain.Listing{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&l.ID,
		&l.UserID,
		&l.BrandID,
		&l.BrandName,
		&l.ModelID,
		&l.ModelName,
		&l.Year,
		&l.Price,
		&l.Mileage,
		&l.VIN,
		&l.Description,
		&l.Status,
		&l.CreatedAt,
		&l.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrListingNotFound
		}
		return nil, fmt.Errorf("failed to get listing by id: %w", err)
	}

	images, err := r.GetImagesByListingID(ctx, l.ID)
	if err != nil {
		return nil, err
	}
	l.Images = images

	return l, nil
}

// выполняет динамический поиск с фильтрами, пагинацией и подсчетом общего числа записей
func (r *ListingRepository) List(ctx context.Context, filter domain.ListingFilter) ([]domain.Listing, int, error) {
	var conditions []string
	var args []interface{}
	argIdx := 1

	if filter.BrandID != nil {
		conditions = append(conditions, fmt.Sprintf("l.brand_id = $%d", argIdx))
		args = append(args, *filter.BrandID)
		argIdx++
	}
	if filter.ModelID != nil {
		conditions = append(conditions, fmt.Sprintf("l.model_id = $%d", argIdx))
		args = append(args, *filter.ModelID)
		argIdx++
	}
	if filter.MinYear != nil {
		conditions = append(conditions, fmt.Sprintf("l.year >= $%d", argIdx))
		args = append(args, *filter.MinYear)
		argIdx++
	}
	if filter.MaxYear != nil {
		conditions = append(conditions, fmt.Sprintf("l.year <= $%d", argIdx))
		args = append(args, *filter.MaxYear)
		argIdx++
	}
	if filter.MinPrice != nil {
		conditions = append(conditions, fmt.Sprintf("l.price >= $%d", argIdx))
		args = append(args, *filter.MinPrice)
		argIdx++
	}
	if filter.MaxPrice != nil {
		conditions = append(conditions, fmt.Sprintf("l.price <= $%d", argIdx))
		args = append(args, *filter.MaxPrice)
		argIdx++
	}
	if filter.Status != nil {
		conditions = append(conditions, fmt.Sprintf("l.status = $%d", argIdx))
		args = append(args, *filter.Status)
		argIdx++
	} else {
		conditions = append(conditions, fmt.Sprintf("l.status = $%d", argIdx))
		args = append(args, domain.ListingStatusActive)
		argIdx++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM listings l %s", whereClause)
	var total int
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count listings: %w", err)
	}
	if total == 0 {
		return []domain.Listing{}, 0, nil
	}

	selectQuery := fmt.Sprintf(`
		SELECT 
			l.id, l.user_id, l.brand_id, b.name AS brand_name,
			l.model_id, m.name AS model_name,
			l.year, l.price, l.mileage, l.vin, l.description, l.status,
			l.created_at, l.updated_at
		FROM listings l
		JOIN car_brands b ON l.brand_id = b.id
		JOIN car_models m ON l.model_id = m.id
		%s
		ORDER BY l.created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIdx, argIdx+1)

	limit := filter.Limit
	if limit <= 0 || limit > 100 {
		limit = 20 //дефолтный лимит
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	args = append(args, limit, offset)
	rows, err := r.pool.Query(ctx, selectQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to fetch listings: %w", err)
	}
	defer rows.Close()

	listings := make([]domain.Listing, 0, limit)
	for rows.Next() {
		var l domain.Listing
		if err := rows.Scan(
			&l.ID,
			&l.UserID,
			&l.BrandID,
			&l.BrandName,
			&l.ModelID,
			&l.ModelName,
			&l.Year,
			&l.Price,
			&l.Mileage,
			&l.VIN,
			&l.Description,
			&l.Status,
			&l.CreatedAt,
			&l.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan listing row: %w", err)
		}
		listings = append(listings, l)
	}

	return listings, total, nil
}

// обновляет поля объявления
func (r *ListingRepository) Update(ctx context.Context, listing *domain.Listing) error {
	query := `
		UPDATE listings
		SET brand_id = $1, model_id = $2, year = $3, price = $4,
		    mileage = $5, vin = $6, description = $7, status = $8, updated_at = $9
		WHERE id = $10
	`

	cmdTag, err := r.pool.Exec(ctx, query,
		listing.BrandID,
		listing.ModelID,
		listing.Year,
		listing.Price,
		listing.Mileage,
		listing.VIN,
		listing.Description,
		listing.Status,
		listing.UpdatedAt,
		listing.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update listing: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return domain.ErrListingNotFound
	}
	return nil
}

// удаляет объявление
func (r *ListingRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM listings WHERE id = $1`

	cmdTag, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete listing: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return domain.ErrListingNotFound
	}
	return nil
}

// сохраняет ссылку на загруженную фотографию
func (r *ListingRepository) AddImage(ctx context.Context, image *domain.ListingImage) error {
	query := `
		INSERT INTO listing_images (id, listing_id, image_url, is_main, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := r.pool.Exec(ctx, query,
		image.ID,
		image.ListingID,
		image.ImageURL,
		image.IsMain,
		image.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to insert listing image: %w", err)
	}
	return nil
}

// возвращает все фотографии конкретного автомобиля
func (r *ListingRepository) GetImagesByListingID(ctx context.Context, listingID uuid.UUID) ([]domain.ListingImage, error) {
	query := `
		SELECT id, listing_id, image_url, is_main, created_at
		FROM listing_images
		WHERE listing_id = $1
		ORDER BY is_main DESC, created_at ASC
	`

	rows, err := r.pool.Query(ctx, query, listingID)
	if err != nil {
		return nil, fmt.Errorf("failed to query listing images: %w", err)
	}
	defer rows.Close()

	var images []domain.ListingImage
	for rows.Next() {
		var img domain.ListingImage
		if err := rows.Scan(&img.ID, &img.ListingID, &img.ImageURL, &img.IsMain, &img.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan image row: %w", err)
		}
		images = append(images, img)
	}
	return images, nil
}

// удаляет запись о фотографии
func (r *ListingRepository) DeleteImage(ctx context.Context, imageID uuid.UUID) error {
	query := `DELETE FROM listing_images WHERE id = $1`
	cmdTag, err := r.pool.Exec(ctx, query, imageID)
	if err != nil {
		return fmt.Errorf("failed to delete image: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return domain.ErrListingNotFound
	}
	return nil
}

// возвращает справочник всех доступных марок
func (r *ListingRepository) GetBrands(ctx context.Context) ([]domain.CarBrand, error) {
	query := `SELECT id, name FROM car_brands ORDER BY name ASC`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get brands: %w", err)
	}
	defer rows.Close()

	var brands []domain.CarBrand
	for rows.Next() {
		var b domain.CarBrand
		if err := rows.Scan(&b.ID, &b.Name); err != nil {
			return nil, fmt.Errorf("failed to scan brand: %w", err)
		}
		brands = append(brands, b)
	}
	return brands, nil
}

// возвращает список моделей выбранной марки
func (r *ListingRepository) GetModelsByBrandID(ctx context.Context, brandID int) ([]domain.CarModel, error) {
	query := `SELECT id, brand_id, name FROM car_models WHERE brand_id = $1 ORDER BY name ASC`

	rows, err := r.pool.Query(ctx, query, brandID)
	if err != nil {
		return nil, fmt.Errorf("failed to get models: %w", err)
	}
	defer rows.Close()

	var models []domain.CarModel
	for rows.Next() {
		var m domain.CarModel
		if err := rows.Scan(&m.ID, &m.BrandID, &m.Name); err != nil {
			return nil, fmt.Errorf("failed to scan model: %w", err)
		}
		models = append(models, m)
	}
	return models, nil
}
