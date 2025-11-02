package telegram

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	telegramcardapp "flash2fy/internal/telegram/application/card"
	telegramuserapp "flash2fy/internal/telegram/application/user"
)

// Bot exposes Telegram commands to manage flashcards.
type Bot struct {
	cardService *telegramcardapp.Service
	userService *telegramuserapp.Service
	client      *bot.Bot
}

// WebhookConfig configures the Telegram bot webhook.
type WebhookConfig struct {
	URL                string
	SecretToken        string
	DropPendingUpdates bool
	AllowedUpdates     []string
}

// New constructs a Telegram bot configured to create cards via chat messages.
func New(token string, cardService *telegramcardapp.Service, userService *telegramuserapp.Service, options ...bot.Option) (*Bot, error) {
	if token == "" {
		return nil, errors.New("telegram bot token must not be empty")
	}
	handler := &updateHandler{
		cardService: cardService,
		userService: userService,
		send: func(ctx context.Context, client *bot.Bot, params *bot.SendMessageParams) error {
			_, err := client.SendMessage(ctx, params)
			return err
		},
	}

	opts := []bot.Option{
		bot.WithDefaultHandler(handler.handle),
	}
	opts = append(opts, options...)

	client, err := bot.New(token, opts...)
	if err != nil {
		return nil, fmt.Errorf("create telegram bot: %w", err)
	}

	return &Bot{
		cardService: cardService,
		userService: userService,
		client:      client,
	}, nil
}

// Start begins polling for Telegram updates and blocks until ctx is cancelled.
func (b *Bot) Start(ctx context.Context) {
	b.client.Start(ctx)
}

// StartWebhook starts processing updates using Telegram webhooks.
func (b *Bot) StartWebhook(ctx context.Context) {
	b.client.StartWebhook(ctx)
}

// ConfigureWebhook registers a webhook with Telegram API.
func (b *Bot) ConfigureWebhook(ctx context.Context, cfg WebhookConfig) error {
	params := &bot.SetWebhookParams{
		URL:                cfg.URL,
		DropPendingUpdates: cfg.DropPendingUpdates,
	}
	if cfg.SecretToken != "" {
		params.SecretToken = cfg.SecretToken
	}
	if len(cfg.AllowedUpdates) > 0 {
		params.AllowedUpdates = cfg.AllowedUpdates
	}

	_, err := b.client.SetWebhook(ctx, params)
	if err != nil {
		return fmt.Errorf("set webhook: %w", err)
	}
	return nil
}

// WebhookHandler returns an http.HandlerFunc to process Telegram webhook requests.
func (b *Bot) WebhookHandler() http.HandlerFunc {
	return b.client.WebhookHandler()
}

type updateHandler struct {
	cardService *telegramcardapp.Service
	userService *telegramuserapp.Service
	send        func(ctx context.Context, client *bot.Bot, params *bot.SendMessageParams) error
}

func (h *updateHandler) handle(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message == nil || update.Message.Text == "" {
		return
	}

	h.dispatch(ctx, b, update)
}
