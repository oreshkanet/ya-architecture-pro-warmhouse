package repository

import (
	"context"
	"fmt"
	"telemetry-service/internal/domain"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
)

type ClickHouseRepo struct {
	conn clickhouse.Conn
}

func NewClickHouseRepo(dsn, database string) (*ClickHouseRepo, error) {
	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr:     []string{dsn},
		Protocol: clickhouse.HTTP,
		Auth: clickhouse.Auth{
			Database: database,
		},
	})
	if err != nil {
		return nil, err
	}

	return &ClickHouseRepo{conn: conn}, nil
}

func (r *ClickHouseRepo) Insert(ctx context.Context, points []domain.TelemetryPoint) error {
	batch, err := r.conn.PrepareBatch(ctx, "INSERT INTO  telemetry (id, sensor_id, device_id, type, unit, value, timestamp)")
	if err != nil {
		return err
	}
	for _, p := range points {
		valueStr := fmt.Sprintf("%v", p.Value)
		err := batch.Append(
			p.ID,
			p.SensorID,
			p.DeviceID,
			p.Type,
			p.Unit,
			valueStr,
			p.Timestamp,
		)
		if err != nil {
			return err
		}
	}
	return batch.Send()
}

func (r *ClickHouseRepo) GetLastBySensorID(ctx context.Context, sensorID string) (*domain.TelemetryPoint, error) {
	var point domain.TelemetryPoint
	row := r.conn.QueryRow(ctx, `
		SELECT id, sensor_id, device_id, type, unit, value, timestamp
		FROM telemetry
		WHERE sensor_id = ?
		ORDER BY timestamp DESC
		LIMIT 1
	`, sensorID)

	var valueStr string
	err := row.Scan(
		&point.ID,
		&point.SensorID,
		&point.DeviceID,
		&point.Type,
		&point.Unit,
		&valueStr,
		&point.Timestamp,
	)
	if err != nil {
		return nil, err
	}
	point.Value = valueStr
	return &point, nil
}

func (r *ClickHouseRepo) GetRaw(ctx context.Context, from, to time.Time, filters map[string]string, limit int) ([]*domain.TelemetryPoint, error) {
	query := `SELECT id, sensor_id, device_id, type, unit, value, timestamp FROM telemetry WHERE timestamp >= ? AND timestamp <= ?`
	args := []interface{}{from, to}

	if sensorID, ok := filters["sensor_id"]; ok && sensorID != "" {
		query += " AND sensor_id = ?"
		args = append(args, sensorID)
	}
	if deviceID, ok := filters["device_id"]; ok && deviceID != "" {
		query += " AND device_id = ?"
		args = append(args, deviceID)
	}
	if typ, ok := filters["type"]; ok && typ != "" {
		query += " AND type = ?"
		args = append(args, typ)
	}

	query += " ORDER BY timestamp DESC LIMIT ?"
	args = append(args, limit)

	rows, err := r.conn.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var points []*domain.TelemetryPoint
	for rows.Next() {
		var p domain.TelemetryPoint
		var valueStr string
		err := rows.Scan(&p.ID, &p.SensorID, &p.DeviceID, &p.Type, &p.Unit, &valueStr, &p.Timestamp)
		if err != nil {
			continue
		}
		p.Value = valueStr
		points = append(points, &p)
	}
	return points, nil
}
