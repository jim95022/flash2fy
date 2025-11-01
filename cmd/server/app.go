package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path"
	"strings"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	telegramapi "github.com/go-telegram/bot"
	_ "github.com/jackc/pgx/v5/stdlib"

	cardhttp "flash2fy/internal/adapters/http/card"
	cardstorage "flash2fy/internal/adapters/storage/card"
	telegram "flash2fy/internal/adapters/telegram"
	cardapp "flash2fy/internal/application/card"
	flashconfig "flash2fy/internal/config"
)

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := flashconfig.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	db, err := connectPostgres(ctx, cfg.Database.URL)
	if err != nil {
		return err
	}
	defer db.Close()

	repo := cardstorage.NewPostgresRepository(db)
	service := cardapp.NewCardService(repo)
	handler := cardhttp.NewHandler(service)

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	if err := setupTelegramWebhook(ctx, r, cfg, service); err != nil {
		return fmt.Errorf("setup telegram webhook: %w", err)
	}

	r.Mount("/v1/cards", handler.Routes())

	srv := &http.Server{
		Addr:    cfg.Server.Addr,
		Handler: r,
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("HTTP server shutdown error: %v", err)
		}
	}()

	log.Printf("card service listening on %s", cfg.Server.Addr)
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("start server: %w", err)
	}

	return nil
}

func setupTelegramWebhook(ctx context.Context, r *chi.Mux, cfg *flashconfig.Config, service *cardapp.CardService) error {
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

	bot, err := telegram.New(cfg.Telegram.BotToken, service, options...)
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

func connectPostgres(ctx context.Context, dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("open postgres connection: %w", err)
	}

	ctxPing, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctxPing); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	return db, nil
}
