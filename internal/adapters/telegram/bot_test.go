package telegram

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	cardstorage "flash2fy/internal/adapters/storage/card"
	telecardstorage "flash2fy/internal/adapters/storage/telegram/card"
	teleuserstorage "flash2fy/internal/adapters/storage/telegram/user"
	userstorage "flash2fy/internal/adapters/storage/user"
	appcardapp "flash2fy/internal/app/application/card"
	appuserapp "flash2fy/internal/app/application/user"
	"flash2fy/internal/app/domain/card"
	telegramcardapp "flash2fy/internal/telegram/application/card"
	telegramuserapp "flash2fy/internal/telegram/application/user"
	telegrmdomain "flash2fy/internal/telegram/domain"
)

func newTelegramServices() (*telegramcardapp.Service, *telegramuserapp.Service, *cardstorage.MemoryRepository, *telecardstorage.MemoryRepository, *userstorage.MemoryRepository, *teleuserstorage.MemoryRepository) {
	appCardRepo := cardstorage.NewMemoryRepository()
	appCardService := appcardapp.NewService(appCardRepo)
	teleCardRepo := telecardstorage.NewMemoryRepository()
	cardService := telegramcardapp.NewService(appCardService, teleCardRepo)

	appUserRepo := userstorage.NewMemoryRepository()
	appUserService := appuserapp.NewService(appUserRepo)
	teleUserRepo := teleuserstorage.NewMemoryRepository()
	userService := telegramuserapp.NewService(appUserService, teleUserRepo)

	return cardService, userService, appCardRepo, teleCardRepo, appUserRepo, teleUserRepo
}

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
	cardService, userService, appCardRepo, teleCardRepo, _, teleUserRepo := newTelegramServices()

	var capturedChatID int64
	var capturedMessage string

	h := &updateHandler{
		cardService: cardService,
		userService: userService,
		send: func(ctx context.Context, _ *bot.Bot, params *bot.SendMessageParams) error {
			if id, ok := params.ChatID.(int64); ok {
				capturedChatID = id
			}
			capturedMessage = params.Text
			return nil
		},
	}

	update := &models.Update{
		Message: &models.Message{
			Chat: models.Chat{ID: 123},
			From: &models.User{ID: 555, FirstName: "John", LastName: "Doe", Username: "john"},
			Text: "Create card",
		},
	}

	h.handleCreateCard(context.Background(), nil, update, update.Message.Text)

	if capturedChatID != update.Message.Chat.ID {
		t.Fatalf("expected chat id %d, got %d", update.Message.Chat.ID, capturedChatID)
	}
	if !strings.Contains(capturedMessage, "Card created") {
		t.Fatalf("expected success message, got %q", capturedMessage)
	}

	cards, err := appCardRepo.FindAll()
	if err != nil {
		t.Fatalf("find all cards: %v", err)
	}
	if len(cards) != 1 {
		t.Fatalf("expected 1 card persisted, got %d", len(cards))
	}

	projection, err := teleCardRepo.FindByCoreID(cards[0].ID)
	if err != nil {
		t.Fatalf("expected telegram projection: %v", err)
	}
	if projection.OwnerTelegramID != update.Message.From.ID {
		t.Fatalf("expected projection owner %d, got %d", update.Message.From.ID, projection.OwnerTelegramID)
	}

	teleUser, err := teleUserRepo.FindByTelegramID(update.Message.From.ID)
	if err != nil {
		t.Fatalf("expected telegram user projection: %v", err)
	}
	if teleUser.Username != "john" {
		t.Fatalf("expected username john, got %s", teleUser.Username)
	}
}

func TestHandleCreateCardRequiresFront(t *testing.T) {
	cardService, userService, appCardRepo, _, _, _ := newTelegramServices()

	h := &updateHandler{
		cardService: cardService,
		userService: userService,
		send:        func(ctx context.Context, _ *bot.Bot, _ *bot.SendMessageParams) error { return nil },
	}

	update := &models.Update{
		Message: &models.Message{
			Chat: models.Chat{ID: 123},
			Text: "   ",
		},
	}

	h.handleCreateCard(context.Background(), nil, update, update.Message.Text)

	cards, err := appCardRepo.FindAll()
	if err != nil {
		t.Fatalf("find all cards: %v", err)
	}
	if len(cards) != 0 {
		t.Fatalf("expected no cards to be created, got %d", len(cards))
	}
}

type failingAppCardService struct{}

func (failingAppCardService) CreateCard(string, string, string) (card.Card, error) {
	return card.Card{}, errors.New("boom")
}

func (failingAppCardService) GetCard(string) (card.Card, error) {
	return card.Card{}, card.ErrNotFound
}

func (failingAppCardService) DeleteCard(string) error { return nil }

type noopTelegramCardRepo struct{}

func (noopTelegramCardRepo) Save(telegrmdomain.Card) (telegrmdomain.Card, error) {
	return telegrmdomain.Card{}, nil
}
func (noopTelegramCardRepo) FindByCoreID(string) (telegrmdomain.Card, error) {
	return telegrmdomain.Card{}, telegrmdomain.ErrCardNotFound
}
func (noopTelegramCardRepo) DeleteByCoreID(string) error { return nil }

func TestHandleCreateCardPropagatesError(t *testing.T) {
	appUserRepo := userstorage.NewMemoryRepository()
	appUserService := appuserapp.NewService(appUserRepo)
	teleUserRepo := teleuserstorage.NewMemoryRepository()
	userService := telegramuserapp.NewService(appUserService, teleUserRepo)

	cardService := telegramcardapp.NewService(failingAppCardService{}, noopTelegramCardRepo{})

	var capturedMessage string
	h := &updateHandler{
		cardService: cardService,
		userService: userService,
		send: func(ctx context.Context, _ *bot.Bot, params *bot.SendMessageParams) error {
			capturedMessage = params.Text
			return nil
		},
	}

	update := &models.Update{
		Message: &models.Message{
			Chat: models.Chat{ID: 111},
			From: &models.User{ID: 999, FirstName: "Jane"},
			Text: "hello",
		},
	}

	h.handleCreateCard(context.Background(), nil, update, update.Message.Text)
	if !strings.Contains(capturedMessage, "Failed to create card") {
		t.Fatalf("expected failure message, got %q", capturedMessage)
	}
}

func TestConfigureWebhook(t *testing.T) {
	const (
		token      = "token"
		webhookURL = "https://example.com/telegram"
		secret     = "s3cret"
	)
	var requestObserved bool

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Skipf("skipping webhook test due to listener error: %v", err)
	}
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
	server.Listener = listener
	server.Start()
	t.Cleanup(server.Close)

	cardService, userService, _, _, _, _ := newTelegramServices()
	tg, err := New(token, cardService, userService, bot.WithSkipGetMe(), bot.WithServerURL(server.URL))
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

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Skipf("skipping webhook error test due to listener error: %v", err)
	}
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseMultipartForm(1 << 20); err != nil {
			t.Fatalf("parse multipart form: %v", err)
		}
		secretReceived = r.FormValue("secret_token")

		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write([]byte(`{"ok":false,"error_code":400,"description":"bad request"}`)); err != nil {
			t.Fatalf("write response: %v", err)
		}
	}))
	server.Listener = listener
	server.Start()
	t.Cleanup(server.Close)

	cardService, userService, _, _, _, _ := newTelegramServices()
	tg, err := New(token, cardService, userService, bot.WithSkipGetMe(), bot.WithServerURL(server.URL))
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
