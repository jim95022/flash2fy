package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "github.com/jackc/pgx/v5/stdlib"

	cardhttp "flash2fy/internal/adapters/http/card"
	cardstorage "flash2fy/internal/adapters/storage/card"
	cardapp "flash2fy/internal/application/card"
)

func main() {
	db := mustConnectPostgres()
	defer db.Close()

	repo := cardstorage.NewPostgresRepository(db)
	service := cardapp.NewCardService(repo)
	handler := cardhttp.NewHandler(service)

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Mount("/v1/cards", handler.Routes())

	log.Println("card service listening on :8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
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
