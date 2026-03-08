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
	// Load environment variables from .env — must happen before any os.Getenv calls
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found, relying on system environment variables")
	}

	// It now returns an error for missing credentials instead of panicking (S2).
	if err := config.InitSupabase(); err != nil {
		log.Fatalf("Failed to initialise Supabase client: %v", err)
	}

	// Logging middleware logs the method, path, and elapsed time for every request
	loggingMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			log.Printf("%s %s", r.Method, r.URL.Path)
			next.ServeHTTP(w, r)
			log.Printf("%s %s completed in %v", r.Method, r.URL.Path, time.Since(start))
		})
	}

	// Token repository uses the admin client so it can bypass RLS for token lookups
	tokenRepo := repository.NewSupabaseTokenRepository(config.SupabaseAdminClient, "booking_tokens")

	// Token service wraps JWT creation/validation with DB-backed one-time-use enforcement
	tokenService := service.NewJWTTokenService(os.Getenv("JWT_SECRET_KEY"), tokenRepo)

	// Booking repository uses the regular (RLS-enforced) client
	repo := repository.NewSupabaseBookingRepository()

	// R1: SetupHTTPRoutes now registers ALL routes including token routes internally.
	// The TokenHandler is created inside SetupHTTPRoutes, so main.go no longer needs
	// to create it or call mux.HandleFunc for /api/generate-link and /booking-form.
	mux := api.SetupHTTPRoutes(repo, tokenService)

	// Root catch-all — returns a simple health-check string for any unmatched path
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("The Furniture Man Moving Houses"))
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // default port for local development
	}

	handler := loggingMiddleware(mux) // wrap the mux with request logging

	log.Printf("Starting server on :%s", port)
	if err := http.ListenAndServe(":"+port, handler); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
