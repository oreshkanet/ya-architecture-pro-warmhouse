package service

import (
	"context"
	"fmt"
	"time"
	"warm-service/internal/client/connector"
	"warm-service/internal/client/telemetry"
	"warm-service/internal/domain"
	"warm-service/internal/repository"

	"github.com/google/uuid"
)

type WarmService struct {
	repo        *repository.WarmRepo
	publisher   *telemetry.Publisher
	monolithURL string
}

func NewWarmService(repo *repository.WarmRepo, pub *telemetry.Publisher, monolithURL string) *WarmService {
	return &WarmService{repo: repo, publisher: pub, monolithURL: monolithURL}
}

func (s *WarmService) Register(ctx context.Context, req domain.RegisterRequest) (*domain.WarmModule, error) {
	id := uuid.New().String()
	now := time.Now()
	module := &domain.WarmModule{
		ID:              id,
		Name:            req.Name,
		Location:        req.Location,
		SerialNumber:    req.SerialNumber,
		IsOn:            false,
		CurrentTemp:     0,
		TargetTemp:      20,
		Status:          "offline",
		FirmwareVersion: req.FirmwareVersion,
		LastSeen:        now,
		IsLegacy:        req.IsLegacy,
		URL:             req.URL,
	}
	if err := s.repo.Create(ctx, module); err != nil {
		return nil, err
	}
	return module, nil
}

func (s *WarmService) GetAll(ctx context.Context, location, status string) ([]domain.WarmModule, error) {
	return s.repo.GetAll(ctx, location, status)
}

func (s *WarmService) GetByID(ctx context.Context, id string) (*domain.WarmModule, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *WarmService) SetTemperature(ctx context.Context, id string, target float64) (*domain.WarmModule, error) {
	module, err := s.repo.GetByID(ctx, id)
	if err != nil || module == nil {
		return nil, fmt.Errorf("module not found")
	}
	// Здесь можно отправить команду на устройство или монолит
	// Для примера — просто обновляем в БД
	module.TargetTemp = target
	module.LastSeen = time.Now()
	if err := s.repo.Update(ctx, module); err != nil {
		return nil, err
	}
	return module, nil
}

func (s *WarmService) Toggle(ctx context.Context, id string, on bool) (*domain.WarmModule, error) {
	module, err := s.repo.GetByID(ctx, id)
	if err != nil || module == nil {
		return nil, fmt.Errorf("module not found")
	}
	module.IsOn = on
	module.LastSeen = time.Now()
	if err := s.repo.Update(ctx, module); err != nil {
		return nil, err
	}
	return module, nil
}

func (s *WarmService) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

// Сбор телеметрии для одного модуля
func (s *WarmService) CollectTelemetryForModule(ctx context.Context, module *domain.WarmModule) error {
	var tr *connector.TemperatureResponse
	var err error

	if module.IsLegacy {
		tr, err = connector.FetchTelemetryFromMonolith(ctx, s.monolithURL, module.ID)
	} else {
		tr, err = connector.FetchTelemetryFromSensor(ctx, module.URL, module.ID)
	}

	if err != nil {
		// Обновляем статус в offline/error
		module.Status = "error"
		module.LastSeen = time.Now()
		s.repo.Update(ctx, module)
		return err
	}

	// Обновляем данные модуля
	module.CurrentTemp = tr.Value
	module.Status = tr.Status
	module.LastSeen = time.Now()
	s.repo.Update(ctx, module)

	// Отправляем в очередь
	telemetryData := map[string]interface{}{
		"module_id":    module.ID,
		"sensor_id":    tr.SensorID,
		"temperature":  tr.Value,
		"unit":         tr.Unit,
		"timestamp":    tr.Timestamp,
		"location":     tr.Location,
		"status":       tr.Status,
		"type":         "warm",
		"collected_at": time.Now().UTC().Format(time.RFC3339),
		"is_legacy":    module.IsLegacy,
	}

	return s.publisher.Publish(module.ID, telemetryData)
}

// Запуск сбора телеметрии по расписанию
func (s *WarmService) StartTelemetryCollection(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Minute)
	go func() {
		for {
			select {
			case <-ctx.Done():
				ticker.Stop()
				return
			case <-ticker.C:
				modules, _ := s.repo.GetAll(ctx, "", "")
				for _, m := range modules {
					go func(m domain.WarmModule) {
						_ = s.CollectTelemetryForModule(ctx, &m)
					}(m)
				}
			}
		}
	}()
}
