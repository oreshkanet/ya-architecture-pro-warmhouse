package repository

import (
	"context"
	"database/sql"
	"device-service/internal/domain"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DeviceRepository struct {
	db *pgxpool.Pool
}

func NewDeviceRepository(db *pgxpool.Pool) *DeviceRepository {
	return &DeviceRepository{db: db}
}

func (r *DeviceRepository) Create(ctx context.Context, d *domain.Device) error {
	query := `
		INSERT INTO devices (user_id, name, device_type, serial_number, location_id, status, firmware_version, connected_at, last_seen)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, connected_at, last_seen`
	return r.db.QueryRow(ctx, query,
		d.UserID, d.Name, d.DeviceType, d.SerialNumber, d.LocationID,
		d.Status, d.FirmwareVersion, d.ConnectedAt, d.LastSeen,
	).Scan(&d.ID, &d.ConnectedAt, &d.LastSeen)
}

func (r *DeviceRepository) GetByUser(ctx context.Context, userID uuid.UUID, filters map[string]interface{}) ([]domain.Device, error) {
	query := "SELECT id, user_id, name, device_type, serial_number, location_id, status, firmware_version, connected_at, last_seen FROM devices WHERE user_id = $1"
	args := []interface{}{userID}
	argID := 2

	if locID, ok := filters["location_id"].(uuid.UUID); ok {
		query += " AND location_id = $" + fmt.Sprint(argID)
		args = append(args, locID)
		argID++
	}
	if devType, ok := filters["type"].(string); ok {
		query += " AND device_type = $" + fmt.Sprint(argID)
		args = append(args, devType)
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var devices []domain.Device
	for rows.Next() {
		var d domain.Device
		var serial, fw *string
		var connectedAt, lastSeen *time.Time
		err := rows.Scan(&d.ID, &d.UserID, &d.Name, &d.DeviceType, &serial, &d.LocationID,
			&d.Status, &fw, &connectedAt, &lastSeen)
		if err != nil {
			return nil, err
		}
		d.SerialNumber = serial
		d.FirmwareVersion = fw
		d.ConnectedAt = connectedAt
		d.LastSeen = lastSeen
		devices = append(devices, d)
	}
	return devices, rows.Err()
}

func (r *DeviceRepository) HasDevicesInLocation(ctx context.Context, locationId uuid.UUID) (bool, error) {
	query := "SELECT id, user_id, name, device_type, serial_number, location_id, status, firmware_version, connected_at, last_seen FROM devices WHERE location_id = $1"
	row := r.db.QueryRow(ctx, query, locationId)

	var id *uuid.UUIDs
	err := row.Scan(&id)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (r *DeviceRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Device, error) {
	query := "SELECT id, user_id, name, device_type, serial_number, location_id, status, firmware_version, connected_at, last_seen FROM devices WHERE id = $1"
	row := r.db.QueryRow(ctx, query, id)

	var d domain.Device
	var serial, fw *string
	var connectedAt, lastSeen *time.Time
	err := row.Scan(&d.ID, &d.UserID, &d.Name, &d.DeviceType, &serial, &d.LocationID,
		&d.Status, &fw, &connectedAt, &lastSeen)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	d.SerialNumber = serial
	d.FirmwareVersion = fw
	d.ConnectedAt = connectedAt
	d.LastSeen = lastSeen
	return &d, nil
}

func (r *DeviceRepository) Update(ctx context.Context, id uuid.UUID, updates map[string]interface{}) (*domain.Device, error) {
	setParts := []string{}
	args := []interface{}{id}
	argID := 2

	if name, ok := updates["name"]; ok {
		setParts = append(setParts, "name = $"+fmt.Sprint(argID))
		args = append(args, name)
		argID++
	}
	if locID, ok := updates["location_id"].(uuid.UUID); ok {
		setParts = append(setParts, "location_id = $"+fmt.Sprint(argID))
		args = append(args, locID)
		argID++
	}
	if status, ok := updates["status"].(domain.DeviceStatus); ok {
		setParts = append(setParts, "status = $"+fmt.Sprint(argID))
		args = append(args, status)
		argID++
	}

	if len(setParts) == 0 {
		return r.GetByID(ctx, id)
	}

	query := "UPDATE devices SET " + strings.Join(setParts, ", ") + " WHERE id = $1 RETURNING id, user_id, name, device_type, serial_number, location_id, status, firmware_version, connected_at, last_seen"
	row := r.db.QueryRow(ctx, query, args...)

	var d domain.Device
	var serial, fw *string
	var connectedAt, lastSeen *time.Time
	err := row.Scan(&d.ID, &d.UserID, &d.Name, &d.DeviceType, &serial, &d.LocationID,
		&d.Status, &fw, &connectedAt, &lastSeen)
	if err != nil {
		return nil, err
	}
	d.SerialNumber = serial
	d.FirmwareVersion = fw
	d.ConnectedAt = connectedAt
	d.LastSeen = lastSeen
	return &d, nil
}

func (r *DeviceRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, "DELETE FROM devices WHERE id = $1", id)
	return err
}
