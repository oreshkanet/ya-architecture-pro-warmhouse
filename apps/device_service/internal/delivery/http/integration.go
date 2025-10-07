package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (h *Handler) getIntegrationDevices(c *gin.Context) {

	userID, err := uuid.Parse(c.Param("user_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	devices, err := h.deviceService.GetDevices(c.Request.Context(), userID, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch devices for integration"})
		return
	}

	type IntegrationDevice struct {
		ID         uuid.UUID `json:"id"`
		Name       string    `json:"name"`
		DeviceType string    `json:"device_type"`
		LocationID uuid.UUID `json:"location_id"`
	}

	result := make([]IntegrationDevice, len(devices))
	for i, d := range devices {
		result[i] = IntegrationDevice{
			ID:         d.ID,
			Name:       d.Name,
			DeviceType: string(d.DeviceType),
			LocationID: d.LocationID,
		}
	}
	c.JSON(http.StatusOK, result)
}
