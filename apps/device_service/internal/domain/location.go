package domain

import "github.com/google/uuid"

type LocationType string

const (
	LocationTypeRoom    LocationType = "room"
	LocationTypeOutdoor LocationType = "outdoor"
	LocationTypeFloor   LocationType = "floor"
)

type Location struct {
	ID               uuid.UUID    `json:"id"`
	UserID           uuid.UUID    `json:"user_id"`
	HouseID          uuid.UUID    `json:"house_id"`
	Name             string       `json:"name"`
	Type             LocationType `json:"type"`
	ParentLocationID *uuid.UUID   `json:"parent_location_id,omitempty"`
}
