package main

import (
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi"
	chimw "github.com/go-chi/chi/middleware"

	"urlShortener/internal/database"
	"urlShortener/internal/handlers"
)

func main() {
	// ensure data dir exists
	_ = os.MkdirAll("data", 0o755)

	// init DB
	store, err := database.Open("data/urls.db")
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer store.Close()

	if err := store.Migrate(); err != nil {
		log.Fatalf("migrate: %v", err)
	}

	// wire handlers
	h := handlers.New(store)

	// router
	r := chi.NewRouter()
	r.Use(chimw.RealIP)
	r.Use(chimw.RequestID)
	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)
	r.Use(chimw.StripSlashes) // <- your request

	// routes
	r.Get("/", h.Home)           // show form
	r.Post("/shorten", h.Create) // create short URL
	r.Get("/{code}", h.Redirect) // redirect

	addr := ":8080"
	log.Printf("listening on %s", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatal(err)
	}
}
