package http

import (
	"device-service/internal/domain"
	"device-service/internal/service"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (h *Handler) sendDeviceCommand(c *gin.Context) {
	var req CommandRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if req.UserId == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	deviceID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid device id"})
		return
	}

	err = h.deviceService.SendCommandToDevice(c.Request.Context(), req.UserId, deviceID, req.Command)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrDeviceNotFound):
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "device_not_found",
				"message": "Устройство не найдено",
			})
		case errors.Is(err, service.ErrUnsupportedDeviceType):
			c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported device type"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to send command: " + err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "command sent"})
}

func (h *Handler) setDeviceValues(c *gin.Context) {
	var req SetValuesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if req.UserId == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	deviceID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid device id"})
		return
	}

	err = h.deviceService.SetValuesToDevice(c.Request.Context(), req.UserId, deviceID, req.Values)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrDeviceNotFound):
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "device_not_found",
				"message": "Устройство не найдено",
			})
		case errors.Is(err, service.ErrUnsupportedDeviceType):
			c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported device type"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to send command: " + err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "command sent"})
}

func (h *Handler) getDevices(c *gin.Context) {
	userID, exists := c.GetQuery("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	filters := make(map[string]interface{})
	if locIDStr := c.Query("location_id"); locIDStr != "" {
		if locID, err := uuid.Parse(locIDStr); err == nil {
			filters["location_id"] = locID
		}
	}
	if devType := c.Query("type"); devType != "" {
		switch devType {
		case "warm", "light", "door", "video", "universal":
			filters["type"] = devType
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid device type"})
			return
		}
	}

	uid, err := uuid.Parse(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}
	devices, err := h.deviceService.GetDevices(c.Request.Context(), uid, filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch devices: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, devices)
}

func (h *Handler) createDevice(c *gin.Context) {
	var req struct {
		Name         string            `json:"name"`
		UserID       uuid.UUID         `json:"user_id"`
		DeviceType   domain.DeviceType `json:"device_type"`
		SerialNumber *string           `json:"serial_number,omitempty"`
		LocationID   uuid.UUID         `json:"location_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if strings.TrimSpace(req.Name) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}
	if req.LocationID == uuid.Nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "location_id is required"})
		return
	}

	device, err := h.deviceService.CreateDevice(c.Request.Context(), req.UserID, service.CreateDeviceRequest{
		Name:         req.Name,
		DeviceType:   req.DeviceType,
		SerialNumber: req.SerialNumber,
		LocationID:   req.LocationID,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create device: " + err.Error()})
		return
	}
	c.JSON(http.StatusCreated, device)
}

func (h *Handler) getDevice(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid device id"})
		return
	}

	device, err := h.deviceService.GetDeviceByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}
	if device == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "device_not_found",
			"message": "Устройство с таким ID не найдено",
			"status":  404,
		})
		return
	}

	userID := c.MustGet("user_id").(uuid.UUID)
	if device.UserID != userID {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "device_not_found",
			"message": "Устройство с таким ID не найдено",
			"status":  404,
		})
		return
	}

	c.JSON(http.StatusOK, device)
}

func (h *Handler) updateDevice(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid device id"})
		return
	}

	var req struct {
		Name       *string              `json:"name,omitempty"`
		LocationID *uuid.UUID           `json:"location_id,omitempty"`
		Status     *domain.DeviceStatus `json:"status,omitempty"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	updates := make(map[string]interface{})
	if req.Name != nil && *req.Name != "" {
		updates["name"] = *req.Name
	}
	if req.LocationID != nil {
		updates["location_id"] = *req.LocationID
	}
	if req.Status != nil {
		switch *req.Status {
		case domain.StatusOnline, domain.StatusOffline, domain.StatusDisabled:
			updates["status"] = *req.Status
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid status value"})
			return
		}
	}

	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no fields to update"})
		return
	}

	device, err := h.deviceService.UpdateDevice(c.Request.Context(), id, updates)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update device"})
		return
	}
	if device == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "device_not_found",
			"message": "Устройство с таким ID не найдено",
			"status":  404,
		})
		return
	}

	userID := c.MustGet("user_id").(uuid.UUID)
	if device.UserID != userID {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "device_not_found",
			"message": "Устройство с таким ID не найдено",
			"status":  404,
		})
		return
	}

	c.JSON(http.StatusOK, device)
}

func (h *Handler) deleteDevice(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid device id"})
		return
	}

	device, err := h.deviceService.GetDeviceByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}

	userID := c.MustGet("user_id").(uuid.UUID)
	if device != nil && device.UserID != userID {
		c.Status(http.StatusNoContent)
		return
	}

	err = h.deviceService.DeleteDevice(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete device"})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) getDeviceStatus(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid device id"})
		return
	}

	device, err := h.deviceService.GetDeviceByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}
	if device == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "device not found"})
		return
	}

	userID := c.MustGet("user_id").(uuid.UUID)
	if device.UserID != userID {
		c.JSON(http.StatusNotFound, gin.H{"error": "device not found"})
		return
	}

	status := DeviceStatusResponse{
		DeviceID: id,
		IsOnline: device.Status == domain.StatusOnline,
		LastSeen: device.LastSeen,
	}

	if !status.IsOnline {
		status.SensorData = nil
	}

	c.JSON(http.StatusOK, status)
}
