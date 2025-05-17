package api

import (
	"MovingCompanyGo/config/service"
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"time"
)

type TokenHandler struct {
	tokenService *service.JWTTokenService
}

func NewTokenHandler(tokenService *service.JWTTokenService) *TokenHandler {
	return &TokenHandler{
		tokenService: tokenService,
	}
}

// GenerateBookingLink creates a new booking form link
func (h *TokenHandler) GenerateBookingLink(w http.ResponseWriter, r *http.Request) {

	// Only allow GET method
	if r.Method != http.MethodGet {
		log.Printf("Error occurred: %v", r.Method)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Generate a token with 24 hour expiry
	tokenID, tokenString, err := h.tokenService.GenerateToken(24 * time.Hour)
	if err != nil {
		log.Printf("Error occurred: %v", err)
		http.Error(w, "Failed to generate booking link", http.StatusInternalServerError)
		return
	}

	// Create the secure link
	baseURL := "http://localhost:8080/booking-form"
	params := url.Values{}
	params.Add("id", tokenID)
	params.Add("token", tokenString)
	bookingURL := baseURL + "?" + params.Encode()

	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)
	encoder.Encode(map[string]interface{}{
		"booking_url": bookingURL,
		"expires_at":  time.Now().Add(24 * time.Hour),
	})
}

// RenderBookingForm validates the token and renders the form
func (h *TokenHandler) RenderBookingForm(w http.ResponseWriter, r *http.Request) {
	// Only allow GET method
	if r.Method != http.MethodGet {
		log.Printf("Error occurred: %v", r.Method)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	tokenID := r.URL.Query().Get("id")
	tokenString := r.URL.Query().Get("token")

	if tokenID == "" || tokenString == "" {
		http.Error(w, "Invalid booking link", http.StatusBadRequest)
		return
	}

	// Validate the token
	validatedTokenID, valid, err := h.tokenService.ValidateToken(tokenString)
	if err != nil || !valid {
		http.Error(w, "This booking link is invalid or has already been used", http.StatusUnauthorized)
		return
	}

	// Parse the index page instead of booking form
	tmpl, err := template.ParseFiles("static/index.html")
	if err != nil {
		http.Error(w, "Error loading index page: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Pass the token data to the booking form
	data := struct {
		TokenID     string
		TokenString string
	}{
		TokenID:     validatedTokenID,
		TokenString: tokenString,
	}

	// Render the booking form with the token data
	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, "Error rendering booking form: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

// SubmitBookingForm processes the form and consumes the token
func (h *TokenHandler) SubmitBookingForm(w http.ResponseWriter, r *http.Request) {
	// Only allow POST method
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Error parsing form data", http.StatusBadRequest)
		return
	}

	tokenString := r.FormValue("token")

	// Consume the token
	valid, err := h.tokenService.ConsumeToken(tokenString)
	if err != nil || !valid {
		http.Error(w, "Invalid booking link or this form has already been submitted", http.StatusUnauthorized)
		return
	}

	// Process the booking form data
	// ...

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Your booking has been successfully submitted",
	})
}
