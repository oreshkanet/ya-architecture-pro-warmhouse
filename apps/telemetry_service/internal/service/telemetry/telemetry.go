package telemetry

import (
	"context"
	"telemetry-service/internal/domain"
	"telemetry-service/internal/repository"
	"time"
)

type TelemetryService struct {
	repo *repository.ClickHouseRepo
}

func NewTelemetryService(repo *repository.ClickHouseRepo) *TelemetryService {
	return &TelemetryService{
		repo: repo,
	}
}

func (s *TelemetryService) GetRaw(ctx context.Context, from, to time.Time, filters map[string]string, limit int) ([]*domain.TelemetryPoint, error) {
	return s.repo.GetRaw(ctx, from, to, filters, limit)
}

func (s *TelemetryService) GetLastBySensorID(ctx context.Context, sensorID string) (*domain.TelemetryPoint, error) {
	return s.repo.GetLastBySensorID(ctx, sensorID)
}
