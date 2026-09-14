package usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/Fedoroff05/auto-backend/internal/domain"
	"github.com/Fedoroff05/auto-backend/pkg/redis"
)

type FavoriteUsecase struct {
	favRepo     domain.FavoriteRepository
	listingRepo domain.ListingRepository
	redisClient *redis.Client
}

func NewFavoriteUsecase(
	favRepo domain.FavoriteRepository,
	listingRepo domain.ListingRepository,
	redisClient *redis.Client,
) *FavoriteUsecase {
	return &FavoriteUsecase{
		favRepo:     favRepo,
		listingRepo: listingRepo,
		redisClient: redisClient,
	}
}

// добавляет в бд и инкрементит счетчик в redis
func (u *FavoriteUsecase) AddToFavorites(ctx context.Context, userID, listingID uuid.UUID) error {
	if _, err := u.listingRepo.GetByID(ctx, listingID); err != nil {
		return err
	}

	added, err := u.favRepo.Add(ctx, userID, listingID)
	if err != nil {
		return err
	}

	if added {
		key := fmt.Sprintf("listing:fav_count:%s", listingID.String())
		_ = u.redisClient.Incr(ctx, key)
	}

	return nil
}

// удаляет из бд и декрементит счетчик в redis
func (u *FavoriteUsecase) RemoveFromFavorites(ctx context.Context, userID, listingID uuid.UUID) error {
	removed, err := u.favRepo.Remove(ctx, userID, listingID)
	if err != nil {
		return err
	}

	if removed {
		key := fmt.Sprintf("listing:fav_count:%s", listingID.String())
		_ = u.redisClient.Decr(ctx, key)
	}

	return nil
}

func (u *FavoriteUsecase) GetUserFavorites(ctx context.Context, userID uuid.UUID) ([]domain.Listing, error) {
	return u.favRepo.GetUserFavorites(ctx, userID)
}

// возвращает количество добавлений в избранное из кэша Redis
func (u *FavoriteUsecase) GetFavoriteCount(ctx context.Context, listingID uuid.UUID) int {
	key := fmt.Sprintf("listing:fav_count:%s", listingID.String())
	count, _ := u.redisClient.GetInt(ctx, key)
	if count < 0 {
		return 0
	}
	return count
}
