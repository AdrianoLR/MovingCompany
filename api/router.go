package api

import (
	"MovingCompanyGo/config/service"
	"MovingCompanyGo/repository"
	"net/http"
	"strings"
)

// Simple authentication middleware to restrict access to API endpoints
/*
func adminAuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Check if the request is coming from our admin page
		referer := r.Header.Get("Referer")

		// Extract the path from the referer URL
		refererPath := ""
		if referer != "" {
			if parsedURL, err := url.Parse(referer); err == nil {
				refererPath = parsedURL.Path
			}
		}

		// Allow access if the request comes from the admin page or is a POST request
		if refererPath == "/admin" || r.Method == http.MethodPost {
			next(w, r)
			return
		}

		// If not from admin and not a POST request, deny access
		http.Error(w, "Unauthorized: Access denied", http.StatusUnauthorized)
	}
}*/

// SetupHTTPRoutes sets up the HTTP routes for the application
func SetupHTTPRoutes(repo repository.BookingRepository, tokenService *service.JWTTokenService) *http.ServeMux {
	mux := http.NewServeMux()
	bookingHandler := NewBookingHandler(repo, tokenService)
	authHandler := NewAuthHandler()

	// Authentication routes
	mux.HandleFunc("/login", authHandler.LoginPageHandler)
	mux.HandleFunc("/api/auth/login", authHandler.AuthenticateHandler)

	// Admin page route with authentication
	mux.HandleFunc("/admin", RequireAuth(authHandler.AdminPageHandler))

	// Public routes
	mux.HandleFunc("/api/submit-booking", bookingHandler.CreateBooking)
	mux.HandleFunc("/submit-booking", bookingHandler.CreateBooking)

	// Booking routes with authentication
	mux.HandleFunc("/api/bookings/", func(w http.ResponseWriter, r *http.Request) {
		// Extract the path to determine which handler to call
		path := strings.TrimPrefix(r.URL.Path, "/api/bookings/")

		// Apply admin authentication middleware
		handler := RequireAuth(func(w http.ResponseWriter, r *http.Request) {
			if path == "" {
				// Handle collection endpoints
				switch r.Method {
				case http.MethodGet:
					bookingHandler.ListBookings(w, r)
				case http.MethodPost:
					bookingHandler.CreateBooking(w, r)
				default:
					http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				}
				return
			}

			// Handle individual booking endpoints
			switch r.Method {
			case http.MethodGet:
				bookingHandler.GetBooking(w, r)
			case http.MethodPut:
				bookingHandler.UpdateBooking(w, r)
			case http.MethodDelete:
				bookingHandler.DeleteBooking(w, r)
			default:
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		})

		handler(w, r)
	})

	return mux
}
