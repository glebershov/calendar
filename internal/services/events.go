package services

import (
	"context"

	"calendar/internal/repos"
)

// EventsService описывает, что нужно хендлерам для работы с событиями.
type EventsService interface {
	CreateEvent(ctx context.Context, e *repos.Event) error
	UpdateEvent(ctx context.Context, e *repos.Event) error
	DeleteEvent(ctx context.Context, id string) error
	ListEvents(ctx context.Context, ownerID string) ([]repos.Event, error)
}

// EventsServiceImpl — реализация сервиса событий.
type EventsServiceImpl struct {
	repo repos.EventsRepo
}

// NewEventsService создаёт новый сервис событий.
func NewEventsService(repo repos.EventsRepo) EventsService {
	return &EventsServiceImpl{
		repo: repo,
	}
}

// CreateEvent создаёт новое событие.
func (s *EventsServiceImpl) CreateEvent(ctx context.Context, e *repos.Event) error {
	return s.repo.CreateEvent(ctx, e)
}

// UpdateEvent обновляет существующее событие.
func (s *EventsServiceImpl) UpdateEvent(ctx context.Context, e *repos.Event) error {
	return s.repo.UpdateEvent(ctx, e)
}

// DeleteEvent удаляет событие по ID.
func (s *EventsServiceImpl) DeleteEvent(ctx context.Context, id string) error {
	return s.repo.DeleteEvent(ctx, id)
}

// ListEvents возвращает все события конкретного владельца.
func (s *EventsServiceImpl) ListEvents(ctx context.Context, ownerID string) ([]repos.Event, error) {
	return s.repo.ListEvents(ctx, ownerID)
}

