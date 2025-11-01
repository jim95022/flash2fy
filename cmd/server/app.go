package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	cardhttp "flash2fy/internal/adapters/http/card"
	cardstorage "flash2fy/internal/adapters/storage/card"
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
