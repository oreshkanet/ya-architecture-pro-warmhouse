package domain

import "time"

type WarmModule struct {
	ID              string    `json:"id"`
	Name            string    `json:"name"`
	Location        string    `json:"location"`
	SerialNumber    string    `json:"serial_number"`
	IsOn            bool      `json:"is_on"`
	CurrentTemp     float64   `json:"current_temperature"`
	TargetTemp      float64   `json:"target_temperature"`
	Status          string    `json:"status"`
	FirmwareVersion *string   `json:"firmware_version"`
	LastSeen        time.Time `json:"last_seen"`
	IsLegacy        bool      `json:"is_legacy"`
	URL             string    `json:"url"`
}

type RegisterRequest struct {
	SerialNumber    string  `json:"serial_number"`
	Name            string  `json:"name"`
	Location        string  `json:"location"`
	FirmwareVersion *string `json:"firmware_version"`
	URL             string  `json:"url"`
	IsLegacy        bool    `json:"is_legacy"`
}

type SetTempRequest struct {
	TargetTemp float64 `json:"target_temperature"`
}
