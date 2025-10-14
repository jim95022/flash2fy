package cardhttp

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"flash2fy/internal/application"
	"flash2fy/internal/domain/card"
)

// CardHandler exposes HTTP endpoints for card operations.
type CardHandler struct {
	service *application.CardService
}

func NewCardHandler(service *application.CardService) *CardHandler {
	return &CardHandler{service: service}
}

func (h *CardHandler) Routes() chi.Router {
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

func (h *CardHandler) createCard(w http.ResponseWriter, r *http.Request) {
	var req cardRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request payload")
		return
	}

	c, err := h.service.CreateCard(req.Front, req.Back)
	if err != nil {
		status := http.StatusInternalServerError
		if err == card.ErrEmptyFront || err == card.ErrEmptyBack {
			status = http.StatusBadRequest
		}
		writeError(w, status, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, toResponse(c))
}

func (h *CardHandler) getCard(w http.ResponseWriter, r *http.Request) {
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

func (h *CardHandler) listCards(w http.ResponseWriter, r *http.Request) {
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

func (h *CardHandler) updateCard(w http.ResponseWriter, r *http.Request) {
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
		case card.ErrEmptyFront, card.ErrEmptyBack:
			status = http.StatusBadRequest
		}
		writeError(w, status, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, toResponse(c))
}

func (h *CardHandler) deleteCard(w http.ResponseWriter, r *http.Request) {
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
