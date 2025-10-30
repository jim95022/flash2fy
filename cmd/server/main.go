package main

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "github.com/jackc/pgx/v5/stdlib"

	cardhttp "flash2fy/internal/adapters/http/card"
	cardstorage "flash2fy/internal/adapters/storage/card"
	cardtelegram "flash2fy/internal/adapters/telegram/cardbot"
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

	if token := cfg.Telegram.BotToken; token != "" {
		bot, err := cardtelegram.New(token, service)
		if err != nil {
			log.Fatalf("failed to initialize telegram bot: %v", err)
		}
		go func() {
			log.Println("telegram bot listening for /add commands")
			bot.Start(ctx)
		}()
	} else {
		log.Println("telegram bot disabled (TELEGRAM_BOT_TOKEN not set)")
	}

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

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
