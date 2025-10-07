package http

import (
	"errors"
	"math/rand"
	"net/http"
	"time"
	"warm-service/internal/domain"
	"warm-service/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

var rng = rand.New(rand.NewSource(time.Now().UnixNano()))

type Handler struct {
	warmService *service.WarmService
}

func NewHandler(ws *service.WarmService) *Handler {
	return &Handler{
		warmService: ws,
	}
}

func (h *Handler) InitRoutes(srv *gin.Engine) {
	srv.GET("health", h.ManageHealth)

	api := srv.Group("/v1")
	{
		api.GET("/warm", h.List)
		api.POST("/warm", h.Register)
		api.DELETE("/warm/:id", h.Delete)
		api.GET("/warm/:id", h.GetByID)
		api.POST("/warm/command", h.RunCommand)
		api.POST("/warm/values", h.SetValues)
		api.PATCH("/warm/:id", h.SetTemp)
		api.POST("/warm/:id/on", h.TurnOn)
		api.POST("/warm/:id/off", h.TurnOff)
	}
}

func (h *Handler) ManageHealth(c *gin.Context) {
	c.AbortWithStatus(http.StatusOK)
}

func (h *Handler) Register(c *gin.Context) {
	var req domain.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_json"})
		return
	}
	module, err := h.warmService.Register(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, module)
}

func (h *Handler) List(c *gin.Context) {
	deviceId := c.Query("device_id")
	location := c.Query("location")
	status := c.Query("status")

	var err error
	var modules []*domain.WarmSensor
	if deviceId != "" {
		modules, err = h.warmService.GetByDeviceID(c.Request.Context(), deviceId)
	} else {
		modules, err = h.warmService.GetAll(c.Request.Context(), location, status)
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, modules)
}

func (h *Handler) GetByID(c *gin.Context) {
	id := c.Param("id")
	if _, err := uuid.Parse(id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_id"})
		return
	}
	module, err := h.warmService.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if module == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not_found"})
		return
	}
	c.JSON(http.StatusOK, module)
}

func (h *Handler) RunCommand(c *gin.Context) {
	var req struct {
		Command  string `json:"command"`
		DeviceId string `json:"device_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	var on bool
	var err error
	switch req.Command {
	case "on":
		on = true
	case "off":
		on = false
	default:
		err = errors.New("unknown command")
	}
	if err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}

	module, err := h.warmService.ToggleByDeviceId(c.Request.Context(), req.DeviceId, on)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not_found"})
		return
	}
	c.JSON(http.StatusOK, module)
}

func (h *Handler) SetValues(c *gin.Context) {
	var req struct {
		DeviceId   string  `json:"device_id"`
		TargetTemp float64 `json:"target_temperature"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if req.TargetTemp < 5 || req.TargetTemp > 30 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "temp_out_of_range"})
		return
	}
	sensors, err := h.warmService.SetTemperatureByDeviceId(c.Request.Context(), req.DeviceId, req.TargetTemp)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not_found"})
		return
	}
	c.JSON(http.StatusOK, sensors)
}

func (h *Handler) SetTemp(c *gin.Context) {
	id := c.Param("id")
	var req domain.SetTempRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_json"})
		return
	}
	if req.TargetTemp < 5 || req.TargetTemp > 30 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "temp_out_of_range"})
		return
	}
	module, err := h.warmService.SetTemperature(c.Request.Context(), id, req.TargetTemp)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not_found"})
		return
	}
	c.JSON(http.StatusOK, module)
}

func (h *Handler) TurnOn(c *gin.Context) {
	id := c.Param("id")
	module, err := h.warmService.Toggle(c.Request.Context(), id, true)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not_found"})
		return
	}
	c.JSON(http.StatusOK, module)
}

func (h *Handler) TurnOff(c *gin.Context) {
	id := c.Param("id")
	module, err := h.warmService.Toggle(c.Request.Context(), id, false)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not_found"})
		return
	}
	c.JSON(http.StatusOK, module)
}

func (h *Handler) Delete(c *gin.Context) {
	id := c.Param("id")
	err := h.warmService.Delete(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}
