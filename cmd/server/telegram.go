package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	telegramapi "github.com/go-telegram/bot"

	telegram "flash2fy/internal/adapters/telegram"
	flashconfig "flash2fy/internal/config"
	telegramcardapp "flash2fy/internal/telegram/application/card"
	telegramuserapp "flash2fy/internal/telegram/application/user"
)

func setupTelegramWebhook(ctx context.Context, r *chi.Mux, cfg *flashconfig.Config, cardService *telegramcardapp.Service, userService *telegramuserapp.Service) error {
	if cfg.Telegram.BotToken == "" {
		log.Println("telegram bot disabled (TELEGRAM_BOT_TOKEN not set)")
		return nil
	}
	if cfg.Telegram.WebhookURL == "" {
		log.Println("telegram bot disabled (TELEGRAM_WEBHOOK_URL not set)")
		return nil
	}

	options := []telegramapi.Option{}
	if secret := cfg.Telegram.WebhookSecret; secret != "" {
		options = append(options, telegramapi.WithWebhookSecretToken(secret))
	}

	bot, err := telegram.New(cfg.Telegram.BotToken, cardService, userService, options...)
	if err != nil {
		return fmt.Errorf("init telegram bot: %w", err)
	}

	webhookURL, webhookPath, err := resolveWebhookEndpoint(cfg.Telegram.WebhookURL, cfg.Telegram.WebhookPath)
	if err != nil {
		return err
	}

	ctxSet, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := bot.ConfigureWebhook(ctxSet, telegram.WebhookConfig{
		URL:                webhookURL,
		SecretToken:        cfg.Telegram.WebhookSecret,
		DropPendingUpdates: true,
	}); err != nil {
		return err
	}

	r.Post(webhookPath, bot.WebhookHandler())

	go func() {
		log.Printf("telegram webhook active on path %s (target %s)", webhookPath, webhookURL)
		bot.StartWebhook(ctx)
	}()

	return nil
}

func resolveWebhookEndpoint(baseURL string, configuredPath string) (string, string, error) {
	if strings.TrimSpace(baseURL) == "" {
		return "", "", errors.New("TELEGRAM_WEBHOOK_URL must not be empty")
	}

	u, err := url.Parse(baseURL)
	if err != nil {
		return "", "", fmt.Errorf("parse TELEGRAM_WEBHOOK_URL: %w", err)
	}
	if u.Scheme == "" || u.Host == "" {
		return "", "", fmt.Errorf("TELEGRAM_WEBHOOK_URL must include scheme and host, got %q", baseURL)
	}

	defaultPath := strings.TrimSpace(configuredPath)
	if defaultPath == "" {
		defaultPath = "/telegram/webhook"
	}
	defaultPath = path.Clean("/" + strings.TrimPrefix(defaultPath, "/"))

	if u.Path != "" && u.Path != "/" {
		defaultPath = path.Clean(u.Path)
	} else {
		u.Path = defaultPath
	}

	return u.String(), defaultPath, nil
}
