package api

import (
	"MovingCompanyGo/config/service"
	"MovingCompanyGo/models"
	"MovingCompanyGo/repository"
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type BookingHandler struct {
	repository   repository.BookingRepository
	tokenService *service.JWTTokenService
}

func NewBookingHandler(repository repository.BookingRepository, tokenService *service.JWTTokenService) *BookingHandler {
	return &BookingHandler{repository: repository, tokenService: tokenService}
}

func (h *BookingHandler) CreateBooking(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		CustomerName   string                `json:"customer_name"`
		Email          string                `json:"email"`
		Phone          string                `json:"phone"`
		PickupAddress  string                `json:"pickup_address"`
		DropAddress    string                `json:"drop_address"`
		PickupDate     string                `json:"pickup_date"`
		FurnitureItems *models.FurnitureItem `json:"furniture_items"`
		Token          string                `json:"token"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	pickupDate, err := time.Parse("2006-01-02T15:04:05", req.PickupDate)
	if err != nil {
		http.Error(w, "Invalid pickup date format. Use YYYY-MM-DDThh:mm:ss", http.StatusBadRequest)
		return
	}

	booking := &models.Booking{
		UserID:        uuid.New().String(),
		CustomerName:  req.CustomerName,
		Email:         req.Email,
		Phone:         req.Phone,
		PickupAddress: req.PickupAddress,
		DropAddress:   req.DropAddress,
		PickupDate:    pickupDate,
		Status:        models.NewBooking().Status,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	// Consume token if provided
	if req.Token != "" && h.tokenService != nil {
		// Validate and consume the token
		valid, err := h.tokenService.ConsumeToken(req.Token)
		if err != nil || !valid {
			http.Error(w, "Invalid or already used token", http.StatusUnauthorized)
			return
		}
	}

	if err := h.repository.Create(r.Context(), booking, req.FurnitureItems); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(booking)
}

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
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(booking)
}

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
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(booking)
}

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
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

func (h *BookingHandler) ListBookings(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	bookings, err := h.repository.List(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(bookings)
}
