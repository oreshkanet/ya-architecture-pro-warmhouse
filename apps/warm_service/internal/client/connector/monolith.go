package connector

import (
	"context"
)

// Аналогично, но вызывает SmartHome Monolith
func FetchTelemetryFromMonolith(ctx context.Context, url string, sensorID string) (*TemperatureResponse, error) {
	return FetchTelemetryFromSensor(ctx, url, sensorID)
}
