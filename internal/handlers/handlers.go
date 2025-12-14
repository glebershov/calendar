package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	"calendar/internal/logger"
	"calendar/internal/repos"
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

// /api/events/{id}
func getIDFromPath(r *http.Request) string {
	path := strings.TrimPrefix(r.URL.Path, "/api/events/")
	path = strings.Trim(path, "/")
	if path == "" || strings.Contains(path, "/") {
		return ""
	}
	return path
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

// Handlers

// CreateEvent — POST /api/events
func (h *Handlers) CreateEvent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req createEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	if req.Title == "" || req.OwnerID == "" {
		writeError(w, http.StatusBadRequest, "title and owner_id are required")
		return
	}

	start, err := time.Parse(time.RFC3339, req.StartTime)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid start_time")
		return
	}
	end, err := time.Parse(time.RFC3339, req.EndTime)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid end_time")
		return
	}

	e := &repos.Event{
		ID:          uuid.New().String(),
		Title:       req.Title,
		Description: req.Description,
		StartTime:   start,
		EndTime:     end,
		OwnerID:     req.OwnerID,
	}

	if err := h.events.CreateEvent(r.Context(), e); err != nil {
		h.log.Error("create event failed", "err", err)
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	resp := eventResponse{
		ID:          e.ID,
		Title:       e.Title,
		Description: e.Description,
		StartTime:   e.StartTime.Format(time.RFC3339),
		EndTime:     e.EndTime.Format(time.RFC3339),
		OwnerID:     e.OwnerID,
	}

	writeJSON(w, http.StatusCreated, resp)
}

// ListEvents — GET /api/events?owner_id=...
func (h *Handlers) ListEvents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	ownerID := r.URL.Query().Get("owner_id")
	if ownerID == "" {
		writeError(w, http.StatusBadRequest, "owner_id is required")
		return
	}

	events, err := h.events.ListEvents(r.Context(), ownerID)
	if err != nil {
		h.log.Error("list events failed", "err", err)
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	resp := make([]eventResponse, 0, len(events))
	for _, e := range events {
		resp = append(resp, eventResponse{
			ID:          e.ID,
			Title:       e.Title,
			Description: e.Description,
			StartTime:   e.StartTime.Format(time.RFC3339),
			EndTime:     e.EndTime.Format(time.RFC3339),
			OwnerID:     e.OwnerID,
		})
	}

	writeJSON(w, http.StatusOK, resp)
}

// UpdateEvent — PUT/PATCH /api/events/{id}
func (h *Handlers) UpdateEvent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut && r.Method != http.MethodPatch {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	id := getIDFromPath(r)
	if id == "" {
		writeError(w, http.StatusBadRequest, "id is required in path")
		return
	}

	var req updateEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	e := &repos.Event{
		ID: id,
	}

	if req.Title != nil {
		e.Title = *req.Title
	}
	if req.Description != nil {
		e.Description = *req.Description
	}
	if req.StartTime != nil {
		start, err := time.Parse(time.RFC3339, *req.StartTime)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid start_time")
			return
		}
		e.StartTime = start
	}
	if req.EndTime != nil {
		end, err := time.Parse(time.RFC3339, *req.EndTime)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid end_time")
			return
		}
		e.EndTime = end
	}
	if req.OwnerID != nil {
		e.OwnerID = *req.OwnerID
	}

	if err := h.events.UpdateEvent(r.Context(), e); err != nil {
		h.log.Error("update event failed", "err", err)
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// DeleteEvent — DELETE /api/events/{id}
func (h *Handlers) DeleteEvent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	id := getIDFromPath(r)
	if id == "" {
		writeError(w, http.StatusBadRequest, "id is required in path")
		return
	}

	if err := h.events.DeleteEvent(r.Context(), id); err != nil {
		h.log.Error("delete event failed", "err", err)
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
