package http

import (
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

var rng = rand.New(rand.NewSource(time.Now().UnixNano()))

type Handler struct {
}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) InitRoutes(srv *gin.Engine) {
	srv.GET("health", h.ManageHealth)
	srv.GET("temperature", h.GetTemperature)
}

func (h *Handler) ManageHealth(c *gin.Context) {
	c.AbortWithStatus(http.StatusOK)
}

func (h *Handler) GetTemperature(c *gin.Context) {

	location := c.Query("location")
	sensorId := c.Query("sensorId")

	// If no location is provided, use a default based on sensor ID
	if location == "" {
		switch sensorId {
		case "1":
			location = "Living Room"
		case "2":
			location = "Bedroom"
		case "3":
			location = "Kitchen"
		default:
			location = "Unknown"
		}
	}

	// If no sensor ID is provided, generate one based on location
	if sensorId == "" {
		switch location {
		case "Living Room":
			sensorId = "1"
		case "Bedroom":
			sensorId = "2"
		case "Kitchen":
			sensorId = "3"
		default:
			sensorId = "0"
		}
	}

	resp := &TemperatureResponse{
		Value:       getRandomTemperature(),
		Unit:        "C",
		Timestamp:   time.Now(),
		Location:    location,
		Status:      "Online",
		SensorID:    sensorId,
		SensorType:  "Warm",
		Description: fmt.Sprintf("Warm sensor - %s", &location),
	}

	c.AbortWithStatusJSON(http.StatusOK, resp)
}

func getRandomTemperature() float64 {
	value := rng.Intn(351) + 100
	return float64(value) / 10.0
}
