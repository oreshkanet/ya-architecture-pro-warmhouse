package http

import (
	"device-service/internal/domain"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (h *Handler) getLocations(c *gin.Context) {
	userID, exists := c.GetQuery("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	uid, err := uuid.Parse(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}
	locations, err := h.locationService.GetAll(c.Request.Context(), uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch locations"})
		return
	}
	c.JSON(http.StatusOK, locations)
}

func (h *Handler) createLocation(c *gin.Context) {
	var req domain.Location
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if req.HouseID == uuid.Nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "house_id is required"})
		return
	}
	if strings.TrimSpace(req.Name) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}
	switch req.Type {
	case domain.LocationTypeRoom, domain.LocationTypeOutdoor, domain.LocationTypeFloor:
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid location type"})
		return
	}

	location, err := h.locationService.Create(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create location: " + err.Error()})
		return
	}
	c.JSON(http.StatusCreated, location)
}

func (h *Handler) updateLocation(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid location id"})
		return
	}

	var req domain.Location
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if req.ID != id {
		c.JSON(http.StatusBadRequest, gin.H{"error": "location id in body must match URL"})
		return
	}
	if req.HouseID == uuid.Nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "house_id is required"})
		return
	}
	if strings.TrimSpace(req.Name) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}
	switch req.Type {
	case domain.LocationTypeRoom, domain.LocationTypeOutdoor, domain.LocationTypeFloor:
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid location type"})
		return
	}

	location, err := h.locationService.Update(c.Request.Context(), id, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update location"})
		return
	}
	if location == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "location not found"})
		return
	}
	c.JSON(http.StatusOK, location)
}

func (h *Handler) deleteLocation(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid location id"})
		return
	}

	err = h.locationService.Delete(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete location"})
		return
	}
	c.Status(http.StatusNoContent)
}
