package api

import (
	"MovingCompanyGo/config/service"
	"MovingCompanyGo/repository"
	"net/http"
	"strconv"
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

	// Handler for generating invoice
	generateInvoiceHandler := RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		// Get booking ID from query parameter
		bookingID := r.URL.Query().Get("booking_id")
		if bookingID == "" {
			http.Error(w, "booking_id parameter is required", http.StatusBadRequest)
			return
		}

		// Get manual pricing parameters
		totalAmountStr := r.URL.Query().Get("total_amount")
		hoursUsedStr := r.URL.Query().Get("hours_used")
		jobDescription := r.URL.Query().Get("job_description")

		// Parse total amount
		var totalAmount float64 = 340.00 // default
		if totalAmountStr != "" {
			if parsed, err := strconv.ParseFloat(totalAmountStr, 64); err == nil {
				totalAmount = parsed
			}
		}

		// Parse hours used
		var hoursUsed float64 = 2.0 // default
		if hoursUsedStr != "" {
			if parsed, err := strconv.ParseFloat(hoursUsedStr, 64); err == nil {
				hoursUsed = parsed
			}
		}

		// Fetch the booking data
		booking, err := repo.GetByID(r.Context(), bookingID)
		if err != nil {
			http.Error(w, "Failed to fetch booking: "+err.Error(), http.StatusNotFound)
			return
		}

		// Fetch the furniture items data
		furnitureItems, err := repo.GetFurnitureItemsByBookingID(r.Context(), bookingID)
		if err != nil {
			http.Error(w, "Failed to fetch furniture items: "+err.Error(), http.StatusNotFound)
			return
		}

		pdfBytes, err := service.GenerateSampleInvoice(booking, furnitureItems, totalAmount, hoursUsed, jobDescription)
		if err != nil {
			http.Error(w, "Failed to generate invoice: "+err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/pdf")
		w.Header().Set("Content-Disposition", "attachment; filename=invoice.pdf")
		w.WriteHeader(http.StatusOK)
		w.Write(pdfBytes)
	})

	mux := http.NewServeMux()
	bookingHandler := NewBookingHandler(repo, tokenService)
	authHandler := NewAuthHandler()

	// Authentication routes
	mux.HandleFunc("/login", authHandler.LoginPageHandler)
	mux.HandleFunc("/api/auth/login", authHandler.AuthenticateHandler)

	// Admin page route with authentication
	mux.HandleFunc("/admin", RequireAuth(authHandler.AdminPageHandler))
	mux.HandleFunc("/admin/generate-invoice", generateInvoiceHandler)

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
