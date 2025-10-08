package http

import (
	"time"

	"github.com/google/uuid"
)

type TemperatureResponse struct {
	Value       float64   `json:"value"`
	Unit        string    `json:"unit"`
	Timestamp   time.Time `json:"timestamp"`
	Location    string    `json:"location"`
	Status      string    `json:"status"`
	SensorID    string    `json:"sensor_id"`
	SensorType  string    `json:"sensor_type"`
	Description string    `json:"description"`
}

type SensorValue struct {
	Value any    `json:"value"`
	Unit  string `json:"unit,omitempty"`
}

type CommandRequest struct {
	UserId  uuid.UUID `json:"user_id"`
	Command string    `json:"command"`
}

type SetValuesRequest struct {
	UserId uuid.UUID      `json:"user_id"`
	Values map[string]any `json:"values,omitempty"`
}

type DeviceStatusResponse struct {
	DeviceID   uuid.UUID              `json:"device_id"`
	IsOnline   bool                   `json:"is_online"`
	LastSeen   *time.Time             `json:"last_seen,omitempty"`
	SensorData map[string]SensorValue `json:"sensor_data,omitempty"`
}
