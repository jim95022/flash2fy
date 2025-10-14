package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	cardhttp "flash2fy/internal/adapters/http"
	"flash2fy/internal/adapters/storage"
	"flash2fy/internal/application"
)

func main() {
	repo := storage.NewMemoryCardRepository()
	service := application.NewCardService(repo)
	handler := cardhttp.NewCardHandler(service)

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Mount("/cards", handler.Routes())

	log.Println("card service listening on :8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
