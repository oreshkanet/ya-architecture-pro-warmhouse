package client

type SetValuesRequest struct {
	DeviceId   string  `json:"device_id"`
	TargetTemp float64 `json:"target_temperature"`
}

type CommandRequest struct {
	Command  string `json:"command"`
	DeviceId string `json:"device_id"`
}
