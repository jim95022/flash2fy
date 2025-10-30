package cardhttp

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	cardapp "flash2fy/internal/application/card"
	"flash2fy/internal/domain/card"
)

// Handler exposes HTTP endpoints for card operations.
type Handler struct {
	service *cardapp.CardService
}

func NewHandler(service *cardapp.CardService) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Post("/", h.createCard)
	r.Get("/", h.listCards)
	r.Get("/{id}", h.getCard)
	r.Put("/{id}", h.updateCard)
	r.Delete("/{id}", h.deleteCard)

	return r
}

type cardRequest struct {
	Front string `json:"front"`
	Back  string `json:"back"`
}

type cardResponse struct {
	ID        string `json:"id"`
	Front     string `json:"front"`
	Back      string `json:"back"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

type errorResponse struct {
	Message string `json:"message"`
}

func (h *Handler) createCard(w http.ResponseWriter, r *http.Request) {
	var req cardRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request payload")
		return
	}

	c, err := h.service.CreateCard(req.Front, req.Back)
	if err != nil {
		status := http.StatusInternalServerError
		if err == card.ErrEmptyFront {
			status = http.StatusBadRequest
		}
		writeError(w, status, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, toResponse(c))
}

func (h *Handler) getCard(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	c, err := h.service.GetCard(id)
	if err != nil {
		status := http.StatusInternalServerError
		if err == card.ErrNotFound {
			status = http.StatusNotFound
		}
		writeError(w, status, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, toResponse(c))
}

func (h *Handler) listCards(w http.ResponseWriter, r *http.Request) {
	cards, err := h.service.ListCards()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	result := make([]cardResponse, 0, len(cards))
	for _, c := range cards {
		result = append(result, toResponse(c))
	}

	writeJSON(w, http.StatusOK, result)
}

func (h *Handler) updateCard(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req cardRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request payload")
		return
	}

	c, err := h.service.UpdateCard(id, req.Front, req.Back)
	if err != nil {
		status := http.StatusInternalServerError
		switch err {
		case card.ErrNotFound:
			status = http.StatusNotFound
		case card.ErrEmptyFront:
			status = http.StatusBadRequest
		}
		writeError(w, status, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, toResponse(c))
}

func (h *Handler) deleteCard(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.service.DeleteCard(id); err != nil {
		status := http.StatusInternalServerError
		if err == card.ErrNotFound {
			status = http.StatusNotFound
		}
		writeError(w, status, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func toResponse(c card.Card) cardResponse {
	return cardResponse{
		ID:        c.ID,
		Front:     c.Front,
		Back:      c.Back,
		CreatedAt: c.CreatedAt.Format(time.RFC3339Nano),
		UpdatedAt: c.UpdatedAt.Format(time.RFC3339Nano),
	}
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, errorResponse{Message: message})
}
