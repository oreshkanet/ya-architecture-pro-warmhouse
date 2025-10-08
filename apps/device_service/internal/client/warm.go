package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type WarmServiceClient struct {
	BaseURL    string
	HTTPClient *http.Client
}

func NewWarmServiceClient(baseURL string) *WarmServiceClient {
	return &WarmServiceClient{
		BaseURL:    baseURL,
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	}
}

// SendCommand отправляет команду устройству в WarmService
func (c *WarmServiceClient) SendCommand(ctx context.Context, deviceID string, cmd string) error {
	url := fmt.Sprintf("%s/v1/warm/command", c.BaseURL)

	payload, err := json.Marshal(&CommandRequest{
		DeviceId: deviceID,
		Command:  cmd,
	})
	if err != nil {
		return fmt.Errorf("marshal command: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("http request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("warm service returned %d", resp.StatusCode)
	}

	return nil
}

// SendCommand отправляет команду устройству в WarmService
func (c *WarmServiceClient) SetTemperature(ctx context.Context, deviceID string, t float64) error {
	url := fmt.Sprintf("%s/v1/warm/values", c.BaseURL)

	payload, err := json.Marshal(&SetValuesRequest{
		DeviceId:   deviceID,
		TargetTemp: t,
	})
	if err != nil {
		return fmt.Errorf("marshal command: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("http request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("warm service returned %d", resp.StatusCode)
	}

	return nil
}
