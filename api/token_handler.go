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
	"os"   // S7: used to read BASE_URL from environment instead of hardcoding localhost
	"time"
)

// TokenHandler manages one-time booking link generation and form rendering.
type TokenHandler struct {
	tokenService service.TokenService // R3: depend on the interface, not the concrete *JWTTokenService
}

// NewTokenHandler constructs a TokenHandler.
// R3: tokenService is the TokenService interface for decoupling and testability.
func NewTokenHandler(tokenService service.TokenService) *TokenHandler {
	return &TokenHandler{
		tokenService: tokenService,
	}
}

// GenerateBookingLink creates a new one-time booking link and returns it as JSON.
func (h *TokenHandler) GenerateBookingLink(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		log.Printf("GenerateBookingLink: unexpected method %s", r.Method)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// R4: pass the request context so the DB insert can respect the request deadline
	tokenID, tokenString, err := h.tokenService.GenerateToken(r.Context(), 24*time.Hour)
	if err != nil {
		log.Printf("GenerateBookingLink: token generation failed: %v", err) // S8: full error to logs
		http.Error(w, "Failed to generate booking link", http.StatusInternalServerError)
		return
	}

	// Pack both the token ID and the signed JWT into a compact, URL-safe payload
	tokenData := map[string]string{
		"id":    tokenID,    // used to look up the DB record on validation
		"token": tokenString, // the signed JWT that the booking form will submit
	}

	jsonData, err := json.Marshal(tokenData)
	if err != nil {
		log.Printf("GenerateBookingLink: JSON marshal failed: %v", err)
		http.Error(w, "Failed to generate booking link", http.StatusInternalServerError)
		return
	}

	// Gzip the JSON to reduce URL length before base64 encoding
	var compressedData bytes.Buffer
	gzipWriter := gzip.NewWriter(&compressedData)
	if _, err := gzipWriter.Write(jsonData); err != nil {
		log.Printf("GenerateBookingLink: gzip write failed: %v", err)
		http.Error(w, "Failed to generate booking link", http.StatusInternalServerError)
		return
	}
	if err := gzipWriter.Close(); err != nil { // Close must be called to flush the gzip trailer
		log.Printf("GenerateBookingLink: gzip close failed: %v", err)
		http.Error(w, "Failed to generate booking link", http.StatusInternalServerError)
		return
	}

	// Base64url-encode the compressed bytes so they are safe to embed in a URL query parameter
	encodedToken := base64.RawURLEncoding.EncodeToString(compressedData.Bytes())

	// S7: Read BASE_URL from env so the link works in any environment.
	// Previously this was hardcoded to "http://localhost:8080" which produced broken links in production.
	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080" // safe fallback for local development only
	}
	bookingURL := baseURL + "/booking-form?t=" + encodedToken

	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false) // prevent the URL's & from being escaped to \u0026 in the JSON output
	encoder.Encode(map[string]interface{}{
		"booking_url": bookingURL,
		"expires_at":  time.Now().Add(24 * time.Hour),
	})
}

// RenderBookingForm validates the one-time token embedded in the URL and renders the booking form.
func (h *TokenHandler) RenderBookingForm(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		log.Printf("RenderBookingForm: unexpected method %s", r.Method)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	encodedToken := r.URL.Query().Get("t")
	if encodedToken == "" {
		http.Error(w, "Invalid booking link", http.StatusBadRequest)
		return
	}

	// Reverse the encoding: base64url → gzip-compressed bytes → JSON
	compressedData, err := base64.RawURLEncoding.DecodeString(encodedToken)
	if err != nil {
		log.Printf("RenderBookingForm: base64 decode failed: %v", err) // S8: full error to logs
		http.Error(w, "Invalid booking link", http.StatusBadRequest)
		return
	}

	gzipReader, err := gzip.NewReader(bytes.NewReader(compressedData))
	if err != nil {
		log.Printf("RenderBookingForm: gzip reader failed: %v", err)
		http.Error(w, "Invalid booking link", http.StatusBadRequest)
		return
	}

	jsonData, err := io.ReadAll(gzipReader)
	if err != nil {
		log.Printf("RenderBookingForm: gzip read failed: %v", err)
		http.Error(w, "Invalid booking link", http.StatusBadRequest)
		return
	}
	if err := gzipReader.Close(); err != nil { // Close flushes and checks the gzip checksum
		log.Printf("RenderBookingForm: gzip close failed: %v", err)
		http.Error(w, "Invalid booking link", http.StatusBadRequest)
		return
	}

	var tokenData map[string]string
	if err := json.Unmarshal(jsonData, &tokenData); err != nil {
		log.Printf("RenderBookingForm: JSON unmarshal failed: %v", err)
		http.Error(w, "Invalid booking link", http.StatusBadRequest)
		return
	}

	tokenID := tokenData["id"]
	tokenString := tokenData["token"]

	if tokenID == "" || tokenString == "" {
		http.Error(w, "Invalid booking link", http.StatusBadRequest) // payload is missing required fields
		return
	}

	// R4: pass the request context so the DB lookup can respect the request deadline
	validatedTokenID, valid, err := h.tokenService.ValidateToken(r.Context(), tokenString)
	if err != nil || !valid {
		http.Error(w, "This booking link is invalid or has already been used", http.StatusUnauthorized)
		return
	}

	tmpl, err := template.ParseFiles("static/index.html")
	if err != nil {
		log.Printf("RenderBookingForm: template parse failed: %v", err) // S8: full error to logs
		http.Error(w, "Error loading booking form", http.StatusInternalServerError) // S8: generic to client
		return
	}

	// Inject the token data so the form can include it in the booking submission
	data := struct {
		TokenID     string
		TokenString string
	}{
		TokenID:     validatedTokenID,
		TokenString: tokenString,
	}

	if err := tmpl.Execute(w, data); err != nil {
		log.Printf("RenderBookingForm: template execute failed: %v", err) // S8: full error to logs
		http.Error(w, "Error rendering booking form", http.StatusInternalServerError) // S8: generic to client
		return
	}
}

// R6: SubmitBookingForm removed — it was never registered in the router and contained only
// a placeholder comment ("// ...") with no real implementation. Token consumption is already
// handled inside BookingHandler.CreateBooking via the Token field in the request body.
