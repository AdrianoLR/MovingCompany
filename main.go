package main

import (
	"log"
	"net/http"

	"MovingCompanyGo/api"
	"MovingCompanyGo/handlers"
	"MovingCompanyGo/repository"
)

func main() {
	// Initialize repository
	repo := repository.NewInMemoryBookingRepository()

	// Initialize handler
	handler := handlers.NewBookingHandler(repo)

	// Setup routes
	mux := api.SetupRoutes(handler)

	// Serve static files
	fs := http.FileServer(http.Dir("static"))
	mux.Handle("/", fs)

	// Start server
	log.Println("Starting server on :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
