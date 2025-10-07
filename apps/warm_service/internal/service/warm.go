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
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
)

type WarmService struct {
	repo        *repository.WarmRepo
	publisher   *telemetry.Publisher
	monolithURL string
	cron        *cron.Cron
}

func NewWarmService(repo *repository.WarmRepo, pub *telemetry.Publisher, monolithURL string) *WarmService {
	return &WarmService{
		repo:        repo,
		publisher:   pub,
		monolithURL: monolithURL,
		cron:        cron.New(),
	}
}

func (s *WarmService) StartTelemetryCollection(ctx context.Context) {
	// Добавляем задачу: каждую минуту
	_, err := s.cron.AddFunc("@every 1m", func() {
		err := s.collectAllTelemetry(ctx)
		if err != nil {
			logrus.Errorf("Failed to collect telemetry: %v", err)
		}
	})
	if err != nil {
		logrus.Fatalf("Failed to schedule telemetry collection: %v", err)
	}

	s.cron.Start()
	logrus.Info("Telemetry collection scheduled: every 1 minute")

	// Остановка при завершении
	go func() {
		<-ctx.Done()
		logrus.Info("Stopping telemetry cron...")
		s.cron.Stop()
	}()
}

// Сбор телеметрии со всех модулей
func (s *WarmService) collectAllTelemetry(ctx context.Context) error {
	sensors, err := s.repo.GetAll(ctx, "", "") // без фильтра — все модули
	if err != nil {
		return fmt.Errorf("failed to fetch sensors: %w", err)
	}

	for _, m := range sensors {
		sensor := m // захватываем копию для горутины
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			var tr *connector.TemperatureResponse
			var err error

			if sensor.IsLegacy {
				tr, err = connector.FetchTelemetryFromMonolith(ctx, sensor.ID, s.monolithURL)
			} else {
				tr, err = connector.FetchTelemetryFromSensor(ctx, sensor.URL, sensor.ID)
			}

			if err != nil {
				logrus.WithError(err).WithField("sensor_id", sensor.ID).Warn("Failed to fetch telemetry")
				// Обновляем статус на error
				sensor.Status = "error"
				sensor.LastSeen = time.Now()
				s.repo.Update(ctx, sensor)
				return
			}

			// Обновляем данные модуля
			sensor.CurrentTemp = tr.Value
			sensor.Status = tr.Status
			sensor.LastSeen = time.Now()
			if err := s.repo.Update(ctx, sensor); err != nil {
				logrus.WithError(err).WithField("sensor_id", sensor.ID).Error("Failed to update sensor")
			}

			// Формируем данные для очереди
			telemetryData := map[string]interface{}{
				"sensor_id":    sensor.ID,
				"device_id":    sensor.DeviceID,
				"temperature":  tr.Value,
				"unit":         tr.Unit,
				"location":     tr.Location,
				"status":       tr.Status,
				"type":         "warm",
				"collected_at": time.Now().UTC().Format(time.RFC3339),
				"is_legacy":    sensor.IsLegacy,
			}

			// Отправляем в очередь
			if err := s.publisher.Publish(sensor.ID, telemetryData); err != nil {
				logrus.WithError(err).WithField("sensor_id", sensor.ID).Error("Failed to publish telemetry")
			} else {
				logrus.WithField("sensor_id", sensor.ID).Info("Telemetry published")
			}
		}()
	}

	return nil
}

func (s *WarmService) Register(ctx context.Context, req domain.RegisterRequest) (*domain.WarmSensor, error) {
	id := uuid.New().String()
	now := time.Now()
	sensor := &domain.WarmSensor{
		ID:              id,
		DeviceID:        req.DeviceID,
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
	if err := s.repo.Create(ctx, sensor); err != nil {
		return nil, err
	}
	return sensor, nil
}

func (s *WarmService) GetAll(ctx context.Context, location, status string) ([]*domain.WarmSensor, error) {
	return s.repo.GetAll(ctx, location, status)
}

func (s *WarmService) GetByID(ctx context.Context, id string) (*domain.WarmSensor, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *WarmService) GetByDeviceID(ctx context.Context, deviceID string) ([]*domain.WarmSensor, error) {
	return s.repo.GetByDeviceID(ctx, deviceID)
}

func (s *WarmService) SetTemperature(ctx context.Context, id string, target float64) (*domain.WarmSensor, error) {
	sensor, err := s.repo.GetByID(ctx, id)
	if err != nil || sensor == nil {
		return nil, fmt.Errorf("sensor not found")
	}
	// Здесь можно отправить команду на устройство или монолит
	// Для примера — просто обновляем в БД
	sensor.TargetTemp = target
	sensor.LastSeen = time.Now()
	if err := s.repo.Update(ctx, sensor); err != nil {
		return nil, err
	}
	return sensor, nil
}

func (s *WarmService) SetTemperatureByDeviceId(ctx context.Context, device_id string, target float64) ([]*domain.WarmSensor, error) {
	sensors, err := s.repo.GetByDeviceID(ctx, device_id)
	if err != nil {
		return nil, err
	}
	for _, v := range sensors {
		s.SetTemperature(ctx, v.ID, target)
	}
	return sensors, nil
}

func (s *WarmService) Toggle(ctx context.Context, id string, on bool) (*domain.WarmSensor, error) {
	sensor, err := s.repo.GetByID(ctx, id)
	if err != nil || sensor == nil {
		return nil, fmt.Errorf("sensor not found")
	}
	sensor.IsOn = on
	sensor.LastSeen = time.Now()
	if err := s.repo.Update(ctx, sensor); err != nil {
		return nil, err
	}
	return sensor, nil
}

func (s *WarmService) ToggleByDeviceId(ctx context.Context, device_id string, on bool) ([]*domain.WarmSensor, error) {
	sensors, err := s.repo.GetByDeviceID(ctx, device_id)
	if err != nil {
		return nil, err
	}
	for _, v := range sensors {
		s.Toggle(ctx, v.ID, on)
	}
	return sensors, nil
}

func (s *WarmService) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

// Сбор телеметрии для одного модуля
func (s *WarmService) CollectTelemetryForModule(ctx context.Context, sensor *domain.WarmSensor) error {
	var tr *connector.TemperatureResponse
	var err error

	if sensor.IsLegacy {
		tr, err = connector.FetchTelemetryFromMonolith(ctx, s.monolithURL, sensor.ID)
	} else {
		tr, err = connector.FetchTelemetryFromSensor(ctx, sensor.URL, sensor.ID)
	}

	if err != nil {
		// Обновляем статус в offline/error
		sensor.Status = "error"
		sensor.LastSeen = time.Now()
		s.repo.Update(ctx, sensor)
		return err
	}

	// Обновляем данные модуля
	sensor.CurrentTemp = tr.Value
	sensor.Status = tr.Status
	sensor.LastSeen = time.Now()
	s.repo.Update(ctx, sensor)

	// Отправляем в очередь
	telemetryData := map[string]interface{}{
		"sensor_id":    sensor.ID,
		"temperature":  tr.Value,
		"unit":         tr.Unit,
		"timestamp":    tr.Timestamp,
		"location":     tr.Location,
		"status":       tr.Status,
		"type":         "warm",
		"collected_at": time.Now().UTC().Format(time.RFC3339),
		"is_legacy":    sensor.IsLegacy,
	}

	return s.publisher.Publish(sensor.ID, telemetryData)
}
