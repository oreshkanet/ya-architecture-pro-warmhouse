package repository

import (
	"context"
	"database/sql"
	"fmt"

	"device-service/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type LocationRepository struct {
	db *pgxpool.Pool
}

func NewLocationRepository(db *pgxpool.Pool) *LocationRepository {
	return &LocationRepository{db: db}
}

// Create создаёт новую локацию
func (r *LocationRepository) Create(ctx context.Context, loc domain.Location) (*domain.Location, error) {
	query := `
		INSERT INTO locations (id, user_id, house_id, name, type, parent_location_id)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id`

	var createdID uuid.UUID
	err := r.db.QueryRow(ctx, query,
		loc.ID,
		loc.UserID,
		loc.HouseID,
		loc.Name,
		loc.Type,
		loc.ParentLocationID,
	).Scan(&createdID)
	if err != nil {
		return nil, fmt.Errorf("failed to insert location: %w", err)
	}

	// Возвращаем полную запись (можно сделать отдельный SELECT, но здесь ID уже известен)
	loc.ID = createdID
	return &loc, nil
}

// GetAll возвращает все локации, принадлежащие пользователю.
func (r *LocationRepository) GetAll(ctx context.Context, userId uuid.UUID) ([]domain.Location, error) {
	query := `
		SELECT id, user_id, house_id, name, type, parent_location_id
		FROM locations
		WHERE user_id = $1
		ORDER BY name`

	rows, err := r.db.Query(ctx, query, userId)
	if err != nil {
		return nil, fmt.Errorf("failed to query locations: %w", err)
	}
	defer rows.Close()

	var locations []domain.Location
	for rows.Next() {
		var loc domain.Location
		var parentID *uuid.UUID
		err := rows.Scan(&loc.ID, &loc.UserID, &loc.HouseID, &loc.Name, &loc.Type, &parentID)
		if err != nil {
			return nil, fmt.Errorf("failed to scan location: %w", err)
		}
		loc.ParentLocationID = parentID
		locations = append(locations, loc)
	}
	return locations, rows.Err()
}

// GetByID возвращает локацию по ID
func (r *LocationRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Location, error) {
	query := `
		SELECT id, user_id, house_id, name, type, parent_location_id
		FROM locations
		WHERE id = $1`

	row := r.db.QueryRow(ctx, query, id)

	var loc domain.Location
	var parentID *uuid.UUID
	err := row.Scan(&loc.ID, &loc.UserID, &loc.HouseID, &loc.Name, &loc.Type, &parentID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get location: %w", err)
	}
	loc.ParentLocationID = parentID
	return &loc, nil
}

// Update обновляет локацию
func (r *LocationRepository) Update(ctx context.Context, id uuid.UUID, loc domain.Location) (*domain.Location, error) {
	query := `
		UPDATE locations
		SET user_id = $2, house_id = $3, name = $4, type = $5, parent_location_id = $6
		WHERE id = $1
		RETURNING id, house_id, name, type, parent_location_id`

	var updated domain.Location
	var parentID *uuid.UUID
	err := r.db.QueryRow(ctx, query,
		id,
		loc.UserID,
		loc.HouseID,
		loc.Name,
		loc.Type,
		loc.ParentLocationID,
	).Scan(&updated.ID, &updated.HouseID, &updated.Name, &updated.Type, &parentID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // не найдено
		}
		return nil, fmt.Errorf("failed to update location: %w", err)
	}
	updated.ParentLocationID = parentID
	return &updated, nil
}

// Delete удаляет локацию
func (r *LocationRepository) Delete(ctx context.Context, id uuid.UUID) error {
	res, err := r.db.Exec(ctx, "DELETE FROM locations WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to delete location: %w", err)
	}
	if res.RowsAffected() == 0 {
		// Можно считать, что локация не существует — но это не ошибка
		return nil
	}
	return nil
}

// Exists проверяет, существует ли локация (полезно при создании устройства)
func (r *LocationRepository) Exists(ctx context.Context, id uuid.UUID) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM locations WHERE id = $1)", id).Scan(&exists)
	return exists, err
}
