package cardhttp

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"

	"flash2fy/internal/adapters/storage"
	cardapp "flash2fy/internal/application/card"
	"flash2fy/internal/domain/card"
)

type httpTestDeps struct {
	service *cardapp.CardService
	handler http.Handler
}

func newHTTPTestDeps() httpTestDeps {
	repo := storage.NewMemoryCardRepository()
	service := cardapp.NewCardService(repo)
	router := chi.NewRouter()
	router.Mount("/v1/cards", NewHandler(service).Routes())
	return httpTestDeps{
		service: service,
		handler: router,
	}
}

func TestCreateCardEndpoint(t *testing.T) {
	deps := newHTTPTestDeps()

	payload := map[string]string{
		"front": "What is Go?",
		"back":  "A programming language",
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/v1/cards", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	deps.handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", rec.Code)
	}

	var resp cardResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Front != payload["front"] || resp.Back != payload["back"] {
		t.Fatalf("unexpected response payload: %+v", resp)
	}
	if resp.ID == "" {
		t.Fatalf("expected ID to be set")
	}
}

func TestCreateCardEndpointValidation(t *testing.T) {
	deps := newHTTPTestDeps()

	payload := map[string]string{
		"front": "",
		"back":  "",
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/v1/cards", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	deps.handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}

	var errResp errorResponse
	if err := json.NewDecoder(rec.Body).Decode(&errResp); err != nil {
		t.Fatalf("failed to decode error response: %v", err)
	}
	if errResp.Message != card.ErrEmptyFront.Error() {
		t.Fatalf("expected error message %q, got %q", card.ErrEmptyFront.Error(), errResp.Message)
	}
}

func TestGetCardEndpoint(t *testing.T) {
	deps := newHTTPTestDeps()

	created, err := deps.service.CreateCard("Front", "Back")
	if err != nil {
		t.Fatalf("setup create failed: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/v1/cards/"+created.ID, nil)
	rec := httptest.NewRecorder()

	deps.handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var resp cardResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.ID != created.ID {
		t.Fatalf("expected ID %s, got %s", created.ID, resp.ID)
	}
}

func TestGetCardEndpointNotFound(t *testing.T) {
	deps := newHTTPTestDeps()

	req := httptest.NewRequest(http.MethodGet, "/v1/cards/missing", nil)
	rec := httptest.NewRecorder()

	deps.handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", rec.Code)
	}
}

func TestUpdateCardEndpoint(t *testing.T) {
	deps := newHTTPTestDeps()

	created, err := deps.service.CreateCard("Old Front", "Old Back")
	if err != nil {
		t.Fatalf("setup create failed: %v", err)
	}

	payload := map[string]string{
		"front": "New Front",
		"back":  "New Back",
	}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPut, "/v1/cards/"+created.ID, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	deps.handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var resp cardResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Front != payload["front"] || resp.Back != payload["back"] {
		t.Fatalf("expected updated payload, got %+v", resp)
	}
}

func TestDeleteCardEndpoint(t *testing.T) {
	deps := newHTTPTestDeps()

	created, err := deps.service.CreateCard("Front", "Back")
	if err != nil {
		t.Fatalf("setup create failed: %v", err)
	}

	req := httptest.NewRequest(http.MethodDelete, "/v1/cards/"+created.ID, nil)
	rec := httptest.NewRecorder()

	deps.handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d", rec.Code)
	}

	_, err = deps.service.GetCard(created.ID)
	if err != card.ErrNotFound {
		t.Fatalf("expected card to be deleted, got error %v", err)
	}
}

func TestListCardsEndpoint(t *testing.T) {
	deps := newHTTPTestDeps()

	_, err := deps.service.CreateCard("Front A", "Back A")
	if err != nil {
		t.Fatalf("setup create failed: %v", err)
	}
	_, err = deps.service.CreateCard("Front B", "Back B")
	if err != nil {
		t.Fatalf("setup create failed: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/v1/cards", nil)
	rec := httptest.NewRecorder()

	deps.handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var resp []cardResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(resp) != 2 {
		t.Fatalf("expected 2 cards, got %d", len(resp))
	}
}
