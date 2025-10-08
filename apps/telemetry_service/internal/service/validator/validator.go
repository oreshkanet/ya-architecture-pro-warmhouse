package validator

import (
	"fmt"
	"telemetry-service/internal/domain"
)

func Validate(point *domain.TelemetryPoint) error {
	if point.SensorID == "" {
		return fmt.Errorf("sensor_id is required")
	}
	if point.Type == "" {
		return fmt.Errorf("type is required")
	}
	if point.Timestamp.IsZero() {
		return fmt.Errorf("timestamp is required")
	}
	// Добавить валидацию по типу: например, temperature → число
	return nil
}
