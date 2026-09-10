package domain

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// описание эксперимента
type ABExperiment struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
}

// сохраненное событие действия пользователя
type AnalyticsEvent struct {
	ID             uuid.UUID       `json:"id"`
	UserID         *uuid.UUID      `json:"user_id,omitempty"`
	ExperimentName string          `json:"experiment_name"`
	Variant        string          `json:"variant"` // "control" или "test"
	EventType      string          `json:"event_type"`
	Payload        json.RawMessage `json:"payload,omitempty"`
	CreatedAt      time.Time       `json:"created_at"`
}

// контракт сохранения и выборки событий
type ABRepository interface {
	TrackEvent(ctx context.Context, event *AnalyticsEvent) error
}
