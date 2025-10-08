package http

import (
	"math/rand"
	"net/http"
	"telemetry-service/internal/service/telemetry"
	"time"

	"github.com/gin-gonic/gin"
)

var rng = rand.New(rand.NewSource(time.Now().UnixNano()))

type Handler struct {
	telemetryService *telemetry.TelemetryService
}

func NewHandler(svc *telemetry.TelemetryService) *Handler {
	return &Handler{
		telemetryService: svc,
	}
}

func (h *Handler) InitRoutes(srv *gin.Engine) {
	srv.GET("health", h.ManageHealth)

	api := srv.Group("/v1")
	{
		api.GET("/telemetry/raw", h.GetRaw)
		api.GET("/telemetry/last", h.GetLast)
		api.GET("/telemetry/aggregated", h.GetAggregated)
	}
}

func (h *Handler) ManageHealth(c *gin.Context) {
	c.AbortWithStatus(http.StatusOK)
}

func (h *Handler) GetRaw(c *gin.Context) {
	ctx := c.Request.Context()

	fromStr := c.Query("from")
	toStr := c.Query("to")
	if fromStr == "" || toStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "from and to are required"})
		return
	}

	from, err := time.Parse(time.RFC3339, fromStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid from format: " + err.Error()})
		return
	}
	to, err := time.Parse(time.RFC3339, toStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid to format: " + err.Error()})
		return
	}

	limit := 100
	if l := c.DefaultQuery("limit", "100"); l != "" {
		// парсинг и ограничение 1–1000
	}

	filters := map[string]string{
		"sensor_id": c.Query("sensor_id"),
		"device_id": c.Query("device_id"),
		"type":      c.Query("type"),
	}

	points, err := h.telemetryService.GetRaw(ctx, from, to, filters, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, points)
}

func (h *Handler) GetLast(c *gin.Context) {
	ctx := c.Request.Context()

	sensorID := c.Query("sensor_id")
	if sensorID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "sensor_id is required"})
		return
	}

	point, err := h.telemetryService.GetLastBySensorID(ctx, sensorID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "error: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, point)
}

// Заглушка для агрегации
func (h *Handler) GetAggregated(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "aggregation not implemented"})
}
