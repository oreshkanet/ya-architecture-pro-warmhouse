package domain

import "time"

type TelemetryPoint struct {
	ID        string      `json:"id,omitempty" clickhouse:"id"`
	SensorID  string      `json:"sensor_id" clickhouse:"sensor_id"`
	DeviceID  string      `json:"device_id" clickhouse:"device_id"`
	Type      string      `json:"type" clickhouse:"type"`
	Unit      string      `json:"unit" clickhouse:"unit"`
	Value     interface{} `json:"value" clickhouse:"value"`
	Timestamp time.Time   `json:"timestamp" clickhouse:"timestamp"`
}
