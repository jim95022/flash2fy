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
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	db := mustConnectPostgres()
	defer db.Close()

	repo := cardstorage.NewPostgresRepository(db)
	service := cardapp.NewCardService(repo)
	handler := cardhttp.NewHandler(service)

	if token := os.Getenv("TELEGRAM_BOT_TOKEN"); token != "" {
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
		Addr:    ":8080",
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

	log.Println("card service listening on :8080")
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("failed to start server: %v", err)
	}
}

func mustConnectPostgres() *sql.DB {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://postgres:postgres@localhost:5432/flash2fy?sslmode=disable"
	}

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
