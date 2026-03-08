package api

import (
	"MovingCompanyGo/config/service"
	"MovingCompanyGo/models"
	"MovingCompanyGo/repository"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

// rateLimiter implements a simple per-IP sliding-window rate limiter.
type rateLimiter struct {
	mu       sync.Mutex             // protects the requests map against concurrent writes
	requests map[string][]time.Time // key: client IP, value: timestamps of requests in the current window
	limit    int                    // maximum requests allowed within the window
	window   time.Duration          // size of the sliding time window
}

// newRateLimiter constructs a rateLimiter with the given limit and window duration.
func newRateLimiter(limit int, window time.Duration) *rateLimiter {
	return &rateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}
}

// allow returns true when the request from ip is within the rate limit.
// It discards timestamps outside the current window before checking the count.
func (rl *rateLimiter) allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	windowStart := now.Add(-rl.window)

	existing := rl.requests[ip]
	var recent []time.Time
	for _, t := range existing {
		if t.After(windowStart) { // keep only timestamps still inside the window
			recent = append(recent, t)
		}
	}

	if len(recent) >= rl.limit { // request count has hit the ceiling for this window
		rl.requests[ip] = recent // store the cleaned slice even when rejecting to avoid unbounded growth
		return false
	}

	rl.requests[ip] = append(recent, now) // record this request and allow it through
	return true
}

// getClientIP extracts the real client IP, preferring the X-Forwarded-For header
// set by reverse proxies (nginx, Cloudflare, etc.) over the raw RemoteAddr.
func getClientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.SplitN(xff, ",", 2)
		return strings.TrimSpace(parts[0])
	}
	ip := r.RemoteAddr
	if i := strings.LastIndex(ip, ":"); i >= 0 {
		ip = ip[:i]
	}
	return ip
}

// limitMiddleware wraps a handler with IP-based rate limiting using this rateLimiter.
func (rl *rateLimiter) limitMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ip := getClientIP(r)
		if !rl.allow(ip) {
			http.Error(w, "Too many requests", http.StatusTooManyRequests) // 429 signals the client to back off
			return
		}
		next(w, r)
	}
}

// bookingFetchResult carries the result of an async booking DB query.
type bookingFetchResult struct {
	booking *models.Booking
	err     error
}

// furnitureFetchResult carries the result of an async furniture DB query.
type furnitureFetchResult struct {
	items *models.FurnitureItem
	err   error
}

// SetupHTTPRoutes registers all application routes and returns the configured ServeMux.
func SetupHTTPRoutes(repo repository.BookingRepository, tokenService service.TokenService) *http.ServeMux {

	// This prevents DB flooding via the booking form and exhaustion of one-time tokens.
	publicLimiter := newRateLimiter(10, time.Minute)

	// Handler for generating and streaming a PDF invoice.
	generateInvoiceHandler := RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		bookingID := r.URL.Query().Get("booking_id")
		if bookingID == "" {
			http.Error(w, "booking_id parameter is required", http.StatusBadRequest)
			return
		}

		totalAmountStr := r.URL.Query().Get("total_amount")
		hoursUsedStr := r.URL.Query().Get("hours_used")
		jobDescription := r.URL.Query().Get("job_description")

		var totalAmount float64 = 340.00 // sensible default when the parameter is absent
		if totalAmountStr != "" {
			parsed, err := strconv.ParseFloat(totalAmountStr, 64)
			if err != nil {
				http.Error(w, "total_amount must be a valid number", http.StatusBadRequest)
				return
			}
			totalAmount = parsed
		}

		var hoursUsed float64 = 2.0
		if hoursUsedStr != "" {
			parsed, err := strconv.ParseFloat(hoursUsedStr, 64)
			if err != nil {
				http.Error(w, "hours_used must be a valid number", http.StatusBadRequest)
				return
			}
			hoursUsed = parsed
		}

		bCh := make(chan bookingFetchResult, 1)   // buffered so the goroutine never blocks on send
		fCh := make(chan furnitureFetchResult, 1) // buffered for the same reason

		go func() {
			b, err := repo.GetByID(r.Context(), bookingID) // r.Context() propagates cancellation if the client disconnects
			bCh <- bookingFetchResult{b, err}              // send result (or error) to the channel
		}()

		go func() {
			f, err := repo.GetFurnitureItemsByBookingID(r.Context(), bookingID)
			fCh <- furnitureFetchResult{f, err}
		}()

		br := <-bCh
		fr := <-fCh

		if br.err != nil {
			log.Printf("generate-invoice: booking fetch failed for %s: %v", bookingID, br.err)
			http.Error(w, "Failed to fetch booking data", http.StatusNotFound)
			return
		}
		if fr.err != nil {
			log.Printf("generate-invoice: furniture fetch failed for %s: %v", bookingID, fr.err)
			http.Error(w, "Failed to fetch booking data", http.StatusNotFound)
			return
		}

		pdfBytes, err := service.GenerateSampleInvoice(br.booking, fr.items, totalAmount, hoursUsed, jobDescription)
		if err != nil {
			log.Printf("generate-invoice: PDF generation failed: %v", err)
			http.Error(w, "Failed to generate invoice", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/pdf")
		w.Header().Set("Content-Disposition", "attachment; filename=invoice.pdf")
		w.WriteHeader(http.StatusOK)
		w.Write(pdfBytes)
	})

	mux := http.NewServeMux()

	// Serve static assets (CSS, JS) from the ./static directory
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
	bookingHandler := NewBookingHandler(repo, tokenService)
	authHandler := NewAuthHandler()

	tokenHandler := NewTokenHandler(tokenService)

	// Authentication routes
	mux.HandleFunc("/login", authHandler.LoginPageHandler)
	mux.HandleFunc("/api/auth/login", authHandler.AuthenticateHandler)

	// Admin routes — protected by RequireAuth
	mux.HandleFunc("/admin", RequireAuth(authHandler.AdminPageHandler))
	mux.HandleFunc("/admin/generate-invoice", generateInvoiceHandler)

	// Public booking form routes — rate limited to prevent token exhaustion and DB flooding (S10)
	mux.HandleFunc("/api/generate-link", RequireAuth(tokenHandler.GenerateBookingLink)) // generate-link is admin-only
	mux.HandleFunc("/booking-form", publicLimiter.limitMiddleware(tokenHandler.RenderBookingForm))

	mux.HandleFunc("/api/submit-booking", publicLimiter.limitMiddleware(bookingHandler.CreateBooking)) // S10: rate limited

	// Booking API routes — all require authentication
	mux.HandleFunc("/api/bookings/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/api/bookings/")

		handler := RequireAuth(func(w http.ResponseWriter, r *http.Request) {
			if path == "" {
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

			// Trailing ID segment — item-level operations
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
