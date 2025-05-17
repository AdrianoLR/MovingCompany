package models

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
)

type Booking struct {
	UserID        string    `json:"user_id" db:"user_id"`
	CustomerName  string    `json:"customer_name" db:"customer_name"`
	Email         string    `json:"email" db:"email"`
	Phone         string    `json:"phone" db:"phone"`
	PickupAddress string    `json:"pickup_address" db:"pickup_address"`
	DropAddress   string    `json:"drop_address" db:"drop_address"`
	PickupDate    time.Time `json:"pickup_date" db:"pickup_date"`
	Status        int       `json:"status" db:"status"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
}

// UnmarshalJSON implements custom JSON unmarshaling for Booking
func (b *Booking) UnmarshalJSON(data []byte) error {
	type Alias Booking
	aux := &struct {
		PickupDate string      `json:"pickup_date"`
		CreatedAt  string      `json:"created_at"`
		UpdatedAt  string      `json:"updated_at"`
		Phone      interface{} `json:"phone"`
		Status     interface{} `json:"status"`
		*Alias
	}{
		Alias: (*Alias)(b),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// Parse the dates using a simpler format
	if aux.PickupDate != "" {
		t, err := time.Parse("2006-01-02T15:04:05", aux.PickupDate)
		if err != nil {
			return err
		}
		b.PickupDate = t
	}
	if aux.CreatedAt != "" {
		t, err := time.Parse("2006-01-02T15:04:05", aux.CreatedAt)
		if err != nil {
			return err
		}
		b.CreatedAt = t
	}
	if aux.UpdatedAt != "" {
		t, err := time.Parse("2006-01-02T15:04:05", aux.UpdatedAt)
		if err != nil {
			return err
		}
		b.UpdatedAt = t
	}

	// Handle phone number that might come as a number
	if aux.Phone != nil {
		switch v := aux.Phone.(type) {
		case string:
			b.Phone = v
		case float64:
			b.Phone = fmt.Sprintf("%.0f", v)
		case int:
			b.Phone = strconv.Itoa(v)
		}
	}

	// Handle status that might come as a string or number
	if aux.Status != nil {
		switch v := aux.Status.(type) {
		case string:
			// Convert status string to int
			switch v {
			case "PENDING":
				b.Status = int(StatusPending)
			case "CONFIRMED":
				b.Status = int(StatusConfirmed)
			case "IN_PROGRESS":
				b.Status = int(StatusInProgress)
			case "COMPLETED":
				b.Status = int(StatusCompleted)
			case "CANCELLED":
				b.Status = int(StatusCancelled)
			default:
				b.Status = int(StatusPending)
			}
		case float64:
			b.Status = int(v)
		case int:
			b.Status = v
		}
	}

	return nil
}

// GetStatusString returns the string representation of the booking status
func (b *Booking) GetStatusString() string {
	switch b.Status {
	case int(StatusPending):
		return "PENDING"
	case int(StatusConfirmed):
		return "CONFIRMED"
	case int(StatusInProgress):
		return "IN_PROGRESS"
	case int(StatusCompleted):
		return "COMPLETED"
	case int(StatusCancelled):
		return "CANCELLED"
	default:
		return "PENDING"
	}
}

type FurnitureItem struct {
	FurnitureID string `json:"furniture_id" db:"furniture_id"`
	// Furniture items
	Chairs          int `json:"chairs" db:"chairs"`
	Table2Seats     int `json:"table_2_seats" db:"table_2_seats"`
	Table3Seats     int `json:"table_3_seats" db:"table_3_seats"`
	Table4PlusSeats int `json:"table_4_plus_seats" db:"table_4_plus_seats"`
	Fridges         int `json:"fridges" db:"fridges"`
	WashingMachines int `json:"washing_machines" db:"washing_machines"`
	Dryers          int `json:"dryers" db:"dryers"`
	Dishwashers     int `json:"dishwashers" db:"dishwashers"`
	Boxes           int `json:"boxes" db:"boxes"`
	PotPlants       int `json:"pot_plants" db:"pot_plants"`
	Mattresses      int `json:"mattresses" db:"mattresses"`
	BedFrames       int `json:"bed_frames" db:"bed_frames"`
	Sofas           int `json:"sofas" db:"sofas"`
}

type BookingStatus int

const (
	StatusPending    BookingStatus = 0
	StatusConfirmed  BookingStatus = 1
	StatusInProgress BookingStatus = 2
	StatusCompleted  BookingStatus = 3
	StatusCancelled  BookingStatus = 4
)

// NewBooking creates a new booking with default values
func NewBooking() *Booking {
	return &Booking{
		UserID:    uuid.New().String(),
		Status:    int(StatusPending),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}
