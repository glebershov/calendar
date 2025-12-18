package repos

import (
	"context"
	"database/sql"
	"time"
)

// Event описывает сущность события календаря.
type Event struct {
	ID          string
	Title       string
	Description string
	StartTime   time.Time
	EndTime     time.Time
	OwnerID     string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// PGEventStorage — реализация хранилища событий поверх PostgreSQL (*sql.DB).
type PGEventStorage struct {
	db *sql.DB
}

// NewPGEventStorage создаёт новое хранилище событий.
func NewPGEventStorage(db *sql.DB) *PGEventStorage {
	return &PGEventStorage{db: db}
}

// CreateEvent добавляет новое событие и заполняет ID/CreatedAt/UpdatedAt.
func (s *PGEventStorage) CreateEvent(ctx context.Context, e *Event) error {
	const query = `
		INSERT INTO events (id, title, description, start_time, end_time, owner_id)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING created_at, updated_at
	`

	return s.db.QueryRowContext(
		ctx,
		query,
		e.ID,
		e.Title,
		e.Description,
		e.StartTime,
		e.EndTime,
		e.OwnerID,
	).Scan(&e.CreatedAt, &e.UpdatedAt)
}

// UpdateEvent изменяет существующее событие по ID.
// Обновляет только те поля, которые были установлены (непустые для строк, не нулевые для времени).
func (s *PGEventStorage) UpdateEvent(ctx context.Context, e *Event) error {
	// Сначала получаем существующее событие
	const selectQuery = `
		SELECT title, description, start_time, end_time, owner_id
		FROM events
		WHERE id = $1
	`

	var existing Event
	err := s.db.QueryRowContext(ctx, selectQuery, e.ID).Scan(
		&existing.Title,
		&existing.Description,
		&existing.StartTime,
		&existing.EndTime,
		&existing.OwnerID,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return sql.ErrNoRows
		}
		return err
	}

	// Определяем, какие поля нужно обновить
	// Используем существующие значения для полей, которые не были переданы
	updateTitle := existing.Title
	updateDescription := existing.Description
	updateStartTime := existing.StartTime
	updateEndTime := existing.EndTime
	updateOwnerID := existing.OwnerID

	// Обновляем только переданные поля (непустые/ненулевые)
	if e.Title != "" {
		updateTitle = e.Title
	}
	if e.Description != "" {
		updateDescription = e.Description
	}
	if !e.StartTime.IsZero() {
		updateStartTime = e.StartTime
	}
	if !e.EndTime.IsZero() {
		updateEndTime = e.EndTime
	}
	if e.OwnerID != "" {
		updateOwnerID = e.OwnerID
	}

	const updateQuery = `
		UPDATE events
		SET
			title       = $1,
			description = $2,
			start_time  = $3,
			end_time    = $4,
			owner_id    = $5,
			updated_at  = NOW()
		WHERE id = $6
	`

	res, err := s.db.ExecContext(
		ctx,
		updateQuery,
		updateTitle,
		updateDescription,
		updateStartTime,
		updateEndTime,
		updateOwnerID,
		e.ID,
	)
	if err != nil {
		return err
	}

	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// DeleteEvent удаляет событие по ID.
func (s *PGEventStorage) DeleteEvent(ctx context.Context, id string) error {
	const query = `DELETE FROM events WHERE id = $1`

	res, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// ListEvents возвращает все события конкретного владельца.
func (s *PGEventStorage) ListEvents(ctx context.Context, ownerID string) ([]Event, error) {
	const query = `
		SELECT
			id,
			title,
			description,
			start_time,
			end_time,
			owner_id,
			created_at,
			updated_at
		FROM events
		WHERE owner_id = $1
		ORDER BY start_time
	`

	rows, err := s.db.QueryContext(ctx, query, ownerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []Event
	for rows.Next() {
		var e Event
		if err := rows.Scan(
			&e.ID,
			&e.Title,
			&e.Description,
			&e.StartTime,
			&e.EndTime,
			&e.OwnerID,
			&e.CreatedAt,
			&e.UpdatedAt,
		); err != nil {
			return nil, err
		}
		events = append(events, e)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return events, nil
}

// GetAllEvents возвращает все события из БД.
func (s *PGEventStorage) GetAllEvents(ctx context.Context) ([]Event, error) {
	const query = `
		SELECT
			id,
			title,
			description,
			start_time,
			end_time,
			owner_id,
			created_at,
			updated_at
		FROM events
		ORDER BY start_time
	`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []Event
	for rows.Next() {
		var e Event
		if err := rows.Scan(
			&e.ID,
			&e.Title,
			&e.Description,
			&e.StartTime,
			&e.EndTime,
			&e.OwnerID,
			&e.CreatedAt,
			&e.UpdatedAt,
		); err != nil {
			return nil, err
		}
		events = append(events, e)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return events, nil
}
