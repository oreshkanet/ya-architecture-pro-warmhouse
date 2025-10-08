package aggregator

import (
	"context"
	"telemetry-service/internal/domain"
	"telemetry-service/internal/repository"
)

type AggregationEngine struct {
	repo *repository.ClickHouseRepo
}

func NewAggregationEngine(repo *repository.ClickHouseRepo) *AggregationEngine {
	return &AggregationEngine{repo: repo}
}

// В простейшем виде — просто передаём данные в репозиторий
// В дальнейшем здесь можно делать буферизацию, оконную агрегацию и т.д.
func (e *AggregationEngine) ProcessBatch(ctx context.Context, points []domain.TelemetryPoint) error {
	// Здесь можно фильтровать, агрегировать, детектировать аномалии
	// В MVP просто сохраняем
	return e.repo.Insert(ctx, points)
}
