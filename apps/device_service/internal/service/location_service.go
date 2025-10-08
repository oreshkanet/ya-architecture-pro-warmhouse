package service

import (
	"context"

	"device-service/internal/domain"
	"device-service/internal/repository"

	"github.com/google/uuid"
)

type LocationService struct {
	locationRepo *repository.LocationRepository
	deviceRepo   *repository.DeviceRepository
}

func NewLocationService(lr *repository.LocationRepository, dr *repository.DeviceRepository) *LocationService {
	return &LocationService{
		locationRepo: lr,
		deviceRepo:   dr,
	}
}

// GetAll возвращает все локации
func (s *LocationService) GetAll(ctx context.Context, userID uuid.UUID) ([]domain.Location, error) {
	return s.locationRepo.GetAll(ctx, userID)
}

// Create создаёт новую локацию
func (s *LocationService) Create(ctx context.Context, loc domain.Location) (*domain.Location, error) {
	// Генерируем ID, если не задан (в OpenAPI он required, но клиент может не прислать)
	if loc.ID == uuid.Nil {
		loc.ID = uuid.New()
	}

	// Проверка house_id — должен быть задан
	if loc.HouseID == uuid.Nil {
		return nil, ErrInvalidInput
	}

	// Проверка parent_location_id (если задан — должен существовать)
	if loc.ParentLocationID != nil {
		exists, err := s.locationRepo.Exists(ctx, *loc.ParentLocationID)
		if err != nil {
			return nil, err
		}
		if !exists {
			return nil, ErrParentLocationNotFound
		}
	}

	return s.locationRepo.Create(ctx, loc)
}

// Update обновляет локацию
func (s *LocationService) Update(ctx context.Context, id uuid.UUID, loc domain.Location) (*domain.Location, error) {
	if id == uuid.Nil {
		return nil, ErrInvalidInput
	}
	if loc.HouseID == uuid.Nil {
		return nil, ErrInvalidInput
	}

	// Проверка существования самой локации
	existing, err := s.locationRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, nil // не найдено
	}

	// Проверка parent_location_id
	if loc.ParentLocationID != nil {
		exists, err := s.locationRepo.Exists(ctx, *loc.ParentLocationID)
		if err != nil {
			return nil, err
		}
		if !exists {
			return nil, ErrParentLocationNotFound
		}
	}

	// Обновляем с тем же ID
	loc.ID = id
	return s.locationRepo.Update(ctx, id, loc)
}

// Delete удаляет локацию
func (s *LocationService) Delete(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return ErrInvalidInput
	}

	hasDevices, err := s.deviceRepo.HasDevicesInLocation(ctx, id)
	if err != nil {
		return err
	}
	if hasDevices {
		return ErrLocationInUse
	}

	return s.locationRepo.Delete(ctx, id)
}
