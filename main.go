package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"MovingCompanyGo/api"
	"MovingCompanyGo/config"
	"MovingCompanyGo/config/service"
	"MovingCompanyGo/repository"

	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found")
	}

	// Initialize Supabase
	if err := config.InitSupabase(); err != nil {
		log.Fatalf("Failed to initialize Supabase client: %v", err)
	}

	// Create HTTP server with logging middleware
	loggingMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			log.Printf("%s %s", r.Method, r.URL.Path)
			next.ServeHTTP(w, r)
			log.Printf("%s %s completed in %v", r.Method, r.URL.Path, time.Since(start))
		})
	}

	// Initialize token repository
	tokenRepo := repository.NewSupabaseTokenRepository(config.SupabaseClient, "booking_tokens") // "tokens" is the table name in Supabase

	// Initialize token service
	tokenService := service.NewJWTTokenService(os.Getenv("JWT_SECRET_KEY"), tokenRepo)

	// Initialize repository
	repo := repository.NewSupabaseBookingRepository()

	// Setup API routes
	mux := api.SetupHTTPRoutes(repo, tokenService)

	// Create token handler
	tokenHandler := api.NewTokenHandler(tokenService)

	// Add token routes
	mux.HandleFunc("/api/generate-link", tokenHandler.GenerateBookingLink)
	mux.HandleFunc("/booking-form", tokenHandler.RenderBookingForm)

	// Serve static files
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("The Furniture Man Moving Houses"))
	})

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Apply logging middleware
	handler := loggingMiddleware(mux)

	log.Printf("Starting server on :%s", port)
	if err := http.ListenAndServe(":"+port, handler); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
