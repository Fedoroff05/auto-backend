package usecase

import (
	"context"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/Fedoroff05/auto-backend/internal/domain"
	"github.com/Fedoroff05/auto-backend/pkg/s3"
)

const (
	maxImageSize = 10 * 1024 * 1024
)

// реализует бизнес-сценарии каталога объявлений
type ListingUsecase struct {
	listingRepo domain.ListingRepository
	s3Client    *s3.Client
}

// конструктор юзкейса объявлений
func NewListingUsecase(listingRepo domain.ListingRepository, s3Client *s3.Client) *ListingUsecase {
	return &ListingUsecase{
		listingRepo: listingRepo,
		s3Client:    s3Client,
	}
}

// создает новое объявление со статусом Active
func (u *ListingUsecase) CreateListing(ctx context.Context, userID uuid.UUID, listing *domain.Listing) (*domain.Listing, error) {
	if listing.Price < 0 || listing.Mileage < 0 || listing.Year < 1900 {
		return nil, domain.ErrInvalidListingData
	}

	now := time.Now().UTC()
	listing.ID = uuid.New()
	listing.UserID = userID
	listing.Status = domain.ListingStatusActive
	listing.CreatedAt = now
	listing.UpdatedAt = now

	if err := u.listingRepo.Create(ctx, listing); err != nil {
		return nil, fmt.Errorf("listing_usecase: failed to create: %w", err)
	}

	createdListing, err := u.listingRepo.GetByID(ctx, listing.ID)
	if err != nil {
		return nil, err
	}

	return createdListing, nil
}

// возвращает объявление по ID
func (u *ListingUsecase) GetListingByID(ctx context.Context, id uuid.UUID) (*domain.Listing, error) {
	return u.listingRepo.GetByID(ctx, id)
}

// ищет объявления по фильтрам с пагинацией
func (u *ListingUsecase) GetListings(ctx context.Context, filter domain.ListingFilter) ([]domain.Listing, int, error) {
	return u.listingRepo.List(ctx, filter)
}

// обновляет объявление с проверкой прав доступа
func (u *ListingUsecase) UpdateListing(ctx context.Context, userID uuid.UUID, userRole domain.Role, listing *domain.Listing) (*domain.Listing, error) {
	existing, err := u.listingRepo.GetByID(ctx, listing.ID)
	if err != nil {
		return nil, err
	}

	if existing.UserID != userID && userRole != domain.RoleAdmin && userRole != domain.RoleModerator {
		return nil, domain.ErrListingForbidden
	}

	existing.BrandID = listing.BrandID
	existing.ModelID = listing.ModelID
	existing.Year = listing.Year
	existing.Price = listing.Price
	existing.Mileage = listing.Mileage
	existing.VIN = listing.VIN
	existing.Description = listing.Description
	existing.UpdatedAt = time.Now().UTC()

	if err := u.listingRepo.Update(ctx, existing); err != nil {
		return nil, fmt.Errorf("listing_usecase: failed to update: %w", err)
	}

	return u.listingRepo.GetByID(ctx, existing.ID)
}

// удаляет объявление с проверкой прав доступа
func (u *ListingUsecase) DeleteListing(ctx context.Context, userID uuid.UUID, userRole domain.Role, listingID uuid.UUID) error {
	existing, err := u.listingRepo.GetByID(ctx, listingID)
	if err != nil {
		return err
	}

	if existing.UserID != userID && userRole != domain.RoleAdmin && userRole != domain.RoleModerator {
		return domain.ErrListingForbidden
	}

	for _, img := range existing.Images {
		_ = u.s3Client.DeleteImage(ctx, img.ImageURL)
	}

	return u.listingRepo.Delete(ctx, listingID)
}

// валидирует файл, заливает в MinIO и сохраняет ссылку
func (u *ListingUsecase) UploadImage(ctx context.Context, userID uuid.UUID, userRole domain.Role, listingID uuid.UUID, fileName string, fileBytes []byte) (*domain.ListingImage, error) {
	existing, err := u.listingRepo.GetByID(ctx, listingID)
	if err != nil {
		return nil, err
	}

	if existing.UserID != userID && userRole != domain.RoleAdmin && userRole != domain.RoleModerator {
		return nil, domain.ErrListingForbidden
	}

	if len(fileBytes) > maxImageSize {
		return nil, domain.ErrImageTooLarge
	}

	contentType := http.DetectContentType(fileBytes)
	if contentType != "image/jpeg" && contentType != "image/png" && contentType != "image/webp" {
		return nil, domain.ErrInvalidImageFormat
	}

	ext := strings.ToLower(filepath.Ext(fileName))
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".webp" {
		return nil, domain.ErrInvalidImageFormat
	}
	imageURL, err := u.s3Client.UploadImage(ctx, fileName, fileBytes, contentType)
	if err != nil {
		return nil, fmt.Errorf("listing_usecase: failed to upload to s3: %w", err)
	}

	isMain := len(existing.Images) == 0
	image := &domain.ListingImage{
		ID:        uuid.New(),
		ListingID: listingID,
		ImageURL:  imageURL,
		IsMain:    isMain,
		CreatedAt: time.Now().UTC(),
	}

	if err := u.listingRepo.AddImage(ctx, image); err != nil {
		_ = u.s3Client.DeleteImage(ctx, imageURL)
		return nil, fmt.Errorf("listing_usecase: failed to save image to db: %w", err)
	}

	return image, nil
}

func (u *ListingUsecase) GetBrands(ctx context.Context) ([]domain.CarBrand, error) {
	return u.listingRepo.GetBrands(ctx)
}

func (u *ListingUsecase) GetModels(ctx context.Context, brandID int) ([]domain.CarModel, error) {
	return u.listingRepo.GetModelsByBrandID(ctx, brandID)
}
