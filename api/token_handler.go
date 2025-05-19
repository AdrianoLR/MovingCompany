package api

import (
	"MovingCompanyGo/config/service"
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"html/template"
	"io"
	"log"
	"net/http"
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

	// Create a compact token representation
	tokenData := map[string]string{
		"id":    tokenID,
		"token": tokenString,
	}

	// Convert to JSON
	jsonData, err := json.Marshal(tokenData)
	if err != nil {
		log.Printf("Error marshaling token data: %v", err)
		http.Error(w, "Failed to generate booking link", http.StatusInternalServerError)
		return
	}

	// Compress the data
	var compressedData bytes.Buffer
	gzipWriter := gzip.NewWriter(&compressedData)
	if _, err := gzipWriter.Write(jsonData); err != nil {
		log.Printf("Error compressing token data: %v", err)
		http.Error(w, "Failed to generate booking link", http.StatusInternalServerError)
		return
	}
	if err := gzipWriter.Close(); err != nil {
		log.Printf("Error closing gzip writer: %v", err)
		http.Error(w, "Failed to generate booking link", http.StatusInternalServerError)
		return
	}

	// Encode to base64url
	encodedToken := base64.RawURLEncoding.EncodeToString(compressedData.Bytes())

	// Create the secure link with a single parameter
	baseURL := "http://localhost:8080/booking-form"
	bookingURL := baseURL + "?t=" + encodedToken

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

	encodedToken := r.URL.Query().Get("t")
	if encodedToken == "" {
		http.Error(w, "Invalid booking link", http.StatusBadRequest)
		return
	}

	// Decode the token
	compressedData, err := base64.RawURLEncoding.DecodeString(encodedToken)
	if err != nil {
		log.Printf("Error decoding token: %v", err)
		http.Error(w, "Invalid booking link", http.StatusBadRequest)
		return
	}

	// Decompress the data
	gzipReader, err := gzip.NewReader(bytes.NewReader(compressedData))
	if err != nil {
		log.Printf("Error creating gzip reader: %v", err)
		http.Error(w, "Invalid booking link", http.StatusBadRequest)
		return
	}

	jsonData, err := io.ReadAll(gzipReader)
	if err != nil {
		log.Printf("Error decompressing token data: %v", err)
		http.Error(w, "Invalid booking link", http.StatusBadRequest)
		return
	}
	if err := gzipReader.Close(); err != nil {
		log.Printf("Error closing gzip reader: %v", err)
		http.Error(w, "Invalid booking link", http.StatusBadRequest)
		return
	}

	// Parse the JSON
	var tokenData map[string]string
	if err := json.Unmarshal(jsonData, &tokenData); err != nil {
		log.Printf("Error unmarshaling token data: %v", err)
		http.Error(w, "Invalid booking link", http.StatusBadRequest)
		return
	}

	tokenID := tokenData["id"]
	tokenString := tokenData["token"]

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
