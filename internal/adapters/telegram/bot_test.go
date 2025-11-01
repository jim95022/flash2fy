package telegram

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	cardapp "flash2fy/internal/application/card"
	"flash2fy/internal/domain/card"
)

func TestSplitCommand(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantCmd  string
		wantBody string
	}{
		{"simple", "/add front | back", "/add", "front | back"},
		{"newline", "/add\nfront\nback", "/add", "front\nback"},
		{"with mention", "/add@flash2fy hello | world", "/add", "hello | world"},
		{"no payload", "/start", "/start", ""},
		{"leading spaces", "   /help   please", "/help", "please"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cmd, body := splitCommand(tc.input)
			if cmd != tc.wantCmd || body != tc.wantBody {
				t.Fatalf("splitCommand(%q) = %q, %q; want %q, %q", tc.input, cmd, body, tc.wantCmd, tc.wantBody)
			}
		})
	}
}

func TestHandleCreateCard(t *testing.T) {
	var (
		capturedChatID   int64
		capturedMessage  string
		capturedFront    string
		capturedOwnerID  string
		createCardCalled bool
	)

	handler := &updateHandler{
		createCard: func(front, back, ownerID string) (card.Card, error) {
			createCardCalled = true
			capturedFront = front
			capturedOwnerID = ownerID
			return card.Card{ID: "id-1", Front: front, Back: back, OwnerID: ownerID}, nil
		},
		send: func(ctx context.Context, _ *bot.Bot, params *bot.SendMessageParams) error {
			if id, ok := params.ChatID.(int64); ok {
				capturedChatID = id
			}
			capturedMessage = params.Text
			return nil
		},
	}
	var client *bot.Bot

	update := &models.Update{
		Message: &models.Message{
			Chat: models.Chat{ID: 123},
			From: &models.User{ID: 555},
			Text: "Create card",
		},
	}

	handler.handleCreateCard(context.Background(), client, update, update.Message.Text)

	if !createCardCalled {
		t.Fatalf("expected CreateCard to be invoked")
	}
	if capturedFront != "Create card" {
		t.Fatalf("expected front 'Create card', got %q", capturedFront)
	}
	if capturedOwnerID != strconv.FormatInt(update.Message.From.ID, 10) {
		t.Fatalf("expected owner %s, got %s", strconv.FormatInt(update.Message.From.ID, 10), capturedOwnerID)
	}
	if capturedChatID != update.Message.Chat.ID {
		t.Fatalf("expected chat ID %d, got %d", update.Message.Chat.ID, capturedChatID)
	}
	if capturedMessage == "" {
		t.Fatalf("expected response message to be sent")
	}
}

func TestHandleCreateCardRequiresFront(t *testing.T) {
	createCalled := false
	handler := &updateHandler{
		createCard: func(front, back, ownerID string) (card.Card, error) {
			createCalled = true
			return card.Card{}, nil
		},
		send: func(ctx context.Context, _ *bot.Bot, params *bot.SendMessageParams) error { return nil },
	}
	var client *bot.Bot

	update := &models.Update{
		Message: &models.Message{
			Chat: models.Chat{ID: 123},
			Text: "   ",
		},
	}

	handler.handleCreateCard(context.Background(), client, update, update.Message.Text)
	if createCalled {
		t.Fatalf("expected CreateCard not to be called for empty message")
	}
}

func TestHandleCreateCardPropagatesError(t *testing.T) {
	expectedErr := errors.New("boom")
	var called bool
	handler := &updateHandler{
		createCard: func(front, back, ownerID string) (card.Card, error) {
			called = true
			return card.Card{}, expectedErr
		},
		send: func(ctx context.Context, _ *bot.Bot, params *bot.SendMessageParams) error { return nil },
	}
	var client *bot.Bot

	update := &models.Update{
		Message: &models.Message{
			Chat: models.Chat{ID: 111},
			From: &models.User{ID: 999},
			Text: "hello",
		},
	}

	handler.handleCreateCard(context.Background(), client, update, update.Message.Text)
	if !called {
		t.Fatalf("expected CreateCard to be called even when service returns error")
	}
}

