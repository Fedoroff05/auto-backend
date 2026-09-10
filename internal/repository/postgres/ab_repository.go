package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Fedoroff05/auto-backend/internal/domain"
)

type ABRepository struct {
	pool *pgxpool.Pool
}

func NewABRepository(pool *pgxpool.Pool) *ABRepository {
	return &ABRepository{pool: pool}
}

func (r *ABRepository) TrackEvent(ctx context.Context, event *domain.AnalyticsEvent) error {
	query := `
		INSERT INTO analytics_events (id, user_id, experiment_name, variant, event_type, payload, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.pool.Exec(ctx, query,
		event.ID,
		event.UserID,
		event.ExperimentName,
		event.Variant,
		event.EventType,
		event.Payload,
		event.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to track analytics event: %w", err)
	}
	return nil
}
