package service

import (
	"context"
	"device-service/internal/client"
	"device-service/internal/domain"
	"device-service/internal/repository"
	"fmt"

	"github.com/google/uuid"
)

type DeviceService struct {
	locationRepo *repository.LocationRepository
	deviceRepo   *repository.DeviceRepository
	warmClient   *client.WarmServiceClient
}

func NewDeviceService(
	lr *repository.LocationRepository,
	dr *repository.DeviceRepository,
	wc *client.WarmServiceClient,
) *DeviceService {
	return &DeviceService{
		locationRepo: lr,
		deviceRepo:   dr,
		warmClient:   wc,
	}
}

func (s *DeviceService) SendCommandToDevice(ctx context.Context, userID uuid.UUID, deviceID uuid.UUID, cmd string) error {
	device, err := s.deviceRepo.GetByID(ctx, deviceID)
	if err != nil {
		return fmt.Errorf("get device: %w", err)
	}
	if device == nil {
		return ErrDeviceNotFound
	}

	if device.UserID != userID {
		return ErrDeviceNotFound // для безопасности — не раскрываем существование
	}

	switch device.DeviceType {
	case domain.DeviceTypeWarm:
		return s.warmClient.SendCommand(ctx, deviceID.String(), cmd)
	case domain.DeviceTypeLight:
		// return s.lightClient.SendCommand(...)
	case domain.DeviceTypeDoor:
		// ...
	case domain.DeviceTypeVideo:
		// ...
	case domain.DeviceTypeUniversal:
		// ...
	default:
		return ErrUnsupportedDeviceType
	}

	return nil
}

func (s *DeviceService) CreateDevice(ctx context.Context, userID uuid.UUID, req CreateDeviceRequest) (*domain.Device, error) {
	device := &domain.Device{
		UserID:       userID,
		Name:         req.Name,
		DeviceType:   req.DeviceType,
		SerialNumber: req.SerialNumber,
		LocationID:   req.LocationID,
		Status:       domain.StatusOffline,
	}

	exists, err := s.locationRepo.Exists(ctx, req.LocationID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrLocationNotFound
	}

	err = s.deviceRepo.Create(ctx, device)
	return device, err
}

func (s *DeviceService) GetDevices(ctx context.Context, userID uuid.UUID, filters map[string]interface{}) ([]domain.Device, error) {
	return s.deviceRepo.GetByUser(ctx, userID, filters)
}

func (s *DeviceService) GetDeviceByID(ctx context.Context, id uuid.UUID) (*domain.Device, error) {
	return s.deviceRepo.GetByID(ctx, id)
}

func (s *DeviceService) UpdateDevice(ctx context.Context, id uuid.UUID, updates map[string]interface{}) (*domain.Device, error) {
	return s.deviceRepo.Update(ctx, id, updates)
}

func (s *DeviceService) DeleteDevice(ctx context.Context, id uuid.UUID) error {
	return s.deviceRepo.Delete(ctx, id)
}

type CreateDeviceRequest struct {
	Name         string
	DeviceType   domain.DeviceType
	SerialNumber *string
	LocationID   uuid.UUID
}
