package usecase

import (
	"context"
	"encoding/json"
	"hash/fnv"
	"time"

	"github.com/google/uuid"

	"github.com/Fedoroff05/auto-backend/internal/domain"
)

type ABUsecase struct {
	repo domain.ABRepository
}

func NewABUsecase(repo domain.ABRepository) *ABUsecase {
	return &ABUsecase{repo: repo}
}

// делит пользователей 50/50 на основе хэша UUID
func (u *ABUsecase) GetVariant(userID uuid.UUID, experimentName string) string {
	hasher := fnv.New32a()
	hasher.Write([]byte(userID.String() + experimentName))
	hashValue := hasher.Sum32()

	if hashValue%2 == 0 {
		return "control"
	}
	return "test_v1"
}

// сохраняет событие клика/действия
func (u *ABUsecase) LogEvent(ctx context.Context, userID *uuid.UUID, expName, variant, eventType string, payload json.RawMessage) error {
	event := &domain.AnalyticsEvent{
		ID:             uuid.New(),
		UserID:         userID,
		ExperimentName: expName,
		Variant:        variant,
		EventType:      eventType,
		Payload:        payload,
		CreatedAt:      time.Now().UTC(),
	}
	return u.repo.TrackEvent(ctx, event)
}
