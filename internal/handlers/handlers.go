package handlers

import (
	"encoding/json"
	"net/http"

	"calendar/internal/logger"
	"calendar/internal/services"
)

type Handlers struct {
	log    logger.Logger
	events services.EventsService
}

func NewHandlers(log logger.Logger, events services.EventsService) *Handlers {
	return &Handlers{
		log:    log,
		events: events,
	}
}

// DTO

type createEventRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	StartTime   string `json:"start_time"` // RFC3339
	EndTime     string `json:"end_time"`   // RFC3339
	OwnerID     string `json:"owner_id"`
}

type updateEventRequest struct {
	Title       *string `json:"title,omitempty"`
	Description *string `json:"description,omitempty"`
	StartTime   *string `json:"start_time,omitempty"` // RFC3339
	EndTime     *string `json:"end_time,omitempty"`   // RFC3339
	OwnerID     *string `json:"owner_id,omitempty"`
}

type eventResponse struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	StartTime   string `json:"start_time"`
	EndTime     string `json:"end_time"`
	OwnerID     string `json:"owner_id"`
}

// Вспомогалки

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// Регистрация маршрутов

func (h *Handlers) RegisterRoutes(mux *http.ServeMux) {
	// список/создание
	mux.HandleFunc("/api/events", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			h.CreateEvent(w, r)
		case http.MethodGet:
			h.ListEvents(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	// обновление/удаление по id
	mux.HandleFunc("/api/events/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPut, http.MethodPatch:
			h.UpdateEvent(w, r)
		case http.MethodDelete:
			h.DeleteEvent(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})
}
