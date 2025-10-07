package repository

import (
	"context"
	"database/sql"
	"fmt"
	"warm-service/internal/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

type WarmRepo struct {
	db *pgxpool.Pool
}

func NewWarmRepo(db *pgxpool.Pool) *WarmRepo {
	return &WarmRepo{db: db}
}

func (r *WarmRepo) Create(ctx context.Context, m *domain.WarmSensor) error {
	query := `
		INSERT INTO warm_sensors (
			id, device_id, name, location, serial_number, is_on, current_temperature,
			target_temperature, status, firmware_version, last_seen, is_legacy, url
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`
	_, err := r.db.Exec(ctx, query,
		m.ID, m.DeviceID, m.Name, m.Location, m.SerialNumber, m.IsOn, m.CurrentTemp,
		m.TargetTemp, m.Status, m.FirmwareVersion, m.LastSeen, m.IsLegacy, m.URL,
	)
	return err
}

func (r *WarmRepo) GetAll(ctx context.Context, location, status string) ([]*domain.WarmSensor, error) {
	query := `SELECT id, device_id, name, location, serial_number, is_on, current_temperature,
		target_temperature, status, firmware_version, last_seen, is_legacy, url
		FROM warm_sensors WHERE true`
	args := []interface{}{}
	i := 1

	if location != "" {
		query += fmt.Sprintf(" AND location = $%d", i)
		args = append(args, location)
		i++
	}
	if status != "" {
		query += fmt.Sprintf(" AND status = $%d", i)
		args = append(args, status)
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var modules []*domain.WarmSensor
	for rows.Next() {
		var m domain.WarmSensor
		var fw *string
		err := rows.Scan(
			&m.ID, &m.DeviceID, &m.Name, &m.Location, &m.SerialNumber, &m.IsOn, &m.CurrentTemp,
			&m.TargetTemp, &m.Status, &fw, &m.LastSeen, &m.IsLegacy, &m.URL,
		)
		if err != nil {
			return nil, err
		}
		m.FirmwareVersion = fw
		modules = append(modules, &m)
	}
	return modules, nil
}

func (r *WarmRepo) GetByID(ctx context.Context, id string) (*domain.WarmSensor, error) {
	query := `SELECT id, device_id, name, location, serial_number, is_on, current_temperature,
		target_temperature, status, firmware_version, last_seen, is_legacy, telemetry_url
		FROM warm_sensors WHERE id = $1`
	var m domain.WarmSensor
	var fw *string
	err := r.db.QueryRow(ctx, query, id).Scan(
		&m.ID, &m.DeviceID, &m.Name, &m.Location, &m.SerialNumber, &m.IsOn, &m.CurrentTemp,
		&m.TargetTemp, &m.Status, &fw, &m.LastSeen, &m.IsLegacy, &m.URL,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	m.FirmwareVersion = fw
	return &m, err
}

func (r *WarmRepo) GetByDeviceID(ctx context.Context, deviceID string) ([]*domain.WarmSensor, error) {
	query := `
		SELECT id, device_id, name, location, serial_number, is_on, current_temperature,
		       target_temperature, status, firmware_version, last_seen, is_legacy, telemetry_url
		FROM warm_modules
		WHERE device_id = $1
	`
	rows, err := r.db.Query(ctx, query, deviceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var modules []*domain.WarmSensor
	for rows.Next() {
		var m domain.WarmSensor
		var fw *string
		err := rows.Scan(
			&m.ID, &m.DeviceID, &m.Name, &m.Location, &m.SerialNumber, &m.IsOn,
			&m.CurrentTemp, &m.TargetTemp, &m.Status, &fw, &m.LastSeen, &m.IsLegacy, &m.URL,
		)
		if err != nil {
			return nil, err
		}
		m.FirmwareVersion = fw
		modules = append(modules, &m)
	}
	return modules, nil
}

func (r *WarmRepo) Update(ctx context.Context, m *domain.WarmSensor) error {
	query := `
		UPDATE warm_sensors SET
			is_on = $2, current_temperature = $3, target_temperature = $4,
			status = $5, last_seen = $6
		WHERE id = $1
	`
	_, err := r.db.Exec(ctx, query,
		m.ID, m.IsOn, m.CurrentTemp, m.TargetTemp, m.Status, m.LastSeen,
	)
	return err
}

func (r *WarmRepo) Delete(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx, "DELETE FROM warm_sensors WHERE id = $1", id)
	return err
}
