package domain

import (
	"time"

	"github.com/google/uuid"
)

type DeviceType string

const (
	DeviceTypeWarm      DeviceType = "warm"
	DeviceTypeLight     DeviceType = "light"
	DeviceTypeDoor      DeviceType = "door"
	DeviceTypeVideo     DeviceType = "video"
	DeviceTypeUniversal DeviceType = "universal"
)

type DeviceStatus string

const (
	StatusOnline   DeviceStatus = "online"
	StatusOffline  DeviceStatus = "offline"
	StatusDisabled DeviceStatus = "disabled"
)

type Device struct {
	ID              uuid.UUID    `json:"id"`
	UserID          uuid.UUID    `json:"-"`
	Name            string       `json:"name"`
	DeviceType      DeviceType   `json:"device_type"`
	SerialNumber    *string      `json:"serial_number,omitempty"`
	LocationID      uuid.UUID    `json:"location_id"`
	Status          DeviceStatus `json:"status"`
	FirmwareVersion *string      `json:"firmware_version,omitempty"`
	ConnectedAt     *time.Time   `json:"connected_at,omitempty"`
	LastSeen        *time.Time   `json:"last_seen,omitempty"`
}

type DeviceCommand struct {
	Action string                 `json:"action"`
	Params map[string]interface{} `json:"params,omitempty"`
}
