package http

import (
	"device-service/internal/service"
	"math/rand"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

var rng = rand.New(rand.NewSource(time.Now().UnixNano()))

type Handler struct {
	deviceService   *service.DeviceService
	locationService *service.LocationService
}

func NewHandler(
	deviceService *service.DeviceService,
	locationService *service.LocationService,
) *Handler {
	return &Handler{
		deviceService:   deviceService,
		locationService: locationService,
	}
}

func (h *Handler) InitRoutes(srv *gin.Engine) {
	srv.GET("health", h.ManageHealth)

	api := srv.Group("/v1")
	{
		api.GET("/devices", h.getDevices)
		api.POST("/devices", h.createDevice)
		api.GET("/devices/:id", h.getDevice)
		api.POST("/devices/:id/command", h.sendDeviceCommand)
		api.POST("/devices/:id/values", h.setDeviceValues)
		api.PATCH("/devices/:id", h.updateDevice)
		api.DELETE("/devices/:id", h.deleteDevice)
		api.GET("/devices/:id/status", h.getDeviceStatus)

		api.GET("/locations", h.getLocations)
		api.POST("/locations", h.createLocation)
		api.PUT("/locations/:id", h.updateLocation)
		api.DELETE("/locations/:id", h.deleteLocation)

		api.GET("/integration/devices", h.getIntegrationDevices)
	}
}

func (h *Handler) ManageHealth(c *gin.Context) {
	c.AbortWithStatus(http.StatusOK)
}
