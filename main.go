package main

import (
	"log"
	"net/http"
	"os"

	"simpleTui/internal/routes"
	"simpleTui/internal/storage"
)

func main() {
	// Read App ID from env. (Put your real ID in env: 3234342324324)
	appID := os.Getenv("OPENEXCHANGERATES_APP_ID")
	if appID == "" {
		log.Fatal("OPENEXCHANGERATES_APP_ID is not set")
	}

	// Create the rates store (fetches & caches OXR rates)
	store := storage.NewRatesStore(appID, "")

	r := routes.NewRouter(store)

	addr := ":8080"
	log.Printf("listening on %s …", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatal(err)
	}
}
