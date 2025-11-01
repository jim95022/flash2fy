package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
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

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := flashconfig.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	db := mustConnectPostgres(cfg.Database.URL)
	defer db.Close()

	repo := cardstorage.NewPostgresRepository(db)
	service := cardapp.NewCardService(repo)
	handler := cardhttp.NewHandler(service)
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	if err := setupTelegramWebhook(ctx, r, cfg, service); err != nil {
		log.Fatalf("failed to configure telegram webhook: %v", err)
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
		log.Fatalf("failed to start server: %v", err)
	}
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

	webhookPath := cfg.Telegram.WebhookPath
	if webhookPath == "" {
		webhookPath = "/telegram/webhook"
	} else if !strings.HasPrefix(webhookPath, "/") {
		webhookPath = "/" + webhookPath
	}

	ctxSet, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := bot.ConfigureWebhook(ctxSet, telegram.WebhookConfig{
		URL:                cfg.Telegram.WebhookURL,
		SecretToken:        cfg.Telegram.WebhookSecret,
		DropPendingUpdates: true,
	}); err != nil {
		return err
	}

	r.Post(webhookPath, bot.WebhookHandler())

	go func() {
		log.Printf("telegram webhook active on path %s (target %s)", webhookPath, cfg.Telegram.WebhookURL)
		bot.StartWebhook(ctx)
	}()

	return nil
}

func mustConnectPostgres(dsn string) *sql.DB {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatalf("failed to open postgres connection: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("failed to ping postgres: %v", err)
	}

	return db
}
