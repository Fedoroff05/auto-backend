package usecase

import (
	"context"
	"errors"

	"github.com/Fedoroff05/auto-backend/internal/domain"
)

type ValuationUsecase struct {
	valuationRepo domain.ValuationRepository
}

func NewValuationUsecase(valuationRepo domain.ValuationRepository) *ValuationUsecase {
	return &ValuationUsecase{valuationRepo: valuationRepo}
}

func (u *ValuationUsecase) EstimatePrice(ctx context.Context, brandID, modelID, year int) (*domain.MarketValuationResult, error) {
	//расчет средней цены аналогичных машин
	avgPrice, count, err := u.valuationRepo.GetAveragePrice(ctx, brandID, modelID, year)
	if err != nil {
		return nil, err
	}

	//если есть хотя бы 1 аналог
	if count > 0 && avgPrice > 0 {
		return &domain.MarketValuationResult{
			BrandID:        brandID,
			ModelID:        modelID,
			TargetYear:     year,
			EstimatedPrice: avgPrice,
			Method:         "average_similar",
			SampleCount:    count,
		}, nil
	}

	//если совпадений нет берем цену первой опубликованной машины данной марки/модели
	firstPrice, err := u.valuationRepo.GetFirstPublishedPrice(ctx, brandID, modelID)
	if err != nil {
		if errors.Is(err, domain.ErrListingNotFound) {
			return nil, errors.New("insufficient market data to estimate price for this car model")
		}
		return nil, err
	}

	return &domain.MarketValuationResult{
		BrandID:        brandID,
		ModelID:        modelID,
		TargetYear:     year,
		EstimatedPrice: firstPrice,
		Method:         "first_published_fallback",
		SampleCount:    1,
	}, nil
}