func TestConfigureWebhook(t *testing.T) {
	const (
		token      = "token"
		webhookURL = "https://example.com/telegram"
		secret     = "s3cret"
	)
	var requestObserved bool

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestObserved = true

		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		wantPath := "/bot" + token + "/setWebhook"
		if r.URL.Path != wantPath {
			t.Fatalf("unexpected path %s, want %s", r.URL.Path, wantPath)
		}
		if err := r.ParseMultipartForm(1 << 20); err != nil {
			t.Fatalf("parse multipart form: %v", err)
		}
		if got := r.FormValue("url"); got != webhookURL {
			t.Fatalf("expected url %q, got %q", webhookURL, got)
		}
		if got := r.FormValue("drop_pending_updates"); got != "true" {
			t.Fatalf("expected drop_pending_updates true, got %q", got)
		}
		if got := r.FormValue("secret_token"); got != secret {
			t.Fatalf("expected secret %q, got %q", secret, got)
		}
		if got := r.FormValue("allowed_updates"); got != `["message","callback_query"]` {
			t.Fatalf("unexpected allowed_updates %q", got)
		}

		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write([]byte(`{"ok":true,"result":true}`)); err != nil {
			t.Fatalf("write response: %v", err)
		}
	}))
	t.Cleanup(server.Close)

	service := cardapp.NewCardService(stubCardRepo{})
	tg, err := New(token, service, bot.WithSkipGetMe(), bot.WithServerURL(server.URL))
	if err != nil {
		t.Fatalf("init telegram bot: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	err = tg.ConfigureWebhook(ctx, WebhookConfig{
		URL:                webhookURL,
		SecretToken:        secret,
		DropPendingUpdates: true,
		AllowedUpdates:     []string{"message", "callback_query"},
	})
	if err != nil {
		t.Fatalf("ConfigureWebhook returned error: %v", err)
	}
	if !requestObserved {
		t.Fatal("expected webhook request to be observed")
	}
}

func TestConfigureWebhookError(t *testing.T) {
	const token = "token"
	var secretReceived string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseMultipartForm(1 << 20); err != nil {
			t.Fatalf("parse multipart form: %v", err)
		}
		secretReceived = r.FormValue("secret_token")

		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write([]byte(`{"ok":false,"error_code":400,"description":"bad request"}`)); err != nil {
			t.Fatalf("write response: %v", err)
		}
	}))
	t.Cleanup(server.Close)

	service := cardapp.NewCardService(stubCardRepo{})
	tg, err := New(token, service, bot.WithSkipGetMe(), bot.WithServerURL(server.URL))
	if err != nil {
		t.Fatalf("init telegram bot: %v", err)
	}

	err = tg.ConfigureWebhook(context.Background(), WebhookConfig{
		URL:         "https://example.com/telegram",
		SecretToken: "",
	})
	if err == nil || !strings.Contains(err.Error(), "bad request") {
		t.Fatalf("expected error containing 'bad request', got %v", err)
	}
	if secretReceived != "" {
		t.Fatalf("expected no secret_token to be sent, got %q", secretReceived)
	}
}

type stubCardRepo struct{}

func (stubCardRepo) Save(c card.Card) (card.Card, error) { return c, nil }

func (stubCardRepo) FindByID(string) (card.Card, error) {
	return card.Card{}, errors.New("not implemented")
}

func (stubCardRepo) FindAll() ([]card.Card, error) { return nil, errors.New("not implemented") }

func (stubCardRepo) Update(card.Card) (card.Card, error) {
	return card.Card{}, errors.New("not implemented")
}

func (stubCardRepo) Delete(string) error { return errors.New("not implemented") }
