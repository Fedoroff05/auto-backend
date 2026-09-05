package domain

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

// статусы
type ListingStatus string

const (
	ListingStatusDraft  ListingStatus = "draft"
	ListingStatusActive ListingStatus = "active"
	ListingStatusSold   ListingStatus = "sold"
	ListingStatusBanned ListingStatus = "banned"
)

// доменные ошибки
var (
	ErrListingNotFound    = errors.New("listing not found")
	ErrListingForbidden   = errors.New("you are not allowed to modify this listing")
	ErrBrandNotFound      = errors.New("car brand not found")
	ErrModelNotFound      = errors.New("car model not found")
	ErrInvalidListingData = errors.New("invalid listing data")
	ErrInvalidImageFormat = errors.New("invalid image format, only jpeg and png are allowed")
	ErrImageTooLarge      = errors.New("image size exceeds maximum limit")
)

// справочник марок автомобилей
type CarBrand struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// справочник моделей автомобилей
type CarModel struct {
	ID      int    `json:"id"`
	BrandID int    `json:"brand_id"`
	Name    string `json:"name"`
}

// фотография объявления
type ListingImage struct {
	ID        uuid.UUID `json:"id"`
	ListingID uuid.UUID `json:"listing_id"`
	ImageURL  string    `json:"image_url"`
	IsMain    bool      `json:"is_main"`
	CreatedAt time.Time `json:"created_at"`
}

// центральная сущность объявления
type Listing struct {
	ID          uuid.UUID      `json:"id"`
	UserID      uuid.UUID      `json:"user_id"`
	BrandID     int            `json:"brand_id"`
	BrandName   string         `json:"brand_name,omitempty"`
	ModelID     int            `json:"model_id"`
	ModelName   string         `json:"model_name,omitempty"`
	Year        int            `json:"year"`
	Price       float64        `json:"price"`
	Mileage     int            `json:"mileage"`
	VIN         *string        `json:"vin,omitempty"`
	Description string         `json:"description"`
	Status      ListingStatus  `json:"status"`
	Images      []ListingImage `json:"images,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

// структура для передачи параметров поиска и пагинации
type ListingFilter struct {
	BrandID  *int           `json:"brand_id"`
	ModelID  *int           `json:"model_id"`
	MinYear  *int           `json:"min_year"`
	MaxYear  *int           `json:"max_year"`
	MinPrice *float64       `json:"min_price"`
	MaxPrice *float64       `json:"max_price"`
	Status   *ListingStatus `json:"status"`
	Limit    int            `json:"limit"`
	Offset   int            `json:"offset"`
}

// контракт, который будет реализовывать бд
type ListingRepository interface {
	Create(ctx context.Context, listing *Listing) error
	GetByID(ctx context.Context, id uuid.UUID) (*Listing, error)
	List(ctx context.Context, filter ListingFilter) ([]Listing, int, error)
	Update(ctx context.Context, listing *Listing) error
	Delete(ctx context.Context, id uuid.UUID) error

	//действия с изображениями
	AddImage(ctx context.Context, image *ListingImage) error
	GetImagesByListingID(ctx context.Context, listingID uuid.UUID) ([]ListingImage, error)
	DeleteImage(ctx context.Context, imageID uuid.UUID) error

	//справочники
	GetBrands(ctx context.Context) ([]CarBrand, error)
	GetModelsByBrandID(ctx context.Context, brandID int) ([]CarModel, error)
}

// сценарии каталога объявлений
type ListingUsecase interface {
	CreateListing(ctx context.Context, userID uuid.UUID, listing *Listing) (*Listing, error)
	GetListingByID(ctx context.Context, id uuid.UUID) (*Listing, error)
	GetListings(ctx context.Context, filter ListingFilter) ([]Listing, int, error)
	UpdateListing(ctx context.Context, userID uuid.UUID, userRole Role, listing *Listing) (*Listing, error)
	DeleteListing(ctx context.Context, userID uuid.UUID, userRole Role, listingID uuid.UUID) error

	UploadImage(ctx context.Context, userID uuid.UUID, userRole Role, listingID uuid.UUID, fileName string, fileBytes []byte) (*ListingImage, error)

	GetBrands(ctx context.Context) ([]CarBrand, error)
	GetModels(ctx context.Context, brandID int) ([]CarModel, error)
}
