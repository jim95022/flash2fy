package cardbot

import (
	"context"
	"errors"
	"strconv"
	"testing"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

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
