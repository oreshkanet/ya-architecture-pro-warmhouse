package connector

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type TemperatureResponse struct {
	Value       float64 `json:"value"`
	Unit        string  `json:"unit"`
	Timestamp   string  `json:"timestamp"`
	Location    string  `json:"location"`
	Status      string  `json:"status"`
	SensorID    string  `json:"sensorId"`
	SensorType  string  `json:"sensorType"`
	Description string  `json:"description"`
}

func FetchTelemetryFromSensor(ctx context.Context, url, sensorID string) (*TemperatureResponse, error) {
	req, _ := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/temperature?sensorId=%s", url, sensorID), nil)
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("non-200 status: %d", resp.StatusCode)
	}

	var tr TemperatureResponse
	if err := json.NewDecoder(resp.Body).Decode(&tr); err != nil {
		return nil, err
	}
	return &tr, nil
}
