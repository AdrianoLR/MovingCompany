package api

import (
	"MovingCompanyGo/config/service"
	"MovingCompanyGo/models"
	"MovingCompanyGo/repository"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/mail"
	"strings"
	"time"

	"github.com/google/uuid"
)

// BookingHandler handles HTTP requests for booking operations.
type BookingHandler struct {
	repository   repository.BookingRepository
	tokenService service.TokenService // R3: depend on the interface, not the concrete *JWTTokenService
}

// NewBookingHandler constructs a BookingHandler.
func NewBookingHandler(repository repository.BookingRepository, tokenService service.TokenService) *BookingHandler {
	return &BookingHandler{
		repository:   repository,
		tokenService: tokenService,
	}
}

type createBookingRequest struct {
	CustomerName   string                `json:"customer_name"`
	Email          string                `json:"email"`
	Phone          string                `json:"phone"`
	PickupAddress  string                `json:"pickup_address"`
	DropAddress    string                `json:"drop_address"`
	PickupDate     string                `json:"pickup_date"`
	FurnitureItems *models.FurnitureItem `json:"furniture_items"`
	Token          string                `json:"token"`
}

// validateCreateBookingRequest checks required fields and enforces maximum lengths.
func validateCreateBookingRequest(req *createBookingRequest) error {
	if strings.TrimSpace(req.CustomerName) == "" {
		return fmt.Errorf("customer_name is required")
	}
	if len(req.CustomerName) > 100 {
		return fmt.Errorf("customer_name must be 100 characters or fewer")
	}

	if strings.TrimSpace(req.Email) == "" {
		return fmt.Errorf("email is required")
	}
	if _, err := mail.ParseAddress(req.Email); err != nil {
		return fmt.Errorf("email is invalid")
	}
	if len(req.Email) > 200 {
		return fmt.Errorf("email must be 200 characters or fewer")
	}

	if strings.TrimSpace(req.Phone) == "" {
		return fmt.Errorf("phone is required")
	}
	if len(req.Phone) > 20 {
		return fmt.Errorf("phone must be 20 characters or fewer")
	}

	if strings.TrimSpace(req.PickupAddress) == "" {
		return fmt.Errorf("pickup_address is required")
	}

	if len(req.PickupAddress) > 300 {
		return fmt.Errorf("pickup_address must be 300 characters or fewer")
	}

	if strings.TrimSpace(req.DropAddress) == "" {
		return fmt.Errorf("drop_address is required")
	}
	if len(req.DropAddress) > 300 {
		return fmt.Errorf("drop_address must be 300 characters or fewer")
	}

	if strings.TrimSpace(req.PickupDate) == "" {
		return fmt.Errorf("pickup_date is required")
	}

	return nil
}

// CreateBooking handles POST requests to create a new booking.
func (h *BookingHandler) CreateBooking(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req createBookingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// S9: Validate all fields before using them — reject bad input at the boundary
	if err := validateCreateBookingRequest(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	pickupDate, err := time.Parse("2006-01-02T15:04:05", req.PickupDate)
	if err != nil {
		http.Error(w, "Invalid pickup date format. Use YYYY-MM-DDThh:mm:ss", http.StatusBadRequest)
		return
	}

	// Consume the one-time token before creating the booking
	if req.Token != "" && h.tokenService != nil {
		valid, err := h.tokenService.ConsumeToken(r.Context(), req.Token)
		if err != nil || !valid {
			http.Error(w, "Invalid or already used token", http.StatusUnauthorized)
			return
		}
	}

	booking := &models.Booking{
		BookingID:     uuid.New().String(),
		CustomerName:  req.CustomerName,
		Email:         req.Email,
		Phone:         req.Phone,
		PickupAddress: req.PickupAddress,
		DropAddress:   req.DropAddress,
		PickupDate:    pickupDate,
		Status:        int(models.StatusPending),
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := h.repository.Create(r.Context(), booking, req.FurnitureItems); err != nil {
		log.Printf("CreateBooking: repository error: %v", err)
		http.Error(w, "Failed to create booking", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(booking)
}

// GetBooking handles GET requests to retrieve a single booking by ID.
func (h *BookingHandler) GetBooking(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	bookingID := r.URL.Query().Get("id")
	if bookingID == "" {
		http.Error(w, "Booking ID is required", http.StatusBadRequest)
		return
	}

	booking, err := h.repository.GetByID(r.Context(), bookingID)
	if err != nil {
		log.Printf("GetBooking: repository error for id %s: %v", bookingID, err)
		http.Error(w, "Booking not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(booking)
}

// UpdateBooking handles PUT requests to update an existing booking.
func (h *BookingHandler) UpdateBooking(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var booking models.Booking
	if err := json.NewDecoder(r.Body).Decode(&booking); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	booking.UpdatedAt = time.Now()

	if err := h.repository.Update(r.Context(), &booking); err != nil {
		log.Printf("UpdateBooking: repository error: %v", err)
		http.Error(w, "Failed to update booking", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(booking)
}

// DeleteBooking handles DELETE requests to remove a booking by ID.
func (h *BookingHandler) DeleteBooking(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	bookingID := r.URL.Query().Get("id")
	if bookingID == "" {
		http.Error(w, "Booking ID is required", http.StatusBadRequest)
		return
	}

	if err := h.repository.Delete(r.Context(), bookingID); err != nil {
		log.Printf("DeleteBooking: repository error for id %s: %v", bookingID, err)
		http.Error(w, "Failed to delete booking", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

// ListBookings handles GET requests to return all bookings.
func (h *BookingHandler) ListBookings(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	bookings, err := h.repository.List(r.Context())
	if err != nil {
		log.Printf("ListBookings: repository error: %v", err)
		http.Error(w, "Failed to list bookings", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(bookings)
}
