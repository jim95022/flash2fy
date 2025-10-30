package cardbot

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	cardapp "flash2fy/internal/application/card"
)

// Bot exposes Telegram commands to manage flashcards.
type Bot struct {
	service *cardapp.CardService
	client  *bot.Bot
}

// New constructs a Telegram bot configured to create cards via /add command.
func New(token string, service *cardapp.CardService, options ...bot.Option) (*Bot, error) {
	if token == "" {
		return nil, errors.New("telegram bot token must not be empty")
	}
	handler := &updateHandler{service: service}

	opts := []bot.Option{
		bot.WithDefaultHandler(handler.handle),
	}
	opts = append(opts, options...)

	client, err := bot.New(token, opts...)
	if err != nil {
		return nil, fmt.Errorf("create telegram bot: %w", err)
	}

	return &Bot{
		service: service,
		client:  client,
	}, nil
}

// Start begins polling for Telegram updates and blocks until ctx is cancelled.
func (b *Bot) Start(ctx context.Context) {
	b.client.Start(ctx)
}

type updateHandler struct {
	service *cardapp.CardService
}

func (h *updateHandler) handle(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message == nil || update.Message.Text == "" {
		return
	}

	h.dispatch(ctx, b, update)
}
